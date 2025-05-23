package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/asek-ll/aecc-server/internal/common"
)

type PlanItemState struct {
	ItemUID        string
	Amount         int
	RequiredAmount int
}

type PlanStepState struct {
	RecipeID int
	Repeats  int
}

type PlanGoal struct {
	ItemUID string
	Amount  int
}

type PlanState struct {
	ID     int
	Status string
	Goals  []PlanGoal

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
		status string NOT NULL
	);

	CREATE TABLE IF NOT EXISTS plan_item_state (
		plan_id INTEGER NOT NULL,
		item_uid string NOT NULL,
		amount integer NOT NULL,
		required_amount integer NOT NULL
	);

	CREATE TABLE IF NOT EXISTS plan_step_state (
		plan_id INTEGER NOT NULL,
		recipe_id integer NOT NULL,
		repeats integer NOT NULL
	);

	CREATE TABLE IF NOT EXISTS plan_goal (
		plan_id INTEGER NOT NULL,
		item_uid string NOT NULL,
		amount integer NOT NULL
	);
	`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &PlansDao{db: db}, nil
}

func (d *PlansDao) GetPlanById(id int) (*PlanState, error) {
	rows, err := d.db.Query("SELECT id, status FROM plan_state WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states, err := readPlanState(rows)
	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, errors.New("plan not found")
	}

	plan := states[0]

	goals, err := d.GetPlanGoals([]int{plan.ID})
	if err != nil {
		return nil, err
	}

	plan.Goals = goals[plan.ID]

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
		err := rows.Scan(&plan.ID, &plan.Status)
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
	rows, err := d.db.Query("SELECT id, status FROM plan_state")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	planStates, err := readPlanState(rows)
	if err != nil {
		return nil, err
	}
	var ids []int

	for _, planState := range planStates {
		ids = append(ids, planState.ID)
	}

	goals, err := d.GetPlanGoals(ids)
	if err != nil {
		return nil, err
	}

	for _, planState := range planStates {
		planState.Goals = goals[planState.ID]
	}

	return planStates, nil
}

func (d *PlansDao) GetPlanItemState(planId int) ([]PlanItemState, error) {
	rows, err := d.db.Query("SELECT item_uid, amount, required_amount FROM plan_item_state WHERE plan_id = ?", planId)
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
		err := rows.Scan(&itemState.ItemUID, &itemState.Amount, &itemState.RequiredAmount)
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

func (d *PlansDao) GetPlanGoals(planIds []int) (map[int][]PlanGoal, error) {
	if len(planIds) == 0 {
		return nil, nil
	}
	rows, err := d.db.Query(fmt.Sprintf("SELECT plan_id, item_uid, amount FROM plan_goal WHERE plan_id IN (?%s)",
		strings.Repeat(",?", len(planIds)-1)), common.ToArgs(planIds)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readPlanGoals(rows)
}

func readPlanGoals(rows *sql.Rows) (map[int][]PlanGoal, error) {
	goals := make(map[int][]PlanGoal)
	for rows.Next() {
		var planId int
		goal := PlanGoal{}
		err := rows.Scan(&planId, &goal.ItemUID, &goal.Amount)
		if err != nil {
			return nil, err
		}
		goals[planId] = append(goals[planId], goal)
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}
	return goals, nil
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

	res, err := tx.Exec("INSERT INTO plan_state (status) VALUES (?)", plan.Status)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	plan.ID = int(id)

	for _, goal := range plan.Goals {
		_, err := tx.Exec("INSERT INTO plan_goal (plan_id, item_uid, amount) VALUES (?, ?, ?)", plan.ID, goal.ItemUID, goal.Amount)
		if err != nil {
			return err
		}
	}

	for _, item := range plan.Items {
		_, err := tx.Exec("INSERT INTO plan_item_state (plan_id, item_uid, amount, required_amount) VALUES (?, ?, ?, ?)",
			plan.ID, item.ItemUID, item.Amount, item.RequiredAmount)
		if err != nil {
			return err
		}
		err = ReserveItem(tx, item.ItemUID, item.Amount)
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

func (d *PlansDao) RemovePlan(planId int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM plan_state WHERE id = ?", planId)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM plan_step_state WHERE plan_id = ?", planId)
	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT item_uid, amount, required_amount FROM plan_item_state WHERE plan_id = ?", planId)
	if err != nil {
		return err
	}
	defer rows.Close()

	item_states, err := readPlanItemState(rows)
	if err != nil {
		return err
	}

	for _, item_state := range item_states {
		err = ReleaseItems(tx, item_state.ItemUID, item_state.Amount)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec("DELETE FROM plan_item_state WHERE plan_id = ?", planId)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM plan_goal WHERE plan_id = ?", planId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *PlansDao) CleanupItems(planId int) error {
	_, err := d.db.Exec("DELETE FROM plan_item_state WHERE plan_id = ? and required_amount = 0 and amount = 0", planId)
	return err
}
