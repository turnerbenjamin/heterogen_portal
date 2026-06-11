// Package handlers contains HTTP handlers for the application.
//
// This file contains a simple error handler for writing AppError responses
package handlers

import (
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

// ErrorHandler uses a template store to write AppErrors to a response
type ErrorHandler struct {
	templateStore TemplateStore
}

// NewErrorHandler is an arguably pointless factory function for creating a new
// error handler
func NewErrorHandler(templateStore TemplateStore) *ErrorHandler {
	return &ErrorHandler{
		templateStore: templateStore,
	}
}

// Write is responsible for writing AppErrors to a response. It handles setting
// the status code and passing error data to the error component template
func (h *ErrorHandler) Write(w http.ResponseWriter, appErr *etc.AppError) error {
	w.WriteHeader(appErr.Code)

	pageConfig := templates.PageConfig{
		ContentOnly: true,
	}

	return h.templateStore.Execute(
		templates.TMPL_COMPONENT_ERRORS,
		w,
		templates.TemplateArgs{
			PageConfig: pageConfig,
			Data:       appErr,
		},
	)
}
