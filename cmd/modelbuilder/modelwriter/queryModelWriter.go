/* Don't look at me, I'm ugly */
package modelwriter

import (
	"fmt"
	"go/format"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderrepo"
)

const (
	MsqlTypeTinyInt        builderrepo.MsqlDataTypeName = "tinyint"
	MsqlTypeNvarchar       builderrepo.MsqlDataTypeName = "nvarchar"
	MsqlTypeInt            builderrepo.MsqlDataTypeName = "int"
	MsqlTypeGeography      builderrepo.MsqlDataTypeName = "geography"
	MsqlTypeDateTimeOffset builderrepo.MsqlDataTypeName = "datetimeoffset"
)

type relationshipType string

const (
	RelationshipOneToMany relationshipType = "1:N"
	RelationshipManyToOne relationshipType = "N:1"
)

type queryModelWriter struct {
	path     string
	metadata *builderrepo.DatabaseSchema
	wtr      *writer
}

func NewQueryModelWriter(
	path string,
	metadata *builderrepo.DatabaseSchema,
) *queryModelWriter {
	return &queryModelWriter{
		path:     path,
		metadata: metadata,
		wtr:      &writer{sb: &strings.Builder{}},
	}
}

func (w *queryModelWriter) Write() {
	buildTableRelationshipData(w.metadata)

	w.writePackageAndStaticTypeDefinitions()
	w.wtr.NewLine()

	w.writeSchemaDefinition()

	for _, table := range w.metadata.Tables {
		w.wtr.NewLine()

		w.writeResourceDefinition(table)
		w.wtr.NewLine()

		w.writeTableMetadataDefinition(table)
		w.wtr.NewLine()

		w.WriteTableModel(table)
		w.wtr.NewLine()

		w.WriteTableModelProjection(table)
	}

	w.wtr.NewLine()
	w.writeTableAccessStructs()

	raw := w.wtr.String()
	formatted, err := format.Source([]byte(raw))
	if err != nil {
		_ = os.WriteFile(w.path+"_raw.go", []byte(raw), 0644)
		log.Fatal(err)

	}

	err = os.WriteFile(w.path, formatted, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// func writeToBuilder(sb *strings.Builder, s string) {
// 	_, err := sb.Write([]byte(s))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

func (w *queryModelWriter) writePackageAndStaticTypeDefinitions() {
	w.wtr.Write(`//GENERATED CODE
// see cmd/modelbuilder/
package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/turnerbenjamin/querystack/querymodel"
)

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
`)
}

func (w *queryModelWriter) writeSchemaDefinition() {
	w.wtr.WriteLine("type schema struct{}")
	w.wtr.NewLine()

	w.wtr.WriteLine("func NewSchema() schema{")
	w.wtr.WriteLine("return schema{}")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	w.wtr.WriteLine("func (s schema) GetResource(resourceName string) (querymodel.Resource, bool){")
	w.wtr.WriteLine("switch resourceName{")

	for _, t := range w.metadata.Tables {
		tableName := t.Name
		resourceName := getModelResourceStructName(tableName)

		w.wtr.WriteLine("case \"%s\":", tableName)
		w.wtr.WriteLine("return %s{}, true", resourceName)
	}
	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return nil, false")

	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}\n")
	w.wtr.NewLine()
}

func (w *queryModelWriter) writeResourceDefinition(tableData *builderrepo.TableMetadata) {
	resourceTypeName := getModelResourceStructName(tableData.Name)

	// Write resource type
	w.wtr.WriteLine("type %s struct{}", resourceTypeName)
	w.wtr.NewLine()

	// Write funcs
	w.writeResourceGetMetadataFunc(resourceTypeName, tableData)
	w.wtr.NewLine()

	w.writeResourceInitProjectionFunc(resourceTypeName, tableData)
	w.wtr.NewLine()

	w.WriteResourceSliceFromJSONFunction(resourceTypeName, tableData)
	w.wtr.NewLine()

	w.WriteResourceTypedSliceFromJSONFunction(resourceTypeName, tableData)
	w.wtr.NewLine()
}

func (w *queryModelWriter) writeResourceGetMetadataFunc(
	resourceTypeName string,
	tableData *builderrepo.TableMetadata,
) {
	metadataIdentifier := modelMetadataStoreIdentifier(tableData.Name)
	w.wtr.WriteLine("func (r %s) GetMetadata() querymodel.TableMetadata {", resourceTypeName)
	w.wtr.WriteLine("return %s", metadataIdentifier)
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) writeResourceInitProjectionFunc(
	resourceTypeName string,
	tableData *builderrepo.TableMetadata,
) {
	modelProjectionIdentifier := getModelProjectionName(tableData.Name)
	w.wtr.WriteLine("func (r %s) InitProjection() querymodel.Projection {", resourceTypeName)
	w.wtr.WriteLine("return new(%s)", modelProjectionIdentifier)
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) WriteResourceSliceFromJSONFunction(resourceTypeName string, tableData *builderrepo.TableMetadata) {
	w.wtr.WriteLine(
		"// SliceFromJSON unmarshals a json array of %s and returns it as a slice",
		tableData.Name,
	)

	w.wtr.WriteLine("func (m %s) SliceFromJSON(", resourceTypeName)
	w.wtr.WriteLine("  jsonData []byte,")
	w.wtr.WriteLine("  projectionNode *querymodel.ProjectionNode,")
	w.wtr.WriteLine(") ([]querymodel.TableModel, error) {")

	w.wtr.WriteLine("if len(jsonData) == 0 {")
	w.wtr.WriteLine("return []querymodel.TableModel{}, nil")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	modelStructName := getModelStructName(tableData.Name)
	w.wtr.WriteLine("var concreteSlice []*%s", modelStructName)

	w.wtr.WriteLine("if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {")
	w.wtr.WriteLine("return nil, err")
	w.wtr.WriteLine("}")

	w.wtr.WriteLine("result := make([]querymodel.TableModel, len(concreteSlice))")
	w.wtr.NewLine()

	w.wtr.WriteLine("for i := range concreteSlice {")
	w.wtr.WriteLine("if err := concreteSlice[i].SetProjection(projectionNode); err != nil{")
	w.wtr.WriteLine("return nil, err")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("result[i] = concreteSlice[i]")
	w.wtr.WriteLine("}")

	w.wtr.WriteLine("return result, nil")
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) WriteResourceTypedSliceFromJSONFunction(resourceTypeName string, tableData *builderrepo.TableMetadata) {
	modelStructName := getModelStructName(tableData.Name)
	w.wtr.WriteLine(
		"// TypedSliceFromJSON unmarshals a json array of %s and returns it as a slice\n",
		tableData.Name,
	)

	w.wtr.WriteLine("func (m %s) TypedSliceFromJSON(", resourceTypeName)
	w.wtr.WriteLine("  jsonData []byte,")
	w.wtr.WriteLine("  projectionNode *querymodel.ProjectionNode,")
	w.wtr.WriteLine(") ([]*%s, error) {", modelStructName)

	w.wtr.WriteLine("if len(jsonData) == 0 {")
	w.wtr.WriteLine("return []*%s{}, nil", modelStructName)
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	w.wtr.WriteLine("var results []*%s", modelStructName)
	w.wtr.WriteLine("if err := json.Unmarshal(jsonData, &results); err != nil {")
	w.wtr.WriteLine("return nil, err")
	w.wtr.WriteLine("}")

	w.wtr.WriteLine("for _, result := range results {")
	w.wtr.WriteLine("if err := result.SetProjection(projectionNode); err != nil{")
	w.wtr.WriteLine("return nil, err")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}")

	w.wtr.WriteLine("return results, nil")
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) writeTableMetadataDefinition(tableData *builderrepo.TableMetadata) {
	primaryKey, columnMapValue := buildColumnMap(tableData)
	if primaryKey == "" {
		log.Fatal(fmt.Errorf("unable to access primary key for table %s", tableData.Name))
	}

	relationshipMapValue := w.buildRelationshipMap(tableData)
	metadataIdentifer := modelMetadataStoreIdentifier(tableData.Name)

	w.wtr.WriteLine(
		"// %s contains the database metadata for the %s.%s table.",
		metadataIdentifer,
		tableData.Schema,
		tableData.Name,
	)

	w.wtr.WriteLine("var %s = querymodel.TableMetadata{", metadataIdentifer)
	w.wtr.WriteLine("Name: \"%s\",", tableData.Name)
	w.wtr.WriteLine("FullyQualifiedName: \"%s.%s\",", tableData.Schema, tableData.Name)

	w.wtr.WriteLine("PrimaryKeyColumn: querymodel.ColumnMetadata{")
	w.wtr.WriteLine("Name: \"%s\",", primaryKey)
	w.wtr.WriteLine("Type: querymodel.DbTypeString,")
	w.wtr.WriteLine("},")

	w.wtr.WriteLine("Columns: %s,", columnMapValue)
	w.wtr.WriteLine("Relationships: %s,", relationshipMapValue)
	w.wtr.WriteLine("}")
	w.wtr.NewLine()
}

