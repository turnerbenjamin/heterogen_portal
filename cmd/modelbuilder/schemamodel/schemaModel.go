package schemamodel

import (
	"fmt"
	"slices"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderrepo"
)

type relationshipType string

const (
	RelationshipOneToMany relationshipType = "1:N"
	RelationshipManyToOne relationshipType = "N:1"
)

type columnType string

const (
	ColumnTypeString columnType = "string"
)

type Column struct {
	Name              string
	DisplayName       string
	Type              columnType
	IsExpansionColumn bool
}

type Relationship struct {
	Id                string
	Type              relationshipType
	ColumnName        string
	ExpansionColumn   *Column
	RelatedTableName  string
	RelatedColumnName string
}

type TableIdentifiers struct {
	ModelStruct      string
	Collection       string
	Resource         string
	Projection       string
	MetadataVar      string
	AccessPolicyName string
}

type ColumnData struct {
	// Map is a mapping from a columns database name to the column data
	Map map[string]*Column

	// Ordered is the collection of columns sorted in alphabetical order - This
	// should be used when writing the model to ensure deterministic output
	Ordered []*Column

	// PrimaryKey is the primary key field for the table
	PrimaryKey *Column

	// NameColumn is the column used as the display name/label for a given record
	// - This model requires that the display name is a string type
	NameColumn *Column
}

type RelationshipData struct {
	// Map is a mapping from a relationship column name to a relationship
	Map map[string]*Relationship

	// Ordered is the collection of relaitonships, sorted by the column name
	Ordered []*Relationship
}

type Table struct {
	// Name is the database name of the table
	Name string

	// FullyQualifiedName is the full database name of the table including the
	// schema
	FullyQualifiedName string

	// DisplayName is the display name of the table
	DisplayName string

	// Identifiers is the collection of identifiers used when building models and
	// data structures for the table
	Identifiers TableIdentifiers

	ColumnData    ColumnData
	Relationships RelationshipData
}

type TableData struct {
	Map     map[string]*Table
	Ordered []*Table
}

type Schema struct {
	TableData TableData
}

func BuildModel(metadata *builderrepo.DatabaseSchema) (*Schema, error) {
	schema := &Schema{
		TableData: TableData{
			Map:     make(map[string]*Table, len(metadata.Tables)),
			Ordered: make([]*Table, 0, len(metadata.Tables)),
		},
	}
	// Build Table
	for _, tableData := range metadata.Tables {
		table, err := buildTable(tableData)
		if err != nil {
			return nil, err
		}
		schema.TableData.Map[table.Name] = table
		schema.TableData.Ordered = append(schema.TableData.Ordered, table)
	}

	slices.SortFunc(schema.TableData.Ordered, func(a *Table, b *Table) int {
		return strings.Compare(a.Name, b.Name)
	})

	return schema, nil
}

func buildTable(tableData *builderrepo.TableMetadata) (*Table, error) {
	columnData, err := buildColumnData(tableData)
	if err != nil {
		return nil, err
	}

	err = validateColumnData(columnData)
	if err != nil {
		return nil, err
	}

	tableNamePascal := snakeToPascal(tableData.Name)
	return &Table{
		Name:               tableData.Name,
		FullyQualifiedName: fmt.Sprintf("%s.%s", tableData.Schema, tableData.Name),
		DisplayName:        snakeToDisplayName(tableData.Name),

		Identifiers: TableIdentifiers{
			ModelStruct:      fmt.Sprintf("%sModel", tableNamePascal),
			Collection:       fmt.Sprintf("%sModels", tableNamePascal),
			Projection:       fmt.Sprintf("%sProjection", tableNamePascal),
			MetadataVar:      fmt.Sprintf("%sMetadata", tableNamePascal),
			Resource:         fmt.Sprintf("%sResource", tableNamePascal),
			AccessPolicyName: fmt.Sprintf("%sAccessPolicy", tableNamePascal),
		},

		ColumnData: columnData,
	}, nil
}

