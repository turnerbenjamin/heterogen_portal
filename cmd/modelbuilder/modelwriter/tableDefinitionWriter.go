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

type Expansion struct {
	Path   string
	Select string
}

type TableColumnDef struct {
	Identifier         string
	ValuePath          string
	FilterPath         string
	Expansion          *Expansion
	DisplayName        string
	PropertyName       string
	BuildInputFunc     string
	BuildTableCellFunc string
	AppColType         schemamodel.AppColType
	Enums              []*schemamodel.Enum
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
	"time"

	"github.com/a-h/templ"
	"github.com/turnerbenjamin/querystack/querymodel"
)

type AppInputType string
type AppDataType string

const (
	AppTypeString     AppInputType = "string"
	AppTypeInteger     AppInputType = "integer"
	AppTypeFloat     AppInputType = "float"
	AppTypeDate       AppInputType = "date"
	AppTypeEnum       AppInputType = "enum"
	AppTypeIdentifier AppInputType = "identifier"
	AppTypeLookup     AppInputType = "lookup"
	AppTypePoint      AppInputType = "point"
)

type AppColType interface {
	GetInputType() AppInputType
	GetDataType() AppDataType
}

type AppColTypeString struct {
	MaxLength int16
}

func (t AppColTypeString) GetInputType() AppInputType {
	return AppTypeString
}

func (t AppColTypeString) GetDataType() AppDataType {
	return AppDataType(AppTypeString)
}

type AppColTypeInteger struct{}

func (t AppColTypeInteger) GetInputType() AppInputType {
	return AppTypeInteger
}

func (t AppColTypeInteger) GetDataType() AppDataType {
	return AppDataType(AppTypeInteger)
}

type AppColTypeFloat struct{}

func (t AppColTypeFloat) GetInputType() AppInputType {
	return AppTypeFloat
}

func (t AppColTypeFloat) GetDataType() AppDataType {
	return AppDataType(AppTypeFloat)
}

type AppColTypeDate struct{}

func (t AppColTypeDate) GetInputType() AppInputType {
	return AppTypeDate
}

func (t AppColTypeDate) GetDataType() AppDataType {
	return AppDataType(AppTypeDate)
}

type AppColTypePoint struct{}

func (t AppColTypePoint) GetInputType() AppInputType {
	return AppTypePoint
}

func (t AppColTypePoint) GetDataType() AppDataType {
	return AppDataType(AppTypePoint)
}

type AppColTypeIdentifier struct{}

func (t AppColTypeIdentifier) GetInputType() AppInputType {
	return AppTypeIdentifier
}

func (t AppColTypeIdentifier) GetDataType() AppDataType {
	return AppDataType(AppTypeIdentifier)
}

type AppColTypeEnum struct {
	TypeName string
	Options  []EnumOption
}

func (t AppColTypeEnum) GetInputType() AppInputType {
	return AppTypeEnum
}

func (t AppColTypeEnum) GetDataType() AppDataType {
	return AppDataType(t.TypeName)
}

type AppColTypeLookup struct {
	TableName string
}

func (t AppColTypeLookup) GetInputType() AppInputType {
	return AppTypeLookup
}

func (t AppColTypeLookup) GetDataType() AppDataType {
	return AppDataType(AppTypeLookup)
}


type UiInputControlBuilder interface {
	StringInput(v *string) templ.Component
	DateInput(v *time.Time) templ.Component
	EnumInput(typeName string, options []EnumOption, v *uint8) templ.Component
	IdentifierInput(v *string) templ.Component
	LookupInput(tableName string, v *string) templ.Component
	IntegerInput(v *int64) templ.Component
	FloatInput(v *float64) templ.Component
	PointInput(v *querymodel.Point) templ.Component
}

type UiTableCellBuilder interface {
	StringCell(v *string) templ.Component
	DateCell(v *time.Time) templ.Component
	EnumCell(v *string) templ.Component
	IdentifierCell(v *string) templ.Component
	LookupCell(v *string) templ.Component
	IntegerCell(v *int64) templ.Component
	FloatCell(v *float64) templ.Component
	PointCell(v *querymodel.Point) templ.Component
}

type TableDefinition[T querymodel.TableModel] struct {
	DisplayName        string
	PluralDisplayName  string
	MainGridColumns    []TableColumn[T]
	MainGridColumnsMap map[string]TableColumn[T]
	EnumTypes          []AppColTypeEnum
}

type Expansion struct {
	Path   string
	Select string
}

