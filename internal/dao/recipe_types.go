package dao

import (
	"database/sql"
	"fmt"
	"log"
)

type RecipeTypesDao struct {
	db *sql.DB
}

type RecipeType struct {
	Name     string
	WorkerID string
}

func NewRecipeTypesDao(db *sql.DB) (*RecipeTypesDao, error) {

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS recipe_types (
		name string NOT NULL,
		worker_id string NOT NULL
	);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}
	return &RecipeTypesDao{db: db}, nil
}

func (d *RecipeTypesDao) InsertRecipeType(rt RecipeType) error {
	_, err := d.db.Exec("INSERT INTO recipe_types(name, worker_id) VALUES(?, ?)", rt.Name, rt.WorkerID)
	return err
}

func readRecipeTypes(rows *sql.Rows) ([]RecipeType, error) {
	var recipeTypes []RecipeType
	for rows.Next() {
		var rt RecipeType
		err := rows.Scan(&rt.Name, &rt.WorkerID)
		if err != nil {
			return nil, err
		}
		recipeTypes = append(recipeTypes, rt)
	}
	return recipeTypes, nil
}

func (d *RecipeTypesDao) GetRecipeTypes() ([]RecipeType, error) {
	rows, err := d.db.Query("SELECT name, worker_id FROM recipe_types")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readRecipeTypes(rows)
}

func (d *RecipeTypesDao) GetRecipeType(typeName string) (*RecipeType, error) {
	rows, err := d.db.Query("SELECT name, worker_id FROM recipe_types WHERE name = ? LIMIT 1", typeName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var rt RecipeType
		err := rows.Scan(&rt.Name, &rt.WorkerID)
		if err != nil {
			return nil, err
		}
		return &rt, nil
	}

	return nil, nil
}

func (d *RecipeTypesDao) DeleteRecipeType(typeName string) error {
	var exists int
	err := d.db.QueryRow("SELECT EXISTS (SELECT 1 FROM recipes WHERE type = ?)", typeName).Scan(&exists)
	if exists == 1 {
		log.Printf("Type %s is used by recipes. Can't delete it", typeName)
		return fmt.Errorf("Type %s is used by recipes. Can't delete it", typeName)
	}
	if err != nil {
		return err
	}
	_, err = d.db.Exec("DELETE FROM recipe_types WHERE name = ?", typeName)
	return err
}