func buildColumnMap(tableData *builderrepo.TableMetadata) (string, string) {
	primaryKey := ""
	colMapWtr := &writer{sb: &strings.Builder{}}

	colMapWtr.WriteLine("map[string]querymodel.ColumnMetadata{")

	// DB Columns
	for _, col := range tableData.Columns {
		colType := getQueryModelType(col.Type, col)
		if col.PrimaryKey == 1 {
			primaryKey = col.Name
		}

		colMapWtr.WriteLine("\"%s\": {", col.Name)
		colMapWtr.WriteLine("Name: \"%s\",", col.Name)
		colMapWtr.WriteLine("Type: %s,", colType)
		colMapWtr.WriteLine("},")
	}
	colMapWtr.Write("}")

	return primaryKey, colMapWtr.String()
}

func (w *queryModelWriter) buildRelationshipMap(tableData *builderrepo.TableMetadata) string {
	relMapWtr := &writer{sb: &strings.Builder{}}

	relMapWtr.WriteLine("map[string]querymodel.RelationshipMetadata{")

	for key, relationship := range tableData.Relationships {
		fromResource := getModelResourceStructName(tableData.Name)
		toResource := getModelResourceStructName(relationship.RelatedTable)

		relMapWtr.WriteLine("%s: {", strLit(key))
		relMapWtr.WriteLine("Id: %s,", strLit(relationship.Id))
		relMapWtr.WriteLine("Type: %s,", relationship.Type)
		relMapWtr.WriteLine("ColumnName: %s,", strLit(relationship.ColumnName))
		relMapWtr.WriteLine("ExpansionColumnName: %s,", strLit(relationship.ExpansionColumnName))
		relMapWtr.WriteLine("From: %s{},", fromResource)
		relMapWtr.WriteLine("To: %s{},", toResource)

		relMapWtr.WriteLine("FromColumn: querymodel.ColumnMetadata{")
		relMapWtr.WriteLine("  Name: %s,", strLit(relationship.LocalColumn))
		relMapWtr.WriteLine("  Type: querymodel.DbTypeString,")
		relMapWtr.WriteLine("},")

		relMapWtr.WriteLine("ToColumn: querymodel.ColumnMetadata{")
		relMapWtr.WriteLine("  Name: %s,", strLit(relationship.ForeignColumn))
		relMapWtr.WriteLine("  Type: querymodel.DbTypeString,")
		relMapWtr.WriteLine("},")

		relMapWtr.WriteLine("},")
	}
	relMapWtr.Write("}")

	return relMapWtr.String()
}

