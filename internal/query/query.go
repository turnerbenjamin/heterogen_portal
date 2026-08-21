package query

import (
	"context"
	"iter"
	"sync"

	"golang.org/x/sync/errgroup"
)

// DbDataTypeName represents a SQL Server data type name supported by the model metadata system.
type DbDataTypeName string

const (
	// DbTypeNvarchar represents the SQL Server nvarchar data type.
	DbTypeNvarchar DbDataTypeName = "nvarchar"

	// DbTypeInt represents the SQL Server int data type.
	DbTypeInt DbDataTypeName = "int"

	// DbTypeFloat represents the SQL Server float data type.
	DbTypeFloat DbDataTypeName = "float"

	// DbTypeGeography represents the SQL Server geography spatial data type.
	DbTypeGeography DbDataTypeName = "geography"

	// DbTypeDateTimeOffset represents the SQL Server datetimeoffset date/time data type.
	DbTypeDateTimeOffset DbDataTypeName = "datetimeoffset"
)

// relationshipType represents different table relationships
type RelationshipType string

const (
	// RelationshipOneToMany represents a 1:N relationship
	RelationshipOneToMany RelationshipType = "1:N"

	// RelationshipManyToOne represents an N:1 relationship
	RelationshipManyToOne RelationshipType = "N:1"
)

type Schema interface {
	GetTableMetadata(tableName string) TableMetadata
}

type TableModel interface {
	// NewSlice unmarshals a json array and returns it as a slice
	NewSlice(
		jsonData []byte,
		projectColumns []string,
	) ([]TableModel, error)

	// SetRelationshipField sets a given relationship field
	SetRelationshipField(relationshipId string, value TableModel) error

	// GetJoinOnValue returns the value of the relevant column for a given relationship
	GetJoinOnValue(relationshipId string) (string, error)

	// GetValueExpression returns a value expression for a given path
	GetValueExpression(path []TraversalStep, columnName string) (ValueExpression, error)

	// IsNil is used to determine if a typed nil pointer contains a nil value
	IsNil() bool
}

type TableMetadata interface {
	GetColumnMetadata(columnName string) ColumnMetadata
	GetRelationshipMetadata(columnName string) RelationshipMetadata
	GetModel() TableModel
	Name() string
	FullyQualifiedName() string
	Columns() iter.Seq[ColumnMetadata]
	ColumnCount() int
	PrimaryKeyField() ColumnMetadata
}

type ColumnMetadata interface {
	Name() string
	Type() DbDataTypeName
}

type RelationshipMetadata interface {
	Id() string
	From() TableMetadata
	To() TableMetadata
	FromColumn() ColumnMetadata
	ToColumn() ColumnMetadata
	Type() RelationshipType
	RelationshipColumn() string
}

type QueryParser interface {
	Parse(queryString string) (operations *Operations, err error)
}

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
		resourceName string,
		queryString string,
		orderByOperation *OrderByOperation,
		lastRecord TableModel,
	) (string, error)

	ParseToken(token string) (*pagingToken, error)
}

type nestedQueryResult struct {
	link    *TraversalStep
	results []TableModel
}

type ExecuteResult struct {
	Count         *int64       `json:"count,omitempty"`
	NextPageToken string       `json:"next_page_token,omitempty"`
	Data          []TableModel `json:"data"`
}

type Query struct {
	ctx               context.Context
	queryString       string
	queryExecutor     QueryExecutor
	tableMetadata     TableMetadata
	AccessPolicy      AccessPolicy
	TableAccessPolicy TableAccessPolicy
	queryBuilder      *sqlQueryBuilder
}

func NewQuery(
	ctx context.Context,
	schema Schema,
	accessPolicy AccessPolicy,
	resourceName string,
	queryString string,
	nextPageTokenBuilder PagingTokenBuilder,
	queryParser QueryParser,
	queryExecutor QueryExecutor,
) (*Query, error) {
	resource := schema.GetTableMetadata(resourceName)
	if resource == nil {
		return nil, bindingErr("the table %s does not exist in the schema", resourceName)
	}

	if accessPolicy == nil {
		return nil, internalErr("access policy cannot be nil")
	}

	resourceAccessPolicy := accessPolicy.GetTableAccessPolicy(resourceName)
	if resourceAccessPolicy == nil {
		return nil, internalErr("unable to find table access policy for table %s", resourceName)
	}

	queryOperations, err := BuildQueryOperations(
		resource,
		accessPolicy,
		queryString,
		queryParser,
		nextPageTokenBuilder,
	)
	if err != nil {
		return nil, err
	}

	queryBuilder, err := newSqlQueryBuilder(
		resource,
		accessPolicy,
		queryOperations,
		nextPageTokenBuilder,
	)
	if err != nil {
		return nil, err
	}

	q := &Query{
		ctx:           ctx,
		queryString:   queryString,
		queryExecutor: queryExecutor,
		tableMetadata: resource,
		queryBuilder:  queryBuilder,
	}
	return q, nil
}

func (q *Query) Execute() (*ExecuteResult, error) {
	resourceModel := q.queryBuilder.rootResource.GetModel()
	if resourceModel == nil {
		return nil, bindingErr(
			"unable to access model for %s",
			q.queryBuilder.rootResource.Name(),
		)
	}

	query, countStatement, err := q.queryBuilder.buildTopLevelQuery()
	if err != nil {
		return nil, err
	}

	json, count, err := q.queryExecutor.ExecuteJsonRequestWithCount(
		q.ctx,
		query.statement,
		countStatement,
		query.args,
	)
	if err != nil {
		return nil, internalErr("query executor failed: %v", err)
	}

	queryResults, stdErr := resourceModel.NewSlice(json, q.queryBuilder.operations.projection)
	if stdErr != nil {
		return nil, internalErr("unable to create model slice: %v", stdErr)
	}

	// Get System Selects
	err = q.populateNestedResults(q.queryBuilder, queryResults)
	if err != nil {
		return nil, err
	}

	// Top level queries set the limit to the requested limit + 1 so that it is
	// possible to determine if a next page of results exists
	var nextPageToken string
	limit := q.queryBuilder.operations.LimitOperation.Limit
	isNextRecord := len(queryResults) > limit
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

	return &ExecuteResult{
		Count:         count,
		NextPageToken: nextPageToken,
		Data:          queryResults,
	}, nil
}

