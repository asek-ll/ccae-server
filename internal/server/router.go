package server

import (
	"embed"
	"io/fs"
	"net/http"

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

func CreateMux() (*http.ServeMux, error) {

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
		tmpls.Render("clients", []string{"index.html.tmpl", "clients.html.tmpl"}, w, nil)
	})

	return mux, nil
}
