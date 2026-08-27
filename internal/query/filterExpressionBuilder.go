package query

type FilterExpression interface {
	IsFilterExpression()
}

type LogicalOperator string

const (
	LogicalAnd LogicalOperator = "and"
	LogicalOr  LogicalOperator = "or"
)

var supportedLogicalOperators = map[string]LogicalOperator{
	"and": LogicalAnd,
	"or":  LogicalOr,
}

type SortDirectionOperator string

const (
	SortDirectionAsc  SortDirectionOperator = "asc"
	SortDirectionDesc SortDirectionOperator = "desc"
)

type ComparisonOperator string

const (
	ComparisonEq         ComparisonOperator = "eq"
	ComparisonNe         ComparisonOperator = "ne"
	ComparisonGt         ComparisonOperator = "gt"
	ComparisonGe         ComparisonOperator = "ge"
	ComparisonLt         ComparisonOperator = "lt"
	ComparisonLe         ComparisonOperator = "le"
	ComparisonIn         ComparisonOperator = "in"
	ComparisonContains   ComparisonOperator = "contains"
	ComparisonStartsWith ComparisonOperator = "startswith"
	ComparisonEndsWith   ComparisonOperator = "endswith"

	// not supported - included for internal negation logic only:
	comparisonNotIn         ComparisonOperator = "notin"
	comparisonNotContains   ComparisonOperator = "notcontains"
	comparisonNotStartsWith ComparisonOperator = "notstartswith"
	comparisonNotEndsWith   ComparisonOperator = "notendswith"
)

var supportedComparisonOperators = map[string]ComparisonOperator{
	"eq":         ComparisonEq,
	"ne":         ComparisonNe,
	"gt":         ComparisonGt,
	"ge":         ComparisonGe,
	"lt":         ComparisonLt,
	"le":         ComparisonLe,
	"in":         ComparisonIn,
	"contains":   ComparisonContains,
	"startswith": ComparisonStartsWith,
	"endswith":   ComparisonEndsWith,
}

type CollectionOperator string

const (
	CollectionAny CollectionOperator = "any"
	CollectionAll CollectionOperator = "all"
)

var supportedCollectionOperators = map[string]CollectionOperator{
	"any": CollectionAny,
	"all": CollectionAll,
}

type LogicalExpression struct {
	Left     FilterExpression `json:"left"`
	Operator LogicalOperator  `json:"operator"`
	Right    FilterExpression `json:"right"`
}

func (e *LogicalExpression) IsFilterExpression() {}

type ComparisonExpression struct {
	ResolvedColumn *ResolvedColumn
	Operator       ComparisonOperator
	Value          ValueExpression

	ExistsPlan *Exists `json:"-"`
}

func (e *ComparisonExpression) IsFilterExpression() {}

func (e *ComparisonExpression) WithValues(operator ComparisonOperator, value ValueExpression) *ComparisonExpression {
	c := *e

	c.Operator = operator
	c.Value = value

	return &c
}

type CollectionExpression struct {
	ResolvedPath     *ResolvedPath
	Operator         CollectionOperator
	FilterExpression FilterExpression

	ExistsPlan *Exists
}

func (e *CollectionExpression) WithValues(operator CollectionOperator, expression FilterExpression) *CollectionExpression {
	c := *e

	c.Operator = operator
	c.FilterExpression = expression

	return &c
}

func (e *CollectionExpression) IsFilterExpression() {}

// TO BE REMOVED
type PropertyPath struct {
	Segments []string `json:"segments"`
}

type filterExpressionBuilder struct {
	rootResource        TableMetadata
	metadataBinder      *metadataBinderNew
	relationshipPlanner *RelationshipPlannerNew
}

func (eb filterExpressionBuilder) NewLogicalExpression(
	left FilterExpression,
	operator LogicalOperator,
	right FilterExpression,
) (*LogicalExpression, error) {
	if left == nil || right == nil {
		return nil, internalErr(
			"unable to build logical expression: left and right must not be nil",
		)
	}

	return &LogicalExpression{
		Left:     left,
		Operator: operator,
		Right:    right,
	}, nil
}

func (eb filterExpressionBuilder) NewComparisonExpression(
	columnPath string,
	operator ComparisonOperator,
	value ValueExpression,
) (*ComparisonExpression, error) {
	column, err := eb.metadataBinder.ResolveColumn(columnPath)
	if err != nil {
		return nil, err
	}

	if len(column.ResolvedPath.Steps) > 0 && column.ResolvedPath.Type != RelationshipManyToOne {
		return nil, bindingErr(
			"invalid value - '%s'. Comparison operations are only supported "+
				"for N:1 and 1:1 relationships. Please use a collection "+
				"operator",
			columnPath,
		)
	}

	exists, err := eb.relationshipPlanner.ProcessExists(column.ResolvedPath)
	if err != nil {
		return nil, err
	}

	return &ComparisonExpression{
		ResolvedColumn: &column,
		Operator:       operator,
		Value:          value,
		ExistsPlan:     exists,
	}, nil
}

func (eb filterExpressionBuilder) NewCollectionExpression(
	resourcePath string,
	operator CollectionOperator,
	filterExpression FilterExpression,
) (*CollectionExpression, error) {
	if filterExpression == nil {
		return nil, internalErr(
			"unable to build collection expression: filter expression cannot be nil",
		)
	}

	resolvedPath, err := eb.metadataBinder.ResolvePath(resourcePath)
	if err != nil {
		return nil, internalErr("unable to build collection expression path: %w", err)
	}

	if resolvedPath.Type != RelationshipOneToMany {
		return nil, bindingErr(
			"invalid value - '%s'. Collection operations are only supported "+
				"for 1:N relationships",
			resolvedPath.Id,
		)
	}

	existsPlan, err := eb.relationshipPlanner.ProcessExists(resolvedPath)
	if err != nil {
		return nil, internalErr("unable to build collection expression exists plan: %w", err)
	}

	return &CollectionExpression{
		ResolvedPath:     &resolvedPath,
		Operator:         operator,
		FilterExpression: filterExpression,
		ExistsPlan:       existsPlan,
	}, nil
}
