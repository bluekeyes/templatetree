package templatetree_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/bluekeyes/templatetree"
)

func Example() {
	files := []*templatetree.File{
		&templatetree.File{
			Name: "base.tmpl",
			Content: strings.TrimSpace(`
Header
{{block "body" .}}Body{{end}}
Footer
`),
		},
		&templatetree.File{
			Name: "a.tmpl",
			Content: strings.TrimSpace(`
{{/* templatetree:extends base.tmpl */}}
{{define "body"}}Body A{{end}}
`),
		},
		&templatetree.File{
			Name: "b.tmpl",
			Content: strings.TrimSpace(`
{{/* templatetree:extends base.tmpl */}}
{{define "body"}}Body B{{end}}
`),
		},
	}

	t, err := templatetree.ParseText(files, nil)
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
