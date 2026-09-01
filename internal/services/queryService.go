package services

import (
	"context"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/query"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryBuilder"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryExecutor"
	"github.com/turnerbenjamin/heterogen_portal/internal/queryAccessPolicies"
)

type QueryRepo interface {
	ExecuteJsonRequest(
		ctx context.Context,
		queryStatementStr string,
		args []any,
	) ([]byte, error)

	ExecuteJsonRequestWithCount(
		ctx context.Context,
		queryStatementStr string,
		countStatementStr string,
		sharedArgs []any,
	) ([]byte, *int64, error)
}

var schemaMetadata = model.NewSchemaMetadata()

type QueryService struct {
	queryRepo            QueryRepo
	queryParser          queryBuilder.QueryParser
	nextPageTokenBuilder queryBuilder.PagingTokenBuilder
}

var accessPolicy = queryAccessPolicies.GetAnonymousAccessPolicy()

func NewQueryService(
	queryRepo QueryRepo,
	queryParser queryBuilder.QueryParser,
	nextPageTokenBuilder queryBuilder.PagingTokenBuilder,
) *QueryService {
	return &QueryService{
		queryRepo:            queryRepo,
		queryParser:          queryParser,
		nextPageTokenBuilder: nextPageTokenBuilder,
	}
}

func (s *QueryService) Execute(ctx context.Context, resource string, queryString string) (*queryExecutor.ExecuteResult, *etc.AppError) {

	q, err := query.NewQuery(
		ctx,
		schemaMetadata,
		accessPolicy,
		resource,
		queryString,
		s.nextPageTokenBuilder,
		s.queryParser,
		s.queryRepo,
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
	return &results, nil
}
