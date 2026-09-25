/* Don't look at me, I'm ugly */
package modelWriter

import (
	"fmt"
	"go/format"
	"log"
	"os"
	"regexp"
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
	w.writeSchemaDefinition()

	for _, table := range w.metadata.Tables {
		w.writeNewLine()
		w.writeResourceDefinition(table)
		w.writeNewLine()
		w.writeTableMetadataDefinition(table)
		w.writeNewLine()
		w.WriteTableModel(table)
		w.writeNewLine()
		w.WriteTableModelProjection(table)
	}
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
	"fmt"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
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

	writeToBuilder(w.sb, ")\n")
}

func (w *modelWriter) writeSchemaDefinition() {

	writeToBuilder(w.sb, "type schema struct{}\n\n")

	writeToBuilder(w.sb, "func NewSchema() schema{\n")
	writeToBuilder(w.sb, "return schema{}\n")
	writeToBuilder(w.sb, "}\n\n")

	writeToBuilder(w.sb, "func (s schema) GetResource(resourceName string) (queryModel.Resource, bool){\n")
	writeToBuilder(w.sb, "switch resourceName{\n")
	for _, t := range w.metadata.Tables {
		tableName := t.Name
		resourceName := getModelResourceStructName(tableName)
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", tableName))
		writeToBuilder(w.sb, fmt.Sprintf("return %s{}, true\n", resourceName))
	}
	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil, false\n")

	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "}\n\n")
}

func (w *modelWriter) writeResourceDefinition(tableData *builderRepo.TableMetadata) {
	resourceTypeName := getModelResourceStructName(tableData.Name)

	// Write resource type
	writeToBuilder(w.sb, fmt.Sprintf("type %s struct{}\n\n", resourceTypeName))

	// Write funcs
	w.writeResourceGetMetadataFunc(resourceTypeName, tableData)
	w.writeNewLine()

	w.writeResourceInitModelFunc(resourceTypeName, tableData)
	w.writeNewLine()

	w.writeResourceInitProjectionFunc(resourceTypeName, tableData)
	w.writeNewLine()
}

func (w *modelWriter) writeResourceGetMetadataFunc(
	resourceTypeName string,
	tableData *builderRepo.TableMetadata,
) {
	metadataIdentifier := modelMetadataStoreIdentifier(tableData.Name)
	writeToBuilder(w.sb, fmt.Sprintf("func (r %s) GetMetadata() queryModel.TableMetadata {\n", resourceTypeName))
	writeToBuilder(w.sb, fmt.Sprintf("return %s\n", metadataIdentifier))
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) writeResourceInitModelFunc(
	resourceTypeName string,
	tableData *builderRepo.TableMetadata,
) {
	modelStructIdentifier := getModelStructName(tableData.Name)
	writeToBuilder(w.sb, fmt.Sprintf("func (r %s) InitModel() queryModel.TableModel { \n", resourceTypeName))
	writeToBuilder(w.sb, fmt.Sprintf("return new(%s)", modelStructIdentifier))
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) writeResourceInitProjectionFunc(
	resourceTypeName string,
	tableData *builderRepo.TableMetadata,
) {
	modelProjectionIdentifier := getModelProjectionName(tableData.Name)
	writeToBuilder(w.sb, fmt.Sprintf("func (r %s) InitProjection() queryModel.Projection { \n", resourceTypeName))
	writeToBuilder(w.sb, fmt.Sprintf("return new(%s)", modelProjectionIdentifier))
	writeToBuilder(w.sb, "}\n")
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
	writeToBuilder(w.sb, fmt.Sprintf("var %s = queryModel.TableMetadata{\n", metadataIdentifer))
	writeToBuilder(w.sb, fmt.Sprintf("Name: \"%s\",\n", tableData.Name))
	writeToBuilder(w.sb, fmt.Sprintf("FullyQualifiedName: \"%s.%s\",\n", tableData.Schema, tableData.Name))

	writeToBuilder(w.sb, "PrimaryKeyColumn: queryModel.ColumnMetadata{\n")
	writeToBuilder(w.sb, fmt.Sprintf("Name: \"%s\",\n", primaryKey))
	writeToBuilder(w.sb, "Type: queryModel.DbTypeString,\n")
	writeToBuilder(w.sb, "},\n")

	writeToBuilder(w.sb, fmt.Sprintf("Columns: %s,\n", columnMapValue))
	writeToBuilder(w.sb, fmt.Sprintf("Relationships: %s,\n", relationshipMapValue))
	writeToBuilder(w.sb, "}\n\n")
}

