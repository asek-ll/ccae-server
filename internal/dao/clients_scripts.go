package dao

import (
	"database/sql"
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
	return &ClientsScriptsDao{db: db}, nil
}

func (d *ClientsScriptsDao) GetClientsScripts() ([]ClientsScript, error) {
	var scripts []ClientsScript
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
		scripts = append(scripts, script)
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

func (d *ClientsScriptsDao) UpdateClientsScript(script ClientsScript) error {
	_, err := d.db.Exec("UPDATE clients_scripts SET content = ?, version = version + 1 WHERE role = ?", script.Content, script.Role)
	return err
}

func (d *ClientsScriptsDao) DeleteClientScript(role string) error {
	_, err := d.db.Exec("DELETE FROM clients_scripts WHERE role = ?", role)
	return err
}
