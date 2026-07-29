package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/query"
)

type QueryService interface {
	Execute(ctx context.Context, resource string, queryString string) ([]query.TableModel, *etc.AppError)
}

type QueryHandler struct {
	service QueryService
}

func NewQueryHandler(service QueryService) *QueryHandler {
	return &QueryHandler{
		service: service,
	}
}

// GetSignOutHandler unsets the app jwt cookie and redirects the user to sign
// out from the auth provider
func (h QueryHandler) ProcessQuery(
	w http.ResponseWriter,
	r *http.Request,
	c *PipelineContext[NoState],
) *etc.AppError {
	resource := r.PathValue("resource")

	decodedQuery, err := url.QueryUnescape(r.URL.RawQuery)
	if err != nil {
		return etc.NewServerError(err)
	}

	res, appErr := h.service.Execute(r.Context(), resource, decodedQuery)
	if appErr != nil {
		return appErr.WithJsonType()
	}
	jsonBody, err := json.Marshal(res)
	if err != nil {
		return etc.NewServerError(err).WithJsonType()
	}

	w.Header().Add("Content-Type", "application-json")
	w.Write(jsonBody)
	return nil
}
