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
	funcs      template.FuncMap
	basePath   string
	loaderFunc func(path string) ([]byte, error)
	cache      map[string]*parse.Tree
}

// NewManager creates a new Manager instance which will handle the templates located
// under the provided basePath.
func NewManager(basePath string, loaderFunc func(path string) ([]byte, error), performCaching bool) *Manager {
	m := Manager{basePath: basePath, loaderFunc: loaderFunc}

	if performCaching {
		m.cache = make(map[string]*parse.Tree)
	}

	return &m
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

// SetFuncs sets the functions that will be made available to the views managed by this Manager
func (m *Manager) SetFuncs(funcs template.FuncMap) {
	m.funcs = funcs
}

// loadTemplates performs the loading of templates into a single template context, as well as take cares of the caching, if enabled.
func (m *Manager) loadTemplates(tpls map[string]string) (*template.Template, error) {
	rt := template.New("")
	rt.Funcs(m.funcs)

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
			c, err := m.loaderFunc(filepath.Join(m.basePath, path))
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
