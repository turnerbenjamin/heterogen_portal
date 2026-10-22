/* Don't look at me, I'm ugly */
package modelwriter

import (
	"fmt"
	"go/format"
	"os"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderrepo"
	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/schemamodel"
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
	metadata *schemamodel.Schema
	wtr      *writer
}

func NewQueryModelWriter(
	path string,
	metadata *schemamodel.Schema,
) *queryModelWriter {
	return &queryModelWriter{
		path:     path,
		metadata: metadata,
		wtr:      &writer{sb: &strings.Builder{}},
	}
}

func (w *queryModelWriter) Write() error {
	w.writePackageAndStaticTypeDefinitions()
	w.wtr.NewLine()

	w.writeSchemaDefinition()

	for _, table := range w.metadata.TableData.Ordered {
		w.wtr.NewLine()

		w.writeResourceDefinition(table)
		w.wtr.NewLine()

		w.writeTableMetadataDefinition(table)
		w.wtr.NewLine()

		err := w.WriteTableModel(table)
		if err != nil {
			return err
		}
		w.wtr.NewLine()

		w.WriteTableModelProjection(table)
	}

	w.wtr.NewLine()
	w.writeTableAccessStructs()

	raw := w.wtr.String()
	formatted, err := format.Source([]byte(raw))
	if err != nil {
		return err
	}

	err = os.WriteFile(w.path, formatted, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (w *queryModelWriter) writePackageAndStaticTypeDefinitions() {
	w.wtr.Write(`//GENERATED CODE
// see cmd/modelbuilder/
package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
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

	for _, t := range w.metadata.TableData.Ordered {
		w.wtr.WriteLine("case \"%s\":", t.Name)
		w.wtr.WriteLine("return %s{}, true", t.Identifiers.Resource)
	}
	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return nil, false")

	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}\n")
	w.wtr.NewLine()
}

func (w *queryModelWriter) writeResourceDefinition(tableData *schemamodel.Table) {
	// Write resource type
	w.wtr.WriteLine("type %s struct{}", tableData.Identifiers.Resource)
	w.wtr.NewLine()

	// Write funcs
	w.writeResourceGetMetadataFunc(tableData)
	w.wtr.NewLine()

	w.writeResourceInitProjectionFunc(tableData)
	w.wtr.NewLine()

	w.WriteResourceSliceFromJSONFunction(tableData)
	w.wtr.NewLine()

	w.WriteResourceTypedSliceFromJSONFunction(tableData)
	w.wtr.NewLine()
}

func (w *queryModelWriter) writeResourceGetMetadataFunc(
	tableData *schemamodel.Table,
) {
	w.wtr.WriteLine(
		"func (r %s) GetMetadata() querymodel.TableMetadata {",
		tableData.Identifiers.Resource,
	)

	w.wtr.WriteLine("return %s", tableData.Identifiers.MetadataVar)
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) writeResourceInitProjectionFunc(
	tableData *schemamodel.Table,
) {
	w.wtr.WriteLine(
		"func (r %s) InitProjection() querymodel.Projection {",
		tableData.Identifiers.Resource,
	)
	w.wtr.WriteLine("return new(%s)", tableData.Identifiers.Projection)
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) WriteResourceSliceFromJSONFunction(
	tableData *schemamodel.Table,
) {
	w.wtr.WriteLine(
		"// SliceFromJSON unmarshals a json array of %s and returns it as a slice",
		tableData.Name,
	)

	w.wtr.WriteLine("func (m %s) SliceFromJSON(", tableData.Identifiers.Resource)
	w.wtr.WriteLine("  jsonData []byte,")
	w.wtr.WriteLine("  projectionNode *querymodel.ProjectionNode,")
	w.wtr.WriteLine(") ([]querymodel.TableModel, error) {")

	w.wtr.WriteLine("if len(jsonData) == 0 {")
	w.wtr.WriteLine("return []querymodel.TableModel{}, nil")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	w.wtr.WriteLine("var concreteSlice []*%s", tableData.Identifiers.ModelStruct)

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

func (w *queryModelWriter) WriteResourceTypedSliceFromJSONFunction(
	tableData *schemamodel.Table,
) {
	w.wtr.WriteLine(
		"// TypedSliceFromJSON unmarshals a json array of %s and returns it as a slice\n",
		tableData.Name,
	)

	w.wtr.WriteLine("func (m %s) TypedSliceFromJSON(", tableData.Identifiers.Resource)
	w.wtr.WriteLine("  jsonData []byte,")
	w.wtr.WriteLine("  projectionNode *querymodel.ProjectionNode,")
	w.wtr.WriteLine(") ([]*%s, error) {", tableData.Identifiers.ModelStruct)

	w.wtr.WriteLine("if len(jsonData) == 0 {")
	w.wtr.WriteLine("return []*%s{}, nil", tableData.Identifiers.ModelStruct)
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	w.wtr.WriteLine("var results []*%s", tableData.Identifiers.ModelStruct)
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

func (w *queryModelWriter) writeTableMetadataDefinition(
	tableData *schemamodel.Table,
) {
	columnMapValue := buildColumnMap(tableData)
	relationshipMapValue := w.buildRelationshipMap(tableData)

	w.wtr.WriteLine(
		"// %s contains the database metadata for the %s table.",
		tableData.Identifiers.MetadataVar,
		tableData.FullyQualifiedName,
	)

	w.wtr.WriteLine("var %s = querymodel.TableMetadata{", tableData.Identifiers.MetadataVar)
	w.wtr.WriteLine("Name: %s,", strLit(tableData.Name))
	w.wtr.WriteLine("FullyQualifiedName: %s,", strLit(tableData.FullyQualifiedName))

	w.wtr.WriteLine("PrimaryKeyColumn: querymodel.ColumnMetadata{")
	w.wtr.WriteLine("Name: %s,", strLit(tableData.ColumnData.PrimaryKey.Name))
	w.wtr.WriteLine("Type: querymodel.%s,", tableData.ColumnData.PrimaryKey.TypeData.QueryModelTypeIdentifier)
	w.wtr.WriteLine("},")

	w.wtr.WriteLine("Columns: %s,", columnMapValue)
	w.wtr.WriteLine("Relationships: %s,", relationshipMapValue)
	w.wtr.WriteLine("}")
	w.wtr.NewLine()
}

func buildColumnMap(tableData *schemamodel.Table) string {
	colMapWtr := &writer{sb: &strings.Builder{}}

	colMapWtr.WriteLine("map[string]querymodel.ColumnMetadata{")

	for _, col := range tableData.ColumnData.Ordered {
		colMapWtr.WriteLine("\"%s\": {", col.Name)
		colMapWtr.WriteLine("Name: \"%s\",", col.Name)
		colMapWtr.WriteLine("Type: querymodel.%s,", col.TypeData.QueryModelTypeIdentifier)
		colMapWtr.WriteLine("},")
	}
	colMapWtr.Write("}")

	return colMapWtr.String()
}

func (w *queryModelWriter) buildRelationshipMap(
	tableData *schemamodel.Table,
) string {
	relMapWtr := &writer{sb: &strings.Builder{}}

	relMapWtr.WriteLine("map[string]querymodel.RelationshipMetadata{")

	for _, relationship := range tableData.RelationshipData.Ordered {
		relMapWtr.WriteLine("%s: {", strLit(relationship.ColumnName))
		relMapWtr.WriteLine("Id: %s,", strLit(relationship.Id))
		relMapWtr.WriteLine("Type: %s,", strLit(string(relationship.Type)))
		relMapWtr.WriteLine("ColumnName: %s,", strLit(relationship.ColumnName))
		relMapWtr.WriteLine("ExpansionColumnName: %s,", strLit(relationship.ExpansionColumn.Name))
		relMapWtr.WriteLine("From: %s{},", tableData.Identifiers.Resource)
		relMapWtr.WriteLine("To: %s{},", relationship.RelatedTable.Identifiers.Resource)

		relMapWtr.WriteLine("FromColumn: querymodel.ColumnMetadata{")
		relMapWtr.WriteLine("  Name: %s,", strLit(relationship.Column.Name))
		relMapWtr.WriteLine("  Type: querymodel.DbTypeString,")
		relMapWtr.WriteLine("},")

		relMapWtr.WriteLine("ToColumn: querymodel.ColumnMetadata{")
		relMapWtr.WriteLine("  Name: %s,", strLit(relationship.RelatedColumn.Name))
		relMapWtr.WriteLine("  Type: querymodel.DbTypeString,")
		relMapWtr.WriteLine("},")

		relMapWtr.WriteLine("},")
	}
	relMapWtr.Write("}")

	return relMapWtr.String()
}

func (w *queryModelWriter) WriteTableModel(tableData *schemamodel.Table) error {
	w.wtr.WriteLine(
		"// %s represents a row from the %s table.",
		tableData.Identifiers.ModelStruct,
		tableData.FullyQualifiedName,
	)
	w.wtr.WriteLine("type %s struct {", tableData.Identifiers.ModelStruct)

	// All cols, including virtual columns
	for _, col := range tableData.ColumnData.All {
		identifier := snakeToPascal(col.Name)
		tag := fmt.Sprintf("`json:%s`", strLit(col.Name))
		w.wtr.WriteLine("%s %s %s", identifier, col.TypeData.GoTypeIdentifier, tag)
	}

	// projection column
	w.wtr.WriteLine("projection %s `json:\"-\"`", tableData.Identifiers.Projection)

	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	w.wtr.WriteLine(
		"type %s []*%s",
		tableData.Identifiers.CollectionType,
		tableData.Identifiers.ModelStruct,
	)
	w.wtr.NewLine()

	w.WriteSetProjectionFunction(tableData)
	w.wtr.NewLine()

	w.WriteModelMarshalJSONFunction(tableData)
	w.wtr.NewLine()

	w.WriteGetValueExpressionFunction(tableData)
	w.wtr.NewLine()

	w.WriteGetRelatedEntityFunction(tableData)
	w.wtr.NewLine()

	err := w.WriteGetValueExpressionPrivateFunction(tableData)
	if err != nil {
		return err
	}

	w.wtr.NewLine()

	w.WriteProjectionFunction(tableData)

	return nil
}

func (w *queryModelWriter) WriteSetProjectionFunction(
	tableData *schemamodel.Table,
) {
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

		w.wtr.WriteLine("var typedProjection *%s", tableData.Identifiers.Projection)
		w.wtr.WriteLine("switch p := projection.(type){")
		w.wtr.WriteLine("case *%s:", tableData.Identifiers.Projection)
		w.wtr.WriteLine("typedProjection = p")
		w.wtr.WriteLine("default:")
		w.wtr.WriteLine("return fmt.Errorf(\"unable to create new slice: invalid projection type received\")")
		w.wtr.WriteLine("}")
		w.wtr.NewLine()
	}

	concreteSetterIdentifier := fmt.Sprintf("set%s", tableData.Identifiers.Projection)

	// Write typing function for model
	receiverParam := "m"

	writeDocComment()
	writeSignature(receiverParam, "*"+tableData.Identifiers.ModelStruct)
	writeProjectionCast()
	w.wtr.WriteLine("return m.%s(*typedProjection, projectionNode.Children)", concreteSetterIdentifier)
	w.wtr.WriteLine("}")

	// write typing function for model collection
	receiverParam = "ms"

	writeDocComment()
	writeSignature(receiverParam, tableData.Identifiers.CollectionType)
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

	w.wtr.WriteLine(
		"func (m *%s) %s(",
		tableData.Identifiers.ModelStruct,
		concreteSetterIdentifier,
	)

	w.wtr.WriteLine("projection %s,", tableData.Identifiers.Projection)
	w.wtr.WriteLine("childNodes map[string]*querymodel.ProjectionNode,")
	w.wtr.WriteLine(") error {")

	w.wtr.WriteLine("m.projection = projection")
	for _, r := range tableData.RelationshipData.Ordered {

		w.wtr.WriteLine("if node, exists := childNodes[%s]; exists {", strLit(r.ExpansionColumn.Name))
		w.wtr.WriteLine("if m.%s != nil {", r.ExpansionColumn.Identifiers.StructProperty)
		w.wtr.WriteLine("if err := m.%s.SetProjection(node); err != nil{", r.ExpansionColumn.Identifiers.StructProperty)
		w.wtr.WriteLine("return err")
		w.wtr.WriteLine("}")
		w.wtr.Write("}")

		if r.Type == "querymodel.RelationshipOneToMany" {
			w.wtr.WriteLine("else{")
			w.wtr.WriteLine(
				"m.%s = %s{}",
				r.ExpansionColumn.Identifiers.StructProperty,
				tableData.Identifiers.CollectionType,
			)
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

func (w *queryModelWriter) WriteModelMarshalJSONFunction(
	tableData *schemamodel.Table,
) {
	w.wtr.WriteLine(
		"// MarshalJSON marshals %s to a json string based on the projection",
		tableData.Identifiers.ModelStruct,
	)
	w.wtr.WriteLine(
		"func (m *%s) MarshalJSON() ([]byte, error) {",
		tableData.Identifiers.ModelStruct,
	)

	w.wtr.WriteLine("var buf bytes.Buffer")
	w.wtr.WriteLine("buf.WriteByte('{')")
	w.wtr.NewLine()

	w.wtr.WriteLine("isFirst := true")

	writeColumn := func(isLastItem bool, column *schemamodel.Column) {
		w.wtr.WriteLine("if m.projection.Has(%s) {", column.Identifiers.Projection)
		w.wtr.WriteLine(
			"err := marshalProperty(&buf, isFirst, %s, m.%s)",
			strLit(column.Name),
			column.Identifiers.StructProperty,
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

	colCount := len(tableData.ColumnData.Ordered)
	for i, column := range tableData.ColumnData.All {
		isLastItem := (i == colCount-1)
		writeColumn(isLastItem, column)
	}

	w.wtr.WriteLine("buf.WriteByte('}')")
	w.wtr.WriteLine("return buf.Bytes(), nil")
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) WriteGetValueExpressionFunction(
	tableData *schemamodel.Table,
) {
	w.wtr.WriteLine("// GetValue returns the value from a given path")

	w.wtr.WriteLine("func (m *%s) GetValue(", tableData.Identifiers.ModelStruct)
	w.wtr.WriteLine(" v querymodel.ValueBuilder,")
	w.wtr.WriteLine(" path []*querymodel.TraversalStep,")
	w.wtr.WriteLine(" columnName string,")
	w.wtr.WriteLine(") (querymodel.Value, error){")

	w.wtr.WriteLine("if len(path) > 0 {")

	if len(tableData.RelationshipData.ManyToOneRelationships) > 0 {
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

func (w *queryModelWriter) WriteGetRelatedEntityFunction(
	tableData *schemamodel.Table,
) {
	if len(tableData.RelationshipData.ManyToOneRelationships) == 0 {
		return
	}

	w.wtr.WriteLine("// getRelatedEntity returns the value from N:1/1:1 relationships as a TableModel")
	w.wtr.WriteLine("// It will return an error for invalid relationships and relationship types")

	w.wtr.WriteLine("func (m *%s) getRelatedEntity(", tableData.Identifiers.ModelStruct)
	w.wtr.WriteLine(" relationship querymodel.RelationshipMetadata,")
	w.wtr.WriteLine(") (querymodel.TableModel, bool, error) {")

	w.wtr.WriteLine("switch relationship.Id {")
	for _, relationship := range tableData.RelationshipData.ManyToOneRelationships {
		w.wtr.WriteLine("case %s:", strLit(relationship.Id))
		w.wtr.WriteLine("v := m.%s", snakeToPascal(relationship.ExpansionColumn.Name))
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

func (w *queryModelWriter) WriteGetValueExpressionPrivateFunction(
	tableData *schemamodel.Table,
) error {
	w.wtr.WriteLine("// getValue returns the value from a given column")

	w.wtr.WriteLine("func (m *%s) getValue(", tableData.Identifiers.ModelStruct)
	w.wtr.WriteLine("  columnName string,")
	w.wtr.WriteLine("  v querymodel.ValueBuilder,")
	w.wtr.WriteLine(") (querymodel.Value, error) {")

	w.wtr.WriteLine("switch columnName{")
	for _, column := range tableData.ColumnData.Ordered {
		valueExpression, err := w.buildGetValueExpression(column)
		if err != nil {
			return err
		}

		w.wtr.WriteLine("case %s:", strLit(column.Name))
		w.wtr.WriteLine("%s", valueExpression)
	}

	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return nil, fmt.Errorf(\"unsupported column: '%%s'\", columnName)")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}")

	return nil
}

func (w *queryModelWriter) WriteProjectionFunction(
	tableData *schemamodel.Table,
) {
	w.wtr.WriteLine("// Project adds a column to the model's projection set")

	w.wtr.WriteLine(
		"func (m *%s) Project(columnName string) error {",
		tableData.Identifiers.ModelStruct,
	)
	w.wtr.WriteLine("return (&m.projection).Add(columnName)")
	w.wtr.WriteLine("}")
}

func (w *queryModelWriter) buildGetValueExpression(columnData *schemamodel.Column) (string, error) {
	columnIdentifier := snakeToPascal(columnData.Name)

	switch columnData.TypeData.AppType.(type) {
	case schemamodel.AppColTypeEnum:
		return fmt.Sprintf("return v.Int(int64(m.%s)), nil", columnIdentifier), nil
	}

	switch columnData.TypeData.GoTypeIdentifier {

	case schemamodel.GoColTypeId:
		return fmt.Sprintf("return v.String(m.%s.String()), nil", columnIdentifier), nil

	case schemamodel.GoColTypeString:
		return fmt.Sprintf("return v.String(m.%s), nil", columnIdentifier), nil

	case schemamodel.GoColTypeInt64:
		return fmt.Sprintf("return v.Int(m.%s), nil", columnIdentifier), nil

	case schemamodel.GoColTypeTime:
		return fmt.Sprintf("return v.DateTime(*m.%s), nil", columnIdentifier), nil

	case schemamodel.GoColTypePoint:
		return fmt.Sprintf(
			"return v.Point(m.%s.Coordinates[0], m.%s.Coordinates[1]), nil",
			columnIdentifier,
			columnIdentifier,
		), nil

	default:
		return "", fmt.Errorf(
			"unable to build value expression: unsupported type: %s",
			columnData.TypeData.GoTypeIdentifier,
		)
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

func (w *queryModelWriter) WriteTableModelProjection(
	tableData *schemamodel.Table,
) error {
	if len(tableData.ColumnData.Ordered) > 64 {
		return fmt.Errorf("model writer currently only supports tables with up to 64 columns")
	}

	// Write the type
	w.wtr.WriteLine(
		"// %s represents column projection for the %s table.",
		tableData.Identifiers.Projection,
		tableData.Name,
	)
	w.wtr.WriteLine("type %s uint64", tableData.Identifiers.Projection)
	w.wtr.NewLine()

	// Write the constants
	w.wtr.WriteLine("const (")
	for i, col := range tableData.ColumnData.All {
		if i == 0 {
			w.wtr.WriteLine(
				"%s %s = 1 << iota",
				col.Identifiers.Projection,
				tableData.Identifiers.Projection,
			)
		} else {
			w.wtr.WriteLine("%s", col.Identifiers.Projection)
		}
	}

	w.wtr.WriteLine(")")
	w.wtr.NewLine()

	//Write setter
	w.wtr.WriteLine("// Add includes a given column within the projection")

	w.wtr.WriteLine("func (p *%s) Add(columnName string) error {", tableData.Identifiers.Projection)

	w.wtr.WriteLine("switch columnName{")
	for _, column := range tableData.ColumnData.All {
		w.wtr.WriteLine("case %s:", strLit(column.Name))
		w.wtr.WriteLine("*p |= %s", column.Identifiers.Projection)
	}

	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return fmt.Errorf(\"unsupported column: '%%s'\", columnName)")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("return nil")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	w.wtr.WriteLine("// Has is used to determine if a given column is in a projection")

	w.wtr.WriteLine(
		"func (p %s) Has(projection %s) bool {",
		tableData.Identifiers.Projection,
		tableData.Identifiers.Projection,
	)
	w.wtr.WriteLine("return p&projection != 0")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// WRITE IS EMPTY
	w.wtr.WriteLine("// IsEmpty is used to determine if there are no projections")
	w.wtr.WriteLine("func (p %s) IsEmpty() bool {", tableData.Identifiers.Projection)
	w.wtr.WriteLine("return p == 0")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	return nil
}

func (w *queryModelWriter) writeTableAccessStructs() {
	// Database Access Policy
	w.wtr.WriteLine("// AccessPolicy defines the access policy for the database")

	w.wtr.WriteLine("type DatabaseAccessPolicy struct {")
	for _, t := range w.metadata.TableData.Ordered {
		w.wtr.WriteLine("%s *%s", t.Identifiers.AccessPolicyName, t.Identifiers.AccessPolicyName)
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

	for _, t := range w.metadata.TableData.Ordered {
		w.wtr.WriteLine("case %s:", strLit(t.Name))
		w.wtr.WriteLine("return p.%s, true", t.Identifiers.AccessPolicyName)
	}
	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return nil, false")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// Individual Table Access Policy struct definitions
	for i, t := range w.metadata.TableData.Ordered {
		if i != 0 {
			w.wtr.NewLine()
		}
		w.writeTableAccessStruct(t)
	}
}

func (w *queryModelWriter) writeTableAccessStruct(t *schemamodel.Table) {
	// Access Policy Struct
	w.wtr.WriteLine(
		"// %s defines an access policy for the %s table",
		t.Identifiers.AccessPolicyName,
		t.Name,
	)

	w.wtr.WriteLine("type %s struct {", t.Identifiers.AccessPolicyName)
	w.wtr.WriteLine("UserCanAccess bool")
	for _, c := range t.ColumnData.Ordered {
		w.wtr.WriteLine("%s bool", snakeToPascal(c.Name))
	}
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// Can Access
	w.wtr.WriteLine("// CanAccess defines, at the table level, if a user can perform any")
	w.wtr.WriteLine("// operations on the %s table. Specific column access policies can restrict", t.Name)
	w.wtr.WriteLine("// access given but cannot override access denied")

	w.wtr.WriteLine("func (p *%s) CanAccess() bool {", t.Identifiers.AccessPolicyName)
	w.wtr.WriteLine("return p.UserCanAccess")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// GetColumn
	w.wtr.WriteLine("// CanAccessColumn returns true if a column can be accessed else false. An error")
	w.wtr.WriteLine("// is returned if the column does not exist on the table")

	w.wtr.WriteLine("func (p *%s) CanAccessColumn(", t.Identifiers.AccessPolicyName)
	w.wtr.WriteLine("  columnName string,")
	w.wtr.WriteLine(") (bool, error) {")

	w.wtr.WriteLine("switch columnName {")
	for _, c := range t.ColumnData.Ordered {
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
