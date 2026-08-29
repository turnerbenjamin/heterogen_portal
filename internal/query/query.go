package query

import (
	"context"
	"sync"

	tkns "github.com/turnerbenjamin/heterogen_portal/internal/query/paginationTokens"
	bldr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryBuilder"
	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	azSqlWriter "github.com/turnerbenjamin/heterogen_portal/internal/query/queryWriters/azSqlWriter"
	"golang.org/x/sync/errgroup"
)

type QueryExecutor interface {
	ExecuteJsonRequest(
		ctx context.Context,
		queryStatementSty string,
		args []any,
	) ([]byte, error)

	ExecuteJsonRequestWithCount(
		ctx context.Context,
		queryStatementStr string,
		countStatementStr string,
		sharedArgs []any,
	) ([]byte, *int64, error)
}

type PagingTokenBuilder interface {
	BuildToken(
		queryDataStore qstore.QueryDataStore,
		lastRecord mdl.TableModel,
	) (string, error)

	ParseToken(token string, valueBuilder mdl.ValueBuilder) (*tkns.PagingToken, error)
}

type nestedQueryResult struct {
	link    mdl.TraversalStep
	results []mdl.TableModel
}

type ExecuteResult struct {
	Count         *int64           `json:"count,omitempty"`
	NextPageToken string           `json:"next_page_token,omitempty"`
	Data          []mdl.TableModel `json:"data"`
}

type Query struct {
	ctx               context.Context
	queryString       string
	queryExecutor     QueryExecutor
	tableMetadata     mdl.TableMetadata
	AccessPolicy      mdl.AccessPolicy
	TableAccessPolicy mdl.TableAccessPolicy
	queryDataStore    qstore.QueryDataStore
	valueBuilder      mdl.ValueBuilder
}

func NewQuery(
	ctx context.Context,
	schema mdl.Schema,
	accessPolicy mdl.AccessPolicy,
	resourceName string,
	queryString string,
	nextPageTokenBuilder PagingTokenBuilder,
	queryParser bldr.QueryParser,
	queryExecutor QueryExecutor,
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
		valueBuilder,
		resource,
		accessPolicy,
	)
	if err != nil {
		return nil, err
	}

	q := &Query{
		ctx:            ctx,
		queryString:    queryString,
		queryExecutor:  queryExecutor,
		tableMetadata:  resource,
		queryDataStore: queryDataStore,
		valueBuilder:   valueBuilder,
	}
	return q, nil
}

func (q *Query) Execute() (*ExecuteResult, error) {
	rootResource := q.queryDataStore.RootResource()
	resourceModel := rootResource.GetModel()
	if resourceModel == nil {
		return nil, qerr.BindingErr(
			"unable to access model for %s",
			rootResource.Name(),
		)
	}

	w, err := azSqlWriter.NewQueryWriter(q.queryDataStore)
	if err != nil {
		return nil, err
	}

	queryStatement := w.WriteQueryStatement()
	countStatement := w.WriteCountStatement()
	args := w.Args()

	json, count, err := q.queryExecutor.ExecuteJsonRequestWithCount(
		q.ctx,
		queryStatement,
		countStatement,
		args,
	)
	if err != nil {
		return nil, qerr.InternalErr("query executor failed: %v", err)
	}

	queryResults, stdErr := resourceModel.NewSlice(json, q.queryDataStore.Projection())
	if stdErr != nil {
		return nil, qerr.InternalErr("unable to create model slice: %v", stdErr)
	}

	err = q.populateNestedResults(q.queryDataStore, queryResults)
	if err != nil {
		return nil, err
	}

	// Top level queries set the limit to the requested limit + 1 so that it is
	// possible to determine if a next page of results exists
	/*
		var nextPageToken string
		limit := q.queryDataStore.Limit()
		isNextRecord := len(queryResults) > int(limit)


		if isNextRecord {
			// If there is a next page of results, trim the sentinel record from the
			// query results
			queryResults = queryResults[0:limit]

			// Generate the next page token using the actual last record
			lastRecord := queryResults[len(queryResults)-1]
			nextPageToken, err = q.queryBuilder.operations.GetPagingToken(lastRecord)
			if err != nil {
				return nil, err
			}
		}
	*/

	return &ExecuteResult{
		Count:         count,
		NextPageToken: "",
		Data:          queryResults,
	}, nil
}

func (q *Query) populateNestedResults(
	s qstore.QueryDataStore,
	queryResults []mdl.TableModel,
) error {
	nestedQueryResults, err := q.getNestedQueryResults(s, queryResults)
	if err != nil {
		return err
	}

	for _, result := range nestedQueryResults {
		link := result.link
		nestedResults := result.results

		switch link.Relationship.Type() {
		case mdl.RelationshipManyToOne:
			err := attachNestedResultsForManyToOneQuery(link.Relationship, queryResults, nestedResults)
			if err != nil {
				return err
			}
		case mdl.RelationshipOneToMany:
			err := attachNestedResultsForOneToManyQuery(link.Relationship, queryResults, nestedResults)
			if err != nil {
				return err
			}
		default:
			return qerr.InternalErr(
				"unsupported relationship type %v",
				link.Relationship.Type(),
			)
		}
	}
	return nil
}

