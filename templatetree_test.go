package templatetree

import (
	html "html/template"
	"strings"
	"testing"
	text "text/template"
)

func TestText(t *testing.T) {
	factory := func(name string) Template[*text.Template] {
		return text.New(name).Funcs(text.FuncMap{
			"Value": func() string { return "Test Value" },
		})
	}

	tmpl, err := Parse("testdata/basic", "*.tmpl", factory)
	if err != nil {
		t.Fatalf("error loading templates: %v", err)
	}

	t.Run("baseTemplate", func(t *testing.T) {
		out := assertRender(t, tmpl, "base.tmpl", nil)
		assertOutput(t, out, []string{"Base", "Child"})
	})

	t.Run("basicInheritance", func(t *testing.T) {
		out := assertRender(t, tmpl, "a.tmpl", nil)
		assertOutput(t, out, []string{"Base", "A"})

		out = assertRender(t, tmpl, "b.tmpl", nil)
		assertOutput(t, out, []string{"Base", "B"})
	})

	t.Run("funcInheritance", func(t *testing.T) {
		out := assertRender(t, tmpl, "funcs.tmpl", nil)
		assertOutput(t, out, []string{"Base", "Test Value"})
	})

	t.Run("nestedTemplate", func(t *testing.T) {
		out := assertRender(t, tmpl, "nested/c.tmpl", nil)
		assertOutput(t, out, []string{"Base", "C", "Default"})
	})

	t.Run("multilevelInheritance", func(t *testing.T) {
		out := assertRender(t, tmpl, "nested/d.tmpl", nil)
		assertOutput(t, out, []string{"Base", "D", "Test Value"})
	})
}

func TestHTML(t *testing.T) {
	factory := func(name string) Template[*html.Template] {
		return html.New(name).Funcs(html.FuncMap{
			"Value": func() string { return "Test Value" },
		})
	}

	tmpl, err := Parse("testdata/basic", "*.tmpl", factory)
	if err != nil {
		t.Fatalf("error loading templates: %v", err)
	}

	htmlTmpl, err := Parse("testdata/html", "*.html.tmpl", factory)
	if err != nil {
		t.Fatalf("error loading templates: %v", err)
	}

	t.Run("basicInheritance", func(t *testing.T) {
		out := assertRender(t, tmpl, "a.tmpl", nil)
		assertOutput(t, out, []string{"Base", "A"})

		out = assertRender(t, tmpl, "b.tmpl", nil)
		assertOutput(t, out, []string{"Base", "B"})
	})

	t.Run("multilevelInheritance", func(t *testing.T) {
		out := assertRender(t, tmpl, "nested/d.tmpl", nil)
		assertOutput(t, out, []string{"Base", "D", "Test Value"})
	})

	t.Run("staticHTML", func(t *testing.T) {
		out := assertRender(t, htmlTmpl, "index.html.tmpl", nil)
		assertOutput(t, out, []string{
			"<!doctype html>",
			"<html>",
			"    <head>",
			"        <title>Index Page</title>",
			"    </head>",
			"    <body>",
			"        <p>This is a test page!</p>",
			"    </body>",
			"</html>",
		})
	})

	t.Run("dynamicHTML", func(t *testing.T) {
		out := assertRender(t, htmlTmpl, "data.html.tmpl", map[string]string{"Word": "Testing"})
		assertOutput(t, out, []string{
			"<!doctype html>",
			"<html>",
			"    <head>",
			"        <title>Data Page | Testing</title>",
			"    </head>",
			"    <body>",
			"        <p>This contains the word <i>Testing</i></p>",
			"    </body>",
			"</html>",
		})
	})
}

func TestDetectCycles(t *testing.T) {
	_, err := Parse("testdata/cycles", "*.tmpl", DefaultTextFactory)
	if err == nil {
		t.Fatal("template cycle was not detected")
	}

	msg := err.Error()
	if strings.Contains(msg, "inheritance cycle in templates") {
		for _, name := range []string{"a.tmpl", "b.tmpl", "c.tmpl"} {
			if !strings.Contains(msg, name) {
				t.Errorf("%q should be part of the cycle but is not:\n%s", name, msg)
			}
		}
	} else {
		t.Errorf("unexpected error encountered while loading templates: %v", err)
	}
}

func render[T StdTemplate](tree Tree[T], name string, data any) (string, error) {
	var b strings.Builder
	if err := tree.ExecuteTemplate(&b, name, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

func assertRender[T StdTemplate](t *testing.T, tree Tree[T], name string, data any) string {
	out, err := render(tree, name, data)
	if err != nil {
		t.Fatalf("error rendering %q: %v", name, err)
		return ""
	}
	return out
}

func assertOutput(t *testing.T, actual string, expected []string) {
	actual = strings.TrimSpace(actual)
	if actual != strings.Join(expected, "\n") {
		var msg strings.Builder
		msg.WriteString("incorrect template output\n")
		msg.WriteString("expected:\n")
		for _, line := range expected {
			msg.WriteString("  ")
			msg.WriteString(line)
			msg.WriteString("\n")
		}
		msg.WriteString("actual:\n")
		for _, line := range strings.Split(actual, "\n") {
			msg.WriteString("  ")
			msg.WriteString(line)
			msg.WriteString("\n")
		}
		t.Error(msg.String())
	}
}
