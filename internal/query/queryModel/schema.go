// Package queryModel contains models used throughout the query package and by
// consuming packages
//
// This file contains the interfaces that allow the query package to work with
// the database schema without reflection
package queryModel

// relationshipType represents different table relationships
type RelationshipType string

const (
	// RelationshipOneToMany represents a 1:N relationship
	RelationshipOneToMany RelationshipType = "1:N"

	// RelationshipManyToOne represents an N:1 relationship
	RelationshipManyToOne RelationshipType = "N:1"
)

// Schema provides access to metadata for tables in a schema.
type Schema interface {

	// GetTableMetadata returns the metadata for the specified table and whether
	// the table was found.
	GetResource(tableName string) (Resource, bool)
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

type ColumnMetadata struct {
	Name string
	Type DbType
}

type RelationshipMetadata struct {
	Id                  string
	Type                RelationshipType
	ColumnName          string
	ExpansionColumnName string
	From                Resource
	To                  Resource
	FromColumn          ColumnMetadata
	ToColumn            ColumnMetadata
}

type TableMetadata struct {
	Name               string
	FullyQualifiedName string
	PrimaryKeyColumn   ColumnMetadata
	Columns            map[string]ColumnMetadata
	Relationships      map[string]RelationshipMetadata
}

type Resource interface {
	GetMetadata() TableMetadata
	InitModel() TableModel
	InitProjection() Projection
}

type Projection interface {
	Add(columnName string) error
	IsEmpty() bool
}

type ProjectionNode struct {
	Projection Projection
	Children   map[string]*ProjectionNode
}
