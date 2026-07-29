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

	w.writeNewLine()
	w.WriteTableMetadataGetter()

	w.writeNewLine()
	w.WriteTableMetadataBinder()
	w.writeNewLine()
	w.writeTableAccessStructs()

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
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"strconv"
	"strings"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/query"
)

var metadataIsBound = false

type schemaMetadata struct{}

func NewSchemaMetadata() *schemaMetadata {
	s := &schemaMetadata{}
	if !metadataIsBound {
		bindMetadata()
	}
	return s
}

// GetTableModel returns a new instance of a given table or nill if
// the table name is invalid
func (s *schemaMetadata) GetTableModel(tableName string) query.TableModel {
	return getTableModel(tableName)
}

// GetTableMetadata returns the metadata for a given table or nill if
// the table name is invalid
func (s *schemaMetadata) GetTableMetadata(tableName string) query.TableMetadata {
	return getTableMetadata(tableName)
}

// columnMetadata describes the metadata associated with a database table column.
type columnMetadata struct {
	name         string
	dbType       query.DbDataTypeName
	maxLength    int
	isPrimaryKey bool
	isRequired   bool
}

// Name returns the database name for the column
func (c *columnMetadata) Name() string {
	return c.name
}

// Type returns the column's type
func (c *columnMetadata) Type() query.DbDataTypeName {
	return c.dbType
}

// relationship describes a foreign key relationship between two database columns.
type relationship struct {
	id               string
	name             string
	relationshipType query.RelationshipType
	from             query.TableMetadata
	to               query.TableMetadata
	fromColumn       query.ColumnMetadata
	toColumn         query.ColumnMetadata
	fromTableName    string
	toTableName      string
	fromColumnName   string
	toColumnName     string
	isInitialised    bool
}

// bindMetadata binds metadata references at runtime to avoid invalid initiation
// cycle due to circular references
func (r *relationship) bindMetadata() {
	r.from = getTableMetadata(r.fromTableName)
	r.fromColumn = r.from.GetColumnMetadata(r.fromColumnName)
	r.to = getTableMetadata(r.toTableName)
	r.toColumn = r.to.GetColumnMetadata(r.toColumnName)

	if r.from == nil || r.fromColumn == nil || r.to == nil || r.toColumn == nil {
		panic(fmt.Sprintf("unable to bind metadata for relationship %s", r.id))
	}
}

// Id returns a unique identifier for a table relationship
func (r *relationship) Id() string {
	return r.id
}

// From returns table metadata for the from table
func (r *relationship) From() query.TableMetadata {
	return r.from
}

// To returns table metadata for the to table
func (r *relationship) To() query.TableMetadata {
	return r.to
}

// From returns table metadata for the from table
func (r *relationship) FromColumn() query.ColumnMetadata {
	return r.fromColumn
}

// To returns table metadata for the to table
func (r *relationship) ToColumn() query.ColumnMetadata {
	return r.toColumn
}

// Type returns the relationship type
func (r *relationship) Type() query.RelationshipType {
	return r.relationshipType
}

// tableMetadata describes the structure and relationships of a database table.
type tableMetadata struct {
	name               string
	schemaName         string
	fullyQualifiedName string
	primaryKey         string
	columns            map[string]*columnMetadata
	columnCount        int
	relationships      map[string]*relationship
}

// bindMetadata binds metadata references at runtime to avoid invalid initiation
// cycle due to circular references
func (t *tableMetadata) bindMetadata() {
	for _, relationship := range t.relationships {
		relationship.bindMetadata()
	}
}

// GetColumnMetadata returns metadata for a given table column; returns nil if
// the column does not exist on the table
func (t *tableMetadata) GetColumnMetadata(columnName string) query.ColumnMetadata {
	c, exists := t.columns[columnName]
	if !exists {
		return nil
	}
	return c
}

// GetRelationshipMetadata returns metadata for a given table column; returns
// nil if the relationship does not exist on the table
func (t *tableMetadata) GetRelationshipMetadata(relationshipName string) query.RelationshipMetadata {
	r, exists := t.relationships[relationshipName]
	if !exists {
		return nil
	}
	return r
}

// GetTableModel returns a model representing the table
func (t *tableMetadata) GetModel() query.TableModel {
	return getTableModel(t.name)
}

// Name returns the table name as it appears in the database
func (t *tableMetadata) Name() string {
	return t.name
}

