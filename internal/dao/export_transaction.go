package dao

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/asek-ll/aecc-server/internal/common"
)

type TXSubject struct {
	ID   int
	Name string
	TXID int
}

type StoredTX struct {
	ID   int
	Data []byte
}

type StoredTXDao struct {
	db *sql.DB
}

func NewStoredTXDao(db *sql.DB) (*StoredTXDao, error) {
	sql := `
	CREATE TABLE IF NOT EXISTS stored_tx_subjects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		txid INTEGER NOT NULL
	);
	CREATE UNIQUE INDEX IF NOT EXISTS stored_tx_subjects_idx ON stored_tx_subjects(name);

	CREATE TABLE IF NOT EXISTS stored_tx (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		data TEXT NOT NULL
	);
	`

	_, err := db.Exec(sql)
	if err != nil {
		return nil, err
	}
	return &StoredTXDao{db: db}, nil
}

func (d *StoredTXDao) InsertTransactionUncommitted(subjects []string, stx *StoredTX) (*sql.Tx, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}

	res, err := tx.Exec("INSERT INTO stored_tx (data) VALUES (?)", stx.Data)
	if err != nil {
		return nil, err
	}

	txid, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	stx.ID = int(txid)

	for _, subject := range subjects {
		_, err := tx.Exec("INSERT INTO stored_tx_subjects (name, txid) VALUES (?, ?)", subject, txid)
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (d *StoredTXDao) FindTransactionIdForSubjects(subjects []string) ([]int, error) {
	rows, err := d.db.Query(fmt.Sprintf(`
	SELECT distinct txid 
	FROM stored_tx_subjects 
	WHERE name IN (?%s)`,
		strings.Repeat(", ?", len(subjects)-1)), common.ToArgs(subjects)...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int

	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (d *StoredTXDao) FindLocks(subjects []string) ([]*StoredTX, error) {
	ids, err := d.FindTransactionIdForSubjects(subjects)
	if err != nil {
		return nil, err
	}

	return d.FindTransactionsByIds(ids)
}

func (d *StoredTXDao) FindTransactionsByIds(ids []int) ([]*StoredTX, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := d.db.Query(fmt.Sprintf(`
	SELECT id, data
	FROM stored_tx 
	WHERE id IN (?%s)`,
		strings.Repeat(", ?", len(ids)-1)), common.ToArgs(ids)...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []*StoredTX

	for rows.Next() {
		var id int
		var data []byte
		err := rows.Scan(&id, &data)
		if err != nil {
			return nil, err
		}
		txs = append(txs, &StoredTX{
			ID:   id,
			Data: data,
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func (d *StoredTXDao) DropTransaction(stx *StoredTX) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM stored_tx WHERE id = ?", stx.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM stored_tx_subjects WHERE txid = ?", stx.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