func buildColumnMap(tableData *builderRepo.TableMetadata) (string, string) {
	primaryKey := ""
	sbColMap := &strings.Builder{}

	writeToBuilder(sbColMap, "map[string]queryModel.ColumnMetadata{\n")

	// DB Columns
	for _, col := range tableData.Columns {
		colType := getColTypeString(col.Type, col.Checks)
		if col.PrimaryKey == 1 {
			primaryKey = col.Name
		}

		writeToBuilder(sbColMap, fmt.Sprintf("\"%s\": {\n", col.Name))
		writeToBuilder(sbColMap, fmt.Sprintf("Name: \"%s\",\n", col.Name))
		writeToBuilder(sbColMap, fmt.Sprintf("Type: %s,\n", colType))
		writeToBuilder(sbColMap, "},\n")
	}

	writeToBuilder(sbColMap, "}")

	return primaryKey, sbColMap.String()
}

func (w *modelWriter) buildRelationshipMap(tableData *builderRepo.TableMetadata) string {
	sbR := &strings.Builder{}

	writeToBuilder(sbR, "map[string]queryModel.RelationshipMetadata{\n")
	for key, relationship := range tableData.Relationships {
		fromResource := getModelResourceStructName(tableData.Name)
		toResource := getModelResourceStructName(relationship.RelatedTable)

		writeToBuilder(sbR, fmt.Sprintf("\"%s\": {\n", key))
		writeToBuilder(sbR, fmt.Sprintf("Id: \"%s\",\n", relationship.Id))
		writeToBuilder(sbR, fmt.Sprintf("Type: %s,\n", relationship.Type))
		writeToBuilder(sbR, fmt.Sprintf("ColumnName: \"%s\",\n", relationship.ColumnName))
		writeToBuilder(sbR, fmt.Sprintf("ExpansionColumnName: \"%s\",\n", relationship.ExpansionColumnName))
		writeToBuilder(sbR, fmt.Sprintf("From: %s{},\n", fromResource))
		writeToBuilder(sbR, fmt.Sprintf("To: %s{},\n", toResource))

		writeToBuilder(sbR, "FromColumn: queryModel.ColumnMetadata{\n")
		writeToBuilder(sbR, fmt.Sprintf("Name: \"%s\",\n", relationship.LocalColumn))
		writeToBuilder(sbR, "Type: queryModel.DbTypeString,\n")
		writeToBuilder(sbR, "},\n")
		writeToBuilder(sbR, "ToColumn: queryModel.ColumnMetadata{\n")
		writeToBuilder(sbR, fmt.Sprintf("Name: \"%s\",\n", relationship.ForeignColumn))
		writeToBuilder(sbR, "Type: queryModel.DbTypeString,\n")
		writeToBuilder(sbR, "},\n")
		writeToBuilder(sbR, "},\n")
	}

	writeToBuilder(sbR, "}")
	return sbR.String()
}

