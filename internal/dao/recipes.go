package dao

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/asek-ll/aecc-server/internal/common"
)

type RecipeItem struct {
	ItemUID string
	Amount  int
	Role    string
	Slot    *int
}

type Recipe struct {
	Id           int
	Name         string
	Type         string
	Results      []RecipeItem
	Ingeredients []RecipeItem
}

type RecipesDao struct {
	db *sql.DB
}

func NewRecipesDao(db *sql.DB) (*RecipesDao, error) {

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS recipes (
		id INTEGER PRIMARY KEY,
		name string NOT NULL,
		type string NOT NULL
	);

	CREATE TABLE IF NOT EXISTS recipe_items (
		recipe_id INTEGER NOT NULL,
		item_uid string NOT NULL,
		amount integer NOT NULL,
		role string NOT NULL,
		slot integer
	);
	CREATE INDEX IF NOT EXISTS recipe_items_idx ON recipe_items(recipe_id);
	`

	_, err := db.Exec(sqlStmt)

	if err != nil {
		return nil, err
	}

	return &RecipesDao{db: db}, nil
}

func readRows(rows *sql.Rows) ([]Recipe, error) {
	recipes := make(map[int]Recipe)
	for rows.Next() {
		var id int
		var name string
		var typ string
		var item_uid string
		var amount int
		var role string
		var slot *int
		err := rows.Scan(&id, &name, &typ, &item_uid, &amount, &role, &slot)
		if err != nil {
			return nil, err
		}
		recipe, ok := recipes[id]
		if !ok {
			recipe = Recipe{
				Id:   id,
				Name: name,
				Type: typ,
			}
			recipes[id] = recipe
		}
		if role == "result" {
			recipe.Results = append(recipe.Results, RecipeItem{
				ItemUID: item_uid,
				Amount:  amount,
				Slot:    slot,
			})
		} else if role == "ingredient" {
			recipe.Ingeredients = append(recipe.Ingeredients, RecipeItem{
				ItemUID: item_uid,
				Amount:  amount,
				Slot:    slot,
			})
		}
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return common.MapValues(recipes), nil
}

func (r *RecipesDao) GetRecipesPage(fromId int) ([]Recipe, error) {

	query := `
	SELECT r.id FROM recipes r
	WHERE r.id >= ?
	ORDER BY r.id
	LIMIT 20
	`

	rows, err := r.db.Query(query, fromId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int

	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	fmt.Println(ids)

	return r.GetRecipesById(ids)
}

func (r *RecipesDao) InsertRecipe(recipe Recipe) error {

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO recipes (name, type) VALUES (?, ?)", recipe.Name, recipe.Type)
	if err != nil {
		return err
	}

	recipeId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	recipe.Id = int(recipeId)

	for _, item := range recipe.Results {
		_, err := tx.Exec("INSERT INTO recipe_items (recipe_id, item_uid, amount, role, slot) VALUES (?, ?, ?, ?, ?)", recipe.Id, item.ItemUID, item.Amount, "result", item.Slot)
		if err != nil {
			return err
		}
	}

	for _, item := range recipe.Ingeredients {
		_, err := tx.Exec("INSERT INTO recipe_items (recipe_id, item_uid, amount, role, slot) VALUES (?, ?, ?, ?, ?)", recipe.Id, item.ItemUID, item.Amount, "ingredient", item.Slot)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()

	return err
}

func (r *RecipesDao) GetRecipeByResult(itemUID string) ([]Recipe, error) {

	query := `
	SELECT r.*, ri.item_uid, ri.amount, ri.role, ri.slot FROM recipes r 
	JOIN recipe_items ri ON r.id = ri.recipe_id 
	WHERE ri.item_uid = ?
	`

	rows, err := r.db.Query(query, itemUID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return readRows(rows)
}

func (r *RecipesDao) GetRecipesById(ids []int) ([]Recipe, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(`
	SELECT r.*, ri.item_uid, ri.amount, ri.role, ri.slot FROM recipes r 
	JOIN recipe_items ri ON r.id = ri.recipe_id 
	WHERE r.id IN (?%s)
	`, strings.Repeat(", ?", len(ids)-1))

	rows, err := r.db.Query(query, common.ToArgs(ids)...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return readRows(rows)
}
