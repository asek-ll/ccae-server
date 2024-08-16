package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type Craft struct {
	ID            int
	PlanID        int
	RecipeType    string
	WorkerID      *string
	Status        string
	Created       time.Time
	RecipeID      int
	Repeats       int
	CommitRepeats int
}

type CraftsDao struct {
	db *sql.DB
}

var COMMITED_CRAFT_STATUS = "COMMITED"
var PENDING_CRAFT_STATUS = "PENDING"

const craftsFieldList string = "id, plan_id, recipe_type, worker_id, status, created, recipe_id, repeats, commit_repeats"

func NewCraftsDao(db *sql.DB) (*CraftsDao, error) {

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS craft (
		id INTEGER PRIMARY KEY,
		plan_id INTEGER NOT NULL,
		recipe_type string NOT NULL,
		worker_id string,
		status string NOT NULL,
		created timestamp NOT NULL,
		recipe_id INTEGER NOT NULL,
		repeats INTEGER NOT NULL,
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

func (d *CraftsDao) GetAllCrafts() ([]*Craft, error) {
	rows, err := d.db.Query(`
	SELECT ` + craftsFieldList + ` 
	FROM craft 
	LIMIT 50`)
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
		err := rows.Scan(&c.ID,
			&c.PlanID,
			&c.RecipeType,
			&c.WorkerID,
			&c.Status,
			&c.Created,
			&c.RecipeID,
			&c.Repeats,
			&c.CommitRepeats,
		)
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

func (d *CraftsDao) InsertCraft(planId int, recipeType string, recipe *Recipe, repeats int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO 
	craft (plan_id, recipe_type, status, created, recipe_id, repeats, commit_repeats) 
	VALUES (?, ?, ?, ?, ?, ?, ?)`,
		planId, recipeType, "PENDING", time.Now(), recipe.ID, repeats, 0)
	if err != nil {
		log.Printf("[ERROR] Insert craft: %v", err)
		return err
	}

	res, err := tx.Exec(`
	UPDATE 
		plan_step_state 
	SET 
		repeats = repeats - ? 
	WHERE 
		plan_id = ? and recipe_id = ? and repeats >= ?`, repeats, planId, recipe.ID, repeats)

	if err != nil {
		return err
	}
	afftected, err := res.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] Plan step update: %v", err)
		return err
	}
	if afftected == 0 {
		return errors.New("Can't find plan step")
	}

	for _, ing := range recipe.Ingredients {
		amount := ing.Amount * repeats
		res, err := tx.Exec(`
		UPDATE plan_item_state 
		SET 
			amount = amount - ?,
			required_amount = required_amount - ?
		WHERE 
			item_uid = ? AND plan_id = ? AND amount >= ?`,
			amount, amount, ing.ItemUID, planId, amount)
		log.Printf("[INFO] Call query with amount: %v, uid: %v, planId: %v", amount, ing.ItemUID, planId)
		if err != nil {
			log.Printf("[ERROR] Plan item step: %v", err)
			return err
		}
		afftected, err := res.RowsAffected()
		if err != nil {
			log.Printf("[ERROR] Affected rows: %v", err)
			return err
		}
		if afftected == 0 {
			return errors.New("Can't acquire ingredients")
		}
	}

	log.Printf("[INFO] Commit transaction INSERT")
	err = tx.Commit()
	if err != nil {
		log.Printf("[ERROR] Commit error: %v", err)
		return err
	}

	return nil
}

func (d *CraftsDao) CommitCraftInOuterTx(tx *sql.Tx, craft *Craft, recipe *Recipe, repeats int) error {
	res, err := tx.Exec(`
	UPDATE craft 
	SET 
		status = 'COMMITED',
		commit_repeats = ? 
	WHERE
		id = ? AND status = 'PENDING' AND commit_repeats = 0`, repeats, craft.ID)
	if err != nil {
		return err
	}

	log.Printf("[WARN] Commit recipe %d", repeats)

	afftected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if afftected != 1 {
		return errors.New("Expected craft in PENDING state")
	}

	for _, ing := range recipe.Ingredients {
		err = ReleaseItems(tx, ing.ItemUID, ing.Amount*repeats)
		if err != nil {
			return err
		}
	}

	craft.CommitRepeats = repeats

	return nil
}

func (d *CraftsDao) CommitCraft(craft *Craft, recipe *Recipe, repeats int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = d.CommitCraftInOuterTx(tx, craft, recipe, repeats)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *CraftsDao) CancelCraft(craft *Craft, recipe *Recipe) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec("DELETE FROM craft WHERE id = ? AND commit_repeats = ?", craft.ID, craft.CommitRepeats)
	if err != nil {
		return err
	}

	afftected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if afftected != 1 {
		return errors.New("Expected craft")
	}

	repeats := craft.Repeats - craft.CommitRepeats
	for _, ing := range recipe.Ingredients {
		err = ReleaseItems(tx, ing.ItemUID, ing.Amount*repeats)
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
	WHERE worker_id = ?
	LIMIT 1`, workerId)
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

func (d *CraftsDao) CompleteCraftInOuterTx(tx *sql.Tx, craft *Craft) error {
	row := tx.QueryRow(`UPDATE craft SET 
	repeats = repeats - commit_repeats, 
	commit_repeats = 0,
	status = 'PENDING'
	WHERE id = ? AND status = 'COMMITED' RETURNING repeats`, craft.ID)

	err := row.Err()
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

func (d *CraftsDao) CompleteCraft(craft *Craft) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = d.CompleteCraftInOuterTx(tx, craft)
	if err != nil {
		return err
	}

	return tx.Commit()
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

func (d *CraftsDao) FindNextByTypes(types []string, workerId string) ([]*Craft, error) {
	if len(types) == 0 {
		return nil, nil
	}
	args := []any{workerId}
	for _, tp := range types {
		args = append(args, tp)
	}

	rows, err := d.db.Query(fmt.Sprintf(`
	SELECT `+craftsFieldList+` 
	FROM craft 
	WHERE worker_id = ? OR recipe_type IN (?%s)
	LIMIT 50`, strings.Repeat(",?", len(types)-1),
	), args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	crafts, err := readCrafts(rows)
	if err != nil {
		return nil, err
	}

	return crafts, nil
}

func (d *CraftsDao) AssignCraftToWorker(craft *Craft, workerId string) (bool, error) {
	res, err := d.db.Exec("UPDATE craft SET worker_id = ? WHERE id = ? AND worker_id IS NULL", workerId, craft.ID)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected == 0 {
		return false, nil
	}
	if affected > 1 {
		return false, errors.New("More than one craft updated")
	}

	return true, nil
}
