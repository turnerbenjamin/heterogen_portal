// Package handlers contains HTTP handlers for the application
package handlers

import (
	"io"

	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

// TemplateStore is an interface wrapping a simple Execute method used to
// write a given template to a writer
type TemplateStore interface {

	// Execute writes the specified template to the writer passing any data
	// provided. Returns any errors encountered when writing
	Execute(
		id templates.TemplateIdentifier,
		w io.Writer,
		data templates.TemplateArgs,
	) error
}
