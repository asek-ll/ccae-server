package dao

import "database/sql"

type ItemReserve struct {
	ItemUID string
	Amount  int
}

type ItemReserveDao struct {
	db *sql.DB
}

func NewItemReserveDao(db *sql.DB) (*ItemReserveDao, error) {

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS item_reserve (
		item_uid string PRIMARY KEY,
		amount integer NOT NULL
	);
	`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &ItemReserveDao{
		db: db,
	}, nil
}

func (d *ItemReserveDao) ReleaseItems(tx *sql.Tx, reserves []ItemReserve) error {
	for _, reserve := range reserves {
		_, err := tx.Exec("UPDATE item_reserve SET amount = amount - ? WHERE item_uid = ?",
			reserve.Amount, reserve.ItemUID)

		if err != nil {
			return err
		}
	}
	return nil
}

func (d *ItemReserveDao) ReserveItems(tx *sql.Tx, reserves []ItemReserve) error {
	for _, reserve := range reserves {
		_, err := tx.Exec(`
		INSERT OR INGORE INTO item_reserve VALUES(?, 0);
		UPDATE item_reserve SET amount = amount + ? WHERE item_uid = ?;
		`, reserve.ItemUID, reserve.Amount, reserve.ItemUID)

		if err != nil {
			return err
		}
	}
	return nil
}

func (d *ItemReserveDao) UpdateItemCount(uid string, count int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT amount FROM item_reserve WHERE item_uid = ?", uid)
	if err != nil {
		return err
	}
	var reserve int
	if rows.Next() {
		err = rows.Scan(&reserve)
		if err != nil {
			return err
		}
	}
	if count <= reserve {
		return nil
	}

	freeAmount := count - reserve

	rows, err = tx.Query("SELECT plan_id, required_amount-amount FROM plan_item_state WHERE item_uid = ? AND required_amount < amount", uid)
	if err != nil {
		return err
	}
	for rows.Next() && freeAmount > 0 {
		var planId, required int

		err = rows.Scan(planId, required)
		if err != nil {
			return err
		}

		toAdd := min(freeAmount, required)
		freeAmount -= toAdd

		_, err = tx.Exec("UPDATE plan_item_state SET amount = amount + ? WHERE plan_id = ? AND item_uid = ?", toAdd, planId, uid)
		if err != nil {
			return err
		}
	}

	if freeAmount < count-reserve {

		_, err = tx.Exec("INSERT OR REPLACE INTO item_reserve VALUES(?, ?)", uid, count-freeAmount)
		if err != nil {
			return err
		}

	}

	return tx.Commit()
}
