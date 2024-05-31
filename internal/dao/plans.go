package dao

import "database/sql"

// const RESULT_ROLE = "result"
// const INGREDIENT_ROLE = "ingredient"

type PlanItem struct {
	ItemUID   string
	Role      string
	ToConsume int
	Consumed  int
	ToCraft   int
	Crafted   int
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
