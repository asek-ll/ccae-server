package dao

import (
	"database/sql"
	"time"
)

type Craft struct {
	planId   int
	workerId int
	status   string
	created  time.Time
	recipeId int64
	repeats  int
}

type CraftsDao struct {
	db *sql.DB
}

func NewCraftsDao(db *sql.DB) (*CraftsDao, error) {
	return nil, nil
}
