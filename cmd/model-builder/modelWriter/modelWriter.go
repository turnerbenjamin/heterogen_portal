/* Don't look at me, I'm ugly */
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
		w.writeNewLine()
		w.WriteTableModelProjection(table)
	}
	w.writeNewLine()
	w.WriteTableModelGetter()

	w.writeNewLine()
	w.WriteTableMetadataGetter()

	w.writeNewLine()
	w.WriteTableProjectionGetter()

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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"strconv"
	"strings"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
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
func (s *schemaMetadata) GetTableModel(tableName string) queryModel.TableModel {
	return getTableModel(tableName)
}

// GetTableMetadata returns the metadata for a given table or nill if
// the table name is invalid
func (s *schemaMetadata) GetTableMetadata(tableName string) queryModel.TableMetadata {
	return getTableMetadata(tableName)
}

// columnMetadata describes the metadata associated with a database table column.
type columnMetadata struct {
	name         string
	dbType       queryModel.DbDataTypeName
	maxLength    int
	isPrimaryKey bool
	isRequired   bool
}

// Name returns the database name for the column
func (c *columnMetadata) Name() string {
	return c.name
}

// Type returns the column's type
func (c *columnMetadata) Type() queryModel.DbDataTypeName {
	return c.dbType
}



// relationship describes a foreign key relationship between two database columns.
type relationship struct {
	id               string
	columnName             string
	expansionColumnName             string
	relationshipType queryModel.RelationshipType
	from             queryModel.TableMetadata
	to               queryModel.TableMetadata
	fromColumn       queryModel.ColumnMetadata
	toColumn         queryModel.ColumnMetadata
	fromTableName    string
	toTableName      string
	fromColumnName   string
	toColumnName     string
	isInitialised    bool
}

// ExpansionColumnName returns the name of the pseudo relationship column on the 
// table used to store expanded results
func (r *relationship) ExpansionColumnName() string {
	return r.expansionColumnName
}

