package queryModel

import (
	"context"
	"time"
)

// DbType represents a SQL Server data type name supported by the model metadata system.
type DbType string

const (
	DbTypeNull     DbType = "null"
	DbTypeString   DbType = "string"
	DbTypeInt      DbType = "int"
	DbTypeFloat    DbType = "float"
	DbTypePoint    DbType = "point"
	DbTypeDateTime DbType = "date/time"
)

type ValueType uint8

const (
	ValueTypeNull ValueType = iota
	ValueTypeString
	ValueTypeInt
	ValueTypeFloat
	ValueTypePoint
	ValueTypeDateTime
	ValueTypeIntList
	ValueTypeStringList
	ValueTypeFloatList
)

type Serialiser interface {
	SerialiseNull()
	SerialiseString(str string)
	SerialiseInt(n int64)
	SerialiseFloat(f float64)

	SerialiseStringList(els []string)
	SerialiseIntList(els []int64)
	SerialiseFloatList(els []float64)
}

type Deserialiser interface {
	DeserialiseString(d []byte) string
	DeserialiseListString(d []byte) []string

	DeserialiseInt(d []byte) (int64, error)
	DeserialiseListInt(d []byte) ([]int64, error)

	DeserialiseFloat(d []byte) (float64, error)
	DeserialiseListFloat(d []byte) ([]float64, error)
}

type ValueBuilder interface {
	Null() Value
	String(str string) Value
	Int(n int64) Value
	Float(f float64) Value
	List(elements []Value) (Value, error)
	Point(longitude float64, latitude float64) Value
	DateTime(dt time.Time) Value

	ExecuteDeserialisation(
		ds Deserialiser,
		t ValueType,
		d []byte,
	) (Value, error)
}

type QueryWriter interface {
	WriteFilterExpressionNull(field string, op ComparisonOperator) error
	WriteFilterExpressionString(field string, op ComparisonOperator, value string) error
	WriteFilterExpressionInt(field string, op ComparisonOperator, value int64) error
	WriteFilterExpressionFloat(field string, op ComparisonOperator, value float64) error
	WriteFilterExpressionPoint(field string, op ComparisonOperator, value Point) error
	WriteFilterExpressionDateTime(field string, op ComparisonOperator, value time.Time) error
	WriteFilterExpressionStringList(field string, op ComparisonOperator, value []string) error
	WriteFilterExpressionIntList(field string, op ComparisonOperator, value []int64) error
	WriteFilterExpressionFloatList(field string, op ComparisonOperator, value []float64) error
	// Placeholder(v any) string

	WriteQueryStatement() string
	WriteCountStatement() string
	Args() []any
}

type Value interface {
	Type() ValueType
	SupportsType(DbType) bool
	SupportsOperator(op ComparisonOperator) bool

	Serialise(s Serialiser)
	WriteFilterExpression(
		writer QueryWriter,
		fieldName string,
		operator ComparisonOperator,
	) error
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
	Value          Value

	ExistsNodes []ExistsNodeNew
}

func (e *ComparisonExpression) IsFilterExpression() {}

func (e *ComparisonExpression) WithValues(operator ComparisonOperator, value Value) *ComparisonExpression {
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

type PayloadSigner interface {
	Sign(secret []byte, data []byte) string
	Verify(secret []byte, value string) (data []byte, ok bool)
}

type Repository interface {
	ExecuteJsonRequest(
		ctx context.Context,
		queryStatementStr string,
		args []any,
	) ([]byte, error)

	ExecuteJsonRequestWithCount(
		ctx context.Context,
		queryStatementStr string,
		countStatementStr string,
		sharedArgs []any,
	) ([]byte, *int64, error)
}