type TableColumn[T querymodel.TableModel] struct{
	DisplayName string
	Type AppColType

	Path string
	Expansion *Expansion

	BuildInput func(T, UiInputControlBuilder) templ.Component
	BuildTableCell func(T, UiTableCellBuilder) templ.Component
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

	slices.SortFunc(
		tableColumnDefinitions,
		func(a *TableColumnDef, b *TableColumnDef) int {
			return strings.Compare(a.DisplayName, b.DisplayName)
		},
	)

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
		if _, exists := excludeFromMainGrid[col.ValuePath]; exists {
			continue
		}
		w.wtr.WriteLine("%s,", col.Identifier)
	}
	w.wtr.WriteLine("},")
	w.wtr.WriteLine(
		"MainGridColumnsMap: map[string]TableColumn[*%s]{",
		table.Identifiers.ModelStruct,
	)
	for _, col := range tableColumnDefinitions {
		if _, exists := excludeFromMainGrid[col.ValuePath]; exists {
			continue
		}
		w.wtr.WriteLine("%s: %s,", strLit(col.ValuePath), col.Identifier)
	}
	w.wtr.WriteLine("},")
	w.wtr.WriteLine("EnumTypes: []AppColTypeEnum{")
	for _, e := range table.Enums {
		w.wtr.WriteLine("{")
		w.wtr.WriteLine("TypeName: %s,", strLit(e.TypeName))
		w.wtr.WriteLine("Options: %s,", e.OptionSetName)
		w.wtr.Write("}")
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
		w.wtr.WriteLine("Path: %s,", strLit(colDef.ValuePath))
		if colDef.Expansion != nil {
			w.wtr.WriteLine("Expansion: &Expansion{")
			w.wtr.WriteLine("  Path: %s,", strLit(colDef.Expansion.Path))
			w.wtr.WriteLine("  Select: %s,", strLit(colDef.Expansion.Select))
			w.wtr.WriteLine("},")
		}
		w.wtr.WriteLine("Type: %s,", appType)
		w.wtr.WriteLine("BuildInput: %s,", colDef.BuildInputFunc)
		w.wtr.WriteLine("BuildTableCell: %s,", colDef.BuildTableCellFunc)
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

	case schemamodel.AppColTypeInteger:
		twtr.Write("AppColTypeNumber {}")

	case schemamodel.AppColTypeDate:
		twtr.Write("AppColTypeDate {}")

	case schemamodel.AppColTypePoint:
		twtr.Write("AppColTypePoint {}")

	case schemamodel.AppColTypeIdentifier:
		twtr.Write("AppColTypeIdentifier {}")

	case schemamodel.AppColTypeEnum:
		twtr.WriteLine("AppColTypeEnum {")
		twtr.WriteLine("TypeName: %s,", strLit(t.EnumData.TypeName))
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

	buildInputFuncVal, err := generateBuilderFunc(
		table,
		traversalSteps,
		colData,
		"UiInputControlBuilder",
		writeCallInputBuilder,
	)
	if err != nil {
		return nil, err
	}

	buildTableCellFuncVal, err := generateBuilderFunc(
		table,
		traversalSteps,
		colData,
		"UiTableCellBuilder",
		writeCallTableCellBuilder,
	)
	if err != nil {
		return nil, err
	}

	enums := make([]*schemamodel.Enum, 0, len(table.Enums))
	for _, e := range table.Enums {
		enums = append(enums, e)
	}
	slices.SortFunc(enums, func(a *schemamodel.Enum, b *schemamodel.Enum) int {
		return strings.Compare(a.DbName, b.DbName)
	})

	return &TableColumnDef{
		Identifier: fmt.Sprintf(
			"%s%sColumn",
			snakeToPascal(table.Name),
			snakeToPascal(strings.ReplaceAll(columnPath, "/", "_")),
		),
		DisplayName:        buildColumnDisplayName(traversalSteps, colData),
		PropertyName:       colData.Identifiers.StructProperty,
		ValuePath:          columnPath,
		BuildInputFunc:     buildInputFuncVal,
		BuildTableCellFunc: buildTableCellFuncVal,
		AppColType:         colData.TypeData.AppType,
		Expansion:          buildExpansionStatement(pathComponents, colData),
		Enums:              enums,
	}, nil

}

func buildColumnDisplayName(
	traversalSteps []*schemamodel.Relationship,
	column *schemamodel.Column,
) string {
	if len(traversalSteps) == 0 {
		return getColumnDisplayName(column)
	}

	step := traversalSteps[0]
	return fmt.Sprintf(
		"%s (%s)",
		getColumnDisplayName(step.Column),
		buildColumnDisplayName(traversalSteps[1:], column),
	)
}

