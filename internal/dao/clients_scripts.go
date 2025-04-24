package dao

import (
	"database/sql"
	"fmt"
)

type ClientsScript struct {
	Role    string
	Content string
	Version int
}

type ClientsScriptsDao struct {
	db *sql.DB
}

func NewClientsScriptsDao(db *sql.DB) (*ClientsScriptsDao, error) {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS clients_scripts (
		role string NOT NULL PRIMARY KEY,
		content string NOT NULL,
		version int NOT NULL DEFAULT 0
	);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &ClientsScriptsDao{db: db}, nil
}

func (d *ClientsScriptsDao) GetClientsScripts() ([]*ClientsScript, error) {
	var scripts []*ClientsScript
	rows, err := d.db.Query("SELECT role, content, version FROM clients_scripts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var script ClientsScript
		if err := rows.Scan(&script.Role, &script.Content, &script.Version); err != nil {
			return nil, err
		}
		scripts = append(scripts, &script)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return scripts, nil
}

func (d *ClientsScriptsDao) GetClientScript(role string) (*ClientsScript, error) {
	var script ClientsScript
	err := d.db.QueryRow("SELECT content, version FROM clients_scripts WHERE role = ?", role).Scan(&script.Content, &script.Version)
	if err != nil {
		return nil, err
	}
	script.Role = role
	return &script, nil
}

func (d *ClientsScriptsDao) CreateClientScript(script *ClientsScript) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	row := tx.QueryRow("SELECT count(*) FROM clients_scripts WHERE role = ?", script.Role)
	if row.Err() != nil {
		return row.Err()
	}
	var count int
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("duplicate role")
	}

	_, err = tx.Exec("INSERT INTO clients_scripts (role, content, version) VALUES (?, ?, 0)", script.Role, script.Content)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *ClientsScriptsDao) UpdateClientsScript(role string, script *ClientsScript) error {
	_, err := d.db.Exec("UPDATE clients_scripts SET role = ?, content = ?, version = version + 1 WHERE role = ?", script.Role, script.Content, role)
	return err
}

func (d *ClientsScriptsDao) DeleteClientScript(role string) error {
	_, err := d.db.Exec("DELETE FROM clients_scripts WHERE role = ?", role)
	return err
}
