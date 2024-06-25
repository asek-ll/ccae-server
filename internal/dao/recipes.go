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
	MaxRepeats  *int
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
		type string NOT NULL,
		max_repeats integer
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

func readRecipes(rows *sql.Rows) ([]*Recipe, error) {
	var recipes []*Recipe
	recipes_by_id := make(map[int]int)
	for rows.Next() {
		var id int
		var name string
		var typ string
		var item_uid *string
		var amount *int
		var role *string
		var slot *int
		var maxRepeats *int
		err := rows.Scan(&id, &name, &typ, &maxRepeats, &item_uid, &amount, &role, &slot)
		if err != nil {
			return nil, err
		}
		recipe_idx, ok := recipes_by_id[id]
		if !ok {
			recipe := Recipe{
				ID:         id,
				Name:       name,
				Type:       typ,
				MaxRepeats: maxRepeats,
			}
			recipe_idx = len(recipes)
			recipes_by_id[id] = recipe_idx
			recipes = append(recipes, &recipe)
		}
		recipe := recipes[recipe_idx]
		if role != nil {
			if *role == RESULT_ROLE {
				recipe.Results = append(recipe.Results, RecipeItem{
					ItemUID: *item_uid,
					Amount:  *amount,
					Role:    *role,
					Slot:    slot,
				})
			} else if *role == INGREDIENT_ROLE {
				recipe.Ingredients = append(recipe.Ingredients, RecipeItem{
					ItemUID: *item_uid,
					Amount:  *amount,
					Role:    *role,
					Slot:    slot,
				})
			}
		}
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	// fmt.Println("REC", recipes[0].Ingredients)

	return recipes, nil
}

func (r *RecipesDao) GetRecipesPage(filter string, fromId int) ([]*Recipe, error) {

	query := `
	SELECT r.id FROM recipes r
	WHERE r.id >= ? AND r.name LIKE ?
	ORDER BY r.id
	LIMIT 20
	`

	rows, err := r.db.Query(query, fromId, "%"+filter+"%")
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

func (r *RecipesDao) InsertRecipe(recipe *Recipe) error {

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO recipes (name, type, max_repeats) VALUES (?, ?, ?)",
		recipe.Name, recipe.Type, recipe.MaxRepeats)
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

func (r *RecipesDao) UpdateRecipe(recipe *Recipe) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
	UPDATE recipes 
	SET 
		name = ?,
		type = ? ,
		max_repeats = ?
	WHERE 
		id = ?`, recipe.Name, recipe.Type, recipe.MaxRepeats, recipe.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM recipe_items WHERE recipe_id = ?", recipe.ID)
	if err != nil {
		return err
	}

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

	return r.GetRecipesById(ids)
}

func (r *RecipesDao) GetRecipesById(ids []int) ([]*Recipe, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(`
	SELECT r.id, r.name, r.type, r.max_repeats, ri.item_uid, ri.amount, ri.role, ri.slot FROM recipes r 
	LEFT JOIN recipe_items ri ON r.id = ri.recipe_id 
	WHERE r.id IN (?%s)
	`, strings.Repeat(", ?", len(ids)-1))

	rows, err := r.db.Query(query, common.ToArgs(ids)...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return readRecipes(rows)
}

func (r *RecipesDao) GetRecipeById(recipeId int) (*Recipe, error) {
	recipes, err := r.GetRecipesById([]int{recipeId})
	if err != nil {
		return nil, err
	}
	if len(recipes) == 0 {
		return nil, fmt.Errorf("Recipe with id %d not found", recipeId)
	}
	return recipes[0], nil
}

func (r *RecipesDao) DeleteRecipe(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM recipes WHERE id = ?", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM recipe_items WHERE recipe_id = ?", id)
	if err != nil {
		return err
	}

	err = tx.Commit()

	return err
}
