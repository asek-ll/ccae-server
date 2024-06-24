package cmd

import (
	"database/sql"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = FillInGameRecipesCommand{}

type FillInGameRecipesCommand struct {
}

func (s FillInGameRecipesCommand) Execute(args []string) error {
	db, err := sql.Open("sqlite3", "data.db")

	if err != nil {
		return err
	}

	_, err = db.Exec(`
	DROP TABLE IF EXISTS imported_recipe;
	DROP TABLE IF EXISTS imported_recipes_ingredient;
	`)

	if err != nil {
		return err
	}

	_, err = dao.NewImportedRecipesDao(db)
	if err != nil {
		return err
	}

	return nil
}
