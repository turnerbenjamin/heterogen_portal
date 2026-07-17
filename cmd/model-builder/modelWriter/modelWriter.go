package modelWriter

import (
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/model-builder/builderRepo"
)

const (
	MsqlTypeNvarchar       builderRepo.MsqlDataTypeName = "nvarchar"
	MsqlTypeInt            builderRepo.MsqlDataTypeName = "int"
	MsqlTypeGeography      builderRepo.MsqlDataTypeName = "geography"
	MsqlTypeDateTimeOffset builderRepo.MsqlDataTypeName = "datetimeoffset"
)

type relationshipType string

const (
	RelationshipOneToMany relationshipType = "1:N"
	RelationshipManyToOne relationshipType = "N:1"
)

type modelWriter struct {
	path     string
	metadata *builderRepo.DatabaseSchema
	sb       *strings.Builder
}

func NewModelWriter(
	path string,
	metadata *builderRepo.DatabaseSchema,
) *modelWriter {
	return &modelWriter{
		path:     path,
		metadata: metadata,
		sb:       &strings.Builder{},
	}
}

func (w *modelWriter) Write() {
	buildTableRelationshipData(w.metadata)

	w.writePackageAndStaticTypeDefinitions()
	w.writeNewLine()
	w.WriteConstants()

	for _, table := range w.metadata.Tables {
		w.writeNewLine()
		w.writeTableMetadataDefinition(table)
		w.writeNewLine()
		w.WriteTableModel(table)
	}
	w.writeNewLine()
	w.WriteTableModelGetter()

	raw := w.sb.String()
	formatted, err := format.Source([]byte(raw))
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(w.path, formatted, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func writeToBuilder(sb *strings.Builder, s string) {
	_, err := sb.Write([]byte(s))
	if err != nil {
		log.Fatal(err)
	}
}

func (w *modelWriter) writeNewLine() {
	writeToBuilder(w.sb, "\n")
}

func (w *modelWriter) writePackageAndStaticTypeDefinitions() {
	writeToBuilder(w.sb, `//GENERATED CODE
// see cmd/model-builder/
package model

import (
	"time"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"errors"
)

// dbDataTypeName represents a SQL Server data type name supported by the model metadata system.
type dbDataTypeName string

const (
	// DbTypeNvarchar represents the SQL Server nvarchar data type.
	DbTypeNvarchar dbDataTypeName = "nvarchar"

	// DbTypeInt represents the SQL Server int data type.
	DbTypeInt dbDataTypeName = "int"

	// DbTypeGeography represents the SQL Server geography spatial data type.
	DbTypeGeography dbDataTypeName = "geography"

	// DbTypeDateTimeOffset represents the SQL Server datetimeoffset date/time data type.
	DbTypeDateTimeOffset dbDataTypeName = "datetimeoffset"
)

// relationshipId represents a unique identifier for a given relationship
type relationshipId string

// relationshipType represents different table relationships
type relationshipType string
const (
	// RelationshipOneToMany represents a 1:N relationship
	RelationshipOneToMany relationshipType = "1:N"

	// RelationshipManyToOne represents an N:1 relationship
	RelationshipManyToOne relationshipType = "N:1"
)

// SupportedDbTypes contains the SQL Server data types supported by the model
// generation and mapping system.
var SupportedDbTypes = map[dbDataTypeName]bool{
	DbTypeNvarchar:  true,
	DbTypeInt:       true,
	DbTypeGeography: true,
	DbTypeDateTimeOffset: true,
}

// ColumnMetadata describes the metadata associated with a database table column.
type ColumnMetadata struct {
	Name         string
	Type         dbDataTypeName
	MaxLength    int
	IsPrimaryKey bool
	IsRequired   bool
}

// Relationship describes a foreign key relationship between two database columns.
type Relationship struct {
	Id					relationshipId
	Type		  		relationshipType
	RelationshipColumn 	string
	RelatedTable  		string
	LocalColumn   		string
	ForeignColumn 		string
}

// TableMetadata describes the structure and relationships of a database table.
type TableMetadata struct {
	TableName          	string
	SchemaName          string
	PrimaryKey    		string
	Columns       		map[string]ColumnMetadata
	Relationships 		map[string]Relationship
}

// TableModel is an interface shared by all database table models
type TableModel interface {
	// GetMetadata returns metadata for the table
	GetMetadata() TableMetadata

	// GetTableName returns the name of the table in the database
	GetTableName() string

	// NewSlice unmarshals a json array and returns it as a slice
	NewSlice(jsonData []byte) ([]TableModel, error)

	// SetRelationshipField sets a given relationship field
	SetRelationshipField(relationshipId relationshipId, value TableModel) error

	// GetJoinOnValue returns the value of the relevant column for a given relationship
	GetJoinOnValue(relationshipId relationshipId) (string, error)
}

// Point represents a geographic point with a GeoJSON-compatible structure.
type Point struct {
    Type string `+"`json:\"type\"`"+`
    Coordinates [2]float64 `+"`json:\"coordinates\"`"+`
}

// UnmarshalJSON parses a point from WKT point data.
func (p *Point) UnmarshalJSON(data []byte) error {
    strData := strings.Trim(string(data), "\"")
    if strData == "null" || strData == "" {
        return nil
    }

    if !strings.HasPrefix(strData, "POINT") {
        return fmt.Errorf("invalid json for point: %s", strData)
    }

    start := strings.Index(strData, "(")
    end := strings.Index(strData, ")")
    if start == -1 || end == -1 || end <= start {
        return fmt.Errorf("invalid WKT point format: %s", strData)
    }

    coordsStr := strings.TrimSpace(strData[start+1:end])
    parts := strings.Fields(coordsStr)
    if len(parts) != 2 {
        return fmt.Errorf("expected 2 coordinates in WKT point, got %d", len(parts))
    }

    long, err := strconv.ParseFloat(parts[0], 64)
    if err != nil {
        return fmt.Errorf("invalid longitude value: %w", err)
    }

    lat, err := strconv.ParseFloat(parts[1], 64)
    if err != nil {
        return fmt.Errorf("invalid latitude value: %w", err)
    }

    p.Type = "Point"
    p.Coordinates = [2]float64{long, lat}
    return nil
}

`)
}

func (w *modelWriter) WriteConstants() {

	writeToBuilder(w.sb, "// Constants representing field names in the database\n")
	writeToBuilder(w.sb, "const(\n")
	for _, table := range w.metadata.Tables {
		for _, column := range table.Columns {
			fieldNameConst := fmt.Sprintf(
				"Col%s%s",
				snakeToPascal(table.Name),
				snakeToPascal(column.Name),
			)
			writeToBuilder(w.sb, fmt.Sprintf("%s = \"%s\"\n", fieldNameConst, column.Name))
		}
	}
	writeToBuilder(w.sb, ")\n\n")

	writeToBuilder(w.sb, "// Constants representing field names in the database\n")
	writeToBuilder(w.sb, "const(\n")
	for _, table := range w.metadata.Tables {
		for _, column := range table.Columns {
			if column.Type != MsqlTypeNvarchar {
				continue
			}
			constraintConst := fmt.Sprintf(
				"MaxLen%s%s",
				snakeToPascal(table.Name),
				snakeToPascal(column.Name),
			)
			writeToBuilder(w.sb, fmt.Sprintf("%s = %d\n", constraintConst, column.MaxLength))
		}
	}

	writeToBuilder(w.sb, ")")
}

func (w *modelWriter) writeTableMetadataDefinition(tableData *builderRepo.TableMetadata) {
	primaryKey, columnMapValue := buildColumnMap(tableData)
	if primaryKey == "" {
		log.Fatal(fmt.Errorf("unable to access primary key for table %s", tableData.Name))
	}

	relationshipMapValue := w.buildRelationshipMap(tableData)
	metadataIdentifer := modelMetadataStoreIdentifier(tableData)

	writeToBuilder(w.sb, fmt.Sprintf(
		"// %s contains the database metadata for the %s.%s table.\n",
		metadataIdentifer,
		tableData.Schema,
		tableData.Name,
	))
	writeToBuilder(w.sb, fmt.Sprintf("var %s = TableMetadata{\n", modelMetadataStoreIdentifier(tableData)))
	writeToBuilder(w.sb, fmt.Sprintf("TableName: \"%s\",\n", tableData.Name))
	writeToBuilder(w.sb, fmt.Sprintf("SchemaName: \"%s\",\n", tableData.Schema))
	writeToBuilder(w.sb, fmt.Sprintf("PrimaryKey: \"%s\",\n", primaryKey))
	writeToBuilder(w.sb, fmt.Sprintf("Columns: %s,\n", columnMapValue))
	writeToBuilder(w.sb, fmt.Sprintf("Relationships: %s,\n", relationshipMapValue))
	writeToBuilder(w.sb, "}\n\n")

	tableNamePascal := snakeToPascal(tableData.Name)
	writeToBuilder(w.sb, fmt.Sprintf(
		"// Get%sMetadata returns the database metadata associated with the %s.%s table.\n",
		tableNamePascal,
		tableData.Name,
		tableNamePascal,
	))
	writeToBuilder(
		w.sb,
		fmt.Sprintf("func Get%sMetadata() TableMetadata {\n", tableNamePascal),
	)

	metadataIdentifier := modelMetadataStoreIdentifier(tableData)
	writeToBuilder(w.sb, fmt.Sprintf("return %s\n", metadataIdentifier))
	writeToBuilder(w.sb, "}")
}

func buildColumnMap(tableData *builderRepo.TableMetadata) (string, string) {
	primaryKey := ""
	sbColMap := &strings.Builder{}

	writeToBuilder(sbColMap, "map[string]ColumnMetadata{\n")

	// DB Columns
	for _, col := range tableData.Columns {
		colType := getColTypeString(col.Type)
		if col.PrimaryKey == 1 {
			primaryKey = col.Name
		}

		writeToBuilder(sbColMap, fmt.Sprintf("\"%s\": {\n", col.Name))
		writeToBuilder(sbColMap, fmt.Sprintf("Name: \"%s\",\n", col.Name))
		writeToBuilder(sbColMap, fmt.Sprintf("Type: %s,\n", colType))
		writeToBuilder(sbColMap, fmt.Sprintf("MaxLength: %d,\n", col.MaxLength))
		writeToBuilder(sbColMap, fmt.Sprintf("IsRequired: %t,\n", !col.IsNullable))
		writeToBuilder(sbColMap, fmt.Sprintf("IsPrimaryKey: %t,\n", col.PrimaryKey == 1))
		writeToBuilder(sbColMap, "},\n")
	}

	writeToBuilder(sbColMap, "}")

	return primaryKey, sbColMap.String()
}

func (w *modelWriter) buildRelationshipMap(tableData *builderRepo.TableMetadata) string {
	sbR := &strings.Builder{}

	writeToBuilder(sbR, "map[string]Relationship{\n")

	for key, relationship := range tableData.Relationships {
		writeToBuilder(sbR, fmt.Sprintf("\"%s\": {\n", key))
		writeToBuilder(sbR, fmt.Sprintf("Id: relationshipId(\"%s\"),\n", relationship.Id))
		writeToBuilder(sbR, fmt.Sprintf("Type: %s,\n", relationship.Type))
		writeToBuilder(sbR, fmt.Sprintf("RelationshipColumn: \"%s\",\n", relationship.RelationshipColumn))
		writeToBuilder(sbR, fmt.Sprintf("RelatedTable: \"%s\",\n", relationship.RelatedTable))
		writeToBuilder(sbR, fmt.Sprintf("LocalColumn: \"%s\",\n", relationship.LocalColumn))
		writeToBuilder(sbR, fmt.Sprintf("ForeignColumn: \"%s\",\n", relationship.ForeignColumn))
		writeToBuilder(sbR, "},\n")
	}

	writeToBuilder(sbR, "}")
	return sbR.String()
}

func getColTypeString(typeName builderRepo.MsqlDataTypeName) string {
	switch typeName {
	case MsqlTypeNvarchar:
		return "DbTypeNvarchar"
	case MsqlTypeInt:
		return "DbTypeInt"
	case MsqlTypeDateTimeOffset:
		return "DbTypeDateTimeOffset"
	case MsqlTypeGeography:
		return "DbTypeGeography"
	default:
		return ""
	}
}

func (w *modelWriter) WriteTableModel(tableData *builderRepo.TableMetadata) {
	structName := getModelStructName(tableData.Name)

	writeToBuilder(w.sb, fmt.Sprintf(
		"// %s represents a row from the %s.%s table.\n",
		structName,
		tableData.Schema,
		tableData.Name,
	))
	writeToBuilder(w.sb, fmt.Sprintf("type %s struct {\n", structName))

	// DB Cols
	for _, col := range tableData.Columns {
		identifier := snakeToPascal(col.Name)

		goType := msqlTypeToGoType(col.Type)
		if goType == "" {
			log.Fatal(fmt.Errorf("unable to convert type to go type: %s", col.Type))
		}

		tag := fmt.Sprintf("`json:\"%s,omitempty\"`", col.Name)
		writeToBuilder(w.sb, fmt.Sprintf("%s %s %s\n", identifier, goType, tag))
	}

	// Relationship Cols
	for _, tableRelationship := range tableData.Relationships {
		tag := fmt.Sprintf("`json:\"%s,omitempty\"`", tableRelationship.RelationshipColumn)
		identifer := snakeToPascal(tableRelationship.RelationshipColumn)
		writeToBuilder(w.sb, fmt.Sprintf("%s %s %s\n", identifer, tableRelationship.RelationshipColumnType, tag))
	}

	writeToBuilder(w.sb, "}\n\n")
	w.WriteTableMetadataGetterFunc(structName, tableData)
	w.writeNewLine()
	w.WriteTableNameGetterFunc(structName, tableData)
	w.writeNewLine()
	w.WriteGetSliceGetterFunction(structName, tableData)
	w.writeNewLine()
	w.WriteRelationshipColumnSetterFunction(structName, tableData)
	w.writeNewLine()
	w.WriteGetJoinOnValueFunction(structName, tableData)
}

func (w *modelWriter) WriteTableMetadataGetterFunc(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, fmt.Sprintf(
		"// GetMetadata returns metadata for the %s.%s table.\n",
		tableData.Schema,
		tableData.Name,
	))
	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) GetMetadata() TableMetadata{\n", modelStructName))
	storeId := modelMetadataStoreIdentifier(tableData)
	writeToBuilder(w.sb, fmt.Sprintf("return %s", storeId))
	writeToBuilder(w.sb, "}")
}

func (w *modelWriter) WriteTableNameGetterFunc(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, fmt.Sprintf(
		"// GetTableName returns the name of the %s.%s table.\n",
		tableData.Schema,
		tableData.Name,
	))
	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) GetTableName() string{\n", modelStructName))
	storeId := modelMetadataStoreIdentifier(tableData)
	writeToBuilder(w.sb, fmt.Sprintf("return %s.TableName", storeId))
	writeToBuilder(w.sb, "}")
}

