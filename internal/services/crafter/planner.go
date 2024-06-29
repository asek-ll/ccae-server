package crafter

import (
	"errors"
	"log"
	"sort"

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
	for _, v := range deps {
		sort.Strings(v)
	}
	uids := common.MapKeys(items)
	sort.Strings(uids)
	orderedItems := common.TopologicalSort(uids, deps)
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

func (p *Planner) GetPlanForItem(goals []Stack) (*Plan, error) {
	uids := make([]string, len(goals))
	for i, goal := range goals {
		uids[i] = goal.ItemID
	}

	log.Printf("[INFO] Goal for uids %v", uids)

	expandState, err := p.expandRecipes(uids)
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

	related := make(map[string]*Related)
	for _, goal := range goals {
		state[goal.ItemID] = -goal.Count
		related[goal.ItemID] = &Related{
			UID:      goal.ItemID,
			Consumed: goal.Count,
		}
	}

	var steps []Step

	for _, item := range expandState.Items {
		if state[item] >= 0 {
			continue
		}

		toCraft := -state[item]

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
				r = &Related{
					UID:           ing.ItemUID,
					StorageAmount: storageCounts[ing.ItemUID],
				}
				related[ing.ItemUID] = r
			}
			r.Consumed += ingredientCount
		}

		for _, ing := range recipe.Results {
			ingredientCount := ing.Amount * repeats
			state[ing.ItemUID] += ingredientCount

			r, e := related[ing.ItemUID]
			if !e {
				r = &Related{
					UID:           ing.ItemUID,
					StorageAmount: storageCounts[ing.ItemUID],
				}
				related[ing.ItemUID] = r
			}
			r.Produced += ingredientCount
		}

		steps = append(steps, Step{
			Recipe:  recipe,
			Repeats: repeats,
		})
	}

	rels := common.MapValues(related)
	sort.Slice(rels, func(i, j int) bool {
		vi := rels[i].StorageAmount - rels[i].Consumed + rels[i].Produced
		vj := rels[j].StorageAmount - rels[j].Consumed + rels[j].Produced

		if vi == vj {

			if rels[i].Consumed == rels[j].Consumed {
				return rels[i].UID < rels[j].UID
			}

			return rels[i].Consumed > rels[j].Consumed
		}
		return vi < vj
	})

	goalItems := make([]dao.RecipeItem, len(goals))
	for i, g := range goals {
		goalItems[i] = dao.RecipeItem{
			ItemUID: g.ItemID,
			Amount:  g.Count,
			Role:    "goal",
		}
	}

	plan := Plan{
		Goals:   goalItems,
		Items:   expandState.Items,
		Steps:   steps,
		Related: rels,
	}

	log.Printf("[INFO] Plan %v", plan)

	return &plan, nil
}
