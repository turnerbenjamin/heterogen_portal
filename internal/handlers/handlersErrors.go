// Package handlers contains HTTP handlers for the application.
//
// This file contains a simple error handler for writing AppError responses
package handlers

import (
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
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
// the status code and passing error data to the error template
func (h *ErrorHandler) Write(
	w http.ResponseWriter,
	r *http.Request,
	appErr *AppError,
) error {
	// Default to handling errors with a the component error template returned
	// to a htmx app
	t := templates.TmplComponentErrors
	pageConfig := templates.PageConfig{
		ContentOnly: true,
	}

	// If the request is not from htmx, return a full page, out of app error
	// template
	if r.Header.Get(constants.HxRequestHeaderRequest) == "" {
		t = templates.TmplPageOutOfAppErr
		pageConfig.ContentOnly = false
	}

	w.WriteHeader(appErr.Code)
	return h.templateStore.Execute(
		t,
		w,
		templates.TemplateArgs{
			PageConfig: pageConfig,
			Data:       appErr,
		},
	)
}
