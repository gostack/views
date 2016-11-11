package view

import (
	"html/template"
	"io"
	"path/filepath"
	"text/template/parse"
)

// Manager handles the loading, parsing and rendering of templates.
// This is the entry point for this library.
type Manager struct {
	// BasePath is the path in which the templates will be loaded relatively from
	BasePath string

	// Loader is the function that translates a path into the actual template content.
	// For example, it can fetch the template from disk, or any other place.
	Loader func(path string) ([]byte, error)

	// Funcs is the map of functions that will be available to the template.
	Funcs template.FuncMap

	// cache holds the cache of parsed templates
	cache map[string]*parse.Tree
}

// EnableCaching will enable caching for this manager.
// This should be used in production environment to avoid reloading and reparsing the same template multiple times.
func (m *Manager) EnableCaching() {
	m.cache = make(map[string]*parse.Tree)
}

// Render takes a relative path to the Manager's basePath, load it as a template then render and write it to w
func (m Manager) Render(w io.Writer, path string, data interface{}) error {
	t, err := m.loadTemplates(map[string]string{"@content": path})
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, "@content", data)
}

// RenderInLayout takes a relative path to a content template and a layout template and render the content within the layout.
// Inside the layout template, you can refer to the content template with {{ template "@content" }}.
func (m Manager) RenderInLayout(w io.Writer, contentPath, layoutPath string, data interface{}) error {
	t, err := m.loadTemplates(map[string]string{
		"@layout":  layoutPath,
		"@content": contentPath,
	})
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, "@layout", data)
}

// loadTemplates performs the loading of templates into a single template context, as well as take cares of the caching, if enabled.
func (m *Manager) loadTemplates(tpls map[string]string) (*template.Template, error) {
	rt := template.New("")
	rt.Funcs(m.Funcs)

	for name, path := range tpls {
		var (
			t   *parse.Tree
			ok  bool
			err error
		)

		if m.cache != nil {
			if t, ok = m.cache[path]; ok {
				_, err = rt.AddParseTree(name, t)
				if err != nil {
					return nil, err
				}
			}
		}

		if t == nil {
			c, err := m.Loader(filepath.Join(m.BasePath, path))
			tpl, err := rt.New(name).Parse(string(c))
			if err != nil {
				return nil, err
			}

			t = tpl.Tree
		}

		if m.cache != nil {
			m.cache[path] = t
		}
	}

	return rt, nil
}