func getColumnDisplayName(column *schemamodel.Column) string {
	switch c := column.TypeData.AppType.(type) {
	case schemamodel.AppColTypeLookup:
		return c.Relationship.ExpansionColumn.DisplayName
	default:
		return column.DisplayName
	}
}

func generateBuilderFunc(
	table *schemamodel.Table,
	traversalSteps []*schemamodel.Relationship,
	column *schemamodel.Column,
	builderTypeName string,
	writeCallToBuilder func(
		w *writer,
		builderIdentifier string,
		valueIdentifier string,
		column *schemamodel.Column,
		valueIsNil bool,
	) error,
) (string, error) {
	switch c := column.TypeData.AppType.(type) {
	case schemamodel.AppColTypeLookup:
		traversalSteps = append(traversalSteps, c.Relationship)
		column = c.Relationship.RelatedTable.PrimaryColumn
	}

	fwtr := &writer{&strings.Builder{}}

	resourceId := "r"
	builderId := "b"
	valueId := "v"

	fwtr.WriteLine(
		"func(%s *%s, %s %s) templ.Component {",
		resourceId,
		table.Identifiers.ModelStruct,
		builderId,
		builderTypeName,
	)

	for i, step := range traversalSteps {
		stepId := fmt.Sprintf("intermediateVal%d", i+1)
		fwtr.WriteLine(
			"%s := %s.%s",
			stepId,
			resourceId,
			step.ExpansionColumn.Identifiers.StructProperty,
		)
		fwtr.WriteLine("if %s == nil {", stepId)
		err := writeCallToBuilder(
			fwtr,
			builderId,
			valueId,
			column,
			true,
		)
		if err != nil {
			return "", err
		}
		fwtr.WriteLine("}")
		fwtr.NewLine()

		resourceId = stepId
	}

	fwtr.WriteLine("%s := %s.%s", valueId, resourceId, column.Identifiers.StructProperty)
	err := writeCallToBuilder(
		fwtr,
		builderId,
		valueId,
		column,
		false,
	)
	if err != nil {
		return "", err
	}

	fwtr.Write("}")

	return fwtr.String(), nil

}

func writeCallInputBuilder(
	w *writer,
	builderIdentifier string,
	valueIdentifier string,
	column *schemamodel.Column,
	valueIsNil bool,
) error {
	switch c := column.TypeData.AppType.(type) {
	case schemamodel.AppColTypeString:
		if valueIsNil {
			w.WriteLine("return %s.StringInput(nil)", builderIdentifier)
			return nil
		}
		w.WriteLine("return %s.StringInput(&%s)", builderIdentifier, valueIdentifier)
		return nil

	case schemamodel.AppColTypeDate:
		if valueIsNil {
			w.WriteLine("return %s.DateInput(nil)", builderIdentifier)
			return nil
		}
		w.WriteLine(
			"return %s.DateInput(%s)",
			builderIdentifier,
			valueIdentifier,
		)
		return nil

	case schemamodel.AppColTypeEnum:
		if valueIsNil {
			w.WriteLine(
				" return %s.EnumInput(%s, %s, nil)",
				builderIdentifier,
				strLit(c.EnumData.TypeName),
				c.EnumData.OptionSetName,
			)
			return nil
		}

		w.WriteLine("_, exists := %s.String()", valueIdentifier)
		w.WriteLine("if !exists{")
		w.WriteLine(
			" return %s.EnumInput(%s, %s, nil)",
			builderIdentifier,
			strLit(c.EnumData.TypeName),
			c.EnumData.OptionSetName,
		)
		w.WriteLine("}")
		w.WriteLine("cv := uint8(v)")
		w.WriteLine(
			" return %s.EnumInput(%s, %s, &cv)",
			builderIdentifier,
			strLit(c.EnumData.TypeName),
			c.EnumData.OptionSetName,
		)
		return nil

	case schemamodel.AppColTypeIdentifier:
		if valueIsNil {
			w.WriteLine("return %s.IdentifierInput(nil)", builderIdentifier)
			return nil
		}

		w.WriteLine("s := %s.String()", valueIdentifier)
		w.WriteLine(
			"return %s.IdentifierInput(&s)",
			builderIdentifier,
		)
		return nil

	case schemamodel.AppColTypeLookup:
		if valueIsNil {
			w.WriteLine("return %s.LookupInput(nil)", builderIdentifier)
			return nil
		}

		w.WriteLine("s := %s.String()", valueIdentifier)
		w.WriteLine(
			"return %s.LookupIdentifier(&s)",
			builderIdentifier,
		)
		return nil

	case schemamodel.AppColTypeInteger:
		if valueIsNil {
			w.WriteLine("return %s.NumberInput(nil)", builderIdentifier)
			return nil
		}

		w.Write("return %s.NumberInput(&%s)", builderIdentifier, valueIdentifier)
		return nil

	case schemamodel.AppColTypePoint:
		if valueIsNil {
			w.WriteLine("return %s.PointInput(nil)", builderIdentifier)
			return nil
		}

		w.Write("return %s.PointInput(%s)", builderIdentifier, valueIdentifier)
		return nil

	default:
		return fmt.Errorf(
			"unable to generate input Builder call: unsupported app type",
		)
	}
}

