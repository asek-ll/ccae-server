package server

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/asek-ll/aecc-server/internal/app"
	"github.com/asek-ll/aecc-server/pkg/template"
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
		err := app.Storage.GetItems()
		if err != nil {
			tmpls.RenderError(err, w)
			return
		}

		tmpls.Render("clients", []string{"index.html.tmpl", "clients.html.tmpl"}, w, nil)
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

	return mux, nil
}