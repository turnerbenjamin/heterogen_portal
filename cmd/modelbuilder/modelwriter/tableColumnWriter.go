package modelwriter

import (
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/schemamodel"
)

var tableConfig = map[string][]string{
	"farm_fields": {
		"business_id/created_by_id/display_name",
	},
	"users": {
		"id",
		"given_name",
		"family_name",
		"display_name",
		"email_address",
		"created_at",
		"modified_at",
	},
	"businesses": {
		"id",
		"reference",
		"trading_name",
		"business_type",
		"cph_number",
		"email_address",
		"website_url",
		"address_line_1",
		"address_line_2",
		"town",
		"county",
		"country",
		"postcode",
		"created_at",
		"created_by_id/display_name",
		"modified_at",
		"modified_by_id/display_name",
	},
}

type EnumOption struct {
}

type TableColumnDef struct {
	Path         string
	DisplayName  string
	PropertyName string
	AppColType   schemamodel.AppColType
	Expansion    string
}

type tableColumnWriter struct {
	path     string
	metadata *schemamodel.Schema
	wtr      *writer
}

func NewTableColumnDefinitionWriter(
	path string,
	metadata *schemamodel.Schema,
) *tableColumnWriter {
	return &tableColumnWriter{
		path:     path,
		metadata: metadata,
		wtr:      &writer{sb: &strings.Builder{}},
	}
}

func (w *tableColumnWriter) Write() error {
	w.writePackageDefinition()

	for k, v := range tableConfig {
		tableData, exists := w.metadata.TableData.Map[k]
		if !exists {
			return fmt.Errorf("invalid table config: cannot find data for %s", k)
		}
		err := w.writeTableConfig(tableData, v)
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

func (w *tableColumnWriter) writePackageDefinition() {
	w.wtr.Write(`//GENERATED CODE
// see cmd/modelbuilder/
package model

import "github.com/turnerbenjamin/querystack/querymodel"

type AppColType interface {
	IsAppColType()
}

type AppColTypeString struct {
	MaxLenth int16
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
	EnumType string
}

func (t AppColTypeEnum) IsAppColType() {}

type AppColTypeLookup struct {
	TableName string
}

func (t AppColTypeLookup) IsAppColType() {}

type TableColumn[T querymodel.TableModel] struct{
	Path string
	DisplayName string
	Type AppColType
	Expansion string
	GetData func(T) string
}
`)
}

func (w *tableColumnWriter) writeTableConfig(
	table *schemamodel.Table,
	columnPaths []string,
) error {
	tableColumnDefinitions := []*TableColumnDef{}
	for _, path := range columnPaths {
		columnDefiniton, err := w.buildTableConfig(table, path)
		if err != nil {
			return err
		}
		tableColumnDefinitions = append(tableColumnDefinitions, columnDefiniton)
	}

	for _, colDef := range tableColumnDefinitions {
		w.wtr.WriteLine(
			"var %s%sColumn =  TableColumn[*%s]{",
			snakeToPascal(table.Name),
			snakeToPascal(strings.ReplaceAll(colDef.Path, "/", "_")),
			table.Identifiers.ModelStruct,
		)

		w.wtr.WriteLine("Path: %s,", strLit(colDef.Path))
		w.wtr.WriteLine("DisplayName: %s,", strLit(colDef.DisplayName))
		// w.wtr.WriteLine("Type: %s,", )
		w.wtr.WriteLine("Expansion: %s,", strLit(colDef.Expansion))

		w.wtr.WriteLine("}")
		w.wtr.NewLine()
	}

	return nil
}

func (w *tableColumnWriter) buildTableConfig(
	table *schemamodel.Table,
	columnPath string,
) (*TableColumnDef, error) {
	pathComponents := strings.Split(columnPath, "/")
	if len(pathComponents) == 0 {
		return nil, fmt.Errorf("unable to build table column def: invalid path: %s", columnPath)
	}

	endResource := table
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
		i++
	}

	columnName := pathComponents[len(pathComponents)-1]
	colData, exists := endResource.ColumnData.Map[columnName]
	if !exists {
		return nil, fmt.Errorf("unable to find column data for: %s", columnName)
	}

	return &TableColumnDef{
		DisplayName:  colData.DisplayName,
		PropertyName: colData.Identifiers.StructProperty,
		Path:         columnPath,
		AppColType:   colData.TypeData.AppType,
		Expansion:    buildExpansionStatement(pathComponents),
	}, nil

}

func buildExpansionStatement(pathComponents []string) string {
	if len(pathComponents) == 1 {
		return ""
	}
	return buildExpansion(pathComponents)
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
