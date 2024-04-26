package dao

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DaoProvider struct {
	Clients *ClientsDao
}

func NewDaoProvider() (*DaoProvider, error) {
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return nil, err
	}
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS clients (
		id string NOT NULL PRIMARY KEY,
		role string NOT NULL,
		online bool NOT NULL,
		last_login timestamp
	);

	UPDATE clients SET online = false;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &DaoProvider{
		Clients: &ClientsDao{db: db},
	}, nil
}
