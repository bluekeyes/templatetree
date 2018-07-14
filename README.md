# templatetree [![GoDoc](https://godoc.org/github.com/bluekeyes/templatetree?status.svg)](http://godoc.org/github.com/bluekeyes/templatetree)

templatetree is a standard library template loader that creates simple template
inheritance trees. Base templates use the `block` or `template` directives to
define sections that are overridden or provided by child templates.

Functions are provided to create both `text/template` and `html/template`
objects.

## Example

TODO

## Stability

While the API is simple, it hasn't seen heavy use yet and may change in the
future. I recommend vendoring this package at a specific commit if you are
concerned about API changes.
