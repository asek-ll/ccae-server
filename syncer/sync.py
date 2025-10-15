#!/usr/bin/env python3
"""
Script to synchronize local Lua script files with remote clients-scripts API.

Usage:
    python sync.py <mode> <base_url> <folder> <password>

Modes:
    pull - Download all scripts from server to local folder
    push - Upload local scripts to server (create, update, delete)

Examples:
    python sync.py pull http://localhost:8080 ./scripts admin_password
    python sync.py push http://localhost:8080 ./scripts admin_password
"""

import argparse
import json
import os
import sys
from pathlib import Path
from typing import Dict, List, Optional

import requests
from requests.auth import HTTPBasicAuth


class ScriptSyncer:
    """Handles synchronization of local scripts with remote API."""

    def __init__(self, base_url: str, password: str):
        """
        Initialize the script syncer.

        Args:
            base_url: Base URL of the API (e.g., http://localhost:8080)
            password: Admin password for authentication
        """
        self.base_url = base_url.rstrip('/')
        self.api_url = f"{self.base_url}/api/v1/clients-scripts"
        self.auth = HTTPBasicAuth('admin', password)
        self.session = requests.Session()
        self.session.auth = self.auth

    def get_remote_scripts(self) -> Dict[str, dict]:
        """
        Fetch all remote scripts from the API.

        Returns:
            Dictionary mapping role names to script data
        """
        try:
            response = self.session.get(f"{self.api_url}/")
            response.raise_for_status()
            scripts = response.json()
            return {script['role']: script for script in scripts}
        except requests.exceptions.RequestException as e:
            print(f"Error fetching remote scripts: {e}", file=sys.stderr)
            sys.exit(1)

    def get_local_scripts(self, folder: Path) -> Dict[str, str]:
        """
        Read all Lua script files from the local folder.

        Args:
            folder: Path to the folder containing script files

        Returns:
            Dictionary mapping role names (filename without .lua) to file content
        """
        if not folder.exists():
            print(f"Error: Folder '{folder}' does not exist", file=sys.stderr)
            sys.exit(1)

        if not folder.is_dir():
            print(f"Error: '{folder}' is not a directory", file=sys.stderr)
            sys.exit(1)

        scripts = {}
        for file_path in folder.glob('*.lua'):
            role = file_path.stem  # filename without extension
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                scripts[role] = content
            except Exception as e:
                print(f"Warning: Could not read {file_path}: {e}", file=sys.stderr)

        return scripts

    def write_local_script(self, folder: Path, role: str, content: str) -> bool:
        """
        Write a script to a local file.

        Args:
            folder: Path to the folder to write the file
            role: Script role/name (will be used as filename)
            content: Script content

        Returns:
            True if successful, False otherwise
        """
        try:
            # Create folder if it doesn't exist
            folder.mkdir(parents=True, exist_ok=True)

            file_path = folder / f"{role}.lua"
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(content)
            print(f"  ✓ Wrote script '{role}' to {file_path}")
            return True
        except Exception as e:
            print(f"  ✗ Failed to write script '{role}': {e}", file=sys.stderr)
            return False

    def create_script(self, role: str, content: str) -> bool:
        """
        Create a new script on the remote server.

        Args:
            role: Script role/name
            content: Script content

        Returns:
            True if successful, False otherwise
        """
        try:
            data = {
                'role': role,
                'content': content,
                'version': 0
            }
            response = self.session.post(
                f"{self.api_url}/",
                json=data,
                headers={'Content-Type': 'application/json'}
            )
            response.raise_for_status()
            print(f"   Created script '{role}'")
            return True
        except requests.exceptions.RequestException as e:
            print(f"   Failed to create script '{role}': {e}", file=sys.stderr)
            return False

    def update_script(self, role: str, content: str, current_version: int) -> bool:
        """
        Update an existing script on the remote server.

        Args:
            role: Script role/name
            content: New script content
            current_version: Current version number

        Returns:
            True if successful, False otherwise
        """
        try:
            data = {
                'role': role,
                'content': content,
                'version': current_version
            }
            response = self.session.post(
                f"{self.api_url}/{role}/",
                json=data,
                headers={'Content-Type': 'application/json'}
            )
            response.raise_for_status()
            print(f"   Updated script '{role}'")
            return True
        except requests.exceptions.RequestException as e:
            print(f"   Failed to update script '{role}': {e}", file=sys.stderr)
            return False

    def delete_script(self, role: str) -> bool:
        """
        Delete a script from the remote server.

        Args:
            role: Script role/name

        Returns:
            True if successful, False otherwise
        """
        try:
            response = self.session.delete(f"{self.api_url}/{role}/")
            response.raise_for_status()
            print(f"   Deleted script '{role}'")
            return True
        except requests.exceptions.RequestException as e:
            print(f"   Failed to delete script '{role}': {e}", file=sys.stderr)
            return False

    def pull(self, folder: Path) -> None:
        """
        Pull all scripts from remote server to local folder.

        Args:
            folder: Path to folder where scripts will be saved
        """
        print(f"Starting pull from server...")
        print(f"API URL: {self.api_url}")
        print(f"Local folder: {folder}")
        print()

        # Get remote scripts
        print("Fetching remote scripts...")
        remote_scripts = self.get_remote_scripts()
        print(f"Found {len(remote_scripts)} remote scripts")
        print()

        # Create folder if it doesn't exist
        folder.mkdir(parents=True, exist_ok=True)

        # Write all scripts to local files
        success_count = 0
        total_count = len(remote_scripts)

        if total_count > 0:
            print(f"Downloading {total_count} scripts:")
            for role, script_data in sorted(remote_scripts.items()):
                if self.write_local_script(folder, role, script_data['content']):
                    success_count += 1

        # Summary
        print()
        print("=" * 50)
        print(f"Pull complete!")
        print(f"Successfully downloaded: {success_count}/{total_count} scripts")
        if success_count < total_count:
            print(f"Failed downloads: {total_count - success_count}")
            sys.exit(1)

    def push(self, folder: Path) -> None:
        """
        Push local scripts to remote server (create, update, delete).

        Args:
            folder: Path to folder containing local script files
        """
        print(f"Starting push to server...")
        print(f"API URL: {self.api_url}")
        print(f"Local folder: {folder}")
        print()

        # Get current state
        print("Fetching remote scripts...")
        remote_scripts = self.get_remote_scripts()
        print(f"Found {len(remote_scripts)} remote scripts")

        print("Reading local scripts...")
        local_scripts = self.get_local_scripts(folder)
        print(f"Found {len(local_scripts)} local scripts")
        print()

        # Determine actions
        local_roles = set(local_scripts.keys())
        remote_roles = set(remote_scripts.keys())

        to_create = local_roles - remote_roles
        to_update = local_roles & remote_roles
        to_delete = remote_roles - local_roles

        # Execute sync operations
        success_count = 0
        total_count = 0

        if to_create:
            print(f"Creating {len(to_create)} new scripts:")
            for role in sorted(to_create):
                total_count += 1
                if self.create_script(role, local_scripts[role]):
                    success_count += 1

        if to_update:
            print(f"\nUpdating {len(to_update)} existing scripts:")
            for role in sorted(to_update):
                # Check if content has changed
                remote_content = remote_scripts[role]['content']
                local_content = local_scripts[role]

                if remote_content != local_content:
                    total_count += 1
                    current_version = remote_scripts[role]['version']
                    if self.update_script(role, local_content, current_version):
                        success_count += 1
                else:
                    print(f"  - Skipped '{role}' (no changes)")

        if to_delete:
            print(f"\nDeleting {len(to_delete)} remote scripts:")
            for role in sorted(to_delete):
                total_count += 1
                if self.delete_script(role):
                    success_count += 1

        # Summary
        print()
        print("=" * 50)
        print(f"Push complete!")
        print(f"Successful operations: {success_count}/{total_count}")
        if success_count < total_count:
            print(f"Failed operations: {total_count - success_count}")
            sys.exit(1)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Synchronize local Lua scripts with remote clients-scripts API',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Modes:
  pull    Download all scripts from server to local folder
  push    Upload local scripts to server (create, update, delete)

Examples:
  %(prog)s pull http://localhost:8080 ./scripts admin_password
  %(prog)s push http://localhost:8080 ./scripts admin_password
        """
    )
    parser.add_argument(
        'mode',
        choices=['pull', 'push'],
        help='Synchronization mode'
    )
    parser.add_argument(
        'base_url',
        help='Base URL of the API (e.g., http://localhost:8080)'
    )
    parser.add_argument(
        'folder',
        help='Path to folder containing/receiving script files'
    )
    parser.add_argument(
        'password',
        help='Admin password for authentication'
    )

    args = parser.parse_args()

    # Convert folder to Path object
    folder = Path(args.folder)

    # Create syncer and run appropriate mode
    syncer = ScriptSyncer(args.base_url, args.password)

    if args.mode == 'pull':
        syncer.pull(folder)
    elif args.mode == 'push':
        syncer.push(folder)


if __name__ == '__main__':
    main()
