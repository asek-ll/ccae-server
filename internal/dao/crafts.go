package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Craft struct {
	ID            int
	PlanID        int
	WorkerID      string
	Status        string
	Created       time.Time
	RecipeID      int
	Repeats       int
	CheckAt       time.Time
	CommitRepeats int
}

type CraftsDao struct {
	db *sql.DB
}

const craftsFieldList string = "id, plan_id, worker_id, status, created, recipe_id, repeats, check_at, commit_repeats"

func NewCraftsDao(db *sql.DB) (*CraftsDao, error) {

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS craft (
		id INTEGER PRIMARY KEY,
		plan_id INTEGER NOT NULL,
		worker_id string NOT NULL,
		status string NOT NULL,
		created timestamp NOT NULL,
		recipe_id INTEGER NOT NULL,
		repeats INTEGER NOT NULL,
		check_at timestamp NOT NULL,
		commit_repeats int NOT NULL
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
	rows, err := d.db.Query(`
	SELECT `+craftsFieldList+` 
	FROM craft 
	WHERE check_at < ?
	LIMIT 50
	`, time.Now())
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
		err := rows.Scan(&c.ID, &c.PlanID, &c.WorkerID, &c.Status, &c.Created, &c.RecipeID, &c.Repeats, &c.CheckAt)
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

func (d *CraftsDao) InsertCraft(planId int, workderId string, recipe *Recipe, repeats int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO 
	craft (plan_id, worker_id, status, created, recipe_id, repeats, check_at, commit_repeats) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		planId, workderId, "PENDING", time.Now(), recipe.ID, repeats, time.Now(), 0)
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
		amount := ing.Amount * repeats
		res, err := tx.Exec(`
		UPDATE plan_item_state SET amount = amount - ?, required_amount = required_amount - ?
		WHERE item_uid = ? AND plan_id = ? AND amount >= ?`,
			amount, amount, ing.ItemUID, planId, amount)
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

func (d *CraftsDao) CommitCraft(craft *Craft, recipe *Recipe, repeats int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec("UPDATE craft SET status = 'COMMITED', commit_repeats = ? WHERE id = ? AND status = 'PENDING'", repeats, craft.ID)
	if err != nil {
		return err
	}

	afftected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if afftected != 1 {
		return errors.New("Expected craft in PENDING state")
	}

	for _, ing := range recipe.Ingredients {
		err = ReleaseItems(tx, ing.ItemUID, ing.Amount*craft.Repeats)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *CraftsDao) FindCurrent(workerId string) (*Craft, error) {
	rows, err := d.db.Query(`
	SELECT `+craftsFieldList+` 
	FROM craft 
	WHERE worker_id = ? AND status = 'COMMITED' AND check_at < ?
	LIMIT 1`, workerId, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	crafts, err := readCrafts(rows)
	if err != nil {
		return nil, err
	}

	if len(crafts) == 0 {
		return nil, nil
	}

	return crafts[0], nil
}

func (d *CraftsDao) FindNext(workerId string) (*Craft, error) {
	rows, err := d.db.Query(`
	SELECT `+craftsFieldList+` 
	FROM craft 
	WHERE worker_id = ? AND status = 'PENDING' AND check_at < ?
	LIMIT 1`, workerId, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	crafts, err := readCrafts(rows)
	if err != nil {
		return nil, err
	}

	if len(crafts) == 0 {
		return nil, nil
	}

	return crafts[0], nil
}

func (d *CraftsDao) CompleteCraft(craft *Craft) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow(`UPDATE craft SET 
	repeats = repeats - commit_repeats, 
	commit_repeats = 0,
	status = 'PENDING',
	check_at = ?
	WHERE id = ? AND status = 'COMMITED' RETURNING repeats`, time.Now(), craft.ID)

	err = row.Err()
	if err != nil {
		return err
	}

	var newRepeats int
	row.Scan(&newRepeats)

	if newRepeats == 0 {

		res, err := tx.Exec("DELETE FROM craft WHERE id = ?", craft.ID)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected != 1 {
			return errors.New("Expected commited craft to complete")
		}
	}
	return nil
}

func (d *CraftsDao) SuspendCraft(craft *Craft) error {
	_, err := d.db.Exec("UPDATE craft SET status = 'PENDING', check_at = ?, commit_repeats = 0 WHERE id = ?",
		time.Now().Add(time.Minute), craft.ID)
	return err
}

func (d *CraftsDao) FindById(craftId int) (*Craft, error) {
	rows, err := d.db.Query(`
	SELECT `+craftsFieldList+` 
	FROM craft 
	WHERE id = ? 
	LIMIT 1`, craftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	crafts, err := readCrafts(rows)
	if err != nil {
		return nil, err
	}

	if len(crafts) == 0 {
		return nil, fmt.Errorf("Craft with id %d not found", craftId)
	}

	return crafts[0], nil
}
