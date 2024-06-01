package crafter

import (
	"errors"

	"github.com/asek-ll/aecc-server/internal/dao"
)

type Crafter struct {
	planner     *Planner
	daoProvider *dao.DaoProvider
}

func NewCrafter(daoProvider *dao.DaoProvider, planner *Planner) *Crafter {
	return &Crafter{
		daoProvider: daoProvider,
		planner:     planner,
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
		maxRepeats := step.Repeats
		for _, ing := range recipe.Ingredients {
			maxRepeats = max(maxRepeats, store[ing.ItemUID]/ing.Amount)
		}
		if maxRepeats > 0 {
			err = c.submitCrafting(plan, recipe, maxRepeats)
			if err != nil {
				return err
			}
			store = make(map[string]int)
			for _, item := range plan.Items {
				store[item.ItemUID] = item.Amount
			}
		}
	}

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
