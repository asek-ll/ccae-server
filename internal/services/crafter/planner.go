package crafter

import (
	"errors"

	"github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
)

type Planner struct {
	daoProvider *dao.DaoProvider
	storage     *storage.Storage
}

func NewPlanner(daoProvider *dao.DaoProvider, storage *storage.Storage) *Planner {
	return &Planner{
		daoProvider: daoProvider,
		storage:     storage,
	}
}

type ExpandState struct {
	Items   []string
	Recipes map[string]*dao.Recipe
}

func (p *Planner) expandRecipes(itemIds []string) (*ExpandState, error) {
	deps := make(map[string][]string)
	items := make(map[string]struct{})
	recipeByResult := make(map[string]*dao.Recipe)

	layer := itemIds
	for len(layer) > 0 {
		var recipesToLoad []string
		for _, itemId := range layer {
			if _, e := items[itemId]; !e {
				recipesToLoad = append(recipesToLoad, itemId)
				items[itemId] = struct{}{}
			}
		}
		recipes, err := p.daoProvider.Recipes.GetRecipesByResults(recipesToLoad)
		if err != nil {
			return nil, err
		}

		nextItems := make(map[string]struct{})
		for _, recipe := range recipes {
			result := recipe.Results[0].ItemUID
			recipeByResult[result] = recipe
			for _, ing := range recipe.Ingredients {
				nextItems[ing.ItemUID] = struct{}{}
				deps[result] = append(deps[result], ing.ItemUID)
			}
		}
		layer = common.MapKeys(nextItems)
	}

	orderedItems := common.TopologicalSort(common.MapKeys(items), deps)
	if len(orderedItems) == 0 {
		return nil, errors.New("Cycle detected")
	}

	plan := &ExpandState{
		Items:   orderedItems,
		Recipes: recipeByResult,
	}

	return plan, nil
}

func ceil(x, y int) int {
	rem := x % y
	if rem == 0 {
		return x / y
	}
	return 1 + ((x - rem) / y)
}

func (p *Planner) GetPlanForItem(uid string, count int) (*Plan, error) {

	expandState, err := p.expandRecipes([]string{uid})
	if err != nil {
		return nil, err
	}

	storageCounts, err := p.storage.GetItemsCount()

	state := make(map[string]int)

	for _, item := range expandState.Items {
		if count, e := storageCounts[item]; e {
			state[item] = count
		}
	}
	var steps []Step

	related := make(map[string]*Related)

	for _, item := range expandState.Items {
		if state[item] >= 0 && item != uid {
			continue
		}

		var toCraft int
		if item == uid {
			toCraft = count
		} else {
			toCraft = -state[item]
		}

		recipe, e := expandState.Recipes[item]
		if !e {
			continue
		}

		repeats := ceil(toCraft, recipe.Results[0].Amount)

		for _, ing := range recipe.Ingredients {
			ingredientCount := ing.Amount * repeats
			state[ing.ItemUID] -= ingredientCount

			r, e := related[ing.ItemUID]
			if !e {
				r = &Related{StorageAmount: storageCounts[ing.ItemUID]}
				related[ing.ItemUID] = r
			}
			r.Consumed += ingredientCount
		}

		for _, ing := range recipe.Results {
			ingredientCount := ing.Amount * repeats
			state[ing.ItemUID] += ingredientCount

			r, e := related[ing.ItemUID]
			if !e {
				r = &Related{StorageAmount: storageCounts[ing.ItemUID]}
				related[ing.ItemUID] = r
			}
			r.Produced += ingredientCount
		}

		steps = append(steps, Step{
			Recipe:  recipe,
			Repeats: repeats,
		})
	}

	plan := Plan{
		Items:   expandState.Items,
		Steps:   steps,
		Related: related,
	}

	return &plan, nil
}
