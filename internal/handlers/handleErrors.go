package handlers

import (
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/logging"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

type errorHandler struct {
	templateStore *templates.Store
}

type ErrorHandler interface {
	Write(http.ResponseWriter, logging.Logger, etc.AppError)
}

type templateError struct {
	ToastError string
	PageErrors []string
}

func NewErrorHandler(templateStore *templates.Store) ErrorHandler {
	return &errorHandler{
		templateStore: templateStore,
	}
}

func (h *errorHandler) Write(w http.ResponseWriter, logger logging.Logger, appErr etc.AppError) {
	w.WriteHeader(appErr.Code())

	te := templateError{appErr.ToastError(), appErr.PageErrors()}

	err := h.templateStore.Execute(
		templates.TMPL_COMPONENT_ERRORS,
		w,
		templates.TemplateArgs{Data: te},
	)
	if err != nil {
		logger.AddKV("error", err.Error())
	}
}