func buildRelationshipData(
	schema *builderrepo.DatabaseSchema,
	table *builderrepo.TableMetadata,
) (RelationshipData, error) {
	data := RelationshipData{
		Map:     map[string]*Relationship{},
		Ordered: []*Relationship{},
	}

	for _, metadata := range schema.Relationships {
		// check table relates to the table
		if metadata.ParentTable != table.Name &&
			metadata.ReferencedTable != table.Name {
			continue
		}

		relType, err := getRelationshipType(table, metadata)
		if err != nil {
			return data, err
		}

		expansionColName, err := relationshipExpansionColumnName(relType, metadata)
		if err != nil {
			return data, err
		}

		expansionColumn := &Column{
			Name:              expansionColName,
			DisplayName:       snakeToDisplayName(expansionColName),
			IsExpansionColumn: true,
		}

		var relationship *Relationship = nil
		if relType == RelationshipManyToOne {
			relationship = &Relationship{
				Id:                relationshipId(metadata),
				Type:              relType,
				ColumnName:        metadata.ParentColumn,
				ExpansionColumn:   expansionColumn,
				RelatedTableName:  metadata.ReferencedTable,
				RelatedColumnName: metadata.ReferencedColumn,
			}
		} else {
			relationship = &Relationship{
				Id:                relationshipId(metadata),
				Type:              relType,
				ColumnName:        expansionColName,
				ExpansionColumn:   expansionColumn,
				RelatedTableName:  metadata.ParentTable,
				RelatedColumnName: metadata.ParentColumn,
			}
		}

		data.Map[relationship.ColumnName] = relationship
		data.Ordered = append(data.Ordered, relationship)

	}

	slices.SortFunc(data.Ordered, func(a *Relationship, b *Relationship) int {
		return strings.Compare(a.ColumnName, b.ColumnName)
	})

	return data, nil
}

func buildColumnData(tableData *builderrepo.TableMetadata) (ColumnData, error) {
	columnData := ColumnData{
		Map:     map[string]*Column{},
		Ordered: make([]*Column, 0, len(tableData.Columns)),
	}

	for _, metadata := range tableData.Columns {
		col := &Column{
			Name:        metadata.Name,
			DisplayName: snakeToDisplayName(metadata.Name),
		}

		columnData.Map[col.Name] = col
		columnData.Ordered = append(columnData.Ordered, col)

		if metadata.PrimaryKey == 1 {
			columnData.PrimaryKey = col
		}
	}

	slices.SortFunc(columnData.Ordered, func(a *Column, b *Column) int {
		return strings.Compare(a.Name, b.Name)
	})

	return columnData, nil
}

func validateColumnData(columnData ColumnData) error {
	if columnData.PrimaryKey == nil {
		return fmt.Errorf("invalid column data: unable to access the primary key")
	}

	return nil
}

func modelMetadataStoreIdentifier(tableName string) string {
	return fmt.Sprintf("%sMetadata", snakeToPascal(tableName))
}

func getModelResourceStructName(tableName string) string {
	return fmt.Sprintf("%sResource", snakeToPascal(tableName))
}

func getRelationshipType(
	tableData *builderrepo.TableMetadata,
	relationship *builderrepo.RelationshipMetadata,
) (relationshipType, error) {
	if relationship.ParentTable == tableData.Name {
		return RelationshipManyToOne, nil
	}

	if relationship.ReferencedTable == tableData.Name {
		return RelationshipOneToMany, nil
	}

	return RelationshipManyToOne, fmt.Errorf(
		"unable to build relationship type: table is not a party to the relationship",
	)
}

func relationshipExpansionColumnName(
	relationshipType relationshipType,
	relationship *builderrepo.RelationshipMetadata,
) (string, error) {
	switch relationshipType {
	case RelationshipManyToOne:
		col := relationship.ParentColumn
		if !strings.HasSuffix(col, "_id") {
			return "", fmt.Errorf("unable to build relationship: 'N' columns must have a _id suffix")
		}
		return strings.TrimSuffix(col, "_id"), nil

	case RelationshipOneToMany:
		return fmt.Sprintf(
			"%s_%s_%s",
			relationship.ParentTable,
			relationship.ReferencedTable,
			relationship.ParentColumn,
		), nil

	default:
		return "", fmt.Errorf("unable to build relationship: invalid relationship type")
	}
}

func relationshipId(relationship *builderrepo.RelationshipMetadata) string {
	return fmt.Sprintf(
		"%s_%s_%s_%s",
		relationship.ParentTable,
		relationship.ParentColumn,
		relationship.ReferencedTable,
		relationship.ReferencedColumn,
	)
}

func snakeToDisplayName(s string) string {
	parts := strings.Split(s, "_")

	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}

	return strings.Join(parts, " ")
}

func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")

	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}

	return strings.Join(parts, "")
}

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")

	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == 0 {
			parts[i] = strings.ToLower(part[:1]) + strings.ToLower(part[1:])
		} else {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	return strings.Join(parts, "")
}
