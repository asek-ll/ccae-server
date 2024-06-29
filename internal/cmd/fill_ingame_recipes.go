package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = FillInGameRecipesCommand{}

type FillInGameRecipesCommand struct {
}

type InGameIngredient struct {
	Item  *string `json:"item"`
	Tag   *string `json:"tag"`
	Count *int    `json:"count"`
	NBT   *string `json:"nbt"`
}

type InGameResult struct {
	Item  string  `json:"item"`
	Count *int    `json:"count"`
	NBT   *string `json:"nbt"`
}

type InGameRecipe struct {
	Ingredients []json.RawMessage `json:"ingredients"`
	Result      json.RawMessage   `json:"result"`
	Width       *int              `json:"w"`
	Height      *int              `json:"h"`
}

type InGameRecipeType struct {
	Title   string         `json:"title"`
	Mod     string         `json:"mod"`
	Recipes []InGameRecipe `json:"recipes"`
}

func (s FillInGameRecipesCommand) Execute(args []string) error {
	db, err := sql.Open("sqlite3", "data.db")

	if err != nil {
		return err
	}

	_, err = db.Exec(`
	DROP TABLE IF EXISTS imported_recipe;
	DROP TABLE IF EXISTS imported_recipe_ingredient;
	`)

	if err != nil {
		return err
	}

	importedDao, err := dao.NewImportedRecipesDao(db)
	if err != nil {
		return err
	}

	jsonFile, err := os.Open("all_recipes.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)
	var data []InGameRecipeType
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}

	for _, rt := range data {
		for _, r := range rt.Recipes {
			isValid := true

			w := 3
			if r.Width != nil {
				w = *r.Width
			}
			h := 3
			if r.Height != nil {
				h = *r.Height
			}

			var importedIngredients []dao.ImportedIngredient
			for slot, is := range r.Ingredients {

				var ings []InGameIngredient
				err := json.Unmarshal(is, &ings)
				if err != nil {
					var ing InGameIngredient
					err = json.Unmarshal(is, &ing)
					if err != nil {
						return err
					}
					ings = append(ings, ing)
				}

				for _, ing := range ings {
					row := slot / w
					column := slot % w

					convertedSlot := (row*3 + column) + 1
					if row >= h {
						// if r.Result.NBT != nil {
						// 	fmt.Println(rt.Title, r.Result, *r.Result.NBT, slot+1, convertedSlot)
						// } else {
						// 	fmt.Println(rt.Title, r.Result, slot+1, convertedSlot)
						// }
						// for _, ing2 := range importedIngredients {
						// 	if ing2.Item != nil {
						// 		fmt.Println(*ing2.Item)
						// 	} else if ing2.ItemTag != nil {
						// 		fmt.Println(*ing2.ItemTag)
						// 	} else {
						// 		fmt.Println("<no item>")
						// 	}
						// }
						// return fmt.Errorf("Invalid wxh format: %d %d", row, h)
						isValid = false
					}

					// if convertedSlot != slot+1 {
					// 	fmt.Printf("Convert %d to %d for %dx%d\n", slot+1, convertedSlot, w, h)
					// }

					if ing.Item == nil && ing.Tag == nil {
						// fmt.Printf("Invalid ingredient for %v\n", r.Result)
						isValid = false
					}

					importedIngredients = append(importedIngredients, dao.ImportedIngredient{
						Slot:    convertedSlot,
						Item:    ing.Item,
						ItemTag: ing.Tag,
						Count:   ing.Count,
						NBT:     ing.NBT,
					})
				}

			}

			if isValid {

				var reses []InGameResult
				err := json.Unmarshal(r.Result, &reses)
				if err != nil {
					var res InGameResult
					err := json.Unmarshal(r.Result, &res)
					if err != nil {
						var res any
						json.Unmarshal(r.Result, &res)
						fmt.Println(res)
						return err
					}
					reses = append(reses, res)
				}
				if len(reses) > 1 {
					fmt.Println(reses)
				}

				err = importedDao.InsertRecipe(dao.ImportedRecipe{
					ResultID:    reses[0].Item,
					ResultCount: reses[0].Count,
					ResultNBT:   reses[0].NBT,
					Ingredients: importedIngredients,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
