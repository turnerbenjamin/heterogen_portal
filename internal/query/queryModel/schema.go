package queryModel

import "iter"

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
		projection Projection,
	) ([]TableModel, error)

	// SetRelationshipField sets a given relationship field
	SetRelationshipField(relationshipId string, value TableModel) error

	// InitRelationshipField initialses 1:N relationship fields to empty arrays
	InitRelationshipField(relationshipId string) error

	// GetJoinOnValue returns the value of the relevant column for a given relationship
	GetJoinOnValue(relationshipId string) (string, error)

	// GetValueExpression returns a value expression for a given path
	GetValueExpression(path []*TraversalStep, columnName string, valueBuilder ValueBuilder) (ValueExpression, error)

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
	InitProjection() (Projection, error)
}

type Projection interface {
	Add(columnName string) error
	IsEmpty() bool
}

type ColumnMetadata interface {
	Name() string
	Type() DbDataTypeName
}

type RelationshipMetadata interface {
	Id() string
	ColumnName() string
	ExpansionColumnName() string
	From() TableMetadata
	To() TableMetadata
	FromColumn() ColumnMetadata
	ToColumn() ColumnMetadata
	Type() RelationshipType
}
