package modelwriter

import (
	"fmt"
	"go/format"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderconfig"
	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/schemamodel"
)

type EnumOption struct {
}

type TableColumnDef struct {
	Identifier   string
	Path         string
	FilterPath   string
	DisplayName  string
	PropertyName string
	GetDataFunc  string
	AppColType   schemamodel.AppColType
	Expansion    string
}

type tableDefinitionWriter struct {
	config   builderconfig.BuilderConfig
	path     string
	metadata *schemamodel.Schema
	wtr      *writer
}

func NewTableDefinitionWriter(
	config builderconfig.BuilderConfig,
	path string,
	metadata *schemamodel.Schema,
) *tableDefinitionWriter {
	return &tableDefinitionWriter{
		config:   config,
		path:     path,
		metadata: metadata,
		wtr:      &writer{sb: &strings.Builder{}},
	}
}

func (w *tableDefinitionWriter) Write() error {
	w.writePackageDefinition()

	for _, t := range w.metadata.TableData.Ordered {
		var tableConfig *builderconfig.TableColumnDefinitionConfig = nil
		if config, exists := w.config.TableDefinitions[t.Name]; exists {
			tableConfig = &config
		}

		err := w.writeTableDefinition(t, tableConfig)
		if err != nil {
			return err
		}
	}

	raw := w.wtr.String()
	formatted, err := format.Source([]byte(raw))
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(w.path, formatted, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (w *tableDefinitionWriter) writePackageDefinition() {
	w.wtr.Write(`//GENERATED CODE
// see cmd/modelbuilder/
package model

import (
	"fmt"

	"github.com/turnerbenjamin/querystack/querymodel"
)

type AppColType interface {
	IsAppColType()
}

type AppColTypeString struct {
	MaxLength int16
}

func (t AppColTypeString) IsAppColType() {}

type AppColTypeNumber struct{}

func (t AppColTypeNumber) IsAppColType() {}

type AppColTypeTime struct{}

func (t AppColTypeTime) IsAppColType() {}

type AppColTypePoint struct{}

func (t AppColTypePoint) IsAppColType() {}

type AppColTypeIdentifier struct{}

func (t AppColTypeIdentifier) IsAppColType() {}

type AppColTypeEnum struct {
	Options []EnumOption
}

func (t AppColTypeEnum) IsAppColType() {}

type AppColTypeLookup struct {
	TableName string
}

func (t AppColTypeLookup) IsAppColType() {}

type TableDefinition [T querymodel.TableModel] struct{
  DisplayName string
	PluralDisplayName string
	MainGridColumns []TableColumn[T]
}

type TableColumn[T querymodel.TableModel] struct{
	Path string
	DisplayName string
	Type AppColType
	Expansion string
	GetData func(T) string
}
`)
}

func (w *tableDefinitionWriter) writeTableDefinition(
	table *schemamodel.Table,
	tableConfig *builderconfig.TableColumnDefinitionConfig,
) error {
	// Build exclude from main grid
	excludeFromMainGrid := map[string]struct{}{}
	if tableConfig != nil && len(tableConfig.ExcludeFromMainGrid) > 0 {
		for _, col := range tableConfig.ExcludeFromMainGrid {
			excludeFromMainGrid[col] = struct{}{}
		}
	}

	// Build column definitions for all standard tables
	tableColumnDefinitions := []*TableColumnDef{}
	for _, column := range table.ColumnData.Ordered {
		columnDefiniton, err := w.buildTableConfig(table, column.Name)
		if err != nil {
			return err
		}
		tableColumnDefinitions = append(tableColumnDefinitions, columnDefiniton)
	}

	// Build columns definitions for any related Columns
	if tableConfig != nil && len(tableConfig.RelatedColumns) > 0 {
		for _, path := range tableConfig.RelatedColumns {
			columnDefiniton, err := w.buildTableConfig(table, path)
			if err != nil {
				return err
			}
			tableColumnDefinitions = append(tableColumnDefinitions, columnDefiniton)
		}
	}

	// write table column definition
	w.wtr.WriteLine(
		"var %sTable = TableDefinition[*%s]{",
		snakeToPascal(table.Name),
		table.Identifiers.ModelStruct,
	)
	w.wtr.WriteLine("DisplayName: %s,", strLit(table.DisplayName))
	w.wtr.WriteLine("PluralDisplayName: %s,", strLit(table.PluralDisplayName))
	w.wtr.WriteLine(
		"MainGridColumns: []TableColumn[*%s]{",
		table.Identifiers.ModelStruct,
	)
	for _, col := range tableColumnDefinitions {
		if _, exists := excludeFromMainGrid[col.Path]; exists {
			continue
		}
		w.wtr.WriteLine("%s,", col.Identifier)
	}
	w.wtr.WriteLine("},")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// Write individual column definitions
	for _, colDef := range tableColumnDefinitions {
		w.wtr.WriteLine(
			"var %s = TableColumn[*%s]{",
			colDef.Identifier,
			table.Identifiers.ModelStruct,
		)

		appType, err := w.writeTypeData(colDef.AppColType)
		if err != nil {
			return err
		}

		w.wtr.WriteLine("DisplayName: %s,", strLit(colDef.DisplayName))
		w.wtr.WriteLine("Path: %s,", strLit(colDef.Path))
		w.wtr.WriteLine("Expansion: %s,", strLit(colDef.Expansion))
		w.wtr.WriteLine("Type: %s,", appType)
		w.wtr.WriteLine("GetData: %s,", colDef.GetDataFunc)
		w.wtr.WriteLine("}")
		w.wtr.NewLine()
	}

	return nil
}

func (w *tableDefinitionWriter) writeTypeData(
	appColType schemamodel.AppColType,
) (string, error) {
	twtr := writer{&strings.Builder{}}

	switch t := appColType.(type) {

	case schemamodel.AppColTypeString:
		twtr.WriteLine("AppColTypeString {")
		twtr.WriteLine("MaxLength: int16(%d),", t.MaxLenth)
		twtr.Write("}")

	case schemamodel.AppColTypeNumber:
		twtr.Write("AppColTypeNumber {}")

	case schemamodel.AppColTypeTime:
		twtr.Write("AppColTypeTime {}")

	case schemamodel.AppColTypePoint:
		twtr.Write("AppColTypePoint {}")

	case schemamodel.AppColTypeIdentifier:
		twtr.Write("AppColTypeIdentifier {}")

	case schemamodel.AppColTypeEnum:
		twtr.WriteLine("AppColTypeEnum {")
		twtr.WriteLine("Options: %s,", t.EnumData.OptionSetName)
		twtr.Write("}")

	case schemamodel.AppColTypeLookup:
		twtr.WriteLine("AppColTypeLookup {")
		twtr.WriteLine("TableName: %s,", strLit(t.Relationship.RelatedTable.Name))
		twtr.Write("}")
	}

	return twtr.String(), nil
}

func (w *tableDefinitionWriter) buildTableConfig(
	table *schemamodel.Table,
	columnPath string,
) (*TableColumnDef, error) {
	pathComponents := strings.Split(columnPath, "/")
	if len(pathComponents) == 0 {
		return nil, fmt.Errorf("unable to build table column def: invalid path: %s", columnPath)
	}

	endResource := table
	traversalSteps := []*schemamodel.Relationship{}
	i := 0
	for i < (len(pathComponents) - 1) {
		intermediateComponent := pathComponents[i]

		var relationship *schemamodel.Relationship = nil
		for _, r := range endResource.RelationshipData.ManyToOneRelationships {
			if r.ColumnName == intermediateComponent {
				relationship = r
				break
			}
		}

		if relationship == nil {
			return nil, fmt.Errorf("unable to find related table from col: %s", intermediateComponent)
		}
		endResource = relationship.RelatedTable
		traversalSteps = append(traversalSteps, relationship)

		i++
	}

	columnName := pathComponents[len(pathComponents)-1]
	colData, exists := endResource.ColumnData.Map[columnName]
	if !exists {
		return nil, fmt.Errorf("unable to find column data for: %s", columnName)
	}

	getDataFuncVal, err := buildGetDataFunctionStatement(
		table,
		traversalSteps,
		colData,
	)
	if err != nil {
		return nil, err
	}

	return &TableColumnDef{
		Identifier: fmt.Sprintf(
			"%s%sColumn",
			snakeToPascal(table.Name),
			snakeToPascal(strings.ReplaceAll(columnPath, "/", "_")),
		),
		DisplayName:  buildColumnDisplayName(traversalSteps, colData),
		PropertyName: colData.Identifiers.StructProperty,
		Path:         columnPath,
		GetDataFunc:  getDataFuncVal,
		AppColType:   colData.TypeData.AppType,
		Expansion:    buildExpansionStatement(pathComponents, colData),
	}, nil

}

func buildColumnDisplayName(
	traversalSteps []*schemamodel.Relationship,
	column *schemamodel.Column,
) string {
	if len(traversalSteps) == 0 {
		return column.DisplayName
	}

	step := traversalSteps[0]
	return fmt.Sprintf(
		"%s(%s)",
		step.Column.DisplayName,
		buildColumnDisplayName(traversalSteps[1:], column),
	)
}

func buildGetDataFunctionStatement(
	table *schemamodel.Table,
	traversalSteps []*schemamodel.Relationship,
	column *schemamodel.Column,
) (string, error) {
	switch c := column.TypeData.AppType.(type) {
	case schemamodel.AppColTypeLookup:
		traversalSteps = append(traversalSteps, c.Relationship)
		column = c.Relationship.RelatedTable.PrimaryColumn
	}

	fwtr := writer{&strings.Builder{}}

	fwtr.WriteLine("func(r *%s) string {", table.Identifiers.ModelStruct)

	resourceId := "r"
	for i, step := range traversalSteps {
		stepId := fmt.Sprintf("intermediateVal%d", i+1)
		fwtr.WriteLine(
			"%s := %s.%s",
			stepId,
			resourceId,
			step.ExpansionColumn.Identifiers.StructProperty,
		)
		fwtr.WriteLine("if %s == nil {", stepId)
		fwtr.WriteLine("return %s", strLit("-"))
		fwtr.WriteLine("}")
		fwtr.NewLine()

		resourceId = stepId
	}

	fwtr.WriteLine("v := %s.%s", resourceId, column.Identifiers.StructProperty)

	switch column.TypeData.AppType.(type) {
	case schemamodel.AppColTypeString:
		fwtr.WriteLine("return v")

	case schemamodel.AppColTypeIdentifier,
		schemamodel.AppColTypeLookup:
		fwtr.WriteLine("return v.String()")

	case schemamodel.AppColTypeEnum:
		fwtr.WriteLine("label, exists := v.String()")
		fwtr.WriteLine("if !exists{")
		fwtr.WriteLine("  return %s", strLit("-"))
		fwtr.WriteLine("}")
		fwtr.WriteLine("return label")

	case schemamodel.AppColTypeNumber:
		fwtr.Write("return strconv.FormatInt(int64(v), 10)")

	case schemamodel.AppColTypePoint:
		fwtr.Write(
			"return fmt.Sprintf(\"[%%f, %%f]\", v.Coordinates[0], v.Coordinates[1])",
		)

	case schemamodel.AppColTypeTime:
		fwtr.Write("return v.Format(%s)", strLit("2 Jan 06 15:04"))

	default:
		return "", fmt.Errorf(
			"unable to build get data function: unsupported app type",
		)

	}

	fwtr.Write("}")

	return fwtr.String(), nil
}

func buildExpansionStatement(pathComponents []string, column *schemamodel.Column) string {
	segments := slices.Clone(pathComponents)
	switch c := column.TypeData.AppType.(type) {
	case schemamodel.AppColTypeLookup:
		segments = append(segments, c.Relationship.RelatedTable.PrimaryColumn.Name)
	}

	if len(segments) == 1 {
		return ""
	}
	return buildExpansion(segments)
}

func buildExpansion(pathComponents []string) string {
	if len(pathComponents) == 1 {
		return fmt.Sprintf("select=%s", pathComponents[0])
	}

	return fmt.Sprintf(
		"expand=%s(%s)",
		pathComponents[0],
		buildExpansion(pathComponents[1:]),
	)
}
