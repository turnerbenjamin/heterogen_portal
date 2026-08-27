package query

import (
	"iter"
	"math"
)

type FilterExpressionBuilder interface {
	NewLogicalExpression(
		left FilterExpression,
		operator LogicalOperator,
		right FilterExpression,
	) (*LogicalExpression, error)

	NewComparisonExpression(
		columnPath string,
		operator ComparisonOperator,
		value ValueExpression,
	) (*ComparisonExpression, error)

	NewCollectionExpression(
		resourcePath string,
		operator CollectionOperator,
		filterExpression FilterExpression,
	) (*CollectionExpression, error)
}

type QueryDataStore interface {
	RootResource() TableMetadata
	RootResourceAccessPolicy() TableAccessPolicy

	QueryString() string
	FilterExpressionBuilder() FilterExpressionBuilder
	IsTopLevelQuery() bool

	JoinCollection() JoinCollection
	RootAlias() string
	GetPathAlias(pathId string) (string, bool)

	Projection() []string // FOR NOW

	AddSelect(columnName string) error
	AddSystemSelect(columnName string) error
	SelectsLen() int
	Selects() iter.Seq2[string, ResolvedColumn]

	AddExpand(relationshipId string) (QueryDataStore, error)
	AddSystemExpand(relationshipId string) (QueryDataStore, error)
	ExpandsLen() int
	Expands() iter.Seq2[string, Expansion]
	GetExpansionByRelationshipId(relationshipId string) (Expansion, bool)

	SetFilterExpression(ex FilterExpression)
	FilterExpression() FilterExpression
	AddAssociatedWithParentFilter(
		linkFromParent TraversalStep,
		joinParentOnValues []string,
	) error // FOR NOW

	SetLimit(limit int) error
	Limit() uint16

	SetDoCount(doCount bool)
	DoCount() bool

	SetPagingToken(string)
	PagingToken() (string, bool)

	AddOrderBy(columnPath string, dir SortDirectionOperator) error
	OrderByLen() int
	OrderBy() iter.Seq[SortingRule]
}

type Expansion struct {
	queryData     QueryDataStore
	traversalStep TraversalStep
}

// queryDataStore is the primary store for all query data. When any operation is
// added, paths are immediately bound to metadata and the relationship plan is
// updated as required. This provides a guarantee to consumers that at any point
// in the process, all metadata is set and the relationship plan is current.
type queryDataStore struct {
	// overview data
	queryString  string
	rootResource TableMetadata
	depth        uint8
	projection   []string // change to actual projection type

	// metadata and relationship binding deps
	metadataBinder          *metadataBinderNew
	relationshipPlanner     *RelationshipPlannerNew
	filterExpressionBuilder filterExpressionBuilder

	// access policies
	accessPolicy             AccessPolicy
	rootResourceAccessPolicy TableAccessPolicy

	// operations
	selects     map[string]ResolvedColumn
	expands     map[string]Expansion
	filter      FilterExpression
	orderBy     []SortingRule
	doCount     bool
	limit       uint16
	pagingToken string
}

func NewQueryDataStore(
	queryString string,
	rootResource TableMetadata,
	accessPolicy AccessPolicy,
) (QueryDataStore, error) {
	depth := uint8(0)
	return newQueryDataStore(queryString, rootResource, accessPolicy, depth)
}

func newQueryDataStore(
	queryString string,
	rootResource TableMetadata,
	accessPolicy AccessPolicy,
	depth uint8,
) (QueryDataStore, error) {
	metadataBinder, err := NewMetadataBinder(rootResource, accessPolicy)
	if err != nil {
		return nil, err
	}

	relationshipPlanner, err := NewRelationshipPlanner(rootResource)
	if err != nil {
		return nil, err
	}

	rootResourceAccessPolicy := accessPolicy.GetTableAccessPolicy(rootResource.Name())
	if rootResourceAccessPolicy == nil {
		return nil, internalErr(
			"unable to create new data store: root resource access policy is nil",
		)
	}

	return &queryDataStore{
		depth:               depth,
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
		},
	}, nil
}

func (qd *queryDataStore) Projection() []string {
	return qd.projection
}

func (qd *queryDataStore) IsTopLevelQuery() bool {
	return qd.depth == 0
}

func (qd *queryDataStore) FilterExpressionBuilder() FilterExpressionBuilder {
	return qd.filterExpressionBuilder
}