func getQueryModelType(typeName builderrepo.MsqlDataTypeName, columnData *builderrepo.ColumnMetadata) string {
	goType := msqlTypeToGoType(typeName, columnData)

	if columnData.EnumName != "" {
		return "querymodel.DbTypeEnum"
	}

	switch goType {
	case "string":
		return "querymodel.DbTypeString"
	case "int64", "uint8":
		return "querymodel.DbTypeInt"
	case "*time.Time":
		return "querymodel.DbTypeDateTime"
	case "*querymodel.Point":
		return "querymodel.DbTypePoint"
	default:
		return ""
	}
}

func (w *queryModelWriter) WriteTableModel(tableData *builderrepo.TableMetadata) {
	structName := getModelStructName(tableData.Name)

	w.wtr.WriteLine(
		"// %s represents a row from the %s.%s table.",
		structName,
		tableData.Schema,
		tableData.Name,
	)
	w.wtr.WriteLine("type %s struct {", structName)

	// DB Cols
	for _, col := range tableData.Columns {
		identifier := snakeToPascal(col.Name)

		goType := msqlTypeToGoType(col.Type, col)
		if goType == "" {
			log.Fatal(fmt.Errorf("unable to convert type to go type: %s(%s)", col.Name, col.Type))
		}

		tag := fmt.Sprintf("`json:%s`", strLit(col.Name))
		w.wtr.WriteLine("%s %s %s", identifier, goType, tag)
	}

	// Relationship Cols
	for _, tableRelationship := range tableData.Relationships {
		tag := fmt.Sprintf("`json:%s`", strLit(tableRelationship.ExpansionColumnName))
		identifer := snakeToPascal(tableRelationship.ExpansionColumnName)
		w.wtr.WriteLine("%s %s %s", identifer, tableRelationship.RelationshipColumnType, tag)
	}

	// projection column
	w.wtr.WriteLine("projection %s `json:\"-\"`", getModelProjectionName(tableData.Name))

	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// collection type
	collectionType := getModelCollectionType(tableData.Name)
	w.wtr.WriteLine("type %s []*%s", collectionType, structName)
	w.wtr.NewLine()

	w.WriteSetProjectionFunction(structName, tableData)
	w.wtr.NewLine()

	w.WriteModelMarshalJSONFunction(structName, tableData)
	w.wtr.NewLine()

	w.WriteGetValueExpressionFunction(structName, tableData)
	w.wtr.NewLine()

	w.WriteGetRelatedEntityFunction(structName, tableData)
	w.wtr.NewLine()

	w.WriteGetValueExpressionPrivateFunction(structName, tableData)
	w.wtr.NewLine()

	w.WriteProjectionFunction(structName, tableData)
}

