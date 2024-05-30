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
			components.Page("ERROR", components.ErrorMessage(err.Error())).Render(r.Context(), w)
			return
		}
	})
}

func loggingMiddleware(log *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, req)
			log.Printf("[INFO] %s %s %s",
				color.RedString(req.Method),
				color.YellowString(req.RequestURI),
				color.CyanString(time.Since(start).String()),
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
		clients, err := app.Daos.Clients.GetClients()
		if err != nil {
			return err
		}

		return components.ClientsPage(clients).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /items/{$}", func(w http.ResponseWriter, r *http.Request) error {
		items, err := app.Storage.GetItems()
		if err != nil {
			return err
		}

		return components.ItemsPage(items).Render(r.Context(), w)
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

		ctx, err := app.Daos.Items.NewDeferedLoader().FromRecipes(item.Recipes).ToContext(r.Context())
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

		return components.CraftingPlanPage(plan).Render(ctx, w)
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

	return mux, nil
}