func (qd *queryDataStore) RootResource() TableMetadata {
	return qd.rootResource
}

func (qd *queryDataStore) RootResourceAccessPolicy() TableAccessPolicy {
	return qd.rootResourceAccessPolicy
}

func (qd *queryDataStore) QueryString() string {
	return qd.queryString
}

func (qd *queryDataStore) GetPathAlias(pathId string) (string, bool) {
	return qd.relationshipPlanner.Aliases.GetAlias(pathId)
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
		qd.selects = make(map[string]ResolvedColumn, 1)
	}

	// Set the selects
	qd.selects[columnName] = c

	// Set the projection if required
	if doProject {
		qd.projection = append(qd.projection, columnName)
	}
	return nil
}

func (qd *queryDataStore) SelectsLen() int {
	return len(qd.selects)
}

func (qd *queryDataStore) Selects() iter.Seq2[string, ResolvedColumn] {
	return func(yield func(string, ResolvedColumn) bool) {
		for k, v := range qd.selects {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (qd *queryDataStore) JoinCollection() JoinCollection {
	return qd.relationshipPlanner.JoinStore
}

func (qd *queryDataStore) RootAlias() string {
	return qd.relationshipPlanner.rootAlias
}

func (qd *queryDataStore) AddExpand(relationshipId string) (QueryDataStore, error) {
	if qd.expands != nil {
		if _, exists := qd.expands[relationshipId]; exists {
			return nil, syntaxErr(
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
	traversalStep, err := qd.metadataBinder.resolveRelationship(qd.rootResource, relationshipId)
	if err != nil {
		return nil, err
	}

	if qd.expands == nil {
		qd.expands = make(map[string]Expansion, 4)
	}

	if qd.depth == math.MaxUint8 {
		return nil, internalErr("unable to add expand as it will cause depth overflow")
	}

	expandedQuery, err := newQueryDataStore(
		qd.queryString,
		traversalStep.Relationship.To(),
		qd.accessPolicy,
		qd.depth+1,
	)
	if err != nil {
		return nil, err
	}

	expansion := Expansion{
		queryData:     expandedQuery,
		traversalStep: traversalStep,
	}

	if _, exists := qd.expands[relationshipId]; exists {
		return nil, internalErr("an expand already exists for %s", relationshipId)
	}

	qd.expands[relationshipId] = expansion

	if doProject {
		qd.projection = append(
			qd.projection,
			traversalStep.Relationship.RelationshipColumn(),
		)
	}

	qd.AddSystemSelect(traversalStep.Relationship.FromColumn().Name())
	expansion.queryData.AddSystemSelect(traversalStep.Relationship.ToColumn().Name())

	return expansion.queryData, nil
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

func (qd *queryDataStore) SetFilterExpression(ex FilterExpression) {
	qd.filter = ex
}

func (qd *queryDataStore) FilterExpression() FilterExpression {
	return qd.filter
}

func (qd *queryDataStore) AddAssociatedWithParentFilter(
	linkFromParent TraversalStep,
	joinParentOnValues []string,
) error {
	assocationFilter, err := qd.filterExpressionBuilder.NewComparisonExpression(
		linkFromParent.Relationship.ToColumn().Name(),
		ComparisonIn,
		&StringListLiteral{Values: joinParentOnValues},
	)
	if err != nil {
		return err
	}

	qd.SetOrAppendFilter(assocationFilter)
	return nil
}

func (qd *queryDataStore) SetOrAppendFilter(ex FilterExpression) {
	if qd.filter == nil {
		qd.filter = ex
	} else {
		qd.filter = &LogicalExpression{
			Left:     qd.filter,
			Operator: LogicalAnd,
			Right:    ex,
		}
	}
}

func (qd *queryDataStore) GetFilter() FilterExpression {
	return qd.filter
}

func (qd *queryDataStore) AddOrderBy(columnPath string, dir SortDirectionOperator) error {
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
		qd.orderBy = make([]SortingRule, 0, 1)
	}

	// add the rule
	qd.orderBy = append(
		qd.orderBy,
		SortingRule{
			ResolvedColumn: c,
			Direction:      dir,
		},
	)
	return nil
}

func (qd *queryDataStore) OrderByLen() int {
	return len(qd.orderBy)
}

func (qd *queryDataStore) OrderBy() iter.Seq[SortingRule] {
	return func(yield func(SortingRule) bool) {
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
		return internalErr("limit must be between 1 and %d", math.MaxUint16)
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