func (w *modelWriter) WriteGetSliceGetterFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, fmt.Sprintf(
		"// NewSlice unmarshals a json array of %s and returns it as a slice\n",
		tableData.Name,
	))

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) NewSlice(jsonData []byte) ([]TableModel, error) {\n", modelStructName))
	writeToBuilder(w.sb, fmt.Sprintf("var concreteSlice []*%s\n", modelStructName))

	writeToBuilder(w.sb, "if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {\n")
	writeToBuilder(w.sb, "return nil, err\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "result := make([]TableModel, len(concreteSlice))\n")
	writeToBuilder(w.sb, "for i := range concreteSlice {\n")
	writeToBuilder(w.sb, "result[i] = concreteSlice[i]\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "return result, nil\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteRelationshipColumnSetterFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, fmt.Sprintf(
		"// SetRelationshipField sets a given relationship field on the %s.%s table\n",
		tableData.Schema,
		tableData.Name,
	))

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) SetRelationshipField(relationshipId relationshipId, value TableModel) error {\n", modelStructName))
	writeToBuilder(w.sb, "switch relationshipId {\n")

	for _, relationship := range tableData.Relationships {
		relationshipColumnIdentifier := snakeToPascal(relationship.RelationshipColumn)
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", relationship.Id))
		writeToBuilder(w.sb, fmt.Sprintf("v, ok := value.(*%s)\n", getModelStructName(relationship.RelatedTable)))
		writeToBuilder(w.sb, "if !ok {\n")
		writeToBuilder(w.sb, "return errors.New(\"unexpected relationship type received\")\n")
		writeToBuilder(w.sb, "}\n")

		if relationship.Type == "RelationshipManyToOne" {
			writeToBuilder(w.sb, fmt.Sprintf("m.%s = v\n\n", relationshipColumnIdentifier))
		}

		if relationship.Type == "RelationshipOneToMany" {
			writeToBuilder(w.sb, fmt.Sprintf("m.%s = append(m.%s, v)\n\n", relationshipColumnIdentifier, relationshipColumnIdentifier))
		}
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return fmt.Errorf(\"unknown relationship: %s\", relationshipId)\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "return nil\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteGetJoinOnValueFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, "// GetJoinOnValue returns the value of the relevant column for a given relationship\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) GetJoinOnValue(relationshipId relationshipId) (string, error){\n", modelStructName))
	writeToBuilder(w.sb, "switch relationshipId {\n")

	for _, relationship := range w.metadata.Relationships {
		if relationship.ParentTable != tableData.Name && relationship.ReferencedTable != tableData.Name {
			continue
		}

		relationshipId := getTableRelationshipIdentifier(relationship)

		joinOnColumn := relationship.ParentColumn
		if relationship.ReferencedTable == tableData.Name {
			joinOnColumn = relationship.ReferencedColumn
		}

		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", relationshipId))
		writeToBuilder(w.sb, fmt.Sprintf("return m.%s, nil\n", snakeToPascal(joinOnColumn)))
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return \"\", fmt.Errorf(\"unknown relationship: %s\", relationshipId)\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteTableModelGetter() {
	writeToBuilder(
		w.sb,
		"// GetTableModel returns a new instance of a given table or nill if \n"+
			"// the table name is invalid\n",
	)
	writeToBuilder(w.sb, "func GetTableModel(tableName string) TableModel{\n")
	writeToBuilder(w.sb, "switch tableName {\n")

	for _, table := range w.metadata.Tables {
		metadataStoreId := modelMetadataStoreIdentifier(table)
		writeToBuilder(w.sb, fmt.Sprintf("case %s.TableName:\n", metadataStoreId))
		writeToBuilder(w.sb, fmt.Sprintf("return new(%s)\n", getModelStructName(table.Name)))
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "}\n")
}

func getModelStructName(tableName string) string {
	return fmt.Sprintf("%sModel", snakeToPascal(tableName))
}

func modelMetadataStoreIdentifier(tableData *builderRepo.TableMetadata) string {
	return fmt.Sprintf("%sMetadata", snakeToCamel(tableData.Name))
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

func msqlTypeToGoType(typeName builderRepo.MsqlDataTypeName) string {
	switch typeName {
	case MsqlTypeNvarchar:
		return "string"
	case MsqlTypeInt:
		return "int"
	case MsqlTypeDateTimeOffset:
		return "*time.Time"
	case MsqlTypeGeography:
		return "*Point"
	default:
		return ""
	}
}

func buildTableRelationshipData(schema *builderRepo.DatabaseSchema) {
	for _, table := range schema.Tables {
		for _, relationship := range schema.Relationships {
			tableRelationship := getTableRelationship(
				table,
				relationship,
			)

			if tableRelationship == nil {
				continue
			}

			key := tableRelationship.RelationshipColumn
			if relationship.ParentTable == table.Name {
				key = tableRelationship.LocalColumn
			}

			if table.Relationships == nil {
				table.Relationships = map[string]*builderRepo.TableRelationship{}
			}
			table.Relationships[key] = tableRelationship
		}
	}
}

func getTableRelationship(
	table *builderRepo.TableMetadata,
	relationship *builderRepo.RelationshipMetadata,
) *builderRepo.TableRelationship {

	if relationship.ParentTable != table.Name && relationship.ReferencedTable != table.Name {
		return nil
	}

	r := &builderRepo.TableRelationship{
		Id: getTableRelationshipIdentifier(relationship),
	}

	if relationship.ParentTable == table.Name {
		if !strings.HasSuffix(relationship.ParentColumn, "_id") {
			log.Fatalf("invalid parent column: %s. Parent columns must be suffixes with _id", relationship.ParentColumn)
		}
		r.RelationshipColumn = strings.TrimSuffix(relationship.ParentColumn, "_id")

		r.Type = "RelationshipManyToOne"
		r.RelatedTable = relationship.ReferencedTable
		r.LocalColumn = relationship.ParentColumn
		r.ForeignColumn = relationship.ReferencedColumn

		r.RelationshipColumnType = fmt.Sprintf("*%s", getModelStructName(r.RelatedTable))
	}

	if relationship.ReferencedTable == table.Name {
		r.RelationshipColumn = fmt.Sprintf(
			"%s_%s_%s",
			relationship.ParentTable,
			relationship.ReferencedTable,
			relationship.ParentColumn,
		)

		r.Type = "RelationshipOneToMany"
		r.RelatedTable = relationship.ParentTable
		r.LocalColumn = relationship.ReferencedColumn
		r.ForeignColumn = relationship.ParentColumn

		r.RelationshipColumnType = fmt.Sprintf("[]*%s", getModelStructName(r.RelatedTable))
	}

	return r
}

func getTableRelationshipIdentifier(relationship *builderRepo.RelationshipMetadata) string {
	return fmt.Sprintf(
		"%s_%s_%s_%s",
		relationship.ParentTable,
		relationship.ParentColumn,
		relationship.ReferencedTable,
		relationship.ReferencedColumn,
	)
}
