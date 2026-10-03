package services

import (
	"context"

	"github.com/turnerbenjamin/heterogen_portal/internal/accesspolicies"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack/querymodel"
)

// QueryExecutor executes resource queries and returns their results.
type QueryExecutor interface {
	Execute(
		ctx context.Context,
		resourceName string,
		accessPolicy querymodel.AccessPolicy,
		queryString string,
	) (*querymodel.ExecuteResult, error)
}

type QueryApiService struct {
	queryExecutor QueryExecutor
}

var schema querymodel.Schema = model.NewSchema()

// type NewServiceQueryExecutor func(
// 	repository querymodel.Repository,
// 	schema querymodel.Schema,
// 	paginationTokenSigner querymodel.PayloadSigner,
// 	paginationTokenSecret []byte,
// 	sqlFlavor querymodel.SqlFlavour,
// 	queryConfig querymodel.QueryConfig,
// ) (QueryExecutor, error)

func NewQueryApiService(
	queryRepo querymodel.Repository,
	queryExecutorFactory NewQueryStackQueryExecutor,
	paginationTokenSigner querymodel.PayloadSigner,
	paginationTokenSecret []byte,

) (*QueryApiService, error) {
	queryExecutor, err := queryExecutorFactory(
		queryRepo,
		schema,
		paginationTokenSigner,
		paginationTokenSecret,
		querymodel.SqlFlavorAzureSql,
		querymodel.QueryConfig{
			DefaultPageSize:   100,
			MaxRecordsPerPage: 100_000,
			MaxDepth:          10,
		},
	)
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
		accesspolicies.AnonymousAccessPolicy,
		queryString,
	)

	if err != nil {
		return nil, err
	}
	return results, nil
}