// FullyQualifiedName returns the fully-qualified table name including the
// schema namespace
func (t *tableMetadata) FullyQualifiedName() string {
	return t.fullyQualifiedName
}

// Columns iterates over all columns associated with the table
func (t *tableMetadata) Columns() iter.Seq[query.ColumnMetadata] {
	return func(yield func(query.ColumnMetadata) bool) {
		for _, col := range t.columns {
			if !yield(col) {
				return
			}
		}
	}
}

// ColumnCount returns the total number of columns associated with the table
func (t *tableMetadata) ColumnCount() int {
	if t.columnCount == -1 {
		t.columnCount = len(t.columns)
	}
	return t.columnCount
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

// ColumnAccessPolicy defines the query access policy for a given database
// column
type ColumnAccessPolicy struct {
	UserCanAccess bool
}

// CanAccess defines, at the column level, if a user can perform any
// operations on that column
func (p ColumnAccessPolicy) CanAccess() bool {
	return p.UserCanAccess
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
	metadataIdentifer := modelMetadataStoreIdentifier(tableData.Name)

	writeToBuilder(w.sb, fmt.Sprintf(
		"// %s contains the database metadata for the %s.%s table.\n",
		metadataIdentifer,
		tableData.Schema,
		tableData.Name,
	))
	writeToBuilder(w.sb, fmt.Sprintf("var %s = &tableMetadata{\n", modelMetadataStoreIdentifier(tableData.Name)))
	writeToBuilder(w.sb, fmt.Sprintf("name: \"%s\",\n", tableData.Name))
	writeToBuilder(w.sb, fmt.Sprintf("schemaName: \"%s\",\n", tableData.Schema))
	writeToBuilder(w.sb, fmt.Sprintf("fullyQualifiedName: \"%s.%s\",\n", tableData.Schema, tableData.Name))
	writeToBuilder(w.sb, fmt.Sprintf("primaryKey: \"%s\",\n", primaryKey))
	writeToBuilder(w.sb, fmt.Sprintf("columns: %s,\n", columnMapValue))
	writeToBuilder(w.sb, fmt.Sprintf("relationships: %s,\n", relationshipMapValue))
	writeToBuilder(w.sb, "columnCount: -1,\n")
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
		fmt.Sprintf("func Get%sMetadata() *tableMetadata {\n", tableNamePascal),
	)

	metadataIdentifier := modelMetadataStoreIdentifier(tableData.Name)
	writeToBuilder(w.sb, fmt.Sprintf("return %s\n", metadataIdentifier))
	writeToBuilder(w.sb, "}")
}

func buildColumnMap(tableData *builderRepo.TableMetadata) (string, string) {
	primaryKey := ""
	sbColMap := &strings.Builder{}

	writeToBuilder(sbColMap, "map[string]*columnMetadata{\n")

	// DB Columns
	for _, col := range tableData.Columns {
		colType := getColTypeString(col.Type)
		if col.PrimaryKey == 1 {
			primaryKey = col.Name
		}

		writeToBuilder(sbColMap, fmt.Sprintf("\"%s\": {\n", col.Name))
		writeToBuilder(sbColMap, fmt.Sprintf("name: \"%s\",\n", col.Name))
		writeToBuilder(sbColMap, fmt.Sprintf("dbType: %s,\n", colType))
		writeToBuilder(sbColMap, fmt.Sprintf("maxLength: %d,\n", col.MaxLength))
		writeToBuilder(sbColMap, fmt.Sprintf("isRequired: %t,\n", !col.IsNullable))
		writeToBuilder(sbColMap, fmt.Sprintf("isPrimaryKey: %t,\n", col.PrimaryKey == 1))
		writeToBuilder(sbColMap, "},\n")
	}

	writeToBuilder(sbColMap, "}")

	return primaryKey, sbColMap.String()
}

func (w *modelWriter) buildRelationshipMap(tableData *builderRepo.TableMetadata) string {
	sbR := &strings.Builder{}

	writeToBuilder(sbR, "map[string]*relationship{\n")
	for key, relationship := range tableData.Relationships {
		writeToBuilder(sbR, fmt.Sprintf("\"%s\": {\n", key))
		writeToBuilder(sbR, fmt.Sprintf("id: \"%s\",\n", relationship.Id))
		writeToBuilder(sbR, fmt.Sprintf("name: \"%s\",\n", relationship.RelationshipColumn))
		writeToBuilder(sbR, fmt.Sprintf("relationshipType: %s,\n", relationship.Type))
		writeToBuilder(sbR, fmt.Sprintf("fromTableName: \"%s\",\n", tableData.Name))
		writeToBuilder(sbR, fmt.Sprintf("toTableName: \"%s\",\n", relationship.RelatedTable))
		writeToBuilder(sbR, fmt.Sprintf("fromColumnName: \"%s\",\n", relationship.LocalColumn))
		writeToBuilder(sbR, fmt.Sprintf("toColumnName: \"%s\",\n", relationship.ForeignColumn))
		writeToBuilder(sbR, "},\n")
	}

	writeToBuilder(sbR, "}")
	return sbR.String()
}

