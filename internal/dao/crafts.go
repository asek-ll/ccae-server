package dao

import (
	"database/sql"
	"errors"
	"time"
)

type Craft struct {
	planId   int
	workerId int
	status   string
	created  time.Time
	recipeId int
	repeats  int
}

type CraftsDao struct {
	db *sql.DB
}

func NewCraftsDao(db *sql.DB) (*CraftsDao, error) {
	return nil, nil
}

func (d *CraftsDao) InsertCraft(planId int, recipe *Recipe, repeats int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO craft (plan_id, worker_id, status, created, recipe_id, repeats) VALUES (?, ?, ?, ?, ?, ?)",
		planId, 0, "pending", time.Now(), recipe.ID, repeats)
	if err != nil {
		return err
	}
	res, err := tx.Exec("UPDATE plan_step_state SET repeats = repeats - ? WHERE plan_id = ? and recipe_id = ? and repeats >= ?", repeats, planId, recipe.ID, repeats)

	if err != nil {
		return err
	}
	afftected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if afftected == 0 {
		return errors.New("Can't find plan step")
	}

	for _, ing := range recipe.Ingredients {
		res, err := tx.Exec("UPDATE plan_item_state SET amount = amount - ? WHERE item_uid = ? AND plan_id = ? AND amount >= ?",
			ing.Amount*repeats, ing.ItemUID, planId, ing.Amount*repeats)
		if err != nil {
			return err
		}
		afftected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if afftected == 0 {
			return errors.New("Can't acquire ingredients")
		}
	}

	return tx.Commit()
}
