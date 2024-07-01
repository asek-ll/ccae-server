package crafter

import (
	"errors"
	"log"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
)

type Crafter struct {
	planner       *Planner
	daoProvider   *dao.DaoProvider
	workerFactory *WorkerFactory
	storage       *storage.Storage
}

func NewCrafter(
	daoProvider *dao.DaoProvider,
	planner *Planner,
	workerFactory *WorkerFactory,
	storage *storage.Storage,
) *Crafter {
	return &Crafter{
		daoProvider:   daoProvider,
		planner:       planner,
		workerFactory: workerFactory,
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
	log.Printf("[INFO] Check for plan %v", store)

	recipesById := make(map[int]*dao.Recipe)
	for _, recipe := range recipes {
		recipesById[recipe.ID] = recipe
	}

	updated := false

	for _, step := range plan.Steps {
		recipe := recipesById[step.RecipeID]

		recipeIngredients := make(map[string]int)
		for _, ing := range recipe.Ingredients {
			recipeIngredients[ing.ItemUID] += ing.Amount
		}

		minRepeats := step.Repeats
		for ing, amount := range recipeIngredients {
			minRepeats = min(minRepeats, store[ing]/amount)
		}

		if minRepeats > 0 {
			log.Printf("[INFO] Submit craft %v", recipe)
			done, err := c.submitCrafting(plan, recipe, minRepeats)
			if err != nil {
				return err
			}
			if done {
				updated = true
				store = make(map[string]int)
				for _, item := range plan.Items {
					store[item.ItemUID] = item.Amount
				}
			}
		}
	}

	if updated {
		err = c.cleanupItems(plan)
		if err != nil {
			return err
		}
	}

	log.Println("[INFO] done ping")
	return nil
}

func (c *Crafter) submitCrafting(plan *dao.PlanState, recipe *dao.Recipe, repeats int) (bool, error) {
	recipeType := recipe.Type
	if recipeType == "" {
		recipeType = "shaped_craft"
	}
	err := c.daoProvider.Crafts.InsertCraft(plan.ID, recipeType, recipe, repeats)
	if err != nil {
		return false, err
	}
	c.workerFactory.Ping(recipeType)

	plan.Items, err = c.daoProvider.Plans.GetPlanItemState(plan.ID)
	return true, nil
}

func (c *Crafter) cleanupItems(plan *dao.PlanState) error {
	return c.daoProvider.Plans.CleanupItems(plan.ID)
}

func (c *Crafter) SchedulePlanForItem(goals []Stack) (*dao.PlanState, error) {
	plan, err := c.planner.GetPlanForItem(goals)
	if err != nil {
		return nil, err
	}

	var planItems []dao.PlanItemState
	for _, rel := range plan.Related {
		resultAmount := rel.StorageAmount + rel.Produced - rel.Consumed
		if resultAmount < 0 {
			return nil, errors.New("Not enough items in storage")
		}
		planItems = append(planItems, dao.PlanItemState{
			ItemUID:        rel.UID,
			Amount:         min(rel.StorageAmount, rel.Consumed),
			RequiredAmount: rel.Consumed,
		})
	}
	var planSteps []dao.PlanStepState

	for _, step := range plan.Steps {
		planSteps = append(planSteps, dao.PlanStepState{
			RecipeID: step.Recipe.ID,
			Repeats:  step.Repeats,
		})
	}

	planGoals := make([]dao.PlanGoal, len(goals))
	for i, goal := range goals {
		planGoals[i] = dao.PlanGoal{
			ItemUID: goal.ItemID,
			Amount:  goal.Count,
		}
	}

	planState := dao.PlanState{
		Status: "SCHEDULED",
		Items:  planItems,
		Steps:  planSteps,
		Goals:  planGoals,
	}

	err = c.daoProvider.Plans.InsertPlan(&planState)
	if err != nil {
		return nil, err
	}

	err = c.CheckNextSteps(&planState)
	if err != nil {
		return nil, err
	}

	return &planState, nil
}

// func (c *Crafter) getWorkerForType(recipeType string) (Worker, error) {
// 	if recipeType == "" {
// 		crafter, err := wsmethods.GetClientForType[*wsmethods.CrafterClient](c.clientManager)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if crafter == nil {
// 			return nil, fmt.Errorf("Crafter not found")
// 		}

// 		return NewShapedCrafter(c.storage, crafter), nil
// 	}
// 	return nil, fmt.Errorf("Worker for type %s not found", recipeType)
// }

// func (c *Crafter) Craft(craftId int) error {
// 	craft, err := c.daoProvider.Crafts.FindById(craftId)
// 	if err != nil {
// 		return err
// 	}
// 	recipe, err := c.daoProvider.Recipes.GetRecipeById(craft.RecipeID)
// 	if err != nil {
// 		return err
// 	}

// 	worker, err := c.getWorkerForType(recipe.Type)
// 	if err != nil {
// 		return err
// 	}

// 	return worker.Craft(recipe, craft.Repeats)
// }
