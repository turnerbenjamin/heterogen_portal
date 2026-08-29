package queryDataStore

import (
	"iter"
	"math"

	"github.com/turnerbenjamin/heterogen_portal/internal/query/metadataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/relationships"
)

type FilterExpressionBuilder interface {
	ValueBuilder() mdl.ValueBuilder

	NewLogicalExpression(
		left mdl.FilterExpression,
		operator mdl.LogicalOperator,
		right mdl.FilterExpression,
	) (*mdl.LogicalExpression, error)

	NewComparisonExpression(
		columnPath string,
		operator mdl.ComparisonOperator,
		value mdl.ValueExpression,
	) (*mdl.ComparisonExpression, error)

	NewCollectionExpression(
		resourcePath string,
		operator mdl.CollectionOperator,
		filterExpression mdl.FilterExpression,
	) (*mdl.CollectionExpression, error)
}

type QueryDataStore interface {
	RootResource() mdl.TableMetadata
	RootResourceAccessPolicy() mdl.TableAccessPolicy

	QueryString() string
	FilterExpressionBuilder() FilterExpressionBuilder
	IsTopLevelQuery() bool

	GetAliasStore() relationships.AliasStore
	JoinCollection() relationships.JoinCollection

	Projection() mdl.Projection

	AddSelect(columnName string) error
	AddSystemSelect(columnName string) error
	SelectsLen() int
	Selects() iter.Seq2[string, mdl.ResolvedColumn]

	AddExpand(relationshipId string) (QueryDataStore, error)
	AddSystemExpand(relationshipId string) (QueryDataStore, error)
	ExpandsLen() int
	Expands() iter.Seq2[string, Expansion]
	GetExpansionByRelationshipId(relationshipId string) (Expansion, bool)

	SetFilterExpression(ex mdl.FilterExpression)
	FilterExpression() mdl.FilterExpression
	AddAssociatedWithParentFilter(
		linkFromParent mdl.TraversalStep,
		joinParentOnValues []mdl.ValueExpression,
	) error // FOR NOW

	SetLimit(limit int) error
	Limit() uint16

	SetDoCount(doCount bool)
	DoCount() bool

	SetPagingToken(string)
	PagingToken() (string, bool)

	AddOrderBy(columnPath string, dir mdl.SortDirectionOperator) error
	OrderByLen() int
	OrderBy() iter.Seq[mdl.SortingRule]
}

type Expansion struct {
	QueryData     QueryDataStore
	TraversalStep mdl.TraversalStep
}

// queryDataStore is the primary store for all query data. When any operation is
// added, paths are immediately bound to metadata and the relationship plan is
// updated as required. This provides a guarantee to consumers that at any point
// in the process, all metadata is set and the relationship plan is current.
type queryDataStore struct {
	// overview data
	queryString  string
	rootResource mdl.TableMetadata
	depth        uint8
	projection   mdl.Projection

	// metadata and relationship binding deps
	metadataBinder          *metadataStore.MetadataBinder
	relationshipPlanner     *relationships.RelationshipPlanner
	filterExpressionBuilder filterExpressionBuilder
	valueBuilder            mdl.ValueBuilder

	// access policies
	accessPolicy             mdl.AccessPolicy
	rootResourceAccessPolicy mdl.TableAccessPolicy

	// operations
	selects     map[string]mdl.ResolvedColumn
	expands     map[string]Expansion
	filter      mdl.FilterExpression
	orderBy     []mdl.SortingRule
	doCount     bool
	limit       uint16
	pagingToken string
}

func NewQueryDataStore(
	queryString string,
	rootResource mdl.TableMetadata,
	accessPolicy mdl.AccessPolicy,
	valueBuilder mdl.ValueBuilder,
) (QueryDataStore, error) {
	depth := uint8(0)
	return newQueryDataStore(queryString, rootResource, accessPolicy, valueBuilder, depth)
}

