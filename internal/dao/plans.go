package dao

import "database/sql"

type PlanItem struct {
	ItemUID   string
	ToConsume int
	Consumed  int
	ToCraft   int
	Crafted   int
}

type PlanStepItems struct {
	ID       int
	PlanID   int
	RecipeID int
	ItemUID  string

	Required       int
	Consumed       int
	MinimumToCraft int

	ToCraft int
}

type Plan struct {
	ID     int
	Status string
	Items  []PlanItem
}

type PlansDao struct {
	db *sql.DB
}

func NewPlansDao(db *sql.DB) *PlansDao {

	return &PlansDao{db: db}
}

func (p *PlansDao) InsertPlan(plan *Plan) error {
	return nil
}
