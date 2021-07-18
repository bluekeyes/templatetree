package templatetree_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/bluekeyes/templatetree"
)

func Example() {
	files := map[string]string{
		"base.tmpl": strings.TrimSpace(`
Header
{{block "body" .}}Body{{end}}
Footer
`),
		"a.tmpl": strings.TrimSpace(`
{{/* templatetree:extends base.tmpl */}}
{{define "body"}}Body A{{end}}
`),
		"b.tmpl": strings.TrimSpace(`
{{/* templatetree:extends base.tmpl */}}
{{define "body"}}Body B{{end}}
`),
	}

	t, err := templatetree.ParseFiles(files, templatetree.TextFactory(nil))
	if err != nil {
		panic(err)
	}

	fmt.Println("--- a.tmpl ---")
	t.ExecuteTemplate(os.Stdout, "a.tmpl", nil)
	fmt.Println()

	fmt.Println("--- b.tmpl ---")
	t.ExecuteTemplate(os.Stdout, "b.tmpl", nil)
	fmt.Println()

	// Output:
	// --- a.tmpl ---
	// Header
	// Body A
	// Footer
	// --- b.tmpl ---
	// Header
	// Body B
	// Footer
}
