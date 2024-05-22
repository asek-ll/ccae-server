package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/asek-ll/aecc-server/internal/app"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/pkg/template"
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

func CreateMux(app *app.App) (*http.ServeMux, error) {

	templatesFs, err := fs.Sub(resources, "resources/templates")
	if err != nil {
		return nil, err
	}

	staticsFs, err := fs.Sub(resources, "resources")
	if err != nil {
		return nil, err
	}

	tmpls := template.NewTemplates(templatesFs)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /static/", createStaticHandler(staticsFs))

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		tmpls.Render("index", []string{"index.html.tmpl"}, w, nil)
	})
	mux.HandleFunc("GET /clients/{$}", func(w http.ResponseWriter, r *http.Request) {
		clients, err := app.Daos.Clients.GetClients()
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}

		tmpls.Render("clients", []string{"index.html.tmpl", "clients.html.tmpl"}, w, clients)
	})

	mux.HandleFunc("GET /items/{$}", func(w http.ResponseWriter, r *http.Request) {
		items, err := app.Storage.GetItems()
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}

		tmpls.Render("items", []string{"index.html.tmpl", "items.html.tmpl", "item-widget.html.tmpl"}, w, items)
	})

	mux.HandleFunc("GET /items/{itemUid}/{$}", func(w http.ResponseWriter, r *http.Request) {
		uid := r.PathValue("itemUid")
		item, err := app.Storage.GetItem(uid)
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}

		tmpls.Render("item", []string{"index.html.tmpl", "item.html.tmpl", "item-widget.html.tmpl"}, w, item)
	})

	mux.HandleFunc("GET /lua/client/{role}", func(w http.ResponseWriter, r *http.Request) {
		role := r.PathValue("role")

		id, err := app.Daos.Seqs.NextId("clientNo")
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}

		tmpls.Render("client.lua", []string{"client.lua.tmpl"}, w, map[string]any{
			"role":  role,
			"wsUrl": "ws://localhost:12526",
			"id":    id,
		})
	})

	formatIngredients := func(r *dao.Recipe) [][]*dao.RecipeItem {
		var rows [][]*dao.RecipeItem
		if r.Type == "" {
			rows = append(rows, make([]*dao.RecipeItem, 3), make([]*dao.RecipeItem, 3), make([]*dao.RecipeItem, 3))
			for _, ri := range r.Ingredients {
				r := ((*ri.Slot) - 1) / 3
				c := ((*ri.Slot) - 1) % 3
				rows[r][c] = &ri
			}
		} else {
			var currentRow []*dao.RecipeItem
			for i, ri := range r.Ingredients {
				if len(currentRow) == 0 {
					currentRow = make([]*dao.RecipeItem, 3)
					rows = append(rows, currentRow)
				}

				c := i % 3

				currentRow[c] = &ri

				if c == 2 {
					currentRow = nil
				}
			}
		}

		return rows
	}

	mux.HandleFunc("GET /recipes/{$}", func(w http.ResponseWriter, r *http.Request) {
		recipes, err := app.Daos.Recipes.GetRecipesPage(0)
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}

		itemsById := make(map[string]*dao.Item)
		for _, r := range recipes {
			for _, i := range r.Ingredients {
				itemsById[i.ItemUID] = nil
			}
			for _, i := range r.Results {
				itemsById[i.ItemUID] = nil
			}
		}

		err = app.Daos.Items.FindItemsIndexed(itemsById)
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}

		tmpls.Render("recipes", []string{"index.html.tmpl", "recipes.html.tmpl", "item-widget.html.tmpl"}, w, map[string]any{
			"recipes":           recipes,
			"items":             itemsById,
			"formatIngredients": formatIngredients,
		})
	})

	mux.HandleFunc("GET /craft-plan/item/{itemUid}/{count}/{$}", func(w http.ResponseWriter, r *http.Request) {
		uid := r.PathValue("itemUid")
		strCount := r.PathValue("count")
		count, err := strconv.Atoi(strCount)
		if err != nil {
			count = 1
		}
		plan, err := app.Planner.GetPlanForItem(uid, count)
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}

		itemsById := make(map[string]*dao.Item)
		for _, id := range plan.Items {
			itemsById[id] = nil
		}

		err = app.Daos.Items.FindItemsIndexed(itemsById)
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}

		tmpls.Render("craft-plan", []string{"index.html.tmpl", "craftplan.html.tmpl", "item-widget.html.tmpl"}, w, map[string]any{
			"plan":              plan,
			"formatIngredients": formatIngredients,
			"items":             itemsById,
		})
	})

	mux.HandleFunc("GET /recipes/new/{$}", func(w http.ResponseWriter, r *http.Request) {
		tmpls.Render("recipes-new", []string{"index.html.tmpl", "create-recipe.html.tmpl"}, w, nil)
	})

	mux.HandleFunc("GET /recipes/{recipeId}/{$}", func(w http.ResponseWriter, r *http.Request) {
		rawRecipeId := r.PathValue("recipeId")
		recipeId, err := strconv.Atoi(rawRecipeId)
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}
		recipes, err := app.Daos.Recipes.GetRecipesById([]int{recipeId})
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}
		if len(recipes) == 0 {
			tmpls.RenderError(fmt.Errorf("Recipe with id '%d' not found", recipeId), w)
			return
		}

		tmpls.Render("recipes-new", []string{"index.html.tmpl", "create-recipe.html.tmpl"}, w, nil)
	})

	mux.HandleFunc("POST /recipes/new/{$}", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}

		recipe, err := app.RecipeManager.CreateRecipeFromParams(r.PostForm)
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}

		_ = recipe

		tmpls.Render("recipes-new", []string{"index.html.tmpl", "create-recipe.html.tmpl"}, w, nil)
	})

	mux.HandleFunc("GET /item-popup/{$}", func(w http.ResponseWriter, r *http.Request) {
		slot := r.URL.Query().Get("slot")
		tmpls.Render("item-popup", []string{"item-popup.html.tmpl"}, w, map[string]string{
			"slot": slot,
		})
	})

	mux.HandleFunc("GET /item-popup/items/{$}", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		items, err := app.Daos.Items.FindByName(filter)
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}
		tmpls.Render("item-popup-items", []string{"item-popup-items.html.tmpl"}, w, items)
	})

	mux.HandleFunc("GET /item-popup/{uid}/{$}", func(w http.ResponseWriter, r *http.Request) {
		slot := r.URL.Query().Get("slot")
		items, err := app.Daos.Items.FindItemsByUids([]string{r.PathValue("uid")})
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}
		tmpls.Render("item-popup-result", []string{"item-popup-result.html.tmpl"}, w, map[string]any{
			"id":   uuid.NewString(),
			"item": items[0],
			"slot": slot,
		})
	})

	return mux, nil
}
