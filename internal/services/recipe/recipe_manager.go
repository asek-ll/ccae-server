package recipe

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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

type CreateRecipeItemParams struct {
	ItemUID string
	Amount  int
	Slot    *int
	Role    string
}

type CreateRecipeParams struct {
	Name  string
	Type  string
	Items []CreateRecipeItemParams
}

func parseCreateRecipeParams(params url.Values) (*CreateRecipeParams, error) {
	name := strings.TrimSpace(params.Get("name"))
	if name == "" {
		return nil, errors.New("Name can't be empty")
	}

	recipeType := params.Get("recipeType")

	items := make(map[string]map[string]string)
	props := []string{"item", "slot", "amount", "role"}
outer:
	for k, v := range params {
		for _, prop := range props {
			prefix := prop + "_"
			if strings.HasPrefix(k, prefix) {
				id := strings.TrimPrefix(k, prefix)
				if i, e := items[id]; !e {
					i = make(map[string]string)
					items[id] = i
				}
				items[id][prop] = v[0]
				continue outer
			}
		}
	}

	var itemsParams []CreateRecipeItemParams

	for _, item := range items {
		itemParams, err := parseItemParams(item)
		if err != nil {
			return nil, err
		}

		itemsParams = append(itemsParams, itemParams)
	}

	return &CreateRecipeParams{
		Name:  name,
		Type:  recipeType,
		Items: itemsParams,
	}, nil
}

func parseItemParams(item map[string]string) (CreateRecipeItemParams, error) {
	var params CreateRecipeItemParams
	itemUID := strings.TrimSpace(item["item"])
	if itemUID == "" {
		return params, errors.New("Item uid not specified")
	}
	params.ItemUID = itemUID

	var err error
	var amount int
	amountStr, e := item["amount"]
	if e {
		amount, err = strconv.Atoi(amountStr)
		if err != nil {
			return params, err
		}
	} else {
		amount = 1
	}
	params.Amount = amount

	slot := strings.TrimSpace(item["slot"])
	if slot != "" {
		slotIdx, err := strconv.Atoi(slot)
		if err != nil {
			return params, fmt.Errorf("Invalid slot index '%s'", slot)
		}
		params.Slot = &slotIdx
	}

	params.Role = strings.TrimSpace(item["role"])
	if params.Role == "" {
		return params, errors.New("Role must be scepcified")
	}

	return params, nil
}

func (m *RecipeManager) CreateRecipeFromParams(values url.Values) (*dao.Recipe, error) {
	params, err := parseCreateRecipeParams(values)
	if err != nil {
		return nil, err
	}

	return m.CreateRecipe(params)
}

func (m *RecipeManager) ParseRecipeFromParams(values url.Values) (*dao.Recipe, error) {
	params, err := parseCreateRecipeParams(values)
	if err != nil {
		return nil, err
	}
	return m.validateCreateParams(params)
}

func (m *RecipeManager) validateCreateParams(params *CreateRecipeParams) (*dao.Recipe, error) {
	var uids []string
	for _, item := range params.Items {
		uids = append(uids, item.ItemUID)
	}

	items, err := m.daoProvider.Items.FindItemsByUids(uids)
	if err != nil {
		return nil, err
	}

	existsItemUids := make(map[string]struct{})
	for _, item := range items {
		existsItemUids[item.UID] = struct{}{}
	}

	for _, uid := range uids {
		if _, e := existsItemUids[uid]; !e {
			return nil, fmt.Errorf("Item with uid '%s' does not exists", uid)
		}
	}

	var results []dao.RecipeItem
	var ingredients []dao.RecipeItem

	for _, item := range params.Items {
		recipeItem := dao.RecipeItem{
			ItemUID: item.ItemUID,
			Amount:  item.Amount,
			Role:    item.Role,
			Slot:    item.Slot,
		}

		switch item.Role {
		case dao.INGREDIENT_ROLE:
			ingredients = append(ingredients, recipeItem)
		case dao.RESULT_ROLE:
			results = append(results, recipeItem)
		default:
			return nil, fmt.Errorf("Unsupported role '%s'", item.Role)
		}
	}

	recipe := dao.Recipe{
		Name:        params.Name,
		Type:        params.Type,
		Results:     results,
		Ingredients: ingredients,
	}

	return &recipe, nil
}

func (m *RecipeManager) CreateRecipe(params *CreateRecipeParams) (*dao.Recipe, error) {

	recipe, err := m.validateCreateParams(params)

	if err != nil {
		return nil, err
	}

	err = m.daoProvider.Recipes.InsertRecipe(recipe)
	if err != nil {
		return nil, err
	}
	return recipe, nil
}
