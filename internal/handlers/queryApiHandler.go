package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/turnerbenjamin/querystack/queryerror"
	"github.com/turnerbenjamin/querystack/querymodel"
)

type QueryService interface {
	ExecuteQuery(
		ctx context.Context,
		resource string,
		queryString string,
	) (*querymodel.ExecuteResult, error)
}

func NewQueryApiHandler(service QueryService) *QueryApiHandler {
	return &QueryApiHandler{
		service: service,
	}
}

type QueryApiHandler struct {
	service QueryService
}

func (h QueryApiHandler) ProcessQuery(
	w http.ResponseWriter,
	r *http.Request,
	c *PipelineContext[NoState],
) *AppError {
	resource := r.PathValue("resource")

	decodedQuery, err := url.QueryUnescape(r.URL.RawQuery)
	if err != nil {
		return NewServerError(err)
	}

	res, err := h.service.ExecuteQuery(
		r.Context(),
		resource,
		decodedQuery,
	)
	if err != nil {
		errType := queryerror.GetErrorCategory(err)
		if errType == queryerror.QueryErrInternalErr ||
			errType == queryerror.QueryErrUnknown {
			return NewServerError(err)
		} else {
			return &AppError{
				Code:       400,
				ToastError: err.Error(),
				innerError: err,
			}
		}
	}

	jsonBody, err := json.Marshal(res)
	if err != nil {
		return NewServerError(err)
	}

	if _, err := w.Write(jsonBody); err != nil {
		return NewServerError(err)
	}

	return nil
}
