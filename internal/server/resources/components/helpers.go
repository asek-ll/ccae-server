package components

import "github.com/asek-ll/aecc-server/internal/dao"

func FormatIngredients(r *dao.Recipe) [][]*dao.RecipeItem {
	var rows [][]*dao.RecipeItem
	if r.Type == "" {
		rows = append(rows, make([]*dao.RecipeItem, 3), make([]*dao.RecipeItem, 3), make([]*dao.RecipeItem, 3))
		for _, ri := range r.Ingredients {
			r := ((*ri.Slot) - 1) / 3
			c := ((*ri.Slot) - 1) % 3
			rows[r][c] = &ri
		}
	} else {
		var currentRow []*dao.RecipeItem
		for i, ri := range r.Ingredients {
			if len(currentRow) == 0 {
				currentRow = make([]*dao.RecipeItem, 3)
				rows = append(rows, currentRow)
			}

			c := i % 3

			currentRow[c] = &ri

			if c == 2 {
				currentRow = nil
			}
		}
	}

	return rows
}
