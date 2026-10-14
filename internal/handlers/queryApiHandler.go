package handlers

import (
	"context"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
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

type QueryApiHandler struct {
	service        QueryService
	decodeUrl      func(s string) (string, error)
	jsonSerialiser JsonSerialiser
}

func NewQueryApiHandler(
	service QueryService,
	decodeUrl func(s string) (string, error),
	jsonSerialiser JsonSerialiser,
) *QueryApiHandler {
	return &QueryApiHandler{
		service:        service,
		decodeUrl:      decodeUrl,
		jsonSerialiser: jsonSerialiser,
	}
}

func (h QueryApiHandler) ProcessQuery(
	w http.ResponseWriter,
	r *http.Request,
	c *PipelineContext[NoState],
) *etc.AppError {
	resource := r.PathValue(constants.UrlParamResource)

	decodedQuery, err := h.decodeUrl(r.URL.RawQuery)
	if err != nil {
		return etc.NewServerError(err)
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
			return etc.NewServerError(err)
		} else {
			return &etc.AppError{
				Code:         400,
				ErrorMessage: err.Error(),
				InnerError:   err,
			}
		}
	}

	jsonBody, err := h.jsonSerialiser.Marshal(res)
	if err != nil {
		return etc.NewServerError(err)
	}

	if _, err := w.Write(jsonBody); err != nil {
		return etc.NewServerError(err)
	}

	return nil
}
