package dao

import (
	"database/sql"
	"errors"
)

type WorkerState struct {
	WorkerKey   string
	WaitCraftID int
}

type WorkerStateDao struct {
	db *sql.DB
}

func NewWorkerStateDao(db *sql.DB) (*WorkerStateDao, error) {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS worker_state (
		worker_key string NOT NULL,
		wait_craft_id int NOT NULL
	);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &WorkerStateDao{db: db}, nil
}

func SetWorkerWaitCraftInOuterTx(tx *sql.Tx, workerKey string, craft *Craft) error {
	_, err := tx.Exec("INSERT INTO worker_state (worker_key, wait_craft_id) VALUES (?, ?)", workerKey, craft.ID)
	return err
}

func (w *WorkerStateDao) CompleteWorkerWait(workerKey string, craft *Craft) error {
	tx, err := w.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec("DELETE FROM worker_state WHERE worker_key = ? AND wait_craft_id = ?", workerKey, craft.ID)

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("no worker wait to complete")
	}

	err = CompleteCraftInOuterTx(tx, craft)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (w *WorkerStateDao) GetWaitCraftID(workerKey string) (*int, error) {
	rows, err := w.db.Query("SELECT wait_craft_id FROM worker_state WHERE worker_key = ?", workerKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var waitCraftID int
	err = rows.Scan(&waitCraftID)
	if err != nil {
		return nil, err
	}

	return &waitCraftID, nil
}

func DeleteWorkerStateInOuterTx(tx *sql.Tx, craftID int) error {
	_, err := tx.Exec("DELETE FROM worker_state WHERE wait_craft_id = ?", craftID)
	return err
}
