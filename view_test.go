package view

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

// loadTemplateCount gets incremented every time loadTemplate() is called
var loadTemplateCount int

// tplData used to test the templates
var tplData = struct {
	Today time.Time
}{
	Today: time.Date(1985, 10, 31, 10, 34, 0, 0, time.UTC),
}

// tplFuncs defines the functions available to the test templates
var tplFuncs = template.FuncMap{
	"formatTime": func(t time.Time) string {
		return t.Format(time.RFC1123)
	},
}

func TestRender(t *testing.T) {
	b := new(bytes.Buffer)
	m := NewManager("_test_assets/templates", loadTemplate, false)
	m.SetFuncs(tplFuncs)

	if err := m.Render(b, "pages/hello.html", tplData); err != nil {
		t.Fatal(err)
	}

	s := b.String()
	e := `<p>Hello!</p>`
	if s != e {
		t.Errorf("unexpected render result:\n%s\n\nexpected:\n%s", s, e)
	}
}

func TestRenderCache(t *testing.T) {
	b := new(bytes.Buffer)
	m := NewManager("_test_assets/templates", loadTemplate, true)
	m.SetFuncs(tplFuncs)

	c := loadTemplateCount + 1
	if err := m.Render(b, "pages/hello.html", tplData); err != nil {
		t.Fatal(err)
	}
	if loadTemplateCount != c {
		t.Error("template not loaded on first call")
	}

	if err := m.Render(b, "pages/hello.html", tplData); err != nil {
		t.Fatal(err)
	}
	if loadTemplateCount != c {
		t.Error("template not cached on subsequent calls")
	}
}

func TestRenderInLayout(t *testing.T) {
	b := new(bytes.Buffer)
	m := NewManager("_test_assets/templates", loadTemplate, false)
	m.SetFuncs(tplFuncs)

	if err := m.RenderInLayout(b, "pages/hello.html", "layouts/application.html", tplData); err != nil {
		t.Fatal(err)
	}

	s := strings.TrimSpace(b.String())
	e := strings.TrimSpace(`
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Views Example</title>
  </head>
  <body>
    <div>Thu, 31 Oct 1985 10:34:00 UTC</div>
    <div class="container">
      <p>Hello!</p>
    </div>
  </body>
</html>
`)
	if s != e {
		t.Errorf("unexpected render result:\n%s\n\nexpected:\n\n%s", s, e)
	}
}

func loadTemplate(path string) ([]byte, error) {
	loadTemplateCount += 1

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}
