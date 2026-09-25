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

type ColumnData struct {
	Name string
	Type DbType
}

type RelationshipData struct {
	Id                  string
	Type                RelationshipType
	ColumnName          string
	ExpansionColumnName string
	From                Resource
	To                  Resource
	FromColumn          ColumnData
	ToColumn            ColumnData
}

type TableData struct {
	Name               string
	FullyQualifiedName string
	PrimaryKeyColumn   ColumnData
	Columns            map[string]ColumnData
	Relationships      map[string]RelationshipData
}

type Resource interface {
	GetMetadata() TableData
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