func writeCallTableCellBuilder(
	w *writer,
	builderIdentifier string,
	valueIdentifier string,
	column *schemamodel.Column,
	valueIsNil bool,
) error {
	switch column.TypeData.AppType.(type) {
	case schemamodel.AppColTypeString:
		if valueIsNil {
			w.WriteLine("return %s.StringCell(nil)", builderIdentifier)
			return nil
		}

		w.WriteLine("return %s.StringCell(&%s)", builderIdentifier, valueIdentifier)
		return nil

	case schemamodel.AppColTypeDate:
		if valueIsNil {
			w.WriteLine(
				"return %s.DateCell(nil)",
				builderIdentifier,
			)
			return nil
		}

		w.WriteLine("if v == nil {")
		w.WriteLine(
			"return %s.DateCell(nil)",
			builderIdentifier,
		)
		w.WriteLine("}")
		w.WriteLine(
			"return %s.DateCell(%s)",
			builderIdentifier,
			valueIdentifier,
		)
		return nil

	case schemamodel.AppColTypeEnum:
		if valueIsNil {
			w.WriteLine("return %s.EnumCell(nil)", builderIdentifier)
			return nil
		}

		w.WriteLine("label, exists := %s.String()", valueIdentifier)
		w.WriteLine("if !exists{")
		w.WriteLine("return %s.EnumCell(nil)", builderIdentifier)
		w.WriteLine("}")
		w.WriteLine("return %s.EnumCell(&label)", builderIdentifier)
		return nil

	case schemamodel.AppColTypeIdentifier:
		if valueIsNil {
			w.WriteLine("return %s.IdentifierCell(nil)", builderIdentifier)
			return nil
		}

		w.WriteLine("s := %s.String()", valueIdentifier)
		w.WriteLine(
			"return %s.IdentifierCell(&s)",
			builderIdentifier,
		)
		return nil

	case schemamodel.AppColTypeLookup:
		if valueIsNil {
			w.WriteLine("return %s.LookupCell(nil)", builderIdentifier)
			return nil
		}

		w.WriteLine("s := %s.String()", valueIdentifier)
		w.WriteLine(
			"return %s.LookupCell(&s)",
			builderIdentifier,
		)
		return nil

	case schemamodel.AppColTypeInteger:
		if valueIsNil {
			w.WriteLine("return %s.IntegerCell(nil)", builderIdentifier)
			return nil
		}

		w.Write("return %s.IntegerCell(&%s)", builderIdentifier, valueIdentifier)
		return nil

	case schemamodel.AppColTypePoint:
		if valueIsNil {
			w.WriteLine("return %s.PointCell(nil)", builderIdentifier)
			return nil
		}

		w.Write("return %s.PointCell(%s)", builderIdentifier, valueIdentifier)
		return nil

	default:
		return fmt.Errorf(
			"unable to generate table cell Builder call: unsupported app type",
		)
	}
}

func buildExpansionStatement(pathComponents []string, column *schemamodel.Column) *Expansion {
	segments := slices.Clone(pathComponents)
	switch c := column.TypeData.AppType.(type) {
	case schemamodel.AppColTypeLookup:
		segments = append(segments, c.Relationship.RelatedTable.PrimaryColumn.Name)
	}

	if len(segments) == 1 {
		return nil
	}

	return &Expansion{
		Path:   buildExpansion(0, segments[0:len(segments)-1]),
		Select: segments[len(segments)-1],
	}
}

func buildExpansion(depth int, pathComponents []string) string {
	operator := ""
	if depth > 1 {
		operator = "expand="
	}

	if len(pathComponents) == 1 {
		return fmt.Sprintf("%s%s", operator, pathComponents[0])
	}

	return fmt.Sprintf(
		"%s%s(%s)",
		operator,
		pathComponents[0],
		buildExpansion(depth+1, pathComponents[1:]),
	)
}
