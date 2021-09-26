package templatetree

import (
	html "html/template"
	"io"
	text "text/template"
)

// Template defines common methods implemented by both *text/template.Template
// and *html/template.Template.
type Template interface {
	Name() string
	Execute(w io.Writer, data interface{}) error
	ExecuteTemplate(w io.Writer, name string, data interface{}) error
}

// template is an adapter interface for stdlib template types
type template interface {
	Unwrap() Template
	Name() string
	Parse(text string) error
}

type textTemplate struct {
	*text.Template
}

func (t textTemplate) Unwrap() Template { return t.Template }

func (t textTemplate) Parse(text string) error {
	_, err := t.Template.Parse(text)
	return err
}

type htmlTemplate struct {
	*html.Template
}

func (t htmlTemplate) Unwrap() Template { return t.Template }

func (t htmlTemplate) Parse(text string) error {
	_, err := t.Template.Parse(text)
	return err
}
