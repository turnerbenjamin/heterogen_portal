package query

import (
	"context"

	bldr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryBuilder"
	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryExecutor"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	azSqlWriter "github.com/turnerbenjamin/heterogen_portal/internal/query/queryWriters/azSqlWriter"
)

type Query struct {
	ctx                  context.Context
	queryString          string
	tableMetadata        mdl.TableMetadata
	AccessPolicy         mdl.AccessPolicy
	TableAccessPolicy    mdl.TableAccessPolicy
	queryDataStore       qstore.QueryDataStore
	valueBuilder         mdl.ValueBuilder
	nextPageTokenBuilder bldr.PagingTokenBuilder
	queryExecutor        *queryExecutor.QueryExecutor
}

// TODO simplify query to wiring only
// Add a query executor factory struct which takes a config - This can then be
// passed to the services as the sole dependency needed for queries. The query
// factory will simply initialise a query using dependencies from the current
// package. Query should have very little in it, it will be responsible for
// wiring and won't be unit tested.
//
// Query executor will change to repository and the execution logic will move to
// a new query executor which will be the main control - Building the query,
// generating statements, sending to the repo, and stitching together results
//
// This should get things set up nicely for unit testing

func NewQuery(
	ctx context.Context,
	schema mdl.Schema,
	accessPolicy mdl.AccessPolicy,
	resourceName string,
	queryString string,
	nextPageTokenBuilder bldr.PagingTokenBuilder,
	queryParser bldr.QueryParser,
	repository queryExecutor.Repository,
) (*Query, error) {
	resource := schema.GetTableMetadata(resourceName)
	if resource == nil {
		return nil, qerr.BindingErr("the table %s does not exist in the schema", resourceName)
	}

	if accessPolicy == nil {
		return nil, qerr.InternalErr("access policy cannot be nil")
	}

	resourceAccessPolicy := accessPolicy.GetTableAccessPolicy(resourceName)
	if resourceAccessPolicy == nil {
		return nil, qerr.InternalErr("unable to find table access policy for table %s", resourceName)
	}

	valueBuilder := azSqlWriter.NewValueBuilder()

	// Build query
	queryDataStore, err := bldr.BuildQuery(
		queryString,
		queryParser,
		nextPageTokenBuilder,
		valueBuilder,
		resource,
		accessPolicy,
	)
	if err != nil {
		return nil, err
	}

	executor := queryExecutor.NewQueryExecutor(
		repository,
		azSqlWriter.NewQueryWriter,
		nextPageTokenBuilder,
		azSqlWriter.NewValueBuilder(),
	)

	q := &Query{
		ctx:                  ctx,
		queryString:          queryString,
		tableMetadata:        resource,
		queryDataStore:       queryDataStore,
		valueBuilder:         valueBuilder,
		nextPageTokenBuilder: nextPageTokenBuilder,
		queryExecutor:        executor,
	}
	return q, nil
}

func (q *Query) Execute() (queryExecutor.ExecuteResult, error) {
	return q.queryExecutor.ExecuteQuery(q.ctx, q.queryDataStore)
}
