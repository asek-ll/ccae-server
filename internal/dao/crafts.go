package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Craft struct {
	ID       int
	PlanID   int
	WorkerID string
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
		worker_id string NOT NULL,
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
	rows, err := d.db.Query("SELECT id, plan_id, worker_id, status, created, recipe_id, repeats FROM craft LIMIT 50")
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
		err := rows.Scan(&c.ID, &c.PlanID, &c.WorkerID, &c.Status, &c.Created, &c.RecipeID, &c.Repeats)
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

	_, err = tx.Exec("INSERT INTO craft (plan_id, worker_id, status, created, recipe_id, repeats) VALUES (?, ?, ?, ?, ?, ?)",
		planId, workderId, "PENDING", time.Now(), recipe.ID, repeats)
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

func (d *CraftsDao) CommitCraft(craft *Craft) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec("UPDATE craft SET status = 'COMMITED' WHERE id = ? AND status = 'PENDING'", craft.ID)
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

	rows, err := tx.Query(`
	SELECT r.*, ri.item_uid, ri.amount, ri.role, ri.slot FROM recipes r 
	LEFT JOIN recipe_items ri ON r.id = ri.recipe_id 
	WHERE r.id = ?`, craft.RecipeID)
	if err != nil {
		return err
	}

	recipes, err := readRecipes(rows)
	if err != nil {
		return err
	}
	if len(recipes) != 1 {
		return errors.New("Expected single recipe")
	}

	recipe := recipes[0]

	for _, ing := range recipe.Ingredients {
		err = ReleaseItems(tx, ing.ItemUID, ing.Amount*craft.Repeats)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *CraftsDao) FindById(craftId int) (*Craft, error) {
	rows, err := d.db.Query("SELECT id, plan_id, worker_id, status, created, recipe_id, repeats FROM craft WHERE id = ? LIMIT 1", craftId)
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
