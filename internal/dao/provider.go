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
	create table if not exists clients (id string not null primary key);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &DaoProvider{
		Clients: &ClientsDao{db: db},
	}, nil
}
