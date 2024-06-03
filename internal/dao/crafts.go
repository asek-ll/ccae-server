package dao

import (
	"database/sql"
	"errors"
	"time"
)

type Craft struct {
	ID       int
	PlanID   int
	WorkerID int
	Status   string
	Created  time.Time
	RecipeID int
	Repeats  int
}

type CraftsDao struct {
	db *sql.DB
}

func NewCraftsDao(db *sql.DB) (*CraftsDao, error) {

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS craft (
		id INTEGER PRIMARY KEY,
		plan_id INTEGER NOT NULL,
		worker_id INTEGER NOT NULL,
		status string NOT NULL,
		created timestamp NOT NULL,
		recipe_id INTEGER NOT NULL,
		repeats INTEGER NOT NULL
	);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &CraftsDao{
		db: db,
	}, nil
}

func (d *CraftsDao) GetCrafts() ([]*Craft, error) {
	rows, err := d.db.Query("SELECT plan_id, worker_id, status, created, recipe_id, repeats FROM craft LIMIT 50")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readCrafts(rows)
}

func readCrafts(rows *sql.Rows) ([]*Craft, error) {
	defer rows.Close()
	var craft []*Craft
	for rows.Next() {
		var c Craft
		err := rows.Scan(&c.PlanID, &c.WorkerID, &c.Status, &c.Created, &c.RecipeID, &c.Repeats)
		if err != nil {
			return nil, err
		}
		craft = append(craft, &c)
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}
	return craft, nil
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
