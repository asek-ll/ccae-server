package dao

import (
	"database/sql"
)

type SeqsDao struct {
	db *sql.DB
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
