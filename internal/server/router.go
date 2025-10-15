package server

import (
	"bufio"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	tmpl "text/template"

	"github.com/asek-ll/aecc-server/internal/app"
	"github.com/asek-ll/aecc-server/internal/build"
	cmn "github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/server/handlers"
	"github.com/asek-ll/aecc-server/internal/server/resources/components"
	"github.com/asek-ll/aecc-server/internal/services/clientscripts"
	"github.com/asek-ll/aecc-server/internal/services/crafter"
	"github.com/asek-ll/aecc-server/internal/services/item"
	"github.com/asek-ll/aecc-server/internal/services/recipe"
	"github.com/asek-ll/aecc-server/internal/ws"
	"github.com/asek-ll/aecc-server/pkg/template"
	"github.com/fatih/color"
	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/avatar"
	lg "github.com/go-pkgz/auth/logger"
	"github.com/go-pkgz/auth/token"
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

func (w *rwWithStatus) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
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

func CreateMux(app *app.App, wsServer *ws.Server) (http.Handler, error) {

	authService := auth.NewService(auth.Opts{
		SecretReader: token.SecretFunc(func(id string) (string, error) {
			return app.ConfigLoader.Config.WebServer.Auth.TokenSecret, nil
		}),
		TokenDuration:  time.Minute * 5,
		CookieDuration: time.Hour * 24,
		Issuer:         "ccae",
		URL:            app.ConfigLoader.Config.WebServer.Url,
		AvatarStore:    avatar.NewLocalFS("/tmp"),
		Validator: token.ValidatorFunc(func(_ string, claims token.Claims) bool {
			return claims.User != nil && slices.Contains(app.ConfigLoader.Config.WebServer.Auth.Admins, claims.User.Name)
		}),
		XSRFIgnoreMethods: []string{"GET"},
		Logger:            lg.Func(func(format string, args ...any) { app.Logger.Printf(format, args...) }),
		AdminPasswd:       app.ConfigLoader.Config.WebServer.Auth.AdminPassword,
	})
	authService.AddProvider("yandex", app.ConfigLoader.Config.WebServer.Auth.OAuthClient, app.ConfigLoader.Config.WebServer.Auth.OAuthSecret)

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

	authRoutes, avaRoutes := authService.Handlers()
	am := authService.Middleware()

	logm := loggingMiddleware(app.Logger)

	static := root.Group()
	static.HandleFunc("GET /static/", createStaticHandler(staticsFs))

	anon := root.Group().Use(logm)
	anon.Handle("/auth/{path...}", authRoutes)
	anon.Handle("/avatar/{path...}", avaRoutes)
	anon.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		components.Page("INDEX").Render(r.Context(), w)
	})
	wsHandler, err := wsServer.CreateHttpHandle()
	if err != nil {
		return nil, err
	}
	anon.HandleFunc("GET /ws/{$}", func(w http.ResponseWriter, r *http.Request) {
		clientSecret := r.Header.Get("X-Client-Secret")

		client, err := app.ClientsService.GetClientBySecret(clientSecret)
		if err != nil {
			log.Printf("[DEBUG] Client connect to WS error: %v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		log.Printf("[DEBUG] Client connect to WS: %v", client)

		if !client.Authorized {
			http.Error(w, "Client is not authorized", http.StatusUnauthorized)
			return
		}

		wsHandler(w, r, client.ID)
	})

	common := root.Group().Use(logm).Use(am.Auth)

	handleFuncWithError(common, "GET /clients/{$}", func(w http.ResponseWriter, r *http.Request) error {
		clients, err := app.Daos.Clients.GetClients()
		if err != nil {
			return err
		}
		code := fmt.Sprintf("wget %s/lua/v3/client/ startup", app.ConfigLoader.Config.WebServer.Url)
		return components.GenClientsPage(clients, code).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /clients/{clientID}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		clientID := r.PathValue("clientID")
		client, err := app.Daos.Clients.GetClientByID(clientID)
		if err != nil {
			return err
		}
		if client == nil {
			return fmt.Errorf("client not found")
		}
		return components.GenClientPage(client).Render(r.Context(), w)
	})

	handleFuncWithError(common, "POST /clients/{clientID}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		clientID := r.PathValue("clientID")
		err = r.ParseForm()
		if err != nil {
			return err
		}
		role := r.FormValue("role")
		label := r.FormValue("label")

		client, err := app.Daos.Clients.GetClientByID(clientID)
		if err != nil {
			return err
		}
		if client == nil {
			return fmt.Errorf("client not found")
		}

		if role != client.Role {
			client.Role = role
		}

		if label != client.Label {
			client.Label = label
		}

		err = app.Daos.Clients.UpdateClient(client)
		if err != nil {
			return err
		}
		return nil
	})

	handleFuncWithError(common, "POST /clients/{clientID}/authorize/{$}", func(w http.ResponseWriter, r *http.Request) error {
		clientID := r.PathValue("clientID")
		err = r.ParseForm()
		if err != nil {
			return err
		}

		client, err := app.Daos.Clients.GetClientByID(clientID)
		if err != nil {
			return err
		}
		if client == nil {
			return fmt.Errorf("client not found")
		}

		return app.Daos.Clients.AuthorizeClient(client)
	})

	handleFuncWithError(common, "POST /clients/{clientID}/delete/{$}", func(w http.ResponseWriter, r *http.Request) error {
		clientID := r.PathValue("clientID")
		err = r.ParseForm()
		if err != nil {
			return err
		}

		client, err := app.Daos.Clients.GetClientByID(clientID)
		if err != nil {
			return err
		}
		if client == nil {
			return fmt.Errorf("client not found")
		}

		return app.Daos.Clients.DeleteClient(client)
	})

	handleFuncWithError(common, "GET /wsclients/{$}", func(w http.ResponseWriter, r *http.Request) error {
		clients := app.ClientsManager.GetClients()
		return components.ClientsPage(clients).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /storageItems/{$}", func(w http.ResponseWriter, r *http.Request) error {
		filter := r.URL.Query().Get("filter")
		view := r.URL.Query().Get("view")

		items, err := app.Storage.GetItems(filter)
		if err != nil {
			return err
		}

		if view == "list" {
			return components.ItemsInventory(items, 9).Render(r.Context(), w)
		}

		return components.ItemsInventoryPage(filter, items).Render(r.Context(), w)
	})

	handleFuncWithError(common, "POST /storageItems/optimize/{$}", func(w http.ResponseWriter, r *http.Request) error {
		err := app.Storage.Optimize()
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", "/storageItems/")
		return nil
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

		itemCount, err := app.Storage.GetItemCount(uid)
		if err != nil {
			return err
		}

		ctx, err := app.Daos.Items.NewDeferedLoader().FromRecipes(item.Recipes).FromRecipes(item.ImportedRecipes).ToContext(r.Context())
		if err != nil {
			return err
		}

		createParams := url.Values{}
		createParams.Add("item_g1", item.Item.UID)
		createParams.Add("amount_g1", "1")
		createParams.Add("role_g1", "goal")
		createUrl := fmt.Sprintf("/craft-plans/new/?%s", createParams.Encode())

		return components.ItemPage(item, createUrl, itemCount).Render(ctx, w)
	})

	handleFuncWithError(common, "GET /lua/client/{role}", func(w http.ResponseWriter, r *http.Request) error {
		role := r.PathValue("role")

		id, err := app.Daos.Seqs.NextId("clientNo")
		if err != nil {
			return err
		}

		return tmpls.Render("client.lua", []string{"client.lua.tmpl"}, w, map[string]any{
			"role":  role,
			"wsUrl": app.ConfigLoader.Config.ClientServer.Url,
			"id":    id,
		})
	})

	anon.HandleFunc("GET /lua/v2/client/{role}/", func(w http.ResponseWriter, r *http.Request) {
		role := r.PathValue("role")

		tmpls.Render("client.lua", []string{"clientV3.lua.tmpl"}, w, map[string]any{
			"wsUrl":   app.ConfigLoader.Config.ClientServer.Url,
			"version": build.Time,
			"role":    role,
		})
	})

	anon.HandleFunc("GET /lua/v3/client/", func(w http.ResponseWriter, r *http.Request) {
		secret := r.URL.Query().Get("secret")
		if secret == "" {
			secret = uuid.NewString()
		}
		params := map[string]any{
			"wsUrl":   app.ConfigLoader.Config.ClientServer.Url,
			"version": build.Time,
			"secret":  secret,
		}

		script, err := app.Daos.ClientsScripts.GetClientScript("bootstrap")
		if err == nil && script != nil {
			params["version"] = strconv.Itoa(script.Version)
			bootstrapTemplate, err := tmpl.New("bootstrap").Parse(script.Content)
			if err == nil {
				err = bootstrapTemplate.Execute(w, params)
				if err != nil {
					w.Write([]byte("ERROR: " + err.Error()))
				}
				return
			}
		}

		tmpls.Render("client.lua", []string{"clientV4.lua.tmpl"}, w, params)
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

	handleFuncWithError(common, "GET /craft-plans/new/{$}", func(w http.ResponseWriter, r *http.Request) error {

		items, err := recipe.ParseItemsParams(r.URL.Query())
		if err != nil {
			return err
		}

		goals := make([]crafter.Stack, len(items))
		for i, item := range items {
			goals[i] = crafter.Stack{ItemID: item.ItemUID, Count: item.Amount}
		}

		plan, err := app.Planner.GetPlanForItem(goals)
		if err != nil {
			return err
		}

		ctx, err := app.Daos.Items.NewDeferedLoader().AddUids(plan.Items).ToContext(r.Context())
		if err != nil {
			return err
		}

		goalsParams := components.RecipeItemsToParams(plan.Goals)
		createUrl := fmt.Sprintf("/craft-plans/new/?%s", goalsParams.Encode())

		return components.CraftingPlanPage(plan, createUrl).Render(ctx, w)
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

		recipeTypes, err := app.Daos.RecipeTypes.GetRecipeTypes()
		if err != nil {
			return err
		}

		return components.Page(fmt.Sprintf("Recipe for %s", recipe.Name), components.CreateRecipeForm(recipe, recipeTypes)).Render(ctx, w)
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
			return fmt.Errorf("recipe with id '%d' not found", recipeId)
		}

		recipe := recipes[0]

		ctx, err := app.Daos.Items.NewDeferedLoader().FromRecipe(recipe).ToContext(r.Context())
		if err != nil {
			return err
		}

		recipeTypes, err := app.Daos.RecipeTypes.GetRecipeTypes()
		if err != nil {
			return err
		}

		return components.Page("Recipe!!!", components.EditRecipeForm(recipe, recipeTypes)).Render(ctx, w)
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

		recipeTypes, err := app.Daos.RecipeTypes.GetRecipeTypes()
		if err != nil {
			return err
		}
		return components.Page(fmt.Sprintf("Recipe for %s", recipe.Name), components.CreateRecipeForm(recipe, recipeTypes)).Render(ctx, w)
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

		w.Header().Add("HX-Location", fmt.Sprintf("/recipes/%d/", recipe.ID))
		return nil
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
			return fmt.Errorf("recipe with id '%d' not found", recipeId)
		}

		err = app.Daos.Recipes.DeleteRecipe(recipeId)
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", "/recipes")
		return nil
	})

	handleFuncWithError(common, "GET /item-popup/{$}", func(w http.ResponseWriter, r *http.Request) error {
		paramMap := make(map[string]string)
		for k := range r.URL.Query() {
			paramMap[k] = r.URL.Query().Get(k)
		}
		data, err := json.Marshal(paramMap)
		if err != nil {
			return err
		}

		return components.ItemPopup(string(data)).Render(r.Context(), w)
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
		mode := r.URL.Query().Get("mode")
		uid := r.PathValue("uid")
		ctx, err := app.Daos.Items.NewDeferedLoader().AddUid(uid).ToContext(r.Context())
		if err != nil {
			return err
		}

		if mode == "recipe-item" {

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
				ItemUID: uid,
				Amount:  1,
				Role:    role,
				Slot:    slotIdx,
			}

			return components.ItemInputs(uuid.NewString(), item).Render(ctx, w)
		}
		return nil
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
			for _, goal := range plan.Goals {
				itemLoader.AddUid(goal.ItemUID)
			}
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

		loader := app.Daos.Items.NewDeferedLoader()
		for _, goal := range plan.Goals {
			loader.AddUid(goal.ItemUID)
		}
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

		var toTransfer []*crafter.Stack
		for _, goal := range plan.Goals {
			toTransfer = append(toTransfer, &crafter.Stack{ItemID: goal.ItemUID, Count: goal.Amount})
		}

		err = app.PlayerManager.SendItems(toTransfer)
		if err != nil {
			if r.URL.Query().Get("force") != "true" {
				return err
			}
		}

		err = app.Daos.Plans.RemovePlan(plan.ID)
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", "/craft-plans/")
		return nil
	})

	handleFuncWithError(common, "POST /craft-plans/new/{$}", func(w http.ResponseWriter, r *http.Request) error {

		items, err := recipe.ParseItemsParams(r.URL.Query())
		if err != nil {
			return err
		}

		goals := make([]crafter.Stack, len(items))
		for i, item := range items {
			goals[i] = crafter.Stack{ItemID: item.ItemUID, Count: item.Amount}
		}

		plan, err := app.Crafter.SchedulePlanForItem(goals)
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", fmt.Sprintf("/craft-plans/%d", plan.ID))
		return nil
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

		hasSelectors := false
		for slot := range slotIngredients {
			slotIngredients[slot] = cmn.Unique(slotIngredients[slot], func(i dao.RecipeItem) string {
				return i.ItemUID
			})
			if len(slotIngredients[slot]) > 1 {
				hasSelectors = true
			}
		}

		if !hasSelectors {
			items := make(map[string]*dao.Item)
			for _, ing := range recipe.Ingredients {
				items[ing.ItemUID] = nil
			}
			err = app.Daos.Items.FindItemsIndexed(items)
			if err != nil {
				return err
			}

			filteredRecipeItems := make([]dao.RecipeItem, 0, len(recipe.Ingredients))
			for _, item := range recipe.Ingredients {
				if items[item.ItemUID] == nil {
					continue
				}
				filteredRecipeItems = append(filteredRecipeItems, item)
			}

			recipe.Ingredients = filteredRecipeItems

			url := components.RecipeToURL(recipe)
			// w.Header().Add("HX-Location", url)
			http.Redirect(w, r, url, http.StatusSeeOther)
			return nil
		}

		return components.Page(fmt.Sprintf("Imported Recipe for %s", recipe.Name), components.ImportRecipeSlotsSelector(recipeId, slotIngredients)).Render(ctx, w)
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

		items := make(map[string]*dao.Item)
		for _, ing := range recipe.Ingredients {
			items[ing.ItemUID] = nil
		}

		err = app.Daos.Items.FindItemsIndexed(items)
		if err != nil {
			return err
		}

		filteredRecipeItems := make([]dao.RecipeItem, 0, len(recipe.Ingredients))
		for _, item := range recipe.Ingredients {
			if uid, ok := uidBySlot[*item.Slot]; ok && uid != item.ItemUID {
				continue
			}
			if items[item.ItemUID] == nil {
				continue
			}
			filteredRecipeItems = append(filteredRecipeItems, item)
		}

		recipe.Ingredients = filteredRecipeItems

		url := components.RecipeToURL(recipe)

		w.Header().Add("HX-Location", url)
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
		var stacks []*crafter.Stack
		for _, ing := range recipe.Ingredients {
			stacks = append(stacks, &crafter.Stack{
				ItemID: ing.ItemUID,
				Count:  ing.Amount,
			})
		}
		for _, ing := range recipe.Catalysts {
			stacks = append(stacks, &crafter.Stack{
				ItemID: ing.ItemUID,
				Count:  ing.Amount,
			})
		}
		err = app.PlayerManager.SendItems(stacks)
		if err != nil {
			return err
		}

		err = app.Daos.Crafts.CommitCraft(craft, recipe, 1)
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

	handleFuncWithError(common, "GET /recipe-types/{$}", func(w http.ResponseWriter, r *http.Request) error {

		types, err := app.Daos.RecipeTypes.GetRecipeTypes()
		if err != nil {
			return err
		}

		return components.RecipeTypesPage(types).Render(r.Context(), w)
	})

	handleFuncWithError(common, "POST /recipe-types/{$}", func(w http.ResponseWriter, r *http.Request) error {

		err = r.ParseForm()
		if err != nil {
			return err
		}

		recipeType := dao.RecipeType{
			Name:     r.FormValue("name"),
			WorkerID: r.FormValue("worker_id"),
		}

		err = app.Daos.RecipeTypes.InsertRecipeType(recipeType)
		if err != nil {
			return err
		}

		http.Redirect(w, r, "/recipe-types/", http.StatusSeeOther)
		return nil
	})

	handleFuncWithError(common, "DELETE /recipe-types/{name}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		name := r.PathValue("name")

		err = app.Daos.RecipeTypes.DeleteRecipeType(name)
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", "/recipe-types")
		return nil
	})

	handleFuncWithError(common, "POST /items/{itemUid}/sendToPlayer/{$}", func(w http.ResponseWriter, r *http.Request) error {
		uid := r.PathValue("itemUid")

		err := r.ParseForm()
		if err != nil {
			return err
		}

		amountStr := r.FormValue("amount")
		amount, err := strconv.Atoi(amountStr)
		if err != nil {
			return err
		}

		return app.PlayerManager.SendItems([]*crafter.Stack{{ItemID: uid, Count: amount}})
	})

	handleFuncWithError(common, "POST /items/{itemUid}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		uid := r.PathValue("itemUid")

		err := r.ParseForm()
		if err != nil {
			return err
		}

		newUid := r.FormValue("newUid")
		displayName := r.FormValue("displayName")

		item, err := app.ItemManager.UpdateItem(uid, &item.ItemParams{
			NewUID:      newUid,
			DisplayName: displayName,
		})

		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", fmt.Sprintf("/items/%s", item.UID))
		return nil
	})

	handleFuncWithError(common, "GET /workers/{$}", func(w http.ResponseWriter, r *http.Request) error {
		workers, err := app.WorkerManager.GetWorkers()
		if err != nil {
			return err
		}
		return components.WorkersPage(workers).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /workers/{key}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		key := r.PathValue("key")

		worker, err := app.WorkerManager.GetWorker(key)
		if err != nil {
			return err
		}

		params := app.WorkerManager.WorkerToParams(worker)

		itemLoader := app.Daos.Items.NewDeferedLoader()
		app.WorkerManager.AddWorkerItemUids(params, itemLoader)
		ctx, err := itemLoader.ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.EditWorkerPage(params).Render(ctx, w)
	})

	handleFuncWithError(common, "GET /workers-new/{$}", func(w http.ResponseWriter, r *http.Request) error {
		params := app.WorkerManager.ParseWorkerParams(url.Values{})

		itemLoader := app.Daos.Items.NewDeferedLoader()
		app.WorkerManager.AddWorkerItemUids(params, itemLoader)
		ctx, err := itemLoader.ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.NewWorkerPage(params).Render(ctx, w)
	})

	handleFuncWithError(common, "GET /workers-new/config/{$}", func(w http.ResponseWriter, r *http.Request) error {
		params := app.WorkerManager.NewWorkerParamsForType(r.URL.Query().Get("type"))
		return components.NewWorkerConfigForm(params).Render(r.Context(), w)
	})

	handleFuncWithError(common, "POST /workers-new/{$}", func(w http.ResponseWriter, r *http.Request) error {
		err = r.ParseForm()
		if err != nil {
			return err
		}
		params := app.WorkerManager.ParseWorkerParams(r.PostForm)
		worker, err := app.WorkerManager.ParseWorker(params)
		if err != nil {
			itemLoader := app.Daos.Items.NewDeferedLoader()
			app.WorkerManager.AddWorkerItemUids(params, itemLoader)
			ctx, err := itemLoader.ToContext(r.Context())
			if err != nil {
				return err
			}
			return components.NewWorkerPage(params).Render(ctx, w)
		}
		err = app.WorkerManager.CreateWorker(worker)
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", fmt.Sprintf("/workers/%s", worker.Key))
		return nil
	})

	handleFuncWithError(common, "POST /workers/{key}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		key := r.PathValue("key")

		err = r.ParseForm()
		if err != nil {
			return err
		}
		params := app.WorkerManager.ParseWorkerParams(r.PostForm)
		worker, err := app.WorkerManager.ParseWorker(params)
		if err != nil {
			errorMessage := err.Error()
			itemLoader := app.Daos.Items.NewDeferedLoader()
			app.WorkerManager.AddWorkerItemUids(params, itemLoader)
			ctx, err := itemLoader.ToContext(r.Context())
			if err != nil {
				return err
			}
			return components.EditWorkerPageContent(params, errorMessage).Render(ctx, w)
		}
		err = app.WorkerManager.UpdateWorker(key, worker)
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", fmt.Sprintf("/workers/%s", key))
		return nil
	})

	handleFuncWithError(common, "DELETE /workers/{key}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		key := r.PathValue("key")
		err := app.WorkerManager.DeleteWorker(key)
		if err != nil {
			return err
		}
		w.Header().Add("HX-Location", "/workers")
		return nil
	})

	handleFuncWithError(common, "GET /remotes/{$}", func(w http.ResponseWriter, r *http.Request) error {
		peripherals, err := app.ModemManager.GetPeripherals()
		if err != nil {
			return err
		}
		return components.PeripheralsPage(peripherals).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /clients-scripts/{$}", func(w http.ResponseWriter, r *http.Request) error {
		scripts, err := app.ScriptsManager.GetScripts()
		if err != nil {
			return err
		}
		return components.ClientsScriptsPage(scripts).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /clients-scripts/{role}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		role := r.PathValue("role")

		script, err := app.ScriptsManager.GetScript(role)
		if err != nil {
			return err
		}
		return components.ClientScriptPage(script).Render(r.Context(), w)
	})

	handleFuncWithError(common, "GET /clients-scripts/{role}/version/{$}", func(w http.ResponseWriter, r *http.Request) error {
		role := r.PathValue("role")

		script, err := app.ScriptsManager.GetScript(role)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(w, "%d", script.Version)
		return err
	})

	anon.HandleFunc("GET /clients-scripts/{role}/content/{$}", func(w http.ResponseWriter, r *http.Request) {
		role := r.PathValue("role")

		script, err := app.ScriptsManager.GetScript(role)
		var content string
		if err == nil {
			content = script.Content
		}

		fmt.Fprintf(w, content)
	})

	handleFuncWithError(common, "POST /clients-scripts/{$}", func(w http.ResponseWriter, r *http.Request) error {
		err = r.ParseForm()
		if err != nil {
			return err
		}
		role := r.PostForm.Get("role")
		err = app.ScriptsManager.CreateScript(&dao.ClientsScript{Role: role})
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", fmt.Sprintf("/clients-scripts/%s/", role))
		return nil
	})

	handleFuncWithError(common, "POST /clients-scripts/{role}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		role := r.PathValue("role")

		err = r.ParseForm()
		if err != nil {
			return err
		}
		newRole := r.PostForm.Get("role")
		content := r.PostForm.Get("content")

		err = app.ScriptsManager.UpdateScript(role, newRole, content)
		if err != nil {
			return err
		}

		w.Header().Add("HX-Replace-Url", fmt.Sprintf("/clients-scripts/%s/", newRole))
		return nil
	})

	handleFuncWithError(common, "DELETE /clients-scripts/{role}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		role := r.PathValue("role")
		err := app.ScriptsManager.DeleteScript(role)
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", "/clients-scripts/")
		return nil
	})

	handleFuncWithError(common, "GET /item-reserves/{$}", func(w http.ResponseWriter, r *http.Request) error {
		reserves, err := app.Daos.ItemReserves.GetActualReserves()
		if err != nil {
			return err
		}

		itemLoader := app.Daos.Items.NewDeferedLoader()
		for _, item := range reserves {
			itemLoader.AddUid(item.ItemUID)
		}

		ctx, err := itemLoader.ToContext(r.Context())
		if err != nil {
			return err
		}

		return components.ItemReservesPage(reserves).Render(ctx, w)
	})

	handleFuncWithError(common, "POST /item-reserves/clear/{$}", func(w http.ResponseWriter, r *http.Request) error {
		err := app.Daos.ItemReserves.ClearActualReserves()
		if err != nil {
			return err
		}

		w.Header().Add("HX-Location", "/item-reserves/")
		return nil
	})

	handleFuncWithError(common, "GET /item-suggest/{$}", handlers.ItemSuggest(app.Daos.Items))

	handleFuncWithError(common, "GET /api/v1/clients-scripts/{$}", func(w http.ResponseWriter, r *http.Request) error {
		scripts, err := app.ScriptsManager.GetScripts()
		if err != nil {
			return err
		}
		var views []*clientscripts.ScriptJsonView
		for _, script := range scripts {
			views = append(views, clientscripts.NewScriptJsonView(script))
		}
		encoder := json.NewEncoder(w)
		return encoder.Encode(views)
	})

	handleFuncWithError(common, "GET /api/v1/clients-scripts/{role}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		role := r.PathValue("role")

		script, err := app.ScriptsManager.GetScript(role)
		if err != nil {
			return err
		}
		if script == nil {
			return errors.New("script not found")
		}
		encoder := json.NewEncoder(w)
		return encoder.Encode(clientscripts.NewScriptJsonView(script))
	})

	handleFuncWithError(common, "POST /api/v1/clients-scripts/{$}", func(w http.ResponseWriter, r *http.Request) error {
		defer r.Body.Close()
		decoder := json.NewDecoder(r.Body)

		var scriptView clientscripts.ScriptJsonView
		err := decoder.Decode(&scriptView)
		if err != nil {
			return err
		}

		err = app.ScriptsManager.CreateScript(scriptView.ToScript())
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusCreated)
		return nil
	})

	handleFuncWithError(common, "POST /api/v1/clients-scripts/{role}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		role := r.PathValue("role")
		defer r.Body.Close()
		decoder := json.NewDecoder(r.Body)

		var scriptView clientscripts.ScriptJsonView
		err := decoder.Decode(&scriptView)
		if err != nil {
			return err
		}

		err = app.ScriptsManager.UpdateScript(role, scriptView.Role, scriptView.Content)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusCreated)
		return nil
	})

	handleFuncWithError(common, "DELETE /api/v1/clients-scripts/{role}/{$}", func(w http.ResponseWriter, r *http.Request) error {
		role := r.PathValue("role")
		err := app.ScriptsManager.DeleteScript(role)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		return nil
	})

	handleFuncWithError(common, "/", func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNotFound)
		return components.Page("Not found").Render(r.Context(), w)
	})

	return mux, nil
}