func (w *queryModelWriter) WriteSetProjectionFunction(modelStructName string, tableData *builderrepo.TableMetadata) {
	writeDocComment := func() {
		w.wtr.WriteLine("// SetProjection sets a projection node and sets projection for itself and any")
		w.wtr.WriteLine("// child nodes")
	}

	writeSignature := func(receiverParam string, typeName string) {
		w.wtr.WriteLine(
			"func (%s %s) SetProjection(projectionNode *querymodel.ProjectionNode) error {",
			receiverParam,
			typeName,
		)
	}

	writeProjectionCast := func() {
		w.wtr.WriteLine("projection := projectionNode.Projection")
		w.wtr.NewLine()

		projectionName := getModelProjectionName(tableData.Name)
		w.wtr.WriteLine("var typedProjection *%s", projectionName)
		w.wtr.WriteLine("switch p := projection.(type){")
		w.wtr.WriteLine("case *%s:", projectionName)
		w.wtr.WriteLine("typedProjection = p")
		w.wtr.WriteLine("default:")
		w.wtr.WriteLine("return fmt.Errorf(\"unable to create new slice: invalid projection type received\")")
		w.wtr.WriteLine("}")
		w.wtr.NewLine()
	}

	collectionType := getModelCollectionType(tableData.Name)
	concreteSetterIdentifier := fmt.Sprintf("set%s", getModelProjectionName(tableData.Name))

	// Write typing function for model
	receiverParam := "m"

	writeDocComment()
	writeSignature(receiverParam, "*"+modelStructName)
	writeProjectionCast()
	w.wtr.WriteLine("return m.%s(*typedProjection, projectionNode.Children)", concreteSetterIdentifier)
	w.wtr.WriteLine("}")

	// write typing function for model collection
	receiverParam = "ms"

	writeDocComment()
	writeSignature(receiverParam, collectionType)
	writeProjectionCast()

	w.wtr.WriteLine("for _, m := range ms{")
	w.wtr.WriteLine(
		"if err := m.%s(*typedProjection, projectionNode.Children); err != nil{",
		concreteSetterIdentifier,
	)
	w.wtr.WriteLine("return err")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("return nil")
	w.wtr.WriteLine("}")

	// write concrete setter function
	w.wtr.WriteLine(
		"// %s sets a projection node and sets projection for itself and any",
		concreteSetterIdentifier,
	)
	w.wtr.WriteLine("// child nodes")

	projectionType := getModelProjectionName(tableData.Name)
	w.wtr.WriteLine(
		"func (m *%s) %s(",
		modelStructName,
		concreteSetterIdentifier,
	)

	w.wtr.WriteLine("projection %s,", projectionType)
	w.wtr.WriteLine("childNodes map[string]*querymodel.ProjectionNode,")
	w.wtr.WriteLine(") error {")

	w.wtr.WriteLine("m.projection = projection")
	for _, r := range tableData.Relationships {
		expansionColumn := snakeToPascal(r.ExpansionColumnName)

		w.wtr.WriteLine("if node, exists := childNodes[%s]; exists {", strLit(r.ExpansionColumnName))
		w.wtr.WriteLine("if m.%s != nil {", expansionColumn)
		w.wtr.WriteLine("if err := m.%s.SetProjection(node); err != nil{", expansionColumn)
		w.wtr.WriteLine("return err")
		w.wtr.WriteLine("}")
		w.wtr.Write("}")

		if r.Type == "querymodel.RelationshipOneToMany" {
			w.wtr.WriteLine("else{")
			w.wtr.WriteLine("m.%s = %s{}", expansionColumn, getModelCollectionType(r.RelatedTable))
			w.wtr.WriteLine("}")
		} else {
			w.wtr.NewLine()
		}
		w.wtr.WriteLine("}")
		w.wtr.NewLine()
	}

	w.wtr.WriteLine("return nil")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()
}

