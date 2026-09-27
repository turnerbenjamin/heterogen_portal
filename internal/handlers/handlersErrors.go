// Package handlers contains HTTP handlers for the application.
//
// This file contains a simple error handler for writing AppError responses
package handlers

import (
	"encoding/json"
	"fmt"
	"mime"
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
	// access content type
	rawContentTypeValue := w.Header().Get("Content-Type")
	contentType, _, err := mime.ParseMediaType(rawContentTypeValue)
	if err != nil {
		return err
	}

	// validate content type is either json or html
	if contentType != ContentTypeJson.String() &&
		contentType != ContentTypeHtml.String() {
		return fmt.Errorf(constants.ErrMsgPatternUnsupportedContentType, contentType)
	}

	// If content type is json, marshal the error and return
	if contentType == ContentTypeJson.String() {
		w.WriteHeader(appErr.Code)
		d, err := json.Marshal(appErr)
		if err != nil {
			return err
		}
		_, err = w.Write(d)
		return err
	}

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

	if appErr.Code == 0 {
		w.WriteHeader(500)
	} else {
		w.WriteHeader(appErr.Code)
	}

	return h.templateStore.Execute(
		t,
		w,
		templates.TemplateArgs{
			PageConfig: pageConfig,
			Data:       appErr,
		},
	)
}
