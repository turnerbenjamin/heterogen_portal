package handlers

import (
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

type ErrorHandler struct {
	templateStore *templates.Store
}

type templateError struct {
	ToastError string
	PageErrors []string
}

func NewErrorHandler(templateStore *templates.Store) *ErrorHandler {
	return &ErrorHandler{
		templateStore: templateStore,
	}
}

func (h *ErrorHandler) Write(w http.ResponseWriter, appErr *etc.AppError) error {
	w.WriteHeader(appErr.Code)

	te := templateError{appErr.ToastError, appErr.PageErrors}

	return h.templateStore.Execute(
		templates.TMPL_COMPONENT_ERRORS,
		w,
		templates.TemplateArgs{Data: te},
	)
}