func getColTypeString(typeName builderRepo.MsqlDataTypeName) string {
	switch typeName {
	case MsqlTypeNvarchar:
		return "query.DbTypeNvarchar"
	case MsqlTypeInt:
		return "query.DbTypeInt"
	case MsqlTypeDateTimeOffset:
		return "query.DbTypeDateTimeOffset"
	case MsqlTypeGeography:
		return "query.DbTypeGeography"
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
	w.WriteGetSliceGetterFunction(structName, tableData)
	w.writeNewLine()
	w.WriteRelationshipColumnSetterFunction(structName, tableData)
	w.writeNewLine()
	w.WriteGetJoinOnValueFunction(structName, tableData)
}

func (w *modelWriter) WriteGetSliceGetterFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, fmt.Sprintf(
		"// NewSlice unmarshals a json array of %s and returns it as a slice\n",
		tableData.Name,
	))

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) NewSlice(jsonData []byte) ([]query.TableModel, error) {\n", modelStructName))
	writeToBuilder(w.sb, fmt.Sprintf("var concreteSlice []*%s\n", modelStructName))

	writeToBuilder(w.sb, "if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {\n")
	writeToBuilder(w.sb, "return nil, err\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "result := make([]query.TableModel, len(concreteSlice))\n")
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

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) SetRelationshipField(relationshipId string, value query.TableModel) error {\n", modelStructName))
	writeToBuilder(w.sb, "switch relationshipId {\n")

	for _, relationship := range tableData.Relationships {
		relationshipColumnIdentifier := snakeToPascal(relationship.RelationshipColumn)
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", relationship.Id))
		writeToBuilder(w.sb, fmt.Sprintf("v, ok := value.(*%s)\n", getModelStructName(relationship.RelatedTable)))
		writeToBuilder(w.sb, "if !ok {\n")
		writeToBuilder(w.sb, "return errors.New(\"unexpected relationship type received\")\n")
		writeToBuilder(w.sb, "}\n")

		if relationship.Type == "query.RelationshipManyToOne" {
			writeToBuilder(w.sb, fmt.Sprintf("m.%s = v\n\n", relationshipColumnIdentifier))
		}

		if relationship.Type == "query.RelationshipOneToMany" {
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

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) GetJoinOnValue(relationshipId string) (string, error){\n", modelStructName))
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
		"// getTableModel returns a new instance of a given table or nill if \n"+
			"// the table name is invalid\n",
	)
	writeToBuilder(w.sb, "func getTableModel(tableName string) query.TableModel{\n")
	writeToBuilder(w.sb, "switch tableName {\n")

	for _, table := range w.metadata.Tables {
		metadataStoreId := modelMetadataStoreIdentifier(table.Name)
		writeToBuilder(w.sb, fmt.Sprintf("case %s.name:\n", metadataStoreId))
		writeToBuilder(w.sb, fmt.Sprintf("return new(%s)\n", getModelStructName(table.Name)))
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteTableMetadataGetter() {
	writeToBuilder(
		w.sb,
		"// getTableMetadata returns the metadata for a given table or nill if \n"+
			"// the table name is invalid\n",
	)
	writeToBuilder(w.sb, "func getTableMetadata(tableName string) query.TableMetadata{\n")
	writeToBuilder(w.sb, "switch tableName {\n")

	for _, table := range w.metadata.Tables {
		metadataStoreId := modelMetadataStoreIdentifier(table.Name)
		writeToBuilder(w.sb, fmt.Sprintf("case %s.name:\n", metadataStoreId))
		writeToBuilder(w.sb, fmt.Sprintf("return %s\n", modelMetadataStoreIdentifier(table.Name)))
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteTableMetadataBinder() {
	writeToBuilder(
		w.sb,
		"// bindMetadata binds table references at runtime to avoid invalid initiation\n"+
			"// cycle due to circular references\n",
	)
	writeToBuilder(w.sb, "func bindMetadata() {\n")
	for _, table := range w.metadata.Tables {
		metadataStoreId := modelMetadataStoreIdentifier(table.Name)
		writeToBuilder(w.sb, fmt.Sprintf("%s.bindMetadata()\n", metadataStoreId))
	}
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) writeTableAccessStructs() {
	// Database Access Policy
	writeToBuilder(w.sb, "// AccessPolicy defines the access policy for the database\n")

	writeToBuilder(w.sb, "type DatabaseAccessPolicy struct {\n")
	for _, t := range w.metadata.Tables {
		tableAccessPolicyName := getTableAccessPolicyName(t.Name)
		writeToBuilder(w.sb, fmt.Sprintf("%s *%s\n", tableAccessPolicyName, tableAccessPolicyName))
	}
	writeToBuilder(w.sb, "}\n")
	w.writeNewLine()

	// Table access policy getter
	writeToBuilder(
		w.sb,
		"// GetTableAccessPolicy returns an access policy for a given table for nil\n"+
			"// if the table does not exist\n",
	)

	writeToBuilder(w.sb, "func (p *DatabaseAccessPolicy) GetTableAccessPolicy(tableName string) query.TableAccessPolicy {\n")
	writeToBuilder(w.sb, "switch tableName {\n")
	for _, t := range w.metadata.Tables {
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", t.Name))
		writeToBuilder(w.sb, fmt.Sprintf("return p.%s\n", getTableAccessPolicyName(t.Name)))
	}
	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "}\n")
	w.writeNewLine()

	// Individual Table Access Policy struct definitions
	for i, t := range w.metadata.Tables {
		if i != 0 {
			w.writeNewLine()
		}
		w.writeTableAccessStruct(t)
	}
}

func (w *modelWriter) writeTableAccessStruct(t *builderRepo.TableMetadata) {
	accessPolicyStructName := getTableAccessPolicyName(t.Name)
	// Access Policy Struct
	writeToBuilder(w.sb, fmt.Sprintf(
		"// %s defines an access policy for the %s table\n",
		accessPolicyStructName,
		t.Name,
	))

	writeToBuilder(w.sb, fmt.Sprintf("type %s struct {\n", accessPolicyStructName))
	writeToBuilder(w.sb, "UserCanAccess bool\n")
	for _, c := range t.Columns {
		writeToBuilder(w.sb, fmt.Sprintf("%s ColumnAccessPolicy\n", snakeToPascal(c.Name)))
	}
	writeToBuilder(w.sb, "}\n")
	w.writeNewLine()

	// Can Access
	writeToBuilder(w.sb, fmt.Sprintf(
		"// CanAccess defines, at the table level, if a user can perform any\n"+
			"// operations on the %s table. Specific column access policies can restrict\n"+
			"// access given but cannot override access denied\n",
		t.Name,
	))
	writeToBuilder(w.sb, fmt.Sprintf("func (p *%s) CanAccess() bool {\n", accessPolicyStructName))
	writeToBuilder(w.sb, "return p.UserCanAccess\n")
	writeToBuilder(w.sb, "}\n")
	w.writeNewLine()

	// GetColumn
	writeToBuilder(w.sb, fmt.Sprintf(
		"// GetColumnAccessPolicy returns an access policy for a specific column on\n"+
			"// the %s table. Returns nil if the column does not exist\n",
		t.Name,
	))
	writeToBuilder(w.sb, fmt.Sprintf("func (p *%s) GetColumnAccessPolicy(columnName string) query.ColumnAccessPolicy {\n", accessPolicyStructName))
	writeToBuilder(w.sb, "switch columnName {\n")
	for _, c := range t.Columns {
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", c.Name))
		writeToBuilder(w.sb, fmt.Sprintf("return p.%s\n", snakeToPascal(c.Name)))
	}
	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "}\n")
	w.writeNewLine()

}

func getTableAccessPolicyName(tableName string) string {
	return fmt.Sprintf("%sAccessPolicy", snakeToPascal(tableName))
}

func getModelStructName(tableName string) string {
	return fmt.Sprintf("%sModel", snakeToPascal(tableName))
}

func modelMetadataStoreIdentifier(tableName string) string {
	return fmt.Sprintf("%sMetadata", snakeToCamel(tableName))
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

		r.Type = "query.RelationshipManyToOne"
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

		r.Type = "query.RelationshipOneToMany"
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
