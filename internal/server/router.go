package server

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/asek-ll/aecc-server/internal/app"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/server/resources/components"
	"github.com/asek-ll/aecc-server/pkg/template"
	"github.com/fatih/color"
	"github.com/google/uuid"
)

//go:embed resources
var resources embed.FS

func createStaticHandler(statics fs.FS) func(w http.ResponseWriter, r *http.Request) {
	fs := http.FileServer(http.FS(statics))
	return func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}
}

func handleFuncWithError(mux *MiddlewaresGroup, pattern string, handler func(w http.ResponseWriter, r *http.Request) error) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("[ERROR] %s", err)
			components.Page("ERROR", components.ErrorMessage(err.Error())).Render(r.Context(), w)
			return
		}
	})
}

type rwWithStatus struct {
	http.ResponseWriter
	status int
}

func (w *rwWithStatus) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func loggingMiddleware(log *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			rw := &rwWithStatus{w, http.StatusOK}
			next.ServeHTTP(rw, req)
			log.Printf("[INFO] %s %s %s %s",
				color.RedString(req.Method),
				color.YellowString(req.RequestURI),
				color.CyanString(time.Since(start).String()),
				color.GreenString(strconv.Itoa(rw.status)),
			)
		})
	}
}

func CreateMux(app *app.App) (http.Handler, error) {

	templatesFs, err := fs.Sub(resources, "resources/templates")
	if err != nil {
		return nil, err
	}

	staticsFs, err := fs.Sub(resources, "resources")
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	root := NewMiddlewareGroup(mux)

	tmpls := template.NewTemplates(templatesFs)

	static := root.Group()
	common := root.Group().Use(loggingMiddleware(app.Logger))

	static.HandleFunc("GET /static/", createStaticHandler(staticsFs))

	handleFuncWithError(common, "GET /{$}", func(w http.ResponseWriter, r *http.Request) error {
		return components.Page("INDEX").Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /clients/{$}", func(w http.ResponseWriter, r *http.Request) error {
		clients := app.ClientsManager.GetClients()
		return components.ClientsPage(clients).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /storageItems/{$}", func(w http.ResponseWriter, r *http.Request) error {
		items, err := app.Storage.GetItems()
		if err != nil {
			return err
		}

		return components.ItemsInventory(items, 9).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /items/{$}", func(w http.ResponseWriter, r *http.Request) error {
		filter := r.URL.Query().Get("filter")
		view := r.URL.Query().Get("view")

		items, err := app.Daos.Items.FindByName(filter)
		if err != nil {
			return err
		}

		if view == "list" {
			return components.ItemsList(items).Render(r.Context(), w)
		}

		return components.ItemsListPage(filter, items).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /playerItems/{$}", func(w http.ResponseWriter, r *http.Request) error {
		items, err := app.PlayerManager.GetItems()
		if err != nil {
			return err
		}
		itemLoader := app.Daos.Items.NewDeferedLoader()
		for _, item := range items {
			itemLoader.AddUid(item.ItemID)
		}

		ctx, err := itemLoader.ToContext(r.Context())
		if err != nil {
			return err
		}

		inv := components.Inventory(items, 4, 9)

		return components.Page("Player", inv).Render(ctx, w)
	})

	handleFuncWithError(common, "POST /playerItems/{slot}/drop/{$}", func(w http.ResponseWriter, r *http.Request) error {
		slotStr := r.PathValue("slot")
		slot, err := strconv.Atoi(slotStr)
		if err != nil {
			return err
		}
		_, err = app.PlayerManager.RemoveItem(slot)
		if err != nil {
			return err
		}

		items, err := app.PlayerManager.GetItems()
		if err != nil {
			return err
		}
		itemLoader := app.Daos.Items.NewDeferedLoader()
		for _, item := range items {
			itemLoader.AddUid(item.ItemID)
		}

		ctx, err := itemLoader.ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.Inventory(items, 4, 9).Render(ctx, w)
	})

	handleFuncWithError(common, "GET /items/{itemUid}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		uid := r.PathValue("itemUid")
		item, err := app.Storage.GetItem(uid)
		if err != nil {
			return err
		}

		ctx, err := app.Daos.Items.NewDeferedLoader().FromRecipes(item.Recipes).FromRecipes(item.ImportedRecipes).ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.ItemPage(item).Render(ctx, w)
	})

	handleFuncWithError(common, "GET /lua/client/{role}", func(w http.ResponseWriter, r *http.Request) error {
		role := r.PathValue("role")

		id, err := app.Daos.Seqs.NextId("clientNo")
		if err != nil {
			return err
		}

		return tmpls.Render("client.lua", []string{"client.lua.tmpl"}, w, map[string]any{
			"role":  role,
			"wsUrl": "ws://localhost:12526",
			"id":    id,
		})
	})

	handleFuncWithError(common, "GET /recipes/{$}", func(w http.ResponseWriter, r *http.Request) error {
		filter := r.URL.Query().Get("filter")
		view := r.URL.Query().Get("view")

		recipes, err := app.Daos.Recipes.GetRecipesPage(filter, 0)
		if err != nil {
			return err
		}

		ctx, err := app.Daos.Items.NewDeferedLoader().FromRecipes(recipes).ToContext(r.Context())
		if err != nil {
			return err
		}

		if view == "list" {
			return components.RecipesList(recipes).Render(ctx, w)
		}

		return components.RecipesPage(filter, recipes).Render(ctx, w)
	})

	handleFuncWithError(common, "GET /craft-plan/item/{itemUid}/{count}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		uid := r.PathValue("itemUid")
		strCount := r.PathValue("count")
		count, err := strconv.Atoi(strCount)
		if err != nil {
			count = 1
		}
		plan, err := app.Planner.GetPlanForItem(uid, count)
		if err != nil {
			return err
		}

		ctx, err := app.Daos.Items.NewDeferedLoader().AddUids(plan.Items).ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.CraftingPlanPage(plan, uid, count).Render(ctx, w)
	})

	handleFuncWithError(common, "GET /recipes/new/{$}", func(w http.ResponseWriter, r *http.Request) error {
		recipe, err := app.RecipeManager.ParseRecipeFromParams(r.URL.Query())
		if err != nil {
			return err
		}

		ctx, err := app.Daos.Items.NewDeferedLoader().FromRecipe(recipe).ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.Page(fmt.Sprintf("Recipe for %s", recipe.Name), components.RecipeForm(recipe)).Render(ctx, w)
	})

	handleFuncWithError(common, "GET /recipes/{recipeId}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		rawRecipeId := r.PathValue("recipeId")
		recipeId, err := strconv.Atoi(rawRecipeId)
		if err != nil {
			return err
		}
		recipes, err := app.Daos.Recipes.GetRecipesById([]int{recipeId})
		if err != nil {
			return err
		}
		if len(recipes) == 0 {
			return fmt.Errorf("Recipe with id '%d' not found", recipeId)
		}

		recipe := recipes[0]

		ctx, err := app.Daos.Items.NewDeferedLoader().FromRecipe(recipe).ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.Page("Recipe!!!", components.RecipeForm(recipe)).Render(ctx, w)
	})

	handleFuncWithError(common, "POST /recipes/{recipeId}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		rawRecipeId := r.PathValue("recipeId")
		recipeId, err := strconv.Atoi(rawRecipeId)
		if err != nil {
			return err
		}

		err = r.ParseForm()
		if err != nil {
			return err
		}

		recipe, err := app.RecipeManager.ParseRecipeFromParams(r.PostForm)
		if err != nil {
			return err
		}

		recipe.ID = recipeId

		err = app.Daos.Recipes.UpdateRecipe(recipe)
		if err != nil {
			return err
		}

		ctx, err := app.Daos.Items.NewDeferedLoader().FromRecipe(recipe).ToContext(r.Context())
		if err != nil {
			return err
		}
		return components.Page(fmt.Sprintf("Recipe for %s", recipe.Name), components.RecipeForm(recipe)).Render(ctx, w)
	})

	handleFuncWithError(common, "POST /recipes/new/{$}", func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			return err
		}

		recipe, err := app.RecipeManager.CreateRecipeFromParams(r.PostForm)
		if err != nil {
			return err
		}

		ctx, err := app.Daos.Items.NewDeferedLoader().FromRecipe(recipe).ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.Page(fmt.Sprintf("Recipe for %s", recipe.Name), components.RecipeForm(recipe)).Render(ctx, w)
	})

	handleFuncWithError(common, "DELETE /recipes/{recipeId}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		rawRecipeId := r.PathValue("recipeId")
		recipeId, err := strconv.Atoi(rawRecipeId)
		if err != nil {
			return err
		}

		recipes, err := app.Daos.Recipes.GetRecipesById([]int{recipeId})
		if err != nil {
			return err
		}
		if len(recipes) == 0 {
			return fmt.Errorf("Recipe with id '%d' not found", recipeId)
		}

		err = app.Daos.Recipes.DeleteRecipe(recipeId)
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", "/recipes")
		return nil
	})

	handleFuncWithError(common, "GET /item-popup/{$}", func(w http.ResponseWriter, r *http.Request) error {
		slot := r.URL.Query().Get("slot")
		role := r.URL.Query().Get("role")

		return components.ItemPopup(slot, role).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /item-popup/items/{$}", func(w http.ResponseWriter, r *http.Request) error {
		filter := r.URL.Query().Get("filter")
		items, err := app.Daos.Items.FindByName(filter)
		if err != nil {
			return err
		}
		return components.ItemPopupItems(items).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /item-popup/{uid}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		slot := r.URL.Query().Get("slot")
		role := r.URL.Query().Get("role")

		var slotIdx *int

		if slot != "" {
			slotValue, err := strconv.Atoi(slot)
			if err != nil {
				return err
			}
			slotIdx = &slotValue
		}

		item := dao.RecipeItem{
			ItemUID: r.PathValue("uid"),
			Amount:  1,
			Role:    role,
			Slot:    slotIdx,
		}

		ctx, err := app.Daos.Items.NewDeferedLoader().AddUid(item.ItemUID).ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.ItemInputs(uuid.NewString(), item).Render(ctx, w)
	})

	handleFuncWithError(common, "PUT /configs/{key}/{value}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		key := r.PathValue("key")
		value := r.PathValue("value")

		err := app.Daos.Configs.SetConfig(key, value)
		if err != nil {
			return err
		}

		options, err := app.Daos.Configs.GetConfigOptions()
		if err != nil {
			return err
		}

		return components.OptionList(options).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /configs/{$}", func(w http.ResponseWriter, r *http.Request) error {
		options, err := app.Daos.Configs.GetConfigOptions()
		if err != nil {
			return err
		}

		return components.Page("Configs", components.OptionList(options)).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /craft-plans/{$}", func(w http.ResponseWriter, r *http.Request) error {
		plans, err := app.Daos.Plans.GetPlans()
		if err != nil {
			return err
		}

		itemLoader := app.Daos.Items.NewDeferedLoader()

		for _, plan := range plans {
			itemLoader.AddUid(plan.GoalItemUID)
		}

		ctx, err := itemLoader.ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.Page("Craft plans", components.PlanList(plans)).Render(ctx, w)
	})

	handleFuncWithError(common, "GET /craft-plans/{planId}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		planIdStr := r.PathValue("planId")
		planId, err := strconv.Atoi(planIdStr)
		if err != nil {
			return err
		}
		plan, err := app.Daos.Plans.GetPlanById(planId)
		if err != nil {
			return err
		}

		loader := app.Daos.Items.NewDeferedLoader().AddUid(plan.GoalItemUID)
		for _, item := range plan.Items {
			loader.AddUid(item.ItemUID)
		}

		ctx, err := loader.ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.Page("Craft plan", components.PlanDetail(plan)).Render(ctx, w)
	})

	handleFuncWithError(common, "DELETE /craft-plans/{planId}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		planIdStr := r.PathValue("planId")
		planId, err := strconv.Atoi(planIdStr)
		if err != nil {
			return err
		}
		plan, err := app.Daos.Plans.GetPlanById(planId)
		if err != nil {
			return err
		}

		err = app.Daos.Plans.RemovePlan(plan.ID)
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", "/craft-plans/")
		return nil
	})

	handleFuncWithError(common, "POST /craft-plans/item/{itemUid}/{count}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		uid := r.PathValue("itemUid")
		strCount := r.PathValue("count")
		count, err := strconv.Atoi(strCount)
		if err != nil {
			count = 1
		}
		plan, err := app.Crafter.SchedulePlanForItem(uid, count)
		if err != nil {
			return err
		}

		// ctx, err := app.Daos.Items.NewDeferedLoader().AddUids(plan.Items).ToContext(r.Context())
		// if err != nil {
		// 	return err
		// }

		w.Header().Add("HX-Location", fmt.Sprintf("/craft-plans/%d", plan.ID))
		return nil
		// return componens.Page("Craft plan", components.Plan(plan)).Render(r.Context(), w)
	})

	handleFuncWithError(common, "POST /craft-plans/{planId}/ping/{$}", func(w http.ResponseWriter, r *http.Request) error {
		planIdStr := r.PathValue("planId")
		planId, err := strconv.Atoi(planIdStr)
		if err != nil {
			return err
		}
		plan, err := app.Daos.Plans.GetPlanById(planId)
		if err != nil {
			return err
		}

		err = app.Crafter.CheckNextSteps(plan)
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", fmt.Sprintf("/craft-plans/%d", plan.ID))
		return nil
	})

	handleFuncWithError(common, "GET /crafts/{$}", func(w http.ResponseWriter, r *http.Request) error {
		crafts, err := app.Daos.Crafts.GetAllCrafts()
		if err != nil {
			return err
		}

		var recipesIds []int
		for _, craft := range crafts {
			recipesIds = append(recipesIds, craft.RecipeID)
		}

		recipes, err := app.Daos.Recipes.GetRecipesById(recipesIds)
		if err != nil {
			return err
		}
		recipesById := make(map[int]*dao.Recipe)
		for _, recipe := range recipes {
			recipesById[recipe.ID] = recipe
		}

		itemLoader := app.Daos.Items.NewDeferedLoader()

		var craftItems []components.CraftItem
		for _, craft := range crafts {
			recipe := recipesById[craft.RecipeID]
			itemLoader.FromRecipe(recipe)

			craftItems = append(craftItems, components.CraftItem{
				Craft:  craft,
				Recipe: recipe,
			})
		}

		ctx, err := itemLoader.ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.Page("Craft plans", components.CraftList(craftItems)).Render(ctx, w)
	})

	handleFuncWithError(common, "GET /imported-recipe/{recipeId}/configure/{$}", func(w http.ResponseWriter, r *http.Request) error {
		rawRecipeId := r.PathValue("recipeId")
		recipeId, err := strconv.Atoi(rawRecipeId)
		if err != nil {
			return err
		}

		recipe, err := app.Daos.ImporetedRecipes.FindRecipeById(recipeId)
		if err != nil {
			return err
		}

		ctx, err := app.Daos.Items.NewDeferedLoader().FromRecipe(recipe).ToContext(r.Context())
		if err != nil {
			return err
		}

		slotIngredients := make(map[int][]dao.RecipeItem)
		for _, item := range recipe.Ingredients {
			slotIngredients[*item.Slot] = append(slotIngredients[*item.Slot], item)
		}

		return components.Page(fmt.Sprintf("Imported Recipe for %s", recipe.Name), components.ImportRecipeSlotsSelector(slotIngredients)).Render(ctx, w)
	})

	handleFuncWithError(common, "POST /imported-recipe/{recipeId}/configure/{$}", func(w http.ResponseWriter, r *http.Request) error {
		rawRecipeId := r.PathValue("recipeId")
		recipeId, err := strconv.Atoi(rawRecipeId)
		if err != nil {
			return err
		}

		recipe, err := app.Daos.ImporetedRecipes.FindRecipeById(recipeId)
		if err != nil {
			return err
		}

		err = r.ParseForm()
		if err != nil {
			return err
		}

		uidBySlot := make(map[int]string)

		for k, vs := range r.PostForm {
			slot, err := strconv.Atoi(k)
			if err != nil {
				return err
			}
			if len(vs) == 0 {
				continue
			}
			uidBySlot[slot] = vs[0]
		}

		filteredRecipeItems := make([]dao.RecipeItem, 0, len(recipe.Ingredients))
		for _, item := range recipe.Ingredients {
			if uid, ok := uidBySlot[*item.Slot]; ok && uid != item.ItemUID {
				continue
			}
			filteredRecipeItems = append(filteredRecipeItems, item)
		}

		recipe.Ingredients = filteredRecipeItems

		url := components.RecipeToURL(recipe)

		http.Redirect(w, r, url, http.StatusSeeOther)

		return nil
	})

	handleFuncWithError(common, "POST /crafts/{id}/commit/{$}", func(w http.ResponseWriter, r *http.Request) error {
		id := r.PathValue("id")
		craftId, err := strconv.Atoi(id)
		if err != nil {
			return err
		}

		craft, err := app.Daos.Crafts.FindById(craftId)
		if err != nil {
			return err
		}

		recipe, err := app.Daos.Recipes.GetRecipeById(craft.RecipeID)
		if err != nil {
			return err
		}

		return app.Daos.Crafts.CommitCraft(craft, recipe, 1)
	})
	handleFuncWithError(common, "POST /crafts/{id}/complete/{$}", func(w http.ResponseWriter, r *http.Request) error {
		id := r.PathValue("id")
		craftId, err := strconv.Atoi(id)
		if err != nil {
			return err
		}

		craft, err := app.Daos.Crafts.FindById(craftId)
		if err != nil {
			return err
		}

		return app.Daos.Crafts.CompleteCraft(craft)
	})

	handleFuncWithError(common, "POST /crafts/{id}/cancel/{$}", func(w http.ResponseWriter, r *http.Request) error {
		id := r.PathValue("id")
		craftId, err := strconv.Atoi(id)
		if err != nil {
			return err
		}

		craft, err := app.Daos.Crafts.FindById(craftId)
		if err != nil {
			return err
		}

		recipe, err := app.Daos.Recipes.GetRecipeById(craft.RecipeID)
		if err != nil {
			return err
		}

		return app.Daos.Crafts.CancelCraft(craft, recipe)
	})

	handleFuncWithError(common, "/", func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNotFound)
		return components.Page("Not found").Render(r.Context(), w)
	})

	return mux, nil
}
