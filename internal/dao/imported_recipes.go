package dao

import (
	"database/sql"
)

type ItemTag struct {
	ID   int
	Name string
}

type ImportedRecipe struct {
	ID          int
	ResultUID   string
	ResultCount int
}

type ImportedRecipesDao struct {
	db *sql.DB
}

func NewImportedRecipesDao(db *sql.DB) (*ImportedRecipesDao, error) {

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS item_tag (
		item_uid string NOT NULL,
		name string NOT NULL
	);

	CREATE TABLE IF NOT EXISTS imported_recipe (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		result_uid string NOT NULL,
		result_count integer NOT NULL
	);

	CREATE TABLE IF NOT EXISTS imported_recipe_ingredient (
		recipe_id integer NOT NULL,
		slot integer NOT NULL,
		item_id integer,
		item_tag integer
	);

	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &ImportedRecipesDao{
		db: db,
	}, nil
}

func (d *ImportedRecipesDao) InsertTag(name string, itemUID string) error {
	_, err := d.db.Exec("INSERT INTO item_tag(name, item_uid) VALUES(?, ?)", name, itemUID)
	return err
}
