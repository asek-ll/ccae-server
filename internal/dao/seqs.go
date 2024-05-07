package dao

import (
	"database/sql"
)

type SeqsDao struct {
	db *sql.DB
}

func NewSeqsDao(db *sql.DB) (*SeqsDao, error) {
	sqlStmt := `

	CREATE TABLE IF NOT EXISTS seqs (
		type string NOT NULL PRIMARY KEY,
		value integer NOT NULL
	);

	INSERT OR IGNORE INTO seqs(type, value) VALUES('clientNo', 0);
	`

	_, err := db.Exec(sqlStmt)

	if err != nil {
		return nil, err
	}

	return &SeqsDao{db: db}, nil
}

func (s *SeqsDao) NextId(typeKey string) (int, error) {

	row := s.db.QueryRow("UPDATE seqs SET value = value + 1 WHERE type = ? RETURNING value", typeKey)
	err := row.Err()
	if err != nil {
		return 0, err
	}

	var id int
	row.Scan(&id)

	return id, nil

}