func (q *Query) populateNestedResults(
	qb *sqlQueryBuilder,
	queryResults []TableModel,
) error {
	nestedQueryResults, err := q.getNestedQueryResults(qb, queryResults)
	if err != nil {
		return err
	}

	for _, result := range nestedQueryResults {
		link := result.link
		nestedResults := result.results

		switch link.Relationship.Type() {
		case RelationshipManyToOne:
			err := attachNestedResultsForManyToOneQuery(link.Relationship, queryResults, nestedResults)
			if err != nil {
				return err
			}
		case RelationshipOneToMany:
			err := attachNestedResultsForOneToManyQuery(link.Relationship, queryResults, nestedResults)
			if err != nil {
				return err
			}
		default:
			return internalErr(
				"unsupported relationship type %v",
				link.Relationship.Type(),
			)
		}
	}
	return nil
}

func (q *Query) getNestedQueryResults(
	qb *sqlQueryBuilder,
	queryResults []TableModel,
) (map[string]*nestedQueryResult, error) {
	if len(queryResults) == 0 || len(qb.nestedQueries) == 0 {
		return nil, nil
	}

	// Execute nested queries asynchronously and collect the results in a map
	nestedQueryResults := make(map[string]*nestedQueryResult)
	var mu sync.Mutex
	g, _ := errgroup.WithContext(q.ctx)
	for _, nestedQuery := range qb.nestedQueries {
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

func (q *Query) executeNestedQueries(qb *sqlQueryBuilder) ([]TableModel, error) {
	resourceModel := qb.rootResource.GetModel()
	if resourceModel == nil {
		return nil, bindingErr(
			"unable to access model for %s",
			qb.rootResource.Name(),
		)
	}

	query, err := qb.build()
	if err != nil {
		return nil, err
	}

	json, err := q.queryExecutor.ExecuteJsonRequest(q.ctx, query.statement, query.args)
	if err != nil {
		return nil, internalErr("query executor failed: %v", err)
	}

	queryResults, stdErr := resourceModel.NewSlice(
		json,
		qb.operations.projection,
	)
	if stdErr != nil {
		return nil, internalErr("unable to create model slice: %v", stdErr)
	}

	err = q.populateNestedResults(qb, queryResults)
	if err != nil {
		return nil, err
	}

	return queryResults, nil
}

func (q *Query) executeNestedQuery(
	nestedQuery *nestedQuery,
	queryResults []TableModel,
) (*nestedQueryResult, error) {
	link := nestedQuery.link
	nestedQueryBuilder := nestedQuery.queryBuilder

	joinOnValues, err := q.getJoinOnValues(link, queryResults)
	if err != nil {
		return nil, err
	}

	err = nestedQueryBuilder.addAssociatedWithParentFilter(
		link,
		joinOnValues,
	)
	if err != nil {
		return nil, err
	}

	results, err := q.executeNestedQueries(nestedQueryBuilder)
	if err != nil {
		return nil, err
	}

	return &nestedQueryResult{
		link:    link,
		results: results,
	}, nil
}

func (q Query) getJoinOnValues(
	link *TraversalStep,
	fromResults []TableModel,
) ([]string, error) {
	seen := map[string]struct{}{}
	joinValues := make([]string, 0, len(fromResults))

	for _, fromResult := range fromResults {
		value, err := fromResult.GetJoinOnValue(link.Relationship.Id())
		if err != nil {
			return nil, err
		}

		if _, exists := seen[value]; !exists {
			joinValues = append(joinValues, value)
		}
		seen[value] = struct{}{}
	}
	return joinValues, nil
}

func attachNestedResultsForManyToOneQuery(
	relationship RelationshipMetadata,
	parentResults,
	nestedResults []TableModel,
) error {
	nestedResultsMap := map[string]TableModel{}
	for _, nestedResult := range nestedResults {
		joinOnValue, err := nestedResult.GetJoinOnValue(relationship.Id())
		if err != nil {
			return internalErr("unable to get join on value: %v", err)
		}
		nestedResultsMap[joinOnValue] = nestedResult
	}

	for _, result := range parentResults {
		joinOnValue, err := result.GetJoinOnValue(relationship.Id())
		if err != nil {
			return internalErr("unable to get join on value: %v", err)
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
	relationship RelationshipMetadata,
	parentResults,
	nestedResults []TableModel,
) error {
	parentResultsMap := map[string]TableModel{}
	for _, parentResult := range parentResults {
		joinOnValue, err := parentResult.GetJoinOnValue(relationship.Id())
		if err != nil {
			return internalErr("unable to get join on value: %v", err)
		}
		parentResultsMap[joinOnValue] = parentResult
	}

	for _, nestedResult := range nestedResults {
		joinOnValue, err := nestedResult.GetJoinOnValue(relationship.Id())
		if err != nil || joinOnValue == "" {
			return internalErr("unable to get join on value: %v", err)
		}
		parent, ok := parentResultsMap[joinOnValue]
		if !ok || parent == nil {
			return internalErr("unable to join query results")
		}
		parent.SetRelationshipField(relationship.Id(), nestedResult)
	}
	return nil
}
