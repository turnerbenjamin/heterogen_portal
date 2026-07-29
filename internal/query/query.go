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
	NewSlice(jsonData []byte) ([]TableModel, error)

	// SetRelationshipField sets a given relationship field
	SetRelationshipField(relationshipId string, value TableModel) error

	// GetJoinOnValue returns the value of the relevant column for a given relationship
	GetJoinOnValue(relationshipId string) (string, error)
}

type TableMetadata interface {
	GetColumnMetadata(columnName string) ColumnMetadata
	GetRelationshipMetadata(columnName string) RelationshipMetadata
	GetModel() TableModel
	Name() string
	FullyQualifiedName() string
	Columns() iter.Seq[ColumnMetadata]
	ColumnCount() int
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
}

type nestedQueryResult struct {
	link    *TraversalStep
	results []TableModel
}

type Query struct {
	ctx               context.Context
	queryExecutor     func(ctx context.Context, statementStr string, args []any) (jsonResult []byte, err error)
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
	operations []QueryOperation,
	exectuteQuery func(ctx context.Context, statementStr string, args []any) (jsonResult []byte, err error),
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

	queryBuilder, err := NewSqlQueryBuilder(
		resource,
		accessPolicy,
		operations,
	)
	if err != nil {
		return nil, err
	}

	q := &Query{
		ctx:           ctx,
		queryExecutor: exectuteQuery,
		tableMetadata: resource,
		queryBuilder:  queryBuilder,
	}
	return q, nil
}

func (q *Query) Execute() ([]TableModel, error) {
	return q.executeQuery(q.queryBuilder)
}

func (q *Query) executeQuery(qb *sqlQueryBuilder) ([]TableModel, error) {
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

	json, err := q.queryExecutor(q.ctx, query.statement, query.args)
	if err != nil {
		return nil, internalErr("query executor failed: %v", err)
	}

	queryResults, stdErr := resourceModel.NewSlice(json)
	if stdErr != nil {
		return nil, internalErr("unable to create model slice: %v", stdErr)
	}

	nestedQueryResults := make(map[string]*nestedQueryResult)
	var mu sync.Mutex

	// Execute any nested queries
	if len(queryResults) > 0 && len(qb.nestedQueries) > 0 {
		g, _ := errgroup.WithContext(q.ctx)

		for _, nestedQuery := range qb.nestedQueries {

			g.Go(func() error {
				link := nestedQuery.link
				nestedQueryBuilder := nestedQuery.queryBuilder

				joinOnValues, err := q.getJoinOnValues(link, queryResults)
				if err != nil {
					return err
				}

				err = nestedQueryBuilder.addAssociatedWithParentFilter(
					link,
					joinOnValues,
				)
				if err != nil {
					return err
				}

				results, err := q.executeQuery(nestedQueryBuilder)
				if err != nil {
					return err
				}

				mu.Lock()
				nestedQueryResults[link.Relationship.Id()] = &nestedQueryResult{
					link:    link,
					results: results,
				}
				mu.Unlock()
				return nil
			})

		}

		err := g.Wait()
		if err != nil {
			return nil, err
		}

		// Attach nested queries to main results
		for _, result := range nestedQueryResults {
			link := result.link
			nestedResults := result.results

			switch link.Relationship.Type() {
			case RelationshipManyToOne:
				err := attachNestedResultsForManyToOneQuery(link.Relationship, queryResults, nestedResults)
				if err != nil {
					return nil, err
				}
			case RelationshipOneToMany:
				err := attachNestedResultsForOneToManyQuery(link.Relationship, queryResults, nestedResults)
				if err != nil {
					return nil, err
				}
			default:
				return nil, internalErr(
					"unsupported relationship type %v",
					link.Relationship.Type(),
				)
			}
		}
	}

	return queryResults, nil
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