func getColTypeString(typeName builderRepo.MsqlDataTypeName, checks []builderRepo.Check) string {
	goType := msqlTypeToGoType(typeName, checks)

	switch goType {
	case "string":
		return "queryModel.DbTypeString"
	case "int64":
		return "queryModel.DbTypeInt"
	case "*time.Time":
		return "queryModel.DbTypeDateTime"
	case "*queryModel.Point":
		return "queryModel.DbTypePoint"
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

		goType := msqlTypeToGoType(col.Type, col.Checks)
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

	// collection type
	collectionType := getModelCollectionType(tableData.Name)
	writeToBuilder(w.sb, fmt.Sprintf("type %s []*%s\n\n", collectionType, structName))

	w.WriteGetSliceGetterFunction(structName, tableData)
	w.writeNewLine()
	w.WriteSetProjectionFunction(structName, tableData)
	w.writeNewLine()
	w.WriteModelMarshalJSONFunction(structName, tableData)
	w.writeNewLine()
	w.WriteGetValueExpressionFunction(structName, tableData)
	w.writeNewLine()
	w.WriteGetRelatedEntityFunction(structName, tableData)
	w.writeNewLine()
	w.WriteGetValueExpressionPrivateFunction(structName, tableData)
	w.writeNewLine()
	w.WriteProjectionFunction(structName, tableData)
}

func (w *modelWriter) WriteGetSliceGetterFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, fmt.Sprintf(
		"// NewSlice unmarshals a json array of %s and returns it as a slice\n",
		tableData.Name,
	))

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) NewSlice(jsonData []byte, projectionNode *queryModel.ProjectionNode) ([]queryModel.TableModel, error) {\n", modelStructName))

	writeToBuilder(w.sb, "if len(jsonData) == 0 {\n")
	writeToBuilder(w.sb, "return []queryModel.TableModel{}, nil\n")
	writeToBuilder(w.sb, "}\n\n")

	writeToBuilder(w.sb, fmt.Sprintf("var concreteSlice []*%s\n", modelStructName))

	writeToBuilder(w.sb, "if err := json.Unmarshal(jsonData, &concreteSlice); err != nil {\n")
	writeToBuilder(w.sb, "return nil, err\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "result := make([]queryModel.TableModel, len(concreteSlice))\n\n")

	writeToBuilder(w.sb, "for i := range concreteSlice {\n")
	writeToBuilder(w.sb, "concreteSlice[i].SetProjection(projectionNode)\n")
	writeToBuilder(w.sb, "result[i] = concreteSlice[i]\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "return result, nil\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteSetProjectionFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeDocComment := func() {
		writeToBuilder(
			w.sb,
			"// SetProjection sets a projection node and sets projection for itself and any \n"+
				"// child nodes\n",
		)
	}

	writeSignature := func(receiverParam string, typeName string) {
		writeToBuilder(w.sb, fmt.Sprintf("func (%s %s) SetProjection(projectionNode *queryModel.ProjectionNode) error {\n", receiverParam, typeName))
	}

	writeProjectionCast := func() {
		writeToBuilder(w.sb, "projection := projectionNode.Projection\n\n")

		projectionName := getModelProjectionName(tableData.Name)
		writeToBuilder(w.sb, fmt.Sprintf("var typedProjection *%s\n", projectionName))
		writeToBuilder(w.sb, "switch p := projection.(type){\n")
		writeToBuilder(w.sb, fmt.Sprintf("case *%s:\n", projectionName))
		writeToBuilder(w.sb, "typedProjection = p\n")
		writeToBuilder(w.sb, "default:\n")
		writeToBuilder(w.sb, "return fmt.Errorf(\"unable to create new slice: invalid projection type received\")\n")
		writeToBuilder(w.sb, "}\n\n")
	}

	collectionType := getModelCollectionType(tableData.Name)
	concreteSetterIdentifier := fmt.Sprintf("set%s", getModelProjectionName(tableData.Name))

	// Write typing function for model
	receiverParam := "m"

	writeDocComment()
	writeSignature(receiverParam, "*"+modelStructName)
	writeProjectionCast()
	writeToBuilder(w.sb, fmt.Sprintf("return m.%s(*typedProjection, projectionNode.Children)\n", concreteSetterIdentifier))
	writeToBuilder(w.sb, "}\n")

	// write typing function for model collection
	receiverParam = "ms"

	writeDocComment()
	writeSignature(receiverParam, collectionType)
	writeProjectionCast()

	writeToBuilder(w.sb, "for _, m := range ms{\n")
	writeToBuilder(w.sb, fmt.Sprintf("if err := m.%s(*typedProjection, projectionNode.Children); err != nil{\n", concreteSetterIdentifier))
	writeToBuilder(w.sb, "return err\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "return nil\n")
	writeToBuilder(w.sb, "}\n")

	// write concrete setter function
	writeToBuilder(w.sb, fmt.Sprintf(
		"// %s sets a projection node and sets projection for itself and any \n"+
			"// child nodes\n",
		concreteSetterIdentifier,
	))

	projectionType := getModelProjectionName(tableData.Name)
	writeToBuilder(w.sb, fmt.Sprintf(
		"func (m *%s) %s(projection %s, childNodes map[string]*queryModel.ProjectionNode) error {\n",
		modelStructName,
		concreteSetterIdentifier,
		projectionType,
	))

	writeToBuilder(w.sb, "m.projection = projection\n")
	for _, r := range tableData.Relationships {
		expansionColumn := snakeToPascal(r.ExpansionColumnName)

		writeToBuilder(w.sb, fmt.Sprintf("if node, exists := childNodes[\"%s\"]; exists {\n", r.ExpansionColumnName))
		writeToBuilder(w.sb, fmt.Sprintf("if m.%s != nil {\n", expansionColumn))
		writeToBuilder(w.sb, fmt.Sprintf("if err := m.%s.SetProjection(node); err != nil{\n", expansionColumn))
		writeToBuilder(w.sb, "return err\n")
		writeToBuilder(w.sb, "}\n")
		writeToBuilder(w.sb, "}")
		if r.Type == "queryModel.RelationshipOneToMany" {
			writeToBuilder(w.sb, "else{\n")
			writeToBuilder(w.sb, fmt.Sprintf("m.%s = %s{}", expansionColumn, getModelCollectionType(r.RelatedTable)))
			writeToBuilder(w.sb, "}\n")
		} else {
			writeToBuilder(w.sb, "\n")
		}
		writeToBuilder(w.sb, "}\n\n")
	}

	writeToBuilder(w.sb, "return nil\n")
	writeToBuilder(w.sb, "}\n\n")
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

func (w *modelWriter) WriteGetValueExpressionFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, "// GetValue returns the value from a given path\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) GetValue(path []*queryModel.TraversalStep, columnName string, v queryModel.ValueBuilder) (queryModel.Value, error){\n", modelStructName))
	writeToBuilder(w.sb, "if len(path) > 0 {\n")
	writeToBuilder(w.sb, "nextStep := path[0]\n")
	writeToBuilder(w.sb, "nextEntity, nextEntityIsNil, err := m.getRelatedEntity(nextStep.Relationship)\n")

	writeToBuilder(w.sb, "if err != nil {\n")
	writeToBuilder(w.sb, "return nil, err\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "if nextEntityIsNil{\n")
	writeToBuilder(w.sb, "return v.Null(), nil\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "return nextEntity.GetValue(path[1:], columnName, v)\n")
	writeToBuilder(w.sb, "}\n")

	writeToBuilder(w.sb, "val, err := m.getValue(columnName, v)\n")
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
	writeToBuilder(w.sb, "// getRelatedEntity returns the value from N:1/1:1 relationships as a TableModel\n")
	writeToBuilder(w.sb, "// It will return an error for invalid relationships and relationship types\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) getRelatedEntity(relationship queryModel.RelationshipMetadata) (queryModel.TableModel, bool, error) {\n", modelStructName))
	writeToBuilder(w.sb, "switch relationship.Id {\n")

	for _, relationship := range tableData.Relationships {
		if relationship.Type != "queryModel.RelationshipManyToOne" {
			continue
		}
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", relationship.Id))
		writeToBuilder(w.sb, fmt.Sprintf("v := m.%s\n", snakeToPascal(relationship.ExpansionColumnName)))
		writeToBuilder(w.sb, "return v, v == nil, nil\n")
	}

	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil, true, fmt.Errorf(\"unable to get related entity: unsupported relationship '%s'\", relationship.Id)\n")
	writeToBuilder(w.sb, "}\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) WriteGetValueExpressionPrivateFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, "// getValue returns the value from a given column\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) getValue(columnName string, v queryModel.ValueBuilder) (queryModel.Value, error) {\n", modelStructName))
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

func (w *modelWriter) WriteProjectionFunction(modelStructName string, tableData *builderRepo.TableMetadata) {
	writeToBuilder(w.sb, "// Project adds a column to the model's projection set\n")

	writeToBuilder(w.sb, fmt.Sprintf("func (m *%s) Project(columnName string) error { \n", modelStructName))
	writeToBuilder(w.sb, "return (&m.projection).Add(columnName)\n")
	writeToBuilder(w.sb, "}\n")
}

func (w *modelWriter) buildGetValueExpression(columnData *builderRepo.ColumnMetadata) string {
	columnIdentifier := snakeToPascal(columnData.Name)
	goType := msqlTypeToGoType(columnData.Type, columnData.Checks)

	switch goType {
	case "string":
		return fmt.Sprintf("return v.String(m.%s), nil\n", columnIdentifier)
	case "int64":
		return fmt.Sprintf("return v.Int(m.%s), nil\n", columnIdentifier)
	case "*time.Time":
		return fmt.Sprintf("return v.DateTime(*m.%s), nil\n", columnIdentifier)
	case "*queryModel.Point":
		return fmt.Sprintf("return v.Point(m.%s.Coordinates[0], m.%s.Coordinates[1]), nil\n", columnIdentifier, columnIdentifier)
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

// func (w *modelWriter) WriteTableModelGetter() {
// 	writeToBuilder(
// 		w.sb,
// 		"// getTableModel returns a new instance of a given table or nill if \n"+
// 			"// the table name is invalid\n",
// 	)
// 	writeToBuilder(w.sb, "func getTableModel(tableName string) queryModel.TableModel{\n")
// 	writeToBuilder(w.sb, "switch tableName {\n")

// 	for _, table := range w.metadata.Tables {
// 		metadataStoreId := modelMetadataStoreIdentifier(table.Name)
// 		writeToBuilder(w.sb, fmt.Sprintf("case %s.name:\n", metadataStoreId))
// 		writeToBuilder(w.sb, fmt.Sprintf("return new(%s)\n", getModelStructName(table.Name)))
// 	}

// 	writeToBuilder(w.sb, "default:\n")
// 	writeToBuilder(w.sb, "return nil\n")
// 	writeToBuilder(w.sb, "}\n")
// 	writeToBuilder(w.sb, "}\n")
// }

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

	writeToBuilder(w.sb, "func (p *DatabaseAccessPolicy) GetTableAccessPolicy(\ntableName string,\n) (queryModel.TableAccessPolicy, bool) {\n")
	writeToBuilder(w.sb, "switch tableName {\n")
	for _, t := range w.metadata.Tables {
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", t.Name))
		writeToBuilder(w.sb, fmt.Sprintf("return p.%s, true\n", getTableAccessPolicyName(t.Name)))
	}
	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, "return nil, false\n")
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
		writeToBuilder(w.sb, fmt.Sprintf("%s bool\n", snakeToPascal(c.Name)))
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
		"// CanAccessColumn returns true if a column can be accessed else false. An error\n"+
			"// is returned if the column does not exist on the table\n",
	))
	writeToBuilder(w.sb, fmt.Sprintf("func (p *%s) CanAccessColumn(\ncolumnName string,\n) (bool, error) {\n", accessPolicyStructName))
	writeToBuilder(w.sb, "switch columnName {\n")
	for _, c := range t.Columns {
		writeToBuilder(w.sb, fmt.Sprintf("case \"%s\":\n", c.Name))
		writeToBuilder(w.sb, fmt.Sprintf("return p.%s, nil\n", snakeToPascal(c.Name)))
	}
	writeToBuilder(w.sb, "default:\n")
	writeToBuilder(w.sb, fmt.Sprintf("return false, fmt.Errorf(\"column %%s does not exists on the %s table\", columnName)\n", t.Name))
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

func getModelCollectionType(tableName string) string {
	modelStructName := getModelStructName(tableName)
	return fmt.Sprintf("%ss", modelStructName)
}

func getModelProjectionName(tableName string) string {
	return fmt.Sprintf("%sProjection", snakeToPascal(tableName))
}

func modelMetadataStoreIdentifier(tableName string) string {
	return fmt.Sprintf("%sMetadata", snakeToCamel(tableName))
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

func msqlTypeToGoType(typeName builderRepo.MsqlDataTypeName, checks []builderRepo.Check) string {
	switch typeName {
	case MsqlTypeNvarchar:
		return "string"
	case MsqlTypeInt:
		return "int64"
	case MsqlTypeDateTimeOffset:
		return "*time.Time"
	case MsqlTypeGeography:
		geometryType := ""
		for _, currentCheck := range checks {
			matches := geometryTypeRegex.FindStringSubmatch(currentCheck.Definition)

			if len(matches) == 2 {
				geometryType = matches[1]
			}
		}
		switch geometryType {
		case "Point":
			return "*queryModel.Point"
		default:
			panic("unsupported geometry type: " + geometryType)
		}

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

		r.RelationshipColumnType = getModelCollectionType(r.RelatedTable)
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
