package services

import (
	"context"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/query"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	"github.com/turnerbenjamin/heterogen_portal/internal/queryAccessPolicies"
)

type QueryService struct {
	queryExecutor query.QueryExecutor
}

var accessPolicy = queryAccessPolicies.GetAnonymousAccessPolicy()

func NewQueryService(
	queryRepo queryModel.Repository,
	paginationTokenSigner queryModel.PayloadSigner,
	paginationTokenSecret []byte,

) (*QueryService, error) {
	queryExecutor, err := query.NewQueryExecutorFactory(query.QueryExecutorConfig{
		Repo:                  queryRepo,
		Schema:                model.NewSchema(),
		AccessPolicy:          accessPolicy,
		PaginationTokenSigner: paginationTokenSigner,
		PaginationTokenSecret: paginationTokenSecret,
		SqlFlavor:             query.SqlFlavorAzureSql,
	})
	if err != nil {
		return nil, err
	}

	return &QueryService{
		queryExecutor: queryExecutor,
	}, nil
}

func (s *QueryService) Execute(ctx context.Context, resource string, queryString string) (*queryModel.ExecuteResult, *etc.AppError) {
	results, err := s.queryExecutor.Execute(
		ctx,
		resource,
		queryString,
	)

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
