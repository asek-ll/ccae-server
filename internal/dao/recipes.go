package dao

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/asek-ll/aecc-server/internal/common"
)

const RESULT_ROLE = "result"
const INGREDIENT_ROLE = "ingredient"

type RecipeItem struct {
	ItemUID string
	Amount  int
	Role    string
	Slot    *int
}

type Recipe struct {
	ID          int
	Name        string
	Type        string
	Results     []RecipeItem
	Ingredients []RecipeItem
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

func readRows(rows *sql.Rows) ([]*Recipe, error) {
	var recipes []*Recipe
	recipes_by_id := make(map[int]int)
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
		recipe_idx, ok := recipes_by_id[id]
		if !ok {
			recipe := Recipe{
				ID:   id,
				Name: name,
				Type: typ,
			}
			recipe_idx = len(recipes)
			recipes_by_id[id] = recipe_idx
			recipes = append(recipes, &recipe)
		}
		recipe := recipes[recipe_idx]
		if role == RESULT_ROLE {
			recipe.Results = append(recipe.Results, RecipeItem{
				ItemUID: item_uid,
				Amount:  amount,
				Slot:    slot,
			})
		} else if role == INGREDIENT_ROLE {
			recipe.Ingredients = append(recipe.Ingredients, RecipeItem{
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

	// fmt.Println("REC", recipes[0].Ingredients)

	return recipes, nil
}

func (r *RecipesDao) GetRecipesPage(fromId int) ([]*Recipe, error) {

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
	recipe.ID = int(recipeId)

	for _, item := range recipe.Results {
		_, err := tx.Exec("INSERT INTO recipe_items (recipe_id, item_uid, amount, role, slot) VALUES (?, ?, ?, ?, ?)",
			recipe.ID, item.ItemUID, item.Amount, RESULT_ROLE, item.Slot)
		if err != nil {
			return err
		}
	}

	for _, item := range recipe.Ingredients {
		_, err := tx.Exec("INSERT INTO recipe_items (recipe_id, item_uid, amount, role, slot) VALUES (?, ?, ?, ?, ?)",
			recipe.ID, item.ItemUID, item.Amount, INGREDIENT_ROLE, item.Slot)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()

	return err
}

func (r *RecipesDao) GetRecipeByResult(itemUID string) ([]*Recipe, error) {

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

func (r *RecipesDao) GetRecipesByResults(itemUIDs []string) ([]*Recipe, error) {
	if len(itemUIDs) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(`
	SELECT ri.recipe_id FROM recipe_items ri 
	WHERE ri.item_uid IN (?%s) AND ri.role = 'result'
	`, strings.Repeat(", ?", len(itemUIDs)-1))

	rows, err := r.db.Query(query, common.ToArgs(itemUIDs)...)
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

	fmt.Println("Search", itemUIDs, "find", len(ids))

	return r.GetRecipesById(ids)
}

func (r *RecipesDao) GetRecipesById(ids []int) ([]*Recipe, error) {
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