func newQueryDataStore(
	queryString string,
	rootResource mdl.TableMetadata,
	accessPolicy mdl.AccessPolicy,
	valueBuilder mdl.ValueBuilder,
	depth uint8,
) (QueryDataStore, error) {
	projection, err := rootResource.InitProjection()
	if err != nil {
		return nil, err
	}

	metadataBinder, err := metadataStore.NewMetadataBinder(rootResource, accessPolicy)
	if err != nil {
		return nil, err
	}

	relationshipPlanner, err := relationships.NewRelationshipPlanner(rootResource)
	if err != nil {
		return nil, err
	}

	rootResourceAccessPolicy := accessPolicy.GetTableAccessPolicy(rootResource.Name())
	if rootResourceAccessPolicy == nil {
		return nil, qerr.InternalErr(
			"unable to create new data store: root resource access policy is nil",
		)
	}

	return &queryDataStore{
		depth:               depth,
		projection:          projection,
		queryString:         queryString,
		rootResource:        rootResource,
		accessPolicy:        accessPolicy,
		metadataBinder:      metadataBinder,
		relationshipPlanner: relationshipPlanner,

		rootResourceAccessPolicy: rootResourceAccessPolicy,
		filterExpressionBuilder: filterExpressionBuilder{
			rootResource:        rootResource,
			metadataBinder:      metadataBinder,
			relationshipPlanner: relationshipPlanner,
			valueBuilder:        valueBuilder,
		},
		valueBuilder: valueBuilder,
	}, nil
}

func (qd *queryDataStore) Projection() mdl.Projection {
	return qd.projection
}

func (qd *queryDataStore) IsTopLevelQuery() bool {
	return qd.depth == 0
}

func (qd *queryDataStore) FilterExpressionBuilder() FilterExpressionBuilder {
	return qd.filterExpressionBuilder
}

func (qd *queryDataStore) RootResource() mdl.TableMetadata {
	return qd.rootResource
}

func (qd *queryDataStore) RootResourceAccessPolicy() mdl.TableAccessPolicy {
	return qd.rootResourceAccessPolicy
}

func (qd *queryDataStore) JoinCollection() relationships.JoinCollection {
	return qd.relationshipPlanner.JoinStore
}

func (qd *queryDataStore) QueryString() string {
	return qd.queryString
}

func (qd *queryDataStore) GetAliasStore() relationships.AliasStore {
	return qd.relationshipPlanner.Aliases
}

func (qd *queryDataStore) AddSelect(columnName string) error {
	return qd.addSelect(columnName, true)
}

func (qd *queryDataStore) AddSystemSelect(columnName string) error {
	return qd.addSelect(columnName, false)
}

func (qd *queryDataStore) addSelect(columnName string, doProject bool) error {
	// Get the metadata
	c, err := qd.metadataBinder.ResolveColumn(columnName)
	if err != nil {
		return err
	}

	// Ensure select initialised
	if qd.selects == nil {
		qd.selects = make(map[string]mdl.ResolvedColumn, 1)
	}

	// Set the selects
	qd.selects[columnName] = c

	// Set the projection if required
	if doProject {
		if err := qd.projection.Add(columnName); err != nil {
			return err
		}
	}
	return nil
}

func (qd *queryDataStore) SelectsLen() int {
	return len(qd.selects)
}

