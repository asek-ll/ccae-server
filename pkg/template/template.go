package template

import (
	"io"
	"io/fs"
	"text/template"
)

type Templates struct {
	resources fs.FS
	cache     map[string]*template.Template
}

func NewTemplates(resources fs.FS) Templates {
	return Templates{
		cache:     make(map[string]*template.Template),
		resources: resources,
	}
}

func (tmpls Templates) Get(name string, patterns []string) (*template.Template, error) {
	if t, e := tmpls.cache[name]; e {
		return t, nil
	}

	tmpl, err := template.ParseFS(tmpls.resources, patterns...)

	if err != nil {
		return nil, err
	}

	tmpls.cache[name] = tmpl
	return tmpl, nil
}

func (tmpls Templates) Render(name string, includes []string, out io.Writer, data any) error {
	tmpl, err := tmpls.Get(name, includes)
	if err != nil {
		return err
	}
	return tmpl.Execute(out, data)
}
