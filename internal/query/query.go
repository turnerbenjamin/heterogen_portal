package query

import (
	"context"
	"fmt"

	"github.com/turnerbenjamin/heterogen_portal/internal/query/paginationTokens"
	bldr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryBuilder"
	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryOrchestrator"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryParser"
	azSqlWriter "github.com/turnerbenjamin/heterogen_portal/internal/query/queryWriters/azSqlWriter"
)

type sqlFlavour string

const SqlFlavorAzureSql sqlFlavour = "azure_sql"

type QueryWriterGetter func(s qstore.QueryDataStore) (w mdl.QueryWriter, err error)
type ValueBuilderGetter func() mdl.ValueBuilder
type QueryParserInitialiser func() bldr.QueryParser

type sqlWriterConfig struct {
	queryWriterGetter  QueryWriterGetter
	ValueBuilderGetter ValueBuilderGetter
}

type queryExecutor struct {
	repository             mdl.Repository
	schema                 mdl.Schema
	accessPolicy           mdl.AccessPolicy
	pagingTokenBuilder     bldr.PagingTokenBuilder
	queryWriterGetter      QueryWriterGetter
	valueBuilderGetter     ValueBuilderGetter
	queryParserInitialiser QueryParserInitialiser
}
type QueryOrchestrator interface {
	Execute(
		ctx context.Context,
		resourceName string,
		queryString string,
	) (queryModel.ExecuteResult, error)
}

type QueryExecutorConfig struct {
	Repo                  mdl.Repository
	Schema                mdl.Schema
	AccessPolicy          mdl.AccessPolicy
	PaginationTokenSigner mdl.PayloadSigner
	PaginationTokenSecret []byte
	SqlFlavor             sqlFlavour
}

func NewQueryExecutorFactory(config QueryExecutorConfig) (QueryOrchestrator, error) {
	pagingTokenBuilder, err := paginationTokens.NewPagingTokenBuilder(
		config.PaginationTokenSigner,
		config.PaginationTokenSecret,
	)
	if err != nil {
		return nil, err
	}

	sqlWriterConfig, err := getSqlWriterConfig(config.SqlFlavor)
	if err != nil {
		return nil, err
	}

	return &queryExecutor{
		repository:             config.Repo,
		schema:                 config.Schema,
		accessPolicy:           config.AccessPolicy,
		pagingTokenBuilder:     pagingTokenBuilder,
		queryWriterGetter:      azSqlWriter.NewQueryWriter,
		valueBuilderGetter:     sqlWriterConfig.ValueBuilderGetter,
		queryParserInitialiser: queryParser.NewQueryParser,
	}, err
}

func (qf *queryExecutor) Execute(
	ctx context.Context,
	resourceName string,
	queryString string,
) (queryModel.ExecuteResult, error) {
	resource := qf.schema.GetTableMetadata(resourceName)
	if resource == nil {
		return queryModel.ExecuteResult{}, qerr.BindingErr("the table %s does not exist in the schema", resourceName)
	}

	resourceAccessPolicy := qf.accessPolicy.GetTableAccessPolicy(resourceName)
	if resourceAccessPolicy == nil {
		return queryModel.ExecuteResult{}, qerr.InternalErr("unable to find table access policy for table %s", resourceName)
	}

	queryParser := qf.queryParserInitialiser()
	valueBuilder := qf.valueBuilderGetter()

	// Build query
	queryDataStore, err := bldr.BuildQuery(
		queryString,
		queryParser,
		qf.pagingTokenBuilder,
		valueBuilder,
		resource,
		qf.accessPolicy,
	)
	if err != nil {
		return queryModel.ExecuteResult{}, err
	}

	executor := queryOrchestrator.NewQueryExecutor(
		qf.repository,
		azSqlWriter.NewQueryWriter,
		qf.pagingTokenBuilder,
		azSqlWriter.NewValueBuilder(),
	)

	return executor.ExecuteQuery(ctx, queryDataStore)
}

func getSqlWriterConfig(flavour sqlFlavour) (sqlWriterConfig, error) {
	switch flavour {
	case SqlFlavorAzureSql:
		return sqlWriterConfig{
			queryWriterGetter:  azSqlWriter.NewQueryWriter,
			ValueBuilderGetter: azSqlWriter.NewValueBuilder,
		}, nil
	default:
		return sqlWriterConfig{}, fmt.Errorf("unsupported sql flavour: %s", flavour)
	}
}