func (qd *queryDataStore) Selects() iter.Seq2[string, mdl.ResolvedColumn] {
	return func(yield func(string, mdl.ResolvedColumn) bool) {
		for k, v := range qd.selects {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (qd *queryDataStore) AddExpand(relationshipId string) (QueryDataStore, error) {
	if qd.expands != nil {
		if _, exists := qd.expands[relationshipId]; exists {
			return nil, qerr.SyntaxErr(
				"an expand operations has already been declared for %s", relationshipId,
			)
		}
	}
	return qd.addExpand(relationshipId, true)
}

func (qd *queryDataStore) AddSystemExpand(relationshipId string) (QueryDataStore, error) {
	return qd.addExpand(relationshipId, false)
}

func (qd *queryDataStore) addExpand(relationshipId string, doProject bool) (QueryDataStore, error) {
	traversalStep, err := qd.metadataBinder.ResolveRelationship(qd.rootResource, relationshipId)
	if err != nil {
		return nil, err
	}

	if qd.expands == nil {
		qd.expands = make(map[string]Expansion, 4)
	}

	if qd.depth == math.MaxUint8 {
		return nil, qerr.InternalErr("unable to add expand as it will cause depth overflow")
	}

	expandedQuery, err := newQueryDataStore(
		qd.queryString,
		traversalStep.Relationship.To(),
		qd.accessPolicy,
		qd.valueBuilder,
		qd.depth+1,
	)
	if err != nil {
		return nil, err
	}

	expansion := Expansion{
		QueryData:     expandedQuery,
		TraversalStep: traversalStep,
	}

	if _, exists := qd.expands[relationshipId]; exists {
		return nil, qerr.InternalErr("an expand already exists for %s", relationshipId)
	}

	qd.expands[traversalStep.Relationship.Id()] = expansion

	if doProject {
		if err := qd.projection.Add(
			traversalStep.Relationship.ExpansionColumnName(),
		); err != nil {
			return nil, err
		}
	}

	qd.AddSystemSelect(traversalStep.Relationship.FromColumn().Name())
	expansion.QueryData.AddSystemSelect(traversalStep.Relationship.ToColumn().Name())

	return expansion.QueryData, nil
}

func (qd *queryDataStore) ExpandsLen() int {
	return len(qd.expands)
}

func (qd *queryDataStore) Expands() iter.Seq2[string, Expansion] {
	return func(yield func(string, Expansion) bool) {
		for k, v := range qd.expands {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (qd *queryDataStore) GetExpansionByRelationshipId(relationshipId string) (Expansion, bool) {
	if qd.expands == nil {
		return Expansion{}, false
	}

	if _, exists := qd.expands[relationshipId]; !exists {
		return Expansion{}, false
	}

	return qd.expands[relationshipId], true
}

func (qd *queryDataStore) SetFilterExpression(ex mdl.FilterExpression) {
	qd.filter = ex
}

func (qd *queryDataStore) FilterExpression() mdl.FilterExpression {
	return qd.filter
}

func (qd *queryDataStore) AddAssociatedWithParentFilter(
	linkFromParent mdl.TraversalStep,
	joinParentOnValues []mdl.ValueExpression,
) error {
	values, err := qd.valueBuilder.List(joinParentOnValues)
	if err != nil {
		return err
	}

	assocationFilter, err := qd.filterExpressionBuilder.NewComparisonExpression(
		linkFromParent.Relationship.ToColumn().Name(),
		mdl.ComparisonIn,
		values,
	)
	if err != nil {
		return err
	}

	qd.SetOrAppendFilter(assocationFilter)
	return nil
}

func (qd *queryDataStore) SetOrAppendFilter(ex mdl.FilterExpression) {
	if qd.filter == nil {
		qd.filter = ex
	} else {
		qd.filter = &mdl.LogicalExpression{
			Left:     qd.filter,
			Operator: mdl.LogicalAnd,
			Right:    ex,
		}
	}
}

func (qd *queryDataStore) GetFilter() mdl.FilterExpression {
	return qd.filter
}

func (qd *queryDataStore) AddOrderBy(columnPath string, dir mdl.SortDirectionOperator) error {
	// Get the metadata
	c, err := qd.metadataBinder.ResolveColumn(columnPath)
	if err != nil {
		return err
	}

	if err := qd.relationshipPlanner.ProcessJoin(c.ResolvedPath); err != nil {
		return err
	}

	// ensure order by field is created
	if qd.orderBy == nil {
		qd.orderBy = make([]mdl.SortingRule, 0, 1)
	}

	// add the rule
	qd.orderBy = append(
		qd.orderBy,
		mdl.SortingRule{
			ResolvedColumn: c,
			Direction:      dir,
		},
	)
	return nil
}

func (qd *queryDataStore) OrderByLen() int {
	return len(qd.orderBy)
}

func (qd *queryDataStore) OrderBy() iter.Seq[mdl.SortingRule] {
	return func(yield func(mdl.SortingRule) bool) {
		for _, rule := range qd.orderBy {
			if !yield(rule) {
				return
			}
		}
	}
}

func (qd *queryDataStore) SetDoCount(doCount bool) {
	qd.doCount = doCount
}

func (qd *queryDataStore) DoCount() bool {
	return qd.doCount
}

func (qd *queryDataStore) SetLimit(limit int) error {
	if limit < 0 || limit > math.MaxUint16 {
		return qerr.InternalErr("limit must be between 1 and %d", math.MaxUint16)
	}
	qd.limit = uint16(limit)
	return nil
}

func (qd *queryDataStore) Limit() uint16 {
	return qd.limit
}

func (qd *queryDataStore) SetPagingToken(pagingToken string) {
	qd.pagingToken = pagingToken
}

func (qd *queryDataStore) PagingToken() (string, bool) {
	if qd.pagingToken == "" {
		return "", false
	}
	return qd.pagingToken, true
}
