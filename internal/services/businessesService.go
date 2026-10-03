package services

import (
	"context"

	"github.com/turnerbenjamin/heterogen_portal/internal/accesspolicies"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack"
	"github.com/turnerbenjamin/querystack/querymodel"
)

// QueryExecutor executes resource queries and returns their results.
type TypedQueryExecutor[T querymodel.TableModel] interface {
}

type NewQueryStackQueryExecutor func(
	repository querymodel.Repository,
	schema querymodel.Schema,
	paginationTokenSigner querymodel.PayloadSigner,
	paginationTokenSecret []byte,
	sqlFlavor querymodel.SqlFlavour,
	queryConfig querymodel.QueryConfig,
) (*querystack.QueryExecutor, error)

type BusinessService struct {
	queryExecutor      *querystack.QueryExecutor
	typedQueryExecutor func(
		executor *querystack.QueryExecutor,
		resource querymodel.TypedResource[*model.BusinessesModel],
		ctx context.Context,
		accessPolicy querymodel.AccessPolicy,
		queryString string,
	) (*querymodel.TypedExecuteResult[*model.BusinessesModel], error)
}

func NewBusinessesService(
	queryRepo querymodel.Repository,
	queryExecutorFactory NewQueryStackQueryExecutor,
	paginationTokenSigner querymodel.PayloadSigner,
	paginationTokenSecret []byte,

) (*BusinessService, error) {
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

	return &BusinessService{
		queryExecutor:      queryExecutor,
		typedQueryExecutor: querystack.Execute[*model.BusinessesModel],
	}, nil
}

func (s *BusinessService) ExecuteQuery(
	ctx context.Context,
	queryString string,
) (*querymodel.TypedExecuteResult[*model.BusinessesModel], error) {
	results, err := s.typedQueryExecutor(
		s.queryExecutor,
		&model.BusinessesResource{},
		ctx,
		accesspolicies.BusinessTableQueryAccessPolicy,
		queryString,
	)

	if err != nil {
		return nil, err
	}
	return results, nil
}
