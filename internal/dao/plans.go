package dao

import (
	"database/sql"
	"errors"
)

type PlanItemState struct {
	ItemUID string
	Amount  int
}

type PlanStepState struct {
	RecipeID int
	Repeats  int
}

type PlanState struct {
	ID          int
	Status      string
	GoalItemUID string
	GoalAmount  int

	Steps []PlanStepState
	Items []PlanItemState
}

type PlansDao struct {
	db *sql.DB
}

func NewPlansDao(db *sql.DB) (*PlansDao, error) {
	sqlStmt := `

	CREATE TABLE IF NOT EXISTS plan_state (
		id INTEGER PRIMARY KEY,
		status string NOT NULL,
		goal_item_uid string NOT NULL,
		goal_amount integer NOT NULL
	);

	CREATE TABLE IF NOT EXISTS plan_item_state (
		plan_id INTEGER NOT NULL,
		item_uid string NOT NULL,
		amount integer NOT NULL
	);

	CREATE TABLE IF NOT EXISTS plan_step_state (
		plan_id INTEGER NOT NULL,
		recipe_id integer NOT NULL,
		repeats integer NOT NULL
	);
	`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &PlansDao{db: db}, nil
}

func (d *PlansDao) GetPlanById(id int) (*PlanState, error) {
	rows, err := d.db.Query("SELECT * FROM plan_state WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states, err := readPlanState(rows)
	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, errors.New("Plan not found")
	}

	plan := states[0]

	items, err := d.GetPlanItemState(plan.ID)
	if err != nil {
		return nil, err
	}
	plan.Items = items

	steps, err := d.GetPlanStepState(plan.ID)
	if err != nil {
		return nil, err
	}
	plan.Steps = steps

	return plan, nil
}

func readPlanState(rows *sql.Rows) ([]*PlanState, error) {
	var planState []*PlanState
	for rows.Next() {
		plan := PlanState{}
		err := rows.Scan(&plan.ID, &plan.Status, &plan.GoalItemUID, &plan.GoalAmount)
		if err != nil {
			return nil, err
		}
		planState = append(planState, &plan)
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}
	return planState, nil
}

func (d *PlansDao) GetPlans() ([]*PlanState, error) {
	rows, err := d.db.Query("SELECT * FROM plan_state")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readPlanState(rows)
}

func (d *PlansDao) GetPlanItemState(planId int) ([]PlanItemState, error) {
	rows, err := d.db.Query("SELECT item_uid, amount FROM plan_item_state WHERE plan_id = ?", planId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readPlanItemState(rows)
}

func readPlanItemState(rows *sql.Rows) ([]PlanItemState, error) {
	var planItemState []PlanItemState
	for rows.Next() {
		itemState := PlanItemState{}
		err := rows.Scan(&itemState.ItemUID, &itemState.Amount)
		if err != nil {
			return nil, err
		}
		planItemState = append(planItemState, itemState)
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}
	return planItemState, nil
}

func (d *PlansDao) GetPlanStepState(planId int) ([]PlanStepState, error) {
	rows, err := d.db.Query("SELECT recipe_id, repeats FROM plan_step_state WHERE plan_id = ?", planId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readPlanStepState(rows)
}

func readPlanStepState(rows *sql.Rows) ([]PlanStepState, error) {
	var planStepState []PlanStepState
	for rows.Next() {
		stepState := PlanStepState{}
		err := rows.Scan(&stepState.RecipeID, &stepState.Repeats)
		if err != nil {
			return nil, err
		}
		planStepState = append(planStepState, stepState)
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}
	return planStepState, nil
}

func (d *PlansDao) InsertPlan(plan *PlanState) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO plan_state (status, goal_item_uid, goal_amount) VALUES (?, ?, ?)", plan.Status, plan.GoalItemUID, plan.GoalAmount)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	plan.ID = int(id)

	for _, item := range plan.Items {
		_, err := tx.Exec("INSERT INTO plan_item_state (plan_id, item_uid, amount) VALUES (?, ?, ?)", plan.ID, item.ItemUID, item.Amount)
		if err != nil {
			return err
		}
	}

	for _, step := range plan.Steps {
		_, err := tx.Exec("INSERT INTO plan_step_state (plan_id, recipe_id, repeats) VALUES (?, ?, ?)", plan.ID, step.RecipeID, step.Repeats)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
