package dao

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DaoProvider struct {
	Clients *ClientsDao
	Seqs    *SeqsDao
	Items   *ItemsDao
}

func NewDaoProvider() (*DaoProvider, error) {
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return nil, err
	}
	sqlStmt := `

	CREATE TABLE IF NOT EXISTS seqs (
		type string NOT NULL PRIMARY KEY,
		value integer NOT NULL
	);

	INSERT OR IGNORE INTO seqs(type, value) VALUES('clientNo', 0);

	CREATE TABLE IF NOT EXISTS clients (
		id string NOT NULL PRIMARY KEY,
		role string NOT NULL,
		online bool NOT NULL,
		last_login timestamp,
		wsclient_id integer
	);

	UPDATE clients SET online = false;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &DaoProvider{
		Clients: &ClientsDao{db: db},
		Seqs:    &SeqsDao{db: db},
		Items:   &ItemsDao{db: db},
	}, nil
}
