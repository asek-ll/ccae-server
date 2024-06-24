package dao

import (
	"database/sql"
	"log"

	"github.com/asek-ll/aecc-server/internal/common"
)

type ImportedIngredient struct {
	Slot    int
	Item    *string
	ItemTag *string
	Count   *int
	NBT     *string
}

type ImportedRecipe struct {
	ID          int
	ResultID    string
	ResultCount *int
	ResultNBT   *string
	Ingredients []ImportedIngredient
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
		result_id string NOT NULL,
		result_count integer,
		result_nbt string
	);

	CREATE TABLE IF NOT EXISTS imported_recipe_ingredient (
		recipe_id integer NOT NULL,
		slot integer NOT NULL,
		item string,
		item_tag string,
		count integer,
		nbt string
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

func (d *ImportedRecipesDao) InsertRecipe(recipe ImportedRecipe) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO imported_recipe(result_id, result_count, result_nbt) VALUES(?, ?, ?)", recipe.ResultID, recipe.ResultCount, recipe.ResultNBT)
	if err != nil {
		return err
	}

	recipeId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	recipe.ID = int(recipeId)

	for _, item := range recipe.Ingredients {
		_, err := tx.Exec("INSERT INTO imported_recipe_ingredient(recipe_id, slot, item, item_tag, count, nbt) VALUES(?, ?, ?, ?, ?, ?)", recipe.ID, item.Slot, item.Item, item.ItemTag, item.Count, item.NBT)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *ImportedRecipesDao) FindRecipeByResult(uid string) error {
	itemId, nbt := common.FromUid(uid)

	var rows *sql.Rows
	var err error
	if nbt == nil {
		rows, err = d.db.Query("SELECT id, result_count WHERE result_id = ? and result_nbt IS NULL", itemId)
	} else {
		rows, err = d.db.Query("SELECT id, result_count WHERE result_id = ? and result_nbt = ?", itemId, nbt)
	}
	if err != nil {
		return err
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	if !rows.Next() {
		return nil
	}

	var id int
	var resultCount int

	err = rows.Scan(&id, &resultCount)
	if err != nil {
		return err
	}

	stmt := `
	SELECT iri.slot, iri.count, iri.item, iri.nbt, it.item_uid
	FROM imported_recipe_ingredient iri ON iri.recipe_id = ir.id
	LEFT JOIN item_tag it ON it.name = iri.item_tag
	WHERE iri.recipe_id = ?
	`

	rows, err = d.db.Query(stmt, id)
	if err != nil {
		return err
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	for rows.Next() {
		var slot int
		var count *int
		var item *string
		var nbt *string
		var uid_from_tag *string
		err = rows.Scan(&slot, &count, &item, &nbt, &uid_from_tag)
		if err != nil {
			return err
		}
		var uid string
		if uid_from_tag != nil {
			uid = *uid_from_tag
		} else if item != nil {
			uid = common.MakeUid(*item, nbt)
		} else {
			log.Printf("[WARN] Invalid ingredient for recipe: %d", id)
			continue
		}
		amount := 1
		if count != nil {
			amount = *count
		}

		_ = RecipeItem{
			ItemUID: uid,
			Amount:  amount,
			Role:    INGREDIENT_ROLE,
			Slot:    &slot,
		}

	}

	return nil
}
