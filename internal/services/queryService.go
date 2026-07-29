package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/query"
	queryaccesspolicies "github.com/turnerbenjamin/heterogen_portal/internal/queryAccessPolicies"
)

type QueryRepo interface {
	Execute(ctx context.Context, statementString string, args []any) (jsonResult []byte, err error)
}

type QueryParser interface {
	Parse(queryString string) (operations []query.QueryOperation, err error)
}

var schemaMetadata = model.NewSchemaMetadata()

type QueryService struct {
	queryRepo   QueryRepo
	queryParser QueryParser
}

var accessPolicy = queryaccesspolicies.GetAnonymousAccessPolicy()

func NewQueryService(queryRepo QueryRepo, queryParser QueryParser) *QueryService {
	return &QueryService{
		queryRepo:   queryRepo,
		queryParser: queryParser,
	}
}

func (s *QueryService) Execute(ctx context.Context, resource string, queryString string) ([]query.TableModel, *etc.AppError) {
	queryOperations, err := s.queryParser.Parse(queryString)
	if err != nil {
		return nil, &etc.AppError{
			Code:         http.StatusBadRequest,
			ErrorMessage: fmt.Sprintf("Unable to parse query string: %s", err.Error()),
		}
	}

	q, err := query.NewQuery(
		ctx,
		schemaMetadata,
		accessPolicy,
		resource,
		queryOperations,
		s.queryRepo.Execute,
	)
	if err != nil {
		return nil, &etc.AppError{
			Code:         http.StatusBadRequest,
			ErrorMessage: err.Error(),
			InnerError:   err,
			ResponseType: etc.ResponseTypeJson,
		}
	}

	results, err := q.Execute()
	if err != nil {
		return nil, &etc.AppError{
			Code:         http.StatusBadRequest,
			ErrorMessage: err.Error(),
			InnerError:   err,
			ResponseType: etc.ResponseTypeJson,
		}
	}

	return results, nil
}
