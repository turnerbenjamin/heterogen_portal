package services

import (
	"context"

	"github.com/turnerbenjamin/heterogen_portal/internal/accesspolicies"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack"
	"github.com/turnerbenjamin/querystack/querymodel"
)

type QueryApiService struct {
	queryExecutor querystack.QueryExecutor
}

func NewQueryApiService(
	queryRepo querymodel.Repository,
	paginationTokenSigner querymodel.PayloadSigner,
	paginationTokenSecret []byte,

) (*QueryApiService, error) {
	queryExecutor, err := querystack.NewQueryExecutorFactory(querystack.QueryExecutorConfig{
		Repo:                  queryRepo,
		Schema:                model.NewSchema(),
		AccessPolicy:          accesspolicies.AnonymousAccessPolicy,
		PaginationTokenSigner: paginationTokenSigner,
		PaginationTokenSecret: paginationTokenSecret,
		SqlFlavor:             querystack.SqlFlavorAzureSql,
	})
	if err != nil {
		return nil, err
	}

	return &QueryApiService{
		queryExecutor: queryExecutor,
	}, nil
}

func (s *QueryApiService) ExecuteQuery(
	ctx context.Context,
	resource string,
	queryString string,
) (*querymodel.ExecuteResult, error) {
	results, err := s.queryExecutor.Execute(
		ctx,
		resource,
		queryString,
	)

	if err != nil {
		return nil, err
	}
	return results, nil
}