// ColumnName returns the name of the column, as it should be referenced in a
// query string
func (r *relationship) ColumnName() string{
	return r.columnName
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
func (r *relationship) From() queryModel.TableMetadata {
	return r.from
}

// To returns table metadata for the to table
func (r *relationship) To() queryModel.TableMetadata {
	return r.to
}

// From returns table metadata for the from table
func (r *relationship) FromColumn() queryModel.ColumnMetadata {
	return r.fromColumn
}

// To returns table metadata for the to table
func (r *relationship) ToColumn() queryModel.ColumnMetadata {
	return r.toColumn
}

// Type returns the relationship type
func (r *relationship) Type() queryModel.RelationshipType {
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
func (t *tableMetadata) GetColumnMetadata(columnName string) queryModel.ColumnMetadata {
	c, exists := t.columns[columnName]
	if !exists {
		return nil
	}
	return c
}

// GetRelationshipMetadata returns metadata for a given table column; returns
// nil if the relationship does not exist on the table
func (t *tableMetadata) GetRelationshipMetadata(relationshipName string) queryModel.RelationshipMetadata {
	r, exists := t.relationships[relationshipName]
	if !exists {
		return nil
	}
	return r
}

// GetTableModel returns a model representing the table
func (t *tableMetadata) GetModel() queryModel.TableModel {
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
func (t *tableMetadata) Columns() iter.Seq[queryModel.ColumnMetadata] {
	return func(yield func(queryModel.ColumnMetadata) bool) {
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

// PrimaryKeyField returns the column metadata for the primary key field
func (t *tableMetadata) PrimaryKeyField() queryModel.ColumnMetadata {
	pk, ok := t.columns[t.primaryKey]
	if !ok || pk == nil {
		panic(fmt.Sprintf("unable to access primary key for %s table", t.name))
	}
	return pk
}

// InitProjection initialises a projection object for the table
func (t *tableMetadata) InitProjection() (queryModel.Projection, error) {
	return initTableProjection(t.name)
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

// String returns the string representation of a Point
func (p *Point) String() string {
	return fmt.Sprintf("POINT (%f %f)", p.Coordinates[0], p.Coordinates[1])
}

// marshalProperty is a helper method to marshal an attribute and value to a json string buffer
func marshalProperty(buf *bytes.Buffer, isFirst bool, columnName string, value any) error {
	if !isFirst {
		buf.WriteByte(',')
	}

	buf.WriteString("\"")
	buf.WriteString(columnName)
	buf.WriteString("\":")

	v, err := json.Marshal(value)
	if err != nil {
		return err
	}

	buf.Write(v)
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
		writeToBuilder(sbR, fmt.Sprintf("columnName: \"%s\",\n", relationship.ColumnName))
		writeToBuilder(sbR, fmt.Sprintf("expansionColumnName: \"%s\",\n", relationship.ExpansionColumnName))
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
		return "queryModel.DbTypeNvarchar"
	case MsqlTypeInt:
		return "queryModel.DbTypeInt"
	case MsqlTypeDateTimeOffset:
		return "queryModel.DbTypeDateTimeOffset"
	case MsqlTypeGeography:
		return "queryModel.DbTypeGeography"
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

		tag := fmt.Sprintf("`json:\"%s\"`", col.Name)
		writeToBuilder(w.sb, fmt.Sprintf("%s %s %s\n", identifier, goType, tag))
	}

	// Relationship Cols
	for _, tableRelationship := range tableData.Relationships {
		tag := fmt.Sprintf("`json:\"%s\"`", tableRelationship.ExpansionColumnName)
		identifer := snakeToPascal(tableRelationship.ExpansionColumnName)
		writeToBuilder(w.sb, fmt.Sprintf("%s %s %s\n", identifer, tableRelationship.RelationshipColumnType, tag))
	}

	// projection column
	writeToBuilder(w.sb, fmt.Sprintf("projection %s `json:\"-\"`\n", getModelProjectionName(tableData.Name)))

	writeToBuilder(w.sb, "}\n\n")
	w.WriteGetSliceGetterFunction(structName, tableData)
	w.writeNewLine()
	w.WriteModelMarshalJSONFunction(structName, tableData)
	w.writeNewLine()
	w.WriteRelationshipColumnSetterFunction(structName, tableData)
	w.writeNewLine()
	w.WriteInitRelationshipFieldFunction(structName, tableData)
	w.writeNewLine()
	w.WriteGetJoinOnValueFunction(structName, tableData)
	w.writeNewLine()
	w.WriteGetValueExpressionFunction(structName, tableData)
	w.writeNewLine()
	w.WriteGetRelatedEntityFunction(structName, tableData)
	w.writeNewLine()
	w.WriteModelIsNilFunciton(structName, tableData)
	w.writeNewLine()
	w.WriteGetValueExpressionPrivateFunction(structName, tableData)
	w.writeNewLine()
	w.WriteSetProjectionFunction(structName, tableData)
}

func (w *modelWriter) WriteGetSliceGetterFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, fmt.Sprintf(
		"// NewSlice unmarshals a json array of %s and returns it as a slice\n",
		tableData.Name,
	))

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) NewSlice(jsonData []byte, projection queryModel.Projection) ([]queryModel.TableModel, error) {\n", modelStructName))

	writeToBuilder(w.sb, "if len(jsonData) == 0 {\n")
	writeToBuilder(w.sb, "return []queryModel.TableModel{}, nil\n")
	writeToBuilder(w.sb, "}\n\n")

	projectionName := getModelProjectionName(tableData.Name)
	writeToBuilder(w.sb, fmt.Sprintf("var typedProjection *%s\n", projectionName))
	writeToBuilder(w.sb, "switch p := projection.(type){\n")
	writeToBuilder(w.sb, fmt.Sprintf("case *%s:\n", projectionName))
	writeToBuilder(w.sb, "typedProjection = p\n")
	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil, fmt.Errorf(\"unable to create new slice: invalid projection type received\")\n")
	writeToBuilder(w.sb, "}\n\n")

	writeToBuilder(w.sb, fmt.Sprintf("var concreteSlice []*%s\n", modelStructName))

	writeToBuilder(w.sb, "if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {\n")
	writeToBuilder(w.sb, "return nil, err\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "result := make([]queryModel.TableModel, len(concreteSlice))\n\n")

	writeToBuilder(w.sb, "for i := range concreteSlice {\n")
	writeToBuilder(w.sb, "concreteSlice[i].projection = *typedProjection\n")
	writeToBuilder(w.sb, "result[i] = concreteSlice[i]\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "return result, nil\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteModelMarshalJSONFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, "// MarshalJSON marshals the model to a json string based on the projection\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) MarshalJSON() ([]byte, error) {\n", modelStructName))
	writeToBuilder(w.sb, "var buf bytes.Buffer\n")
	writeToBuilder(w.sb, "buf.WriteByte('{')\n\n")

	writeToBuilder(w.sb, "isFirst := true\n")

	writeColumn := func(columnName string) {
		projectionIdentifier := getColProjectionId(tableData.Name, columnName)
		writeToBuilder(w.sb, fmt.Sprintf("if m.projection.Has(%s) {\n", projectionIdentifier))
		writeToBuilder(w.sb, fmt.Sprintf(
			"err := marshalProperty(&buf, isFirst, \"%s\", m.%s)\n",
			columnName,
			snakeToPascal(columnName),
		))
		writeToBuilder(w.sb, "if err != nil {\n")
		writeToBuilder(w.sb, "return nil, err\n")
		writeToBuilder(w.sb, "}\n")
		writeToBuilder(w.sb, "isFirst = false\n")

		writeToBuilder(w.sb, "}\n\n")
	}

	for _, column := range tableData.Columns {
		writeColumn(column.Name)
	}

	for _, relationship := range tableData.Relationships {
		writeColumn(relationship.ExpansionColumnName)
	}

	writeToBuilder(w.sb, "buf.WriteByte('}')\n")
	writeToBuilder(w.sb, "return buf.Bytes(), nil\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteRelationshipColumnSetterFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, fmt.Sprintf(
		"// SetRelationshipField sets a given relationship field on the %s.%s table\n",
		tableData.Schema,
		tableData.Name,
	))

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) SetRelationshipField(relationshipId string, value queryModel.TableModel) error {\n", modelStructName))
	writeToBuilder(w.sb, "switch relationshipId {\n")

	for _, relationship := range tableData.Relationships {
		relationshipColumnIdentifier := snakeToPascal(relationship.ExpansionColumnName)
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", relationship.Id))
		writeToBuilder(w.sb, fmt.Sprintf("v, ok := value.(*%s)\n", getModelStructName(relationship.RelatedTable)))
		writeToBuilder(w.sb, "if !ok {\n")
		writeToBuilder(w.sb, "return errors.New(\"unexpected relationship type received\")\n")
		writeToBuilder(w.sb, "}\n")

		if relationship.Type == "queryModel.RelationshipManyToOne" {
			writeToBuilder(w.sb, fmt.Sprintf("m.%s = v\n\n", relationshipColumnIdentifier))
		}

		if relationship.Type == "queryModel.RelationshipOneToMany" {
			writeToBuilder(w.sb, fmt.Sprintf("m.%s = append(m.%s, v)\n\n", relationshipColumnIdentifier, relationshipColumnIdentifier))
		}
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return fmt.Errorf(\"unknown relationship: %s\", relationshipId)\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "return nil\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteInitRelationshipFieldFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, fmt.Sprintf(
		"// InitRelationshipField initialses 1:N relationship fields to empty arrays on\n// the %s.%s table\n",
		tableData.Schema,
		tableData.Name,
	))

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) InitRelationshipField(relationshipId string) error {\n", modelStructName))
	writeToBuilder(w.sb, "switch relationshipId {\n")

	// Write do nothing code for N:1 relationships
	countN1 := 0
	for _, relationship := range tableData.Relationships {
		if relationship.Type != "queryModel.RelationshipManyToOne" {
			continue
		}
		countN1++

		if countN1 == 1 {
			writeToBuilder(w.sb, "case ")
		} else {
			writeToBuilder(w.sb, ", ")
		}
		writeToBuilder(w.sb, fmt.Sprintf("\"%s\"", relationship.Id))
	}
	if countN1 > 0 {
		writeToBuilder(w.sb, ":\n")
		writeToBuilder(w.sb, "return nil\n")
	}

	// Write array setters for 1:N relationships
	for _, relationship := range tableData.Relationships {
		if relationship.Type != "queryModel.RelationshipOneToMany" {
			continue
		}

		relationshipColumnIdentifier := snakeToPascal(relationship.ExpansionColumnName)
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", relationship.Id))
		writeToBuilder(w.sb, fmt.Sprintf("m.%s = []*%s{}\n", relationshipColumnIdentifier, getModelStructName(relationship.RelatedTable)))
		writeToBuilder(w.sb, "return nil\n")
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return fmt.Errorf(\"unknown relationship: %s\", relationshipId)\n")
	writeToBuilder(w.sb, "}\n")
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

func (w *modelWriter) WriteGetValueExpressionFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, "// GetValueExpression returns the value from a given path as a ValueExpression\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) GetValueExpression(path []*queryModel.TraversalStep, columnName string, v queryModel.ValueBuilder) (queryModel.ValueExpression, error){\n", modelStructName))
	writeToBuilder(w.sb, "if len(path) > 0 {\n")
	writeToBuilder(w.sb, "nextStep := path[0]\n")
	writeToBuilder(w.sb, "nextEntity, err := m.getRelatedEntity(nextStep.Relationship)\n")

	writeToBuilder(w.sb, "if err != nil {\n")
	writeToBuilder(w.sb, "return nil, err\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "if nextEntity.IsNil(){\n")
	writeToBuilder(w.sb, "return v.Null(), nil\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "return nextEntity.GetValueExpression(path[1:], columnName, v)\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "val, err := m.getValueExpression(columnName, v)\n")
	writeToBuilder(w.sb, "if err != nil {\n")
	writeToBuilder(w.sb, "return nil, err\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "if val == nil{\n")
	writeToBuilder(w.sb, "return v.Null(), nil\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "return val, nil\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteGetRelatedEntityFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, "// GetRelatedEntity returns the value from N:1/1:1 relationships as a TableModel\n")
	writeToBuilder(w.sb, "// It will return an error for invalid relationships and relationship types\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) getRelatedEntity(relationship queryModel.RelationshipMetadata) (queryModel.TableModel, error) {\n", modelStructName))
	writeToBuilder(w.sb, "switch relationship.Id() {\n")

	for _, relationship := range tableData.Relationships {
		if relationship.Type != "queryModel.RelationshipManyToOne" {
			continue
		}
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", relationship.Id))
		writeToBuilder(w.sb, fmt.Sprintf("return m.%s, nil\n", snakeToPascal(relationship.ExpansionColumnName)))
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil, fmt.Errorf(\"unable to get related entity: unsupported relationship '%s'\", relationship.Id())\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteModelIsNilFunciton(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, "// IsNil is used to determine if a typed nil pointer contains a nil value\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) IsNil() bool {\n", modelStructName))
	writeToBuilder(w.sb, "return m == nil\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteGetValueExpressionPrivateFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, "// getValueExpression returns the value from a given column as a value expression\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) getValueExpression(columnName string, v queryModel.ValueBuilder) (queryModel.ValueExpression, error) {\n", modelStructName))
	writeToBuilder(w.sb, "switch columnName{\n")

	for _, column := range tableData.Columns {
		getValueExpression := w.buildGetValueExpression(column)
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", column.Name))
		writeToBuilder(w.sb, getValueExpression)
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil, fmt.Errorf(\"unsupported column: '%s'\", columnName)\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteSetProjectionFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, "// Project adds a column to the model's projection set\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) Project(columnName string) error { \n", modelStructName))
	writeToBuilder(w.sb, "return (&m.projection).Add(columnName)\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) buildGetValueExpression(columnData *builderRepo.ColumnMetadata) string {
	columnIdentifier := snakeToPascal(columnData.Name)

	nullCheck := func(columnIdentifier string) string {
		return fmt.Sprintf("if m.%s == nil {\n return v.Null(), nil }\n", columnIdentifier)
	}

	switch columnData.Type {
	case MsqlTypeNvarchar:
		return fmt.Sprintf("return v.String(m.%s), nil\n", columnIdentifier)
	case MsqlTypeInt:
		return fmt.Sprintf("return v.Int(m.%s), nil\n", columnIdentifier)
	case MsqlTypeDateTimeOffset, MsqlTypeGeography:
		return fmt.Sprintf("%s return v.String(m.%s.String()), nil\n", nullCheck(columnIdentifier), columnIdentifier)
	default:
		return ""
	}
}

type BusinessesProjection uint64

const (
	businessesProjectionId BusinessesProjection = 1 << iota
	businessesProjectionReference
	businessesProjectionTradingName
)

func (p *BusinessesProjection) Add(columnName string) error {
	switch columnName {
	case "id":
		*p |= businessesProjectionId
	case "reference":
		*p |= businessesProjectionReference
	case "trading_name":
		*p |= businessesProjectionTradingName
	default:
		return fmt.Errorf("unsupported column: '%s'", columnName)
	}
	return nil
}

func (p BusinessesProjection) Has(columnName string) bool {
	switch columnName {
	case "id":
		return p&businessesProjectionId != 0
	case "reference":
		return p&businessesProjectionReference != 0
	case "trading_name":
		return p&businessesProjectionTradingName != 0
	default:
		return false
	}
}

func (w *modelWriter) WriteTableModelProjection(tableData *builderRepo.TableMetadata) {
	projectionName := getModelProjectionName(tableData.Name)

	if len(tableData.Columns) > 64 {
		panic("model writer currently only supports tables with up to 64 columns")
	}

	// Write the type
	writeToBuilder(w.sb, fmt.Sprintf(
		"// %s represents column projection for the %s table.\n",
		projectionName,
		tableData.Name,
	))
	writeToBuilder(w.sb, fmt.Sprintf("type %s uint64\n", projectionName))
	w.writeNewLine()

	// Write the constants
	writeToBuilder(w.sb, "const (\n")
	for i, col := range tableData.Columns {
		identifier := getColProjectionId(tableData.Name, col.Name)

		if i == 0 {
			writeToBuilder(w.sb, fmt.Sprintf("%s %s = 1 << iota\n", identifier, projectionName))
		} else {
			writeToBuilder(w.sb, fmt.Sprintf("%s\n", identifier))
		}
	}

	for _, rel := range tableData.Relationships {
		identifier := getColProjectionId(tableData.Name, rel.ExpansionColumnName)
		writeToBuilder(w.sb, fmt.Sprintf("%s\n", identifier))
	}

	writeToBuilder(w.sb, ")\n")
	w.writeNewLine()

	//Write setter
	writeToBuilder(w.sb, "// Add includes a given column within the projection\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (p *%s) Add(columnName string) error {\n", projectionName))
	writeToBuilder(w.sb, "switch columnName{\n")

	for _, column := range tableData.Columns {
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", column.Name))
		writeToBuilder(w.sb, fmt.Sprintf("*p |= %s\n", getColProjectionId(tableData.Name, column.Name)))
	}

	for _, rel := range tableData.Relationships {
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", rel.ExpansionColumnName))
		writeToBuilder(w.sb, fmt.Sprintf("*p |= %s\n", getColProjectionId(tableData.Name, rel.ExpansionColumnName)))
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return fmt.Errorf(\"unsupported column: '%s'\", columnName)\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "return nil\n")
	writeToBuilder(w.sb, "}\n")
	w.writeNewLine()

	writeToBuilder(w.sb, "// Has is used to determine if a given column is in a projection\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (p %s) Has(projection %s) bool {\n", projectionName, projectionName))
	writeToBuilder(w.sb, "return p&projection != 0\n")
	writeToBuilder(w.sb, "}\n")
	w.writeNewLine()

	// WRITE IS EMPTY
	writeToBuilder(w.sb, "// IsEmpty is used to determine if there are no projections\n")
	writeToBuilder(w.sb, fmt.Sprintf("func (p %s) IsEmpty() bool {\n", projectionName))
	writeToBuilder(w.sb, "return p == 0\n")
	writeToBuilder(w.sb, "}\n")
	w.writeNewLine()
}

func (w *modelWriter) WriteTableModelGetter() {
	writeToBuilder(
		w.sb,
		"// getTableModel returns a new instance of a given table or nill if \n"+
			"// the table name is invalid\n",
	)
	writeToBuilder(w.sb, "func getTableModel(tableName string) queryModel.TableModel{\n")
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
	writeToBuilder(w.sb, "func getTableMetadata(tableName string) queryModel.TableMetadata{\n")
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

func (w *modelWriter) WriteTableProjectionGetter() {
	writeToBuilder(
		w.sb,
		"// func initTableProjection initialises a projection for the table\n",
	)
	writeToBuilder(w.sb, "func initTableProjection(tableName string) (queryModel.Projection, error){\n")
	writeToBuilder(w.sb, "switch tableName {\n")

	for _, table := range w.metadata.Tables {
		metadataStoreId := modelMetadataStoreIdentifier(table.Name)
		projectionName := getModelProjectionName(table.Name)

		writeToBuilder(w.sb, fmt.Sprintf("case %s.name:\n", metadataStoreId))
		writeToBuilder(w.sb, fmt.Sprintf("p := %s(0)\n", projectionName))
		writeToBuilder(w.sb, "return &p, nil\n")
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil, fmt.Errorf(\"unsupported table: '%s'\", tableName)\n")
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

	writeToBuilder(w.sb, "func (p *DatabaseAccessPolicy) GetTableAccessPolicy(tableName string) queryModel.TableAccessPolicy {\n")
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
	writeToBuilder(w.sb, fmt.Sprintf("func (p *%s) GetColumnAccessPolicy(columnName string) queryModel.ColumnAccessPolicy {\n", accessPolicyStructName))
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

func getModelProjectionName(tableName string) string {
	return fmt.Sprintf("%sProjection", snakeToPascal(tableName))
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
		return "int64"
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

			if relationship.ParentTable == table.Name {
				tableRelationship.ColumnName = tableRelationship.LocalColumn
			} else {
				tableRelationship.ColumnName = tableRelationship.ExpansionColumnName
			}

			if table.Relationships == nil {
				table.Relationships = map[string]*builderRepo.TableRelationship{}
			}
			table.Relationships[tableRelationship.ColumnName] = tableRelationship
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
		r.ExpansionColumnName = strings.TrimSuffix(relationship.ParentColumn, "_id")

		r.Type = "queryModel.RelationshipManyToOne"
		r.RelatedTable = relationship.ReferencedTable
		r.LocalColumn = relationship.ParentColumn
		r.ForeignColumn = relationship.ReferencedColumn

		r.RelationshipColumnType = fmt.Sprintf("*%s", getModelStructName(r.RelatedTable))
	}

	if relationship.ReferencedTable == table.Name {
		r.ExpansionColumnName = fmt.Sprintf(
			"%s_%s_%s",
			relationship.ParentTable,
			relationship.ReferencedTable,
			relationship.ParentColumn,
		)

		r.Type = "queryModel.RelationshipOneToMany"
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

func getColProjectionId(tableName, columnName string) string {
	return fmt.Sprintf("%sProjection%s", snakeToCamel(tableName), snakeToPascal(columnName))
}
