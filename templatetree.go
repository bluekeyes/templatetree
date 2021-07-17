package templatetree

import (
	"fmt"
	html "html/template"
	"io"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
	text "text/template"
)

const (
	// CommentTagExtends is the tag used in template comments to mark a
	// template's parent.
	CommentTagExtends = "templatetree:extends"

	// DefaultRootTemplateName is the name of the root template when none is
	// provided by the caller.
	DefaultRootTemplateName = "[templatetree:root]"
)

// File is an unparsed template definition.
type File struct {
	// Name is the template name. Child templates use this in the extends tag
	// and it identifies the template in ExecuteTemplate calls.
	Name string

	// Content is the unparsed template definition.
	Content string
}

// LoadText recursively loads all text templates in dir with names matching
// pattern, respecting inheritance. If the root template is nil, a new default
// template is used. Templates are named as normalized paths relative to dir.
func LoadText(dir, pattern string, root *text.Template) (TextTree, error) {
	return ParseTextFS(os.DirFS(dir), pattern, root)
}

// ParseTextFS recursively parses all text templates in fsys with names
// matching pattern, respecting inheritance. If the root template is nil, a new
// default template is used. Templates are named by their paths in fsys.
func ParseTextFS(fsys fs.FS, pattern string, root *text.Template) (TextTree, error) {
	files, err := loadFiles(fsys, pattern)
	if err != nil {
		return nil, err
	}
	return ParseText(files, root)
}

// ParseText parses files into a TextTree, respecting inheritance. If the root
// template is nil, a new default template is used.
func ParseText(files []*File, root *text.Template) (TextTree, error) {
	if root == nil {
		root = text.New(DefaultRootTemplateName)
	}

	tree := make(TextTree)
	return tree, parseAll(files, textTemplate{root}, func(name string, t template) {
		tree[name] = t.(textTemplate).Template
	})
}

// LoadHTML recursively loads all HTML templates in dir with names matching
// pattern, respecting inheritance. If the root template is nil, a new default
// template is used. Templates are named as normalized paths relative to dir.
func LoadHTML(dir, pattern string, root *html.Template) (HTMLTree, error) {
	return ParseHTMLFS(os.DirFS(dir), pattern, root)
}

// ParseHTMLFS recursively loads all HTML templates in fsys with names matching
// pattern, respecting inheritance. If the root template is nil, a new default
// template is used. Templates are named by their paths in fsys.
func ParseHTMLFS(fsys fs.FS, pattern string, root *html.Template) (HTMLTree, error) {
	files, err := loadFiles(fsys, pattern)
	if err != nil {
		return nil, err
	}
	return ParseHTML(files, root)
}

// ParseHTML parses files into a HTMLTree, respecting inheritance. If the root
// template is nil, a new default template is used.
func ParseHTML(files []*File, root *html.Template) (HTMLTree, error) {
	if root == nil {
		root = html.New(DefaultRootTemplateName)
	}

	tree := make(HTMLTree)
	return tree, parseAll(files, htmlTemplate{root}, func(name string, t template) {
		tree[name] = t.(htmlTemplate).Template
	})
}

// TextTree is a hierarchy of text templates, mapping name to template.
type TextTree map[string]*text.Template

// ExecuteTemplate renders the template with the given name. See the
// text/template package for more details.
func (tree TextTree) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	if tmpl, ok := tree[name]; ok {
		return tmpl.Execute(wr, data)
	}
	return fmt.Errorf("templatetree: no template %q", name)
}

// HTMLTree is a hierarchy of text templates, mapping name to template.
type HTMLTree map[string]*html.Template

// ExecuteTemplate renders the template with the given name. See the
// html/template package for more details.
func (tree HTMLTree) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	if tmpl, ok := tree[name]; ok {
		return tmpl.Execute(wr, data)
	}
	return fmt.Errorf("templatetree: no template %q", name)
}

// adapter for text/template and html/template
type template interface {
	Name() string
	Clone() (template, error)
	Parse(string) error
}

type textTemplate struct {
	*text.Template
}

func (t textTemplate) Clone() (template, error) {
	nt, err := t.Template.Clone()
	return textTemplate{nt}, err
}

func (t textTemplate) Parse(content string) error {
	_, err := t.Template.Parse(content)
	return err
}

type htmlTemplate struct {
	*html.Template
}

func (t htmlTemplate) Clone() (template, error) {
	nt, err := t.Template.Clone()
	return htmlTemplate{nt}, err
}

func (t htmlTemplate) Parse(content string) error {
	_, err := t.Template.Parse(content)
	return err
}

type node struct {
	name     string
	content  string
	template template
	parent   *node
}

func loadFiles(templates fs.FS, pattern string) ([]*File, error) {
	var files []*File
	walkFn := func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		match, err := path.Match(pattern, path.Base(name))
		if err != nil {
			return err
		}
		if match {
			b, err := fs.ReadFile(templates, name)
			if err != nil {
				return err
			}
			files = append(files, &File{Name: name, Content: string(b)})
		}
		return nil
	}

	if err := fs.WalkDir(templates, ".", walkFn); err != nil {
		return nil, err
	}
	return files, nil
}

func parseAll(files []*File, root template, register func(string, template)) error {
	nodes := make(map[string]*node)
	for _, f := range files {
		nodes[f.Name] = &node{name: f.Name, content: f.Content}
	}

	// create links between parents and children
	for _, n := range nodes {
		parent := parseHeader(n.content)
		if parent != "" {
			if p, ok := nodes[parent]; ok {
				n.parent = p
			} else {
				return fmt.Errorf("templatetree: template %q extends unknown template %s", n.name, parent)
			}
		}
	}

	// parse templates from the root nodes in/down
	for {
		n := findNext(nodes)
		if n == nil {
			break
		}
		delete(nodes, n.name)

		base := root
		if n.parent != nil {
			base = n.parent.template
		}

		t, err := base.Clone()
		if err != nil {
			return err
		}
		if err := t.Parse(n.content); err != nil {
			return formatParseError(n, t, err)
		}

		n.template = t
		register(n.name, t)
	}

	// check for cycles
	if len(nodes) > 0 {
		var names []string
		for _, n := range nodes {
			names = append(names, strconv.Quote(n.name))
		}
		return fmt.Errorf("templatetree: inheritance cycle in templates [%s]", strings.Join(names, ", "))
	}

	return nil
}

func findNext(nodes map[string]*node) *node {
	for _, n := range nodes {
		if n.parent == nil || n.parent.template != nil {
			return n
		}
	}
	return nil
}

func parseHeader(content string) (parent string) {
	prefix := "{{/* " + CommentTagExtends + " "
	if !strings.HasPrefix(content, prefix) {
		return
	}

	idx := strings.Index(content[len(prefix):], " */}}")
	if idx < 0 {
		return
	}

	parent = content[len(prefix) : len(prefix)+idx]
	return
}

// The current template API doesn't provide a way to change names, so try to
// edit the error message so the correct name appears for users. This is dirty,
// but is strictly for usability, not correctness.
func formatParseError(n *node, t template, err error) error {
	msg := err.Error()
	old := "template: " + t.Name()
	if strings.HasPrefix(msg, old) {
		return fmt.Errorf("template: %s%s", n.name, strings.TrimPrefix(msg, old))
	}
	return err
}
