package crafter

import (
	"errors"
	"fmt"
	"log"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type Crafter struct {
	planner       *Planner
	daoProvider   *dao.DaoProvider
	clientManager *wsmethods.ClientsManager
	storage       *storage.Storage
}

func NewCrafter(
	daoProvider *dao.DaoProvider,
	planner *Planner,
	clientManager *wsmethods.ClientsManager,
	storage *storage.Storage,
) *Crafter {
	return &Crafter{
		daoProvider:   daoProvider,
		planner:       planner,
		clientManager: clientManager,
		storage:       storage,
	}
}

func (c *Crafter) CheckNextSteps(plan *dao.PlanState) error {
	var recipesIds []int
	for _, step := range plan.Steps {
		recipesIds = append(recipesIds, step.RecipeID)
	}

	recipes, err := c.daoProvider.Recipes.GetRecipesById(recipesIds)
	if err != nil {
		return err
	}

	store := make(map[string]int)
	for _, item := range plan.Items {
		store[item.ItemUID] = item.Amount
	}

	recipesById := make(map[int]*dao.Recipe)
	for _, recipe := range recipes {
		recipesById[recipe.ID] = recipe
	}

	for _, step := range plan.Steps {
		recipe := recipesById[step.RecipeID]
		minRepeats := step.Repeats
		for _, ing := range recipe.Ingredients {
			minRepeats = min(minRepeats, store[ing.ItemUID]/ing.Amount)
		}
		if minRepeats > 0 {
			log.Printf("[INFO] Submit craft %v", recipe)
			err = c.submitCrafting(plan, recipe, minRepeats)
			if err != nil {
				return err
			}
			store = make(map[string]int)
			for _, item := range plan.Items {
				store[item.ItemUID] = item.Amount
			}
		}
	}

	log.Println("[INFO] done ping")
	return nil
}

func (c *Crafter) submitCrafting(plan *dao.PlanState, recipe *dao.Recipe, repeats int) error {
	err := c.daoProvider.Crafts.InsertCraft(plan.ID, recipe, repeats)
	if err != nil {
		return err
	}

	plan.Items, err = c.daoProvider.Plans.GetPlanItemState(plan.ID)
	return nil
}

func (c *Crafter) SchedulePlanForItem(uid string, count int) (*dao.PlanState, error) {
	plan, err := c.planner.GetPlanForItem(uid, count)
	if err != nil {
		return nil, err
	}

	var planItems []dao.PlanItemState
	for _, rel := range plan.Related {
		if rel.ResultAmount < 0 {
			return nil, errors.New("Not enough items in storage")
		}
		if rel.ResultAmount < rel.StorageAmount {
			planItems = append(planItems, dao.PlanItemState{
				ItemUID: rel.ItemUID,
				Amount:  rel.StorageAmount - rel.ResultAmount,
			})
		}
	}
	var planSteps []dao.PlanStepState

	for _, step := range plan.Steps {
		planSteps = append(planSteps, dao.PlanStepState{
			RecipeID: step.Recipe.ID,
			Repeats:  step.Repeats,
		})
	}

	planState := dao.PlanState{
		Status:      "SCHEDULED",
		Items:       planItems,
		Steps:       planSteps,
		GoalItemUID: uid,
		GoalAmount:  count,
	}

	err = c.daoProvider.Plans.InsertPlan(&planState)
	if err != nil {
		return nil, err
	}

	return &planState, nil
}

func (c *Crafter) getWorkerForType(recipeType string) (Worker, error) {
	if recipeType == "" {
		crafter, err := wsmethods.GetClientForType[*wsmethods.CrafterClient](c.clientManager)
		if err != nil {
			return nil, err
		}
		if crafter == nil {
			return nil, fmt.Errorf("Crafter not found")
		}

		return NewShapedCrafter(c.storage, crafter), nil
	}
	return nil, fmt.Errorf("Worker for type %s not found", recipeType)
}

func (c *Crafter) Craft(craftId int) error {
	craft, err := c.daoProvider.Crafts.FindById(craftId)
	if err != nil {
		return err
	}
	recipe, err := c.daoProvider.Recipes.GetRecipeById(craft.RecipeID)
	if err != nil {
		return err
	}

	worker, err := c.getWorkerForType(recipe.Type)
	if err != nil {
		return err
	}

	return worker.Craft(recipe, craft.Repeats)
}
