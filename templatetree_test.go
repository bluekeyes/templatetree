package templatetree

import (
	"strings"
	"testing"
	text "text/template"
)

func TestText(t *testing.T) {
	factory := TextFactory(func(name string) *text.Template {
		return text.New("root").Funcs(text.FuncMap{
			"Value": func() string { return "Test Value" },
		})
	})

	tmpl, err := Parse("testdata/basic", "*.tmpl", factory)
	if err != nil {
		t.Fatalf("error loading templates: %v", err)
	}

	t.Run("baseTemplate", func(t *testing.T) {
		out := assertRender(t, tmpl, "base.tmpl")
		assertOutput(t, out, []string{"Base", "Child"})
	})

	t.Run("basicInheritance", func(t *testing.T) {
		out := assertRender(t, tmpl, "a.tmpl")
		assertOutput(t, out, []string{"Base", "A"})

		out = assertRender(t, tmpl, "b.tmpl")
		assertOutput(t, out, []string{"Base", "B"})
	})

	t.Run("funcInheritance", func(t *testing.T) {
		out := assertRender(t, tmpl, "funcs.tmpl")
		assertOutput(t, out, []string{"Base", "Test Value"})
	})

	t.Run("nestedTemplate", func(t *testing.T) {
		out := assertRender(t, tmpl, "nested/c.tmpl")
		assertOutput(t, out, []string{"Base", "C", "Default"})
	})

	t.Run("multilevelInheritance", func(t *testing.T) {
		out := assertRender(t, tmpl, "nested/d.tmpl")
		assertOutput(t, out, []string{"Base", "D", "Test Value"})
	})
}

func TestDetectCycles(t *testing.T) {
	_, err := Parse("testdata/cycles", "*.tmpl", TextFactory(nil))
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

func render(tree Tree, name string) (string, error) {
	var b strings.Builder
	if err := tree.ExecuteTemplate(&b, name, nil); err != nil {
		return "", err
	}
	return b.String(), nil
}

func assertRender(t *testing.T, tree Tree, name string) string {
	out, err := render(tree, name)
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