func (q *Query) getNestedQueryResults(
	s qstore.QueryDataStore,
	queryResults []mdl.TableModel,
) (map[string]*nestedQueryResult, error) {
	if s.ExpandsLen() == 0 || len(queryResults) == 0 {
		return nil, nil
	}

	// Execute nested queries asynchronously and collect the results in a map
	nestedQueryResults := make(map[string]*nestedQueryResult)
	var mu sync.Mutex
	g, _ := errgroup.WithContext(q.ctx)
	for _, nestedQuery := range s.Expands() {
		g.Go(func() error {
			nestedResults, err := q.executeNestedQuery(nestedQuery, queryResults)
			if err != nil {
				return err
			}

			mu.Lock()
			nestedQueryResults[nestedResults.link.Relationship.Id()] = nestedResults
			mu.Unlock()

			return nil
		})
	}

	// Wait for all nested queries to resolve
	err := g.Wait()
	if err != nil {
		return nil, err
	}

	// return results map
	return nestedQueryResults, nil
}

func (q *Query) executeNestedQueries(s qstore.QueryDataStore) ([]mdl.TableModel, error) {
	rootResource := s.RootResource()
	resourceModel := rootResource.GetModel()
	if resourceModel == nil {
		return nil, qerr.BindingErr(
			"unable to access model for %s",
			rootResource.Name(),
		)
	}

	w, err := azSqlWriter.NewQueryWriter(s)
	if err != nil {
		return nil, err
	}

	queryStatement := w.WriteQueryStatement()
	args := w.Args()

	json, err := q.queryExecutor.ExecuteJsonRequest(q.ctx, queryStatement, args)
	if err != nil {
		return nil, qerr.InternalErr("query executor failed: %v", err)
	}

	queryResults, stdErr := resourceModel.NewSlice(
		json,
		s.Projection(),
	)
	if stdErr != nil {
		return nil, qerr.InternalErr("unable to create model slice: %v", stdErr)
	}

	err = q.populateNestedResults(s, queryResults)
	if err != nil {
		return nil, err
	}

	return queryResults, nil
}

func (q *Query) executeNestedQuery(
	nestedQuery qstore.Expansion,
	queryResults []mdl.TableModel,
) (*nestedQueryResult, error) {
	link := nestedQuery.TraversalStep
	queryData := nestedQuery.QueryData

	joinOnValues, err := q.getJoinOnValues(link, queryResults)
	if err != nil {
		return nil, err
	}

	err = queryData.AddAssociatedWithParentFilter(
		link,
		joinOnValues,
	)
	if err != nil {
		return nil, err
	}

	results, err := q.executeNestedQueries(queryData)
	if err != nil {
		return nil, err
	}

	return &nestedQueryResult{
		link:    link,
		results: results,
	}, nil
}

func (q Query) getJoinOnValues(
	link mdl.TraversalStep,
	fromResults []mdl.TableModel,
) ([]mdl.ValueExpression, error) {
	seen := map[string]struct{}{}
	joinValues := make([]mdl.ValueExpression, 0, len(fromResults))

	for _, fromResult := range fromResults {
		value, err := fromResult.GetJoinOnValue(link.Relationship.Id())
		if err != nil {
			return nil, err
		}

		if _, exists := seen[value]; !exists {
			joinValues = append(joinValues, q.valueBuilder.String(value))
		}
		seen[value] = struct{}{}
	}
	return joinValues, nil
}

func attachNestedResultsForManyToOneQuery(
	relationship mdl.RelationshipMetadata,
	parentResults,
	nestedResults []mdl.TableModel,
) error {
	nestedResultsMap := map[string]mdl.TableModel{}
	for _, nestedResult := range nestedResults {
		joinOnValue, err := nestedResult.GetJoinOnValue(relationship.Id())
		if err != nil {
			return qerr.InternalErr("unable to get join on value: %v", err)
		}
		nestedResultsMap[joinOnValue] = nestedResult
	}

	for _, result := range parentResults {
		joinOnValue, err := result.GetJoinOnValue(relationship.Id())
		if err != nil {
			return qerr.InternalErr("unable to get join on value: %v", err)
		}
		related, ok := nestedResultsMap[joinOnValue]
		if !ok || related == nil {
			continue
		}
		result.SetRelationshipField(relationship.Id(), related)
	}
	return nil
}

func attachNestedResultsForOneToManyQuery(
	relationship mdl.RelationshipMetadata,
	parentResults,
	nestedResults []mdl.TableModel,
) error {
	parentResultsMap := map[string]mdl.TableModel{}
	for _, parentResult := range parentResults {
		joinOnValue, err := parentResult.GetJoinOnValue(relationship.Id())
		if err != nil {
			return qerr.InternalErr("unable to get join on value: %v", err)
		}
		parentResultsMap[joinOnValue] = parentResult
		if err := parentResult.InitRelationshipField(relationship.Id()); err != nil {
			return err
		}
	}

	for _, nestedResult := range nestedResults {
		joinOnValue, err := nestedResult.GetJoinOnValue(relationship.Id())
		if err != nil || joinOnValue == "" {
			return qerr.InternalErr("unable to get join on value: %v", err)
		}
		parent, ok := parentResultsMap[joinOnValue]
		if !ok || parent == nil {
			return qerr.InternalErr("unable to join query results")
		}
		if err := parent.SetRelationshipField(relationship.Id(), nestedResult); err != nil {
			return err
		}
	}
	return nil
}
