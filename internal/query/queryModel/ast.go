package queryModel

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

type LiteralType uint8

const (
	LiteralTypeNull LiteralType = iota
	LiteralTypeString
	LiteralTypeInt
	LiteralTypeFloat
	LiteralTypeStringList
	LiteralTypeIntList
	LiteralTypeFloatList
)

type QueryWriter interface {
	Write(statement string, args ...any)
	Placeholder(v any) string
}

type ValueExpression interface {
	GetTypeName() string
	GetType() LiteralType
	IsCompatibleWithComparisonOperator(op ComparisonOperator) bool
	IsSupportedByDbType(dbType DbDataTypeName) bool
	WriteFilterExpression(w QueryWriter, fieldName string, op ComparisonOperator) error
}

type ResolvedPath struct {
	Id            string
	StartResource TableMetadata
	Steps         []*TraversalStep
	EndResource   TableMetadata
	Type          RelationshipType
}

type TraversalStep struct {
	SubPathId    string
	Relationship RelationshipMetadata
}

// All columns, for select and order by will be represented by this type
type ResolvedColumn struct {
	ResolvedPath ResolvedPath
	Metadata     ColumnMetadata
}

type FilterExpression interface {
	IsFilterExpression()
}

type LogicalOperator string

const (
	LogicalAnd LogicalOperator = "and"
	LogicalOr  LogicalOperator = "or"
)

var SupportedLogicalOperators = map[string]LogicalOperator{
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
	ComparisonNotIn         ComparisonOperator = "notin"
	ComparisonNotContains   ComparisonOperator = "notcontains"
	ComparisonNotStartsWith ComparisonOperator = "notstartswith"
	ComparisonNotEndsWith   ComparisonOperator = "notendswith"
)

var SupportedComparisonOperators = map[string]ComparisonOperator{
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

var SupportedCollectionOperators = map[string]CollectionOperator{
	"any": CollectionAny,
	"all": CollectionAll,
}

type LogicalExpression struct {
	Left     FilterExpression `json:"left"`
	Operator LogicalOperator  `json:"operator"`
	Right    FilterExpression `json:"right"`
}

func (e *LogicalExpression) IsFilterExpression() {}

type ExistsNode interface {
	Step() TraversalStep
	Alias() string
	ParentAlias() string
	Next() ExistsNode
}

type ExistsNodeNew struct {
	Step        *TraversalStep
	Alias       string
	ParentAlias string
}

type ComparisonExpression struct {
	ResolvedColumn *ResolvedColumn
	Operator       ComparisonOperator
	Value          ValueExpression

	ExistsNodes []ExistsNodeNew
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

	ExistsNodes []ExistsNodeNew
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

type SortingRule struct {
	ResolvedColumn ResolvedColumn
	Direction      SortDirectionOperator
}
