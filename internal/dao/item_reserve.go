package dao

import (
	"database/sql"
	"fmt"
	"log"
)

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

func ReleaseItems(tx *sql.Tx, uid string, amount int) error {
	_, err := tx.Exec("UPDATE item_reserve SET amount = amount - ? WHERE item_uid = ?", amount, uid)

	return err
}

func ReserveItem(tx *sql.Tx, uid string, amount int) error {
	_, err := tx.Exec(`
	INSERT OR IGNORE INTO item_reserve VALUES(?, 0);
	UPDATE item_reserve SET amount = amount + ? WHERE item_uid = ?;
	`, uid, amount, uid)

	return err
}

func (d *ItemReserveDao) UpdateItemCount(uid string, count int) ([]int, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT amount FROM item_reserve WHERE item_uid = ?", uid)
	if err != nil {
		return nil, err
	}
	var reserve int
	if rows.Next() {
		err = rows.Scan(&reserve)
		if err != nil {
			rows.Close()
			return nil, err
		}
	}
	rows.Close()
	if count <= reserve {
		return nil, nil
	}

	freeAmount := count - reserve

	rows, err = tx.Query("SELECT plan_id, required_amount-amount FROM plan_item_state WHERE item_uid = ? AND required_amount > amount", uid)
	if err != nil {
		return nil, err
	}

	var planIds []int
	for rows.Next() && freeAmount > 0 {
		log.Printf("[INFO] DETECT TO update %s", uid)
		var planId, required int

		err = rows.Scan(&planId, &required)
		if err != nil {
			rows.Close()
			return nil, err
		}

		toAdd := min(freeAmount, required)
		freeAmount -= toAdd

		res, err := tx.Exec("UPDATE plan_item_state SET amount = amount + ? WHERE plan_id = ? AND item_uid = ?", toAdd, planId, uid)
		if err != nil {
			rows.Close()
			return nil, err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return nil, err
		}
		if affected != 1 {
			return nil, fmt.Errorf("Expected 1 updated plaan items, got %d", affected)
		}

		planIds = append(planIds, planId)
	}
	rows.Close()

	if freeAmount < count-reserve {

		_, err = tx.Exec("INSERT OR REPLACE INTO item_reserve VALUES(?, ?)", uid, count-freeAmount)
		if err != nil {
			return nil, err
		}

	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return planIds, nil
}

func (d *ItemReserveDao) GetReserves() ([]ItemReserve, error) {
	rows, err := d.db.Query(`
	SELECT ir.item_uid, ir.amount 
	FROM item_reserve ir
	JOIN plan_item_state pis ON pis.item_uid = ir.item_uid 
	WHERE pis.amount < pis.required_amount
	LIMIT 50`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []ItemReserve

	for rows.Next() {
		var itemReserve ItemReserve
		err := rows.Scan(&itemReserve.ItemUID, &itemReserve.Amount)
		if err != nil {
			return nil, err
		}
		result = append(result, itemReserve)
	}

	return result, nil
}
