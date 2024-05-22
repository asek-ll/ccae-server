package recipe

import (
	"errors"
	"strconv"

	"github.com/asek-ll/aecc-server/internal/dao"
)

type RecipeManager struct {
	daoProvider *dao.DaoProvider
}

func NewRecipeManager(daoProvider *dao.DaoProvider) *RecipeManager {
	return &RecipeManager{
		daoProvider: daoProvider,
	}
}

type CreateRecipeParams struct {
	Name  string
	Type  string
	Items []map[string]string
}

func (m *RecipeManager) CreateRecipe(params CreateRecipeParams) error {

	for _, item := range params.Items {
		itemUID, e := item["item"]
		if !e {
			return errors.New("Item uid not specified")
		}

		var err error
		var amount int
		amountStr, e := item["amount"]
		if e {
			amount, err = strconv.Atoi(amountStr)
			if err != nil {
				return err
			}
		}

		slotKey, e := item["slot"]
		if e {

		}
	}

}
