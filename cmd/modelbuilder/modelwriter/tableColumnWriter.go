package modelwriter

import (
	"fmt"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderrepo"
)

var tableConfig = map[string][]string{
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

type DataType string

const (
	DataTypeString DataType = "string"
	DataTypeNumber DataType = "number"
	DataTypeEnum   DataType = "enum"
	DataTypeTime   DataType = "time"
	DataTypePoint  DataType = "point"
)

type EnumOption struct {
}

type TableColumnDef struct {
	Path        string
	DisplayName string
	DataType    DataType
	Expansion   string
	EnumOptions []string
}

type tableData struct {
	columns       map[string]*builderrepo.ColumnMetadata
	relationships map[string]*builderrepo.TableRelationship
}

type tableColumnWriter struct {
	path     string
	metadata *builderrepo.DatabaseSchema
	w        *writer
	tableMap map[string]*tableData
}

func newTableColumnWriter(
	path string,
	metadata *builderrepo.DatabaseSchema,
) *tableColumnWriter {
	tableMap := map[string]*tableData{}
	for _, table := range metadata.Tables {
		colMap := map[string]*builderrepo.ColumnMetadata{}
		for _, col := range table.Columns {
			colMap[col.Name] = col
		}

		relMap := map[string]*builderrepo.TableRelationship{}
		for _, rel := range table.Relationships {
			relMap[rel.LocalColumn] = rel
		}

		tableMap[table.Name] = &tableData{
			columns:       colMap,
			relationships: relMap,
		}
	}

	return &tableColumnWriter{
		path:     path,
		metadata: metadata,
		w:        &writer{sb: &strings.Builder{}},
		tableMap: tableMap,
	}
}

func (w *tableColumnWriter) Write() {
	w.writePackageDefinition()

	for k, v := range tableConfig {
		tableData, exists := w.tableMap[k]
		if !exists {
			err := fmt.Errorf("invalid table config: cannot find data for %s", k)
			panic(err)
		}
		w.writeTableConfig(tableData, v)
	}
}

func (w *tableColumnWriter) writePackageDefinition() {
	w.w.Write(`//GENERATED CODE
// see cmd/modelbuilder/
package model

type DataType string

const (
	DataTypeString DataType = "string"
	DataTypeNumber DataType = "number"
	DataTypeEnum DataType = "enum"
	DataTypeTime   DataType = "time"
	DataTypePoint  DataType = "point"
)

type TableColumn struct{
	Path string
	DisplayName string
	DataType DataType	
	Expansion string
	GetData func(T) string
}
`)
}

func (w *tableColumnWriter) writeTableConfig(
	table *tableData,
	columnPaths []string,
) {
	// for _, path := range columnPaths {
	// 	config, err := w.buildTableConfig(table, path)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	w.wr
	// }
}

func (w *tableColumnWriter) buildTableConfig(
	table *tableData,
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
		relatedData, exists := endResource.relationships[intermediateComponent]
		if !exists {
			return nil, fmt.Errorf("unable to find related table from col: %s", intermediateComponent)
		}

		relatedTableData, exists := w.tableMap[relatedData.Name]
		if !exists {
			return nil, fmt.Errorf("unable to find table data for: %s", relatedData.Name)
		}
		endResource = relatedTableData
	}

	columnName := pathComponents[len(pathComponents)-1]
	colData, exists := endResource.columns[columnName]
	if !exists {
		return nil, fmt.Errorf("unable to find column data for: %s", columnName)
	}

	return &TableColumnDef{
		Path:        columnPath,
		DisplayName: snakeToDisplayName(colData.Name),
		Expansion:   buildExpansionStatement(pathComponents),
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