func (w *queryModelWriter) WriteModelMarshalJSONFunction(modelStructName string, tableData *builderrepo.TableMetadata) {
	w.wtr.WriteLine("// MarshalJSON marshals the model to a json string based on the projection")

	w.wtr.WriteLine("func (m *%s) MarshalJSON() ([]byte, error) {", modelStructName)
	w.wtr.WriteLine("var buf bytes.Buffer")
	w.wtr.WriteLine("buf.WriteByte('{')")
	w.wtr.NewLine()

	w.wtr.WriteLine("isFirst := true")

	writeColumn := func(isLastItem bool, columnName string) {
		projectionIdentifier := getColProjectionId(tableData.Name, columnName)
		w.wtr.WriteLine("if m.projection.Has(%s) {", projectionIdentifier)
		w.wtr.WriteLine(
			"err := marshalProperty(&buf, isFirst, %s, m.%s)",
			strLit(columnName),
			snakeToPascal(columnName),
		)
		w.wtr.WriteLine("if err != nil {")
		w.wtr.WriteLine("return nil, err")
		w.wtr.WriteLine("}")

		if !isLastItem {
			w.wtr.WriteLine("isFirst = false")
		}

		w.wtr.WriteLine("}")
		w.wtr.NewLine()
	}

	colCount := len(tableData.Columns)
	relCount := len(tableData.Relationships)

	for i, column := range tableData.Columns {
		isLastItem := (i == colCount-1) && relCount == 0
		writeColumn(isLastItem, column.Name)
	}

	i := 0
	for _, relationship := range tableData.Relationships {
		isLastItem := (i == relCount-1)
		writeColumn(isLastItem, relationship.ExpansionColumnName)
		i++
	}

	w.wtr.WriteLine("buf.WriteByte('}')")
	w.wtr.WriteLine("return buf.Bytes(), nil")
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) WriteGetValueExpressionFunction(modelStructName string, tableData *builderrepo.TableMetadata) {
	w.wtr.WriteLine("// GetValue returns the value from a given path")

	w.wtr.WriteLine("func (m *%s) GetValue(", modelStructName)
	w.wtr.WriteLine(" v querymodel.ValueBuilder,")
	w.wtr.WriteLine(" path []*querymodel.TraversalStep,")
	w.wtr.WriteLine(" columnName string,")
	w.wtr.WriteLine(") (querymodel.Value, error){")

	w.wtr.WriteLine("if len(path) > 0 {")

	if hasManyToOneRelationships(tableData) {
		w.wtr.WriteLine("nextStep := path[0]")
		w.wtr.WriteLine("nextEntity, nextEntityIsNil, err := m.getRelatedEntity(nextStep.Relationship)")

		w.wtr.WriteLine("if err != nil {")
		w.wtr.WriteLine("return nil, err")
		w.wtr.WriteLine("}")

		w.wtr.WriteLine("if nextEntityIsNil{")
		w.wtr.WriteLine("return v.Null(), nil")
		w.wtr.WriteLine("}")

		w.wtr.WriteLine("return nextEntity.GetValue(v, path[1:], columnName)")
	} else {
		w.wtr.WriteLine("return nil, fmt.Errorf(\"unable to get value: invalid path\")")
	}

	w.wtr.WriteLine("}")

	w.wtr.WriteLine("val, err := m.getValue(columnName, v)")
	w.wtr.WriteLine("if err != nil {")
	w.wtr.WriteLine("return nil, err")
	w.wtr.WriteLine("}")

	w.wtr.WriteLine("if val == nil{")
	w.wtr.WriteLine("return v.Null(), nil")
	w.wtr.WriteLine("}")

	w.wtr.WriteLine("return val, nil")
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) WriteGetRelatedEntityFunction(modelStructName string, tableData *builderrepo.TableMetadata) {
	if !hasManyToOneRelationships(tableData) {
		return
	}

	w.wtr.WriteLine("// getRelatedEntity returns the value from N:1/1:1 relationships as a TableModel")
	w.wtr.WriteLine("// It will return an error for invalid relationships and relationship types")

	w.wtr.WriteLine("func (m *%s) getRelatedEntity(", modelStructName)
	w.wtr.WriteLine(" relationship querymodel.RelationshipMetadata,")
	w.wtr.WriteLine(") (querymodel.TableModel, bool, error) {")

	w.wtr.WriteLine("switch relationship.Id {")
	for _, relationship := range tableData.Relationships {
		if relationship.Type != "querymodel.RelationshipManyToOne" {
			continue
		}
		w.wtr.WriteLine("case %s:", strLit(relationship.Id))
		w.wtr.WriteLine("v := m.%s", snakeToPascal(relationship.ExpansionColumnName))
		w.wtr.WriteLine("return v, v == nil, nil")
	}

	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return nil, true, fmt.Errorf(")
	w.wtr.WriteLine("  \"unable to get related entity: unsupported relationship '%%s'\",")
	w.wtr.WriteLine(" relationship.Id,")
	w.wtr.WriteLine(")")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) WriteGetValueExpressionPrivateFunction(modelStructName string, tableData *builderrepo.TableMetadata) {
	w.wtr.WriteLine("// getValue returns the value from a given column")

	w.wtr.WriteLine("func (m *%s) getValue(", modelStructName)
	w.wtr.WriteLine("  columnName string,")
	w.wtr.WriteLine("  v querymodel.ValueBuilder,")
	w.wtr.WriteLine(") (querymodel.Value, error) {")

	w.wtr.WriteLine("switch columnName{")
	for _, column := range tableData.Columns {
		w.wtr.WriteLine("case %s:", strLit(column.Name))
		w.wtr.WriteLine("%s", w.buildGetValueExpression(column))
	}

	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return nil, fmt.Errorf(\"unsupported column: '%%s'\", columnName)")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) WriteProjectionFunction(modelStructName string, tableData *builderrepo.TableMetadata) {
	w.wtr.WriteLine("// Project adds a column to the model's projection set")

	w.wtr.WriteLine("func (m *%s) Project(columnName string) error {", modelStructName)
	w.wtr.WriteLine("return (&m.projection).Add(columnName)")
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) buildGetValueExpression(columnData *builderrepo.ColumnMetadata) string {
	columnIdentifier := snakeToPascal(columnData.Name)
	goType := msqlTypeToGoType(columnData.Type, columnData)

	if columnData.EnumName != "" {
		return fmt.Sprintf("return v.Int(int64(m.%s)), nil", columnIdentifier)
	}

	switch goType {
	case "string":
		return fmt.Sprintf("return v.String(m.%s), nil", columnIdentifier)
	case "int64":
		return fmt.Sprintf("return v.Int(m.%s), nil", columnIdentifier)
	case "*time.Time":
		return fmt.Sprintf("return v.DateTime(*m.%s), nil", columnIdentifier)
	case "*querymodel.Point":
		return fmt.Sprintf("return v.Point(m.%s.Coordinates[0], m.%s.Coordinates[1]), nil", columnIdentifier, columnIdentifier)
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

func (w *queryModelWriter) WriteTableModelProjection(tableData *builderrepo.TableMetadata) {
	projectionName := getModelProjectionName(tableData.Name)

	if len(tableData.Columns) > 64 {
		panic("model writer currently only supports tables with up to 64 columns")
	}

	// Write the type
	w.wtr.WriteLine(
		"// %s represents column projection for the %s table.",
		projectionName,
		tableData.Name,
	)
	w.wtr.WriteLine("type %s uint64", projectionName)
	w.wtr.NewLine()

	// Write the constants
	w.wtr.WriteLine("const (")
	for i, col := range tableData.Columns {
		identifier := getColProjectionId(tableData.Name, col.Name)

		if i == 0 {
			w.wtr.WriteLine("%s %s = 1 << iota", identifier, projectionName)
		} else {
			w.wtr.WriteLine("%s", identifier)
		}
	}

	for _, rel := range tableData.Relationships {
		identifier := getColProjectionId(tableData.Name, rel.ExpansionColumnName)
		w.wtr.WriteLine("%s", identifier)
	}

	w.wtr.WriteLine(")")
	w.wtr.NewLine()

	//Write setter
	w.wtr.WriteLine("// Add includes a given column within the projection")

	w.wtr.WriteLine("func (p *%s) Add(columnName string) error {", projectionName)

	w.wtr.WriteLine("switch columnName{")
	for _, column := range tableData.Columns {
		w.wtr.WriteLine("case %s:", strLit(column.Name))
		w.wtr.WriteLine("*p |= %s", getColProjectionId(tableData.Name, column.Name))
	}

	for _, rel := range tableData.Relationships {
		w.wtr.WriteLine("case %s:", strLit(rel.ExpansionColumnName))
		w.wtr.WriteLine("*p |= %s", getColProjectionId(tableData.Name, rel.ExpansionColumnName))
	}

	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return fmt.Errorf(\"unsupported column: '%%s'\", columnName)")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("return nil")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	w.wtr.WriteLine("// Has is used to determine if a given column is in a projection")

	w.wtr.WriteLine("func (p %s) Has(projection %s) bool {", projectionName, projectionName)
	w.wtr.WriteLine("return p&projection != 0")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// WRITE IS EMPTY
	w.wtr.WriteLine("// IsEmpty is used to determine if there are no projections")
	w.wtr.WriteLine("func (p %s) IsEmpty() bool {", projectionName)
	w.wtr.WriteLine("return p == 0")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()
}

func (w *queryModelWriter) writeTableAccessStructs() {
	// Database Access Policy
	w.wtr.WriteLine("// AccessPolicy defines the access policy for the database")

	w.wtr.WriteLine("type DatabaseAccessPolicy struct {")
	for _, t := range w.metadata.Tables {
		tableAccessPolicyName := getTableAccessPolicyName(t.Name)
		w.wtr.WriteLine("%s *%s", tableAccessPolicyName, tableAccessPolicyName)
	}
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// Table access policy getter
	w.wtr.WriteLine("// GetTableAccessPolicy returns an access policy for a given table for nil")
	w.wtr.WriteLine("// if the table does not exist")

	w.wtr.WriteLine("func (p *DatabaseAccessPolicy) GetTableAccessPolicy(")
	w.wtr.WriteLine("  tableName string,")
	w.wtr.WriteLine(") (querymodel.TableAccessPolicy, bool) {")

	w.wtr.WriteLine("switch tableName {")
	for _, t := range w.metadata.Tables {
		w.wtr.WriteLine("case %s:", strLit(t.Name))
		w.wtr.WriteLine("return p.%s, true", getTableAccessPolicyName(t.Name))
	}
	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return nil, false")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// Individual Table Access Policy struct definitions
	for i, t := range w.metadata.Tables {
		if i != 0 {
			w.wtr.NewLine()
		}
		w.writeTableAccessStruct(t)
	}
}

func (w *queryModelWriter) writeTableAccessStruct(t *builderrepo.TableMetadata) {
	accessPolicyStructName := getTableAccessPolicyName(t.Name)
	// Access Policy Struct
	w.wtr.WriteLine(
		"// %s defines an access policy for the %s table",
		accessPolicyStructName,
		t.Name,
	)

	w.wtr.WriteLine("type %s struct {", accessPolicyStructName)
	w.wtr.WriteLine("UserCanAccess bool")
	for _, c := range t.Columns {
		w.wtr.WriteLine("%s bool", snakeToPascal(c.Name))
	}
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// Can Access
	w.wtr.WriteLine("// CanAccess defines, at the table level, if a user can perform any")
	w.wtr.WriteLine("// operations on the %s table. Specific column access policies can restrict", t.Name)
	w.wtr.WriteLine("// access given but cannot override access denied")

	w.wtr.WriteLine("func (p *%s) CanAccess() bool {", accessPolicyStructName)
	w.wtr.WriteLine("return p.UserCanAccess")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// GetColumn
	w.wtr.WriteLine("// CanAccessColumn returns true if a column can be accessed else false. An error")
	w.wtr.WriteLine("// is returned if the column does not exist on the table")

	w.wtr.WriteLine("func (p *%s) CanAccessColumn(", accessPolicyStructName)
	w.wtr.WriteLine("  columnName string,")
	w.wtr.WriteLine(") (bool, error) {")

	w.wtr.WriteLine("switch columnName {")
	for _, c := range t.Columns {
		w.wtr.WriteLine("case %s:", strLit(c.Name))
		w.wtr.WriteLine("return p.%s, nil", snakeToPascal(c.Name))
	}
	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return false, fmt.Errorf(")
	w.wtr.WriteLine("  \"column %%s does not exists on the %s table\",", t.Name)
	w.wtr.WriteLine("  columnName,")
	w.wtr.WriteLine(")")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

}

func getTableAccessPolicyName(tableName string) string {
	return fmt.Sprintf("%sAccessPolicy", snakeToPascal(tableName))
}

func getModelStructName(tableName string) string {
	return fmt.Sprintf("%sModel", snakeToPascal(tableName))
}

func getModelCollectionType(tableName string) string {
	modelStructName := getModelStructName(tableName)
	return fmt.Sprintf("%ss", modelStructName)
}

func getModelProjectionName(tableName string) string {
	return fmt.Sprintf("%sProjection", snakeToPascal(tableName))
}

func modelMetadataStoreIdentifier(tableName string) string {
	return fmt.Sprintf("%sMetadata", snakeToPascal(tableName))
}

func getModelResourceStructName(tableName string) string {
	return fmt.Sprintf("%sResource", snakeToPascal(tableName))
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

var geometryTypeRegex = regexp.MustCompile(
	`(?i)STGeometryType\]\(\)\s*=\s*'([^']+)'`,
)

func msqlTypeToGoType(typeName builderrepo.MsqlDataTypeName, columnData *builderrepo.ColumnMetadata) string {
	switch typeName {
	case MsqlTypeNvarchar:
		return "string"
	case MsqlTypeInt:
		return "int64"
	case MsqlTypeTinyInt:
		if columnData.EnumName != "" {
			return enumTypeName(columnData.EnumName)
		} else {
			return "uint8"
		}
	case MsqlTypeDateTimeOffset:
		return "*time.Time"
	case MsqlTypeGeography:
		geometryType := ""
		for _, currentCheck := range columnData.Checks {
			matches := geometryTypeRegex.FindStringSubmatch(currentCheck.Definition)

			if len(matches) == 2 {
				geometryType = matches[1]
			}
		}
		switch geometryType {
		case "Point":
			return "*querymodel.Point"
		default:
			panic("unsupported geometry type: " + geometryType)
		}

	default:
		return ""
	}
}

func buildTableRelationshipData(schema *builderrepo.DatabaseSchema) {
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
				table.Relationships = map[string]*builderrepo.TableRelationship{}
			}
			table.Relationships[tableRelationship.ColumnName] = tableRelationship
		}
	}
}

func getTableRelationship(
	table *builderrepo.TableMetadata,
	relationship *builderrepo.RelationshipMetadata,
) *builderrepo.TableRelationship {

	if relationship.ParentTable != table.Name && relationship.ReferencedTable != table.Name {
		return nil
	}

	r := &builderrepo.TableRelationship{
		Id: getTableRelationshipIdentifier(relationship),
	}

	if relationship.ParentTable == table.Name {
		if !strings.HasSuffix(relationship.ParentColumn, "_id") {
			log.Fatalf("invalid parent column: %s. Parent columns must be suffixes with _id", relationship.ParentColumn)
		}
		r.ExpansionColumnName = strings.TrimSuffix(relationship.ParentColumn, "_id")

		r.Type = "querymodel.RelationshipManyToOne"
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

		r.Type = "querymodel.RelationshipOneToMany"
		r.RelatedTable = relationship.ParentTable
		r.LocalColumn = relationship.ReferencedColumn
		r.ForeignColumn = relationship.ParentColumn

		r.RelationshipColumnType = getModelCollectionType(r.RelatedTable)
	}

	return r
}

func getTableRelationshipIdentifier(relationship *builderrepo.RelationshipMetadata) string {
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

func hasManyToOneRelationships(tableData *builderrepo.TableMetadata) bool {
	for _, r := range tableData.Relationships {
		if r.Type == "querymodel.RelationshipManyToOne" {
			return true
		}
	}
	return false
}
