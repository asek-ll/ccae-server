package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/asek-ll/aecc-server/internal/dao"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jessevdk/go-flags"
)

type Recipe struct {
	Input  []any  `json:"input"`
	Output []any  `json:"output"`
	Type   string `json:"type"`
}

var _ flags.Commander = FillRecipesCommand{}

type FillRecipesCommand struct {
}

func toRecipeItem(item any) (*dao.RecipeItem, error) {
	if item == nil {
		return nil, nil
	}
	var recipeItem dao.RecipeItem
	if uid, ok := item.(string); ok {
		recipeItem.ItemUID = uid
		recipeItem.Amount = 1
		return &recipeItem, nil
	}

	if itemMap, ok := item.(map[string]any); ok {
		recipeItem.ItemUID = itemMap["item"].(string)
		recipeItem.Amount = int(itemMap["count"].(float64))
		return &recipeItem, nil
	}

	return nil, fmt.Errorf("invalid item")
}

func (s FillRecipesCommand) Execute(args []string) error {

	db, err := sql.Open("sqlite3", "data.db")

	if err != nil {
		return err
	}

	_, err = db.Exec(`
	DROP TABLE IF EXISTS recipe_items;
	DROP TABLE IF EXISTS recipes;
	`)

	if err != nil {
		return err
	}

	recipesDao, err := dao.NewRecipesDao(db)
	if err != nil {
		return err
	}

	jsonFile, err := os.Open("recipes.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)
	var data map[string]Recipe
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}

	for key, d := range data {
		resultItem, err := toRecipeItem(key)
		if err != nil {
			return err
		}
		if resultItem == nil {
			return fmt.Errorf("invalid result item")
		}

		recipe := dao.Recipe{}
		recipe.Type = d.Type
		recipe.Name = key

		for _, result := range d.Output {
			recipeItem, err := toRecipeItem(result)
			if err != nil {
				return err
			}
			if recipeItem == nil {
				continue
			}
			recipe.Results = append(recipe.Results, *recipeItem)
		}

		for i, input := range d.Input {
			recipeItem, err := toRecipeItem(input)
			if err != nil {
				return err
			}
			if recipeItem == nil {
				continue
			}
			item := *recipeItem
			item.Role = "ingredient"
			if d.Type == "" {
				slot := i + 1
				item.Slot = &slot
			}
			recipe.Ingredients = append(recipe.Ingredients, item)
		}

		err = recipesDao.InsertRecipe(recipe)
		if err != nil {
			return err
		}
	}

	return nil
}
