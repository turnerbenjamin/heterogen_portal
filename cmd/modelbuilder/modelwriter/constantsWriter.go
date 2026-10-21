package modelwriter

import (
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/schemamodel"
)

type constantsWriter struct {
	path     string
	metadata *schemamodel.Schema
	wtr      *writer
}

func NewConstantsWriter(
	path string,
	metadata *schemamodel.Schema,
) *constantsWriter {
	return &constantsWriter{
		path:     path,
		metadata: metadata,
		wtr:      &writer{&strings.Builder{}},
	}
}

func (w *constantsWriter) Write() {
	w.writePackageDefinition()
	w.writeConstants()

	raw := w.wtr.String()
	formatted, err := format.Source([]byte(raw))
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(w.path, formatted, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func (w *constantsWriter) writeConstants() {

	w.writeColumnNameConstants()
	w.wtr.NewLine()

	// w.writeColumnTypeConstants()
	// w.wtr.NewLine()

	w.writeColumnDisplayNameConstants()
	w.wtr.NewLine()

	w.writeRelationshipNameConstants()
	w.wtr.NewLine()

	w.writeRelationshipDisplayNameConstants()
	w.wtr.NewLine()

	w.writeMaxLengthConstants()
}

func (w *constantsWriter) writePackageDefinition() {
	w.wtr.WriteLine(`//GENERATED CODE
// see cmd/modelbuilder/
package model

type DataType string

const (
	DataTypeString DataType = "string"
	DataTypeNumber DataType = "number"
	DataTypeEnum DataType = "enum"
	DataTypeTime   DataType = "time"
	DataTypePoint  DataType = "point"
)`)
}

func (w *constantsWriter) writeColumnNameConstants() {
	w.wtr.WriteLine("// Constants representing column names in the database")
	w.wtr.WriteLine("const(")
	for i, table := range w.metadata.TableData.Ordered {
		if i != 0 {
			w.wtr.NewLine()
		}

		for _, column := range table.ColumnData.Ordered {
			fieldNameConst := fmt.Sprintf(
				"Col%s%s",
				snakeToPascal(table.Name),
				snakeToPascal(column.Name),
			)
			w.wtr.WriteLine("%s = %s", fieldNameConst, strLit(column.Name))
		}
	}
	w.wtr.WriteLine(")")
}

func (w *constantsWriter) writeColumnDisplayNameConstants() {
	w.wtr.WriteLine("// Constants representing column display names in the database")
	w.wtr.WriteLine("const(")
	for i, table := range w.metadata.TableData.Ordered {
		if i != 0 {
			w.wtr.NewLine()
		}

		for _, column := range table.ColumnData.Ordered {
			fieldNameConst := fmt.Sprintf(
				"ColDisplayName%s%s",
				snakeToPascal(table.Name),
				snakeToPascal(column.Name),
			)
			w.wtr.WriteLine("%s = %s", fieldNameConst, strLit(snakeToDisplayName(column.Name)))
		}
	}
	w.wtr.WriteLine(")")
}

// func (w *constantsWriter) writeColumnTypeConstants() {
// 	w.wtr.WriteLine("// Constants representing column data types")
// 	w.wtr.WriteLine("const(")
// 	for i, table := range w.metadata.TableData.Ordered {
// 		if i != 0 {
// 			w.wtr.NewLine()
// 		}

// 		for _, column := range table.ColumnData.Ordered {
// 			dataType := msqlTypeToDataTypeString(column)
// 			if dataType == "" {
// 				log.Fatal(fmt.Errorf("unable to convert type to go type: %s(%s)", column.Name, column.Type))
// 			}

// 			fieldNameConst := fmt.Sprintf(
// 				"ColType%s%s",
// 				snakeToPascal(table.Name),
// 				snakeToPascal(column.Name),
// 			)
// 			w.wtr.WriteLine("%s = %s", fieldNameConst, dataType)
// 		}
// 	}
// 	w.wtr.WriteLine(")")
// }

func (w *constantsWriter) writeRelationshipNameConstants() {
	w.wtr.WriteLine("// Constants representing relationship columns names")
	w.wtr.WriteLine("const(")
	for i, table := range w.metadata.TableData.Ordered {
		if i != 0 {
			w.wtr.NewLine()
		}

		for _, relationship := range table.RelationshipData.ManyToOneRelationships {
			fieldNameConst := fmt.Sprintf(
				"RelCol%s%s",
				snakeToPascal(table.Name),
				snakeToPascal(relationship.ColumnName),
			)
			w.wtr.WriteLine("%s = %s", fieldNameConst, strLit(relationship.ColumnName))
		}
	}
	w.wtr.WriteLine(")")
}

func (w *constantsWriter) writeRelationshipDisplayNameConstants() {
	w.wtr.WriteLine("// Constants representing relationship columns display names")
	w.wtr.WriteLine("const(")
	for i, table := range w.metadata.TableData.Ordered {
		if i != 0 {
			w.wtr.NewLine()
		}

		for _, relationship := range table.RelationshipData.ManyToOneRelationships {
			fieldNameConst := fmt.Sprintf(
				"RelColDisplayName%s%s",
				snakeToPascal(table.Name),
				snakeToPascal(relationship.ColumnName),
			)
			w.wtr.WriteLine("%s = %s", fieldNameConst, strLit(relationship.ExpansionColumn.DisplayName))
		}
	}
	w.wtr.WriteLine(")")
}

func (w *constantsWriter) writeMaxLengthConstants() {
	w.wtr.WriteLine("// Constants representing column max lengths")
	w.wtr.WriteLine("const(")
	for _, table := range w.metadata.TableData.Ordered {
		for _, column := range table.ColumnData.Ordered {
			switch col := column.TypeData.AppType.(type) {
			case schemamodel.AppColTypeString:
				constraintConst := fmt.Sprintf(
					"MaxLen%s%s",
					snakeToPascal(table.Name),
					snakeToPascal(column.Name),
				)
				w.wtr.WriteLine("%s = %d", constraintConst, col.MaxLenth)
			}
		}
	}

	w.wtr.WriteLine(")")
}

func snakeToDisplayName(s string) string {
	parts := strings.Split(s, "_")

	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}

	return strings.Join(parts, " ")
}
