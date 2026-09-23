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
		projectionNode *ProjectionNode,
	) ([]TableModel, error)

	// GetValueExpression returns a value expression for a given path
	GetValue(path []*TraversalStep, columnName string, valueBuilder ValueBuilder) (Value, error)
}

type TableMetadata interface {
	GetColumnMetadata(columnName string) ColumnMetadata
	GetRelationshipMetadata(columnName string) RelationshipMetadata
	GetModel() TableModel
	Name() string
	FullyQualifiedName() string
	Columns() iter.Seq[ColumnMetadata]
	PrimaryKeyField() ColumnMetadata
	InitProjection() (Projection, error)
}

type Projection interface {
	Add(columnName string) error
	IsEmpty() bool
}

type ProjectionNode struct {
	Projection Projection
	Children   map[string]*ProjectionNode
}

type ColumnMetadata interface {
	Name() string
	Type() DbType
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
