package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/query"
)

type QueryRepo interface {
	Execute(ctx context.Context, query string) (jsonResult []byte, err error)
}

type QueryParser interface {
	Parse(queryString string) (operations []query.QueryOperation, err error)
}

type QueryService struct {
	queryRepo   QueryRepo
	queryParser QueryParser
}

func NewQueryService(queryRepo QueryRepo, queryParser QueryParser) *QueryService {
	return &QueryService{
		queryRepo:   queryRepo,
		queryParser: queryParser,
	}
}

func (s *QueryService) Execute(ctx context.Context, resource string, queryString string) ([]model.TableModel, *etc.AppError) {
	queryOperations, err := s.queryParser.Parse(queryString)
	if err != nil {
		return nil, &etc.AppError{
			Code:         http.StatusBadRequest,
			ErrorMessage: fmt.Sprintf("Unable to parse query string: %s", err.Error()),
		}
	}

	q, appErr := query.NewQuery(ctx, s.queryRepo.Execute, resource, queryOperations)
	if appErr != nil {
		return nil, appErr
	}
	results, err := q.Execute()
	if err != nil {
		return nil, etc.NewServerError(err)
	}

	return results, nil
}
