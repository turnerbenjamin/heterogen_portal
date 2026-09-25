package query

import (
	"context"
	"fmt"

	"github.com/turnerbenjamin/heterogen_portal/internal/query/paginationTokens"
	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryParser"
	qplan "github.com/turnerbenjamin/heterogen_portal/internal/query/queryPlanner"
	azSqlWriter "github.com/turnerbenjamin/heterogen_portal/internal/query/queryWriters/azSqlWriter"
	valuebuilder "github.com/turnerbenjamin/heterogen_portal/internal/query/valueBuilder"
)

type sqlFlavour string

const SqlFlavorAzureSql sqlFlavour = "azure_sql"

type QueryWriterGetter func(s qstore.QueryDataStore) mdl.QueryWriter
type QueryParserInitialiser func() qplan.QueryParser

type queryExecutor struct {
	repository             mdl.Repository
	queryConfig            mdl.QueryConfig
	schema                 mdl.Schema
	accessPolicy           mdl.AccessPolicy
	pagingTokenBuilder     qplan.PagingTokenBuilder
	queryWriterGetter      QueryWriterGetter
	queryParserInitialiser QueryParserInitialiser
}
type QueryExecutor interface {
	Execute(
		ctx context.Context,
		resourceName string,
		queryString string,
	) (*queryModel.ExecuteResult, error)
}

type QueryExecutorConfig struct {
	Repo                  mdl.Repository
	Schema                mdl.Schema
	AccessPolicy          mdl.AccessPolicy
	PaginationTokenSigner mdl.PayloadSigner
	PaginationTokenSecret []byte
	SqlFlavor             sqlFlavour
	queryConfig           mdl.QueryConfig
}

func NewQueryExecutorFactory(config QueryExecutorConfig) (QueryExecutor, error) {
	pagingTokenBuilder, err := paginationTokens.NewPagingTokenBuilder(
		config.PaginationTokenSigner,
		config.PaginationTokenSecret,
	)
	if err != nil {
		return nil, err
	}

	sqlWriterGetter, err := getSqlWriter(config.SqlFlavor)

	if err != nil {
		return nil, err
	}

	return &queryExecutor{
		repository:             config.Repo,
		schema:                 config.Schema,
		accessPolicy:           config.AccessPolicy,
		pagingTokenBuilder:     pagingTokenBuilder,
		queryWriterGetter:      sqlWriterGetter,
		queryParserInitialiser: queryParser.NewQueryParser,
		queryConfig:            mdl.QueryConfigWithDefaults(config.queryConfig),
	}, err
}

func (qf *queryExecutor) Execute(
	ctx context.Context,
	resourceName string,
	queryString string,
) (*queryModel.ExecuteResult, error) {
	rootResource, exists := qf.schema.GetResource(resourceName)
	if !exists {
		return nil, qerr.BindingErr("the table %s does not exist in the schema", resourceName)
	}

	queryParser := qf.queryParserInitialiser()
	valueBuilder := valuebuilder.NewValueBuilder()

	// Plan query
	s, err := qplan.PlanQuery(
		qf.queryConfig,
		queryString,
		queryParser,
		qf.pagingTokenBuilder,
		valueBuilder,
		rootResource,
		qf.accessPolicy,
	)
	if err != nil {
		return nil, err
	}

	// Write and execute the query
	w := qf.queryWriterGetter(s)
	json, count, err := qf.executeQueryStatement(ctx, s, w)

	// Parse the query into a table model
	queryResults, stdErr := rootResource.SliceFromJSON(
		json,
		s.ProjectionNode(),
	)
	if stdErr != nil {
		return nil, qerr.InternalErr("unable to create model slice: %v", stdErr)
	}

	// Build the next page token
	nextPageToken, err := qf.getNextPageToken(s, &queryResults)
	if err != nil {
		return nil, err
	}

	// return the results
	return &mdl.ExecuteResult{
		Count:         count,
		NextPageToken: nextPageToken,
		Data:          queryResults,
	}, nil
}

func getSqlWriter(flavour sqlFlavour) (QueryWriterGetter, error) {
	switch flavour {
	case SqlFlavorAzureSql:
		return azSqlWriter.NewQueryWriter, nil
	default:
		return nil, fmt.Errorf("unsupported sql flavour: %s", flavour)
	}
}

func (e *queryExecutor) getNextPageToken(
	s qstore.QueryDataStore,
	queryResults *[]mdl.TableModel,
) (string, error) {
	results := *queryResults
	limit := s.Limit()
	isNextRecord := len(results) > int(limit)

	if !isNextRecord {
		return "", nil
	}
	lastRecord := results[limit-1]
	*queryResults = results[0:limit]

	return e.pagingTokenBuilder.BuildToken(
		s,
		lastRecord,
	)
}

func (e *queryExecutor) executeQueryStatement(
	ctx context.Context,
	s qstore.QueryDataStore,
	w mdl.QueryWriter,
) ([]byte, *uint64, error) {
	queryStatement, err := w.WriteQueryStatement()
	if err != nil {
		return nil, nil, err
	}

	var json []byte = nil
	var count *uint64 = nil
	if s.DoCount() {
		countStatement, err := w.WriteCountStatement()
		if err != nil {
			return nil, nil, err
		}

		json, count, err = e.repository.ExecuteJsonRequestWithCount(
			ctx,
			queryStatement,
			countStatement,
		)
	} else {
		json, err = e.repository.ExecuteJsonRequest(ctx, queryStatement)
	}
	if err != nil {
		return nil, nil, qerr.InternalErr("query executor failed: %v", err)
	}
	return json, count, nil
}
