package services

import (
	"context"

	"github.com/turnerbenjamin/heterogen_portal/internal/accesspolicies"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack"
	"github.com/turnerbenjamin/querystack/querymodel"
)

type UsersService struct {
	queryExecutor      *querystack.QueryExecutor
	typedQueryExecutor func(
		executor *querystack.QueryExecutor,
		resource querymodel.TypedResource[*model.UsersModel],
		ctx context.Context,
		accessPolicy querymodel.AccessPolicy,
		queryString string,
	) (*querymodel.TypedExecuteResult[*model.UsersModel], error)
}

func NewUsersService(
	queryRepo querymodel.Repository,
	queryExecutorFactory NewQueryStackQueryExecutor,
	paginationTokenSigner querymodel.PayloadSigner,
	paginationTokenSecret []byte,

) (*UsersService, error) {
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

	return &UsersService{
		queryExecutor:      queryExecutor,
		typedQueryExecutor: querystack.Execute[*model.UsersModel],
	}, nil
}

func (s *UsersService) ExecuteQuery(
	ctx context.Context,
	queryString string,
) (*querymodel.TypedExecuteResult[*model.UsersModel], error) {
	results, err := s.typedQueryExecutor(
		s.queryExecutor,
		&model.UsersResource{},
		ctx,
		accesspolicies.UsersTableQueryAccessPolicy,
		queryString,
	)

	if err != nil {
		return nil, err
	}
	return results, nil
}
