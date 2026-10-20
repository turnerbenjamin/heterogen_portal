package modelwriter

import (
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderrepo"
)

type constantsWriter struct {
	path     string
	metadata *builderrepo.DatabaseSchema
	wtr      *writer
}

func NewConstantsWriter(
	path string,
	metadata *builderrepo.DatabaseSchema,
) *constantsWriter {
	return &constantsWriter{
		path:     path,
		metadata: metadata,
		wtr:      &writer{&strings.Builder{}},
	}
}

func (w *constantsWriter) Write() {
	buildTableRelationshipData(w.metadata)
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

	w.writeColumnTypeConstants()
	w.wtr.NewLine()

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
	for i, table := range w.metadata.Tables {
		if i != 0 {
			w.wtr.NewLine()
		}

		for _, column := range table.Columns {
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
	for i, table := range w.metadata.Tables {
		if i != 0 {
			w.wtr.NewLine()
		}

		for _, column := range table.Columns {
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

func (w *constantsWriter) writeColumnTypeConstants() {
	w.wtr.WriteLine("// Constants representing column data types")
	w.wtr.WriteLine("const(")
	for i, table := range w.metadata.Tables {
		if i != 0 {
			w.wtr.NewLine()
		}

		for _, column := range table.Columns {
			dataType := msqlTypeToDataTypeString(column)
			if dataType == "" {
				log.Fatal(fmt.Errorf("unable to convert type to go type: %s(%s)", column.Name, column.Type))
			}

			fieldNameConst := fmt.Sprintf(
				"ColType%s%s",
				snakeToPascal(table.Name),
				snakeToPascal(column.Name),
			)
			w.wtr.WriteLine("%s = %s", fieldNameConst, dataType)
		}
	}
	w.wtr.WriteLine(")")
}

func (w *constantsWriter) writeRelationshipNameConstants() {
	w.wtr.WriteLine("// Constants representing relationship columns names")
	w.wtr.WriteLine("const(")
	for i, table := range w.metadata.Tables {
		if i != 0 {
			w.wtr.NewLine()
		}

		for _, relationship := range table.Relationships {
			if relationship.Type != "querymodel.RelationshipManyToOne" {
				continue
			}

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
	for i, table := range w.metadata.Tables {
		if i != 0 {
			w.wtr.NewLine()
		}

		for _, relationship := range table.Relationships {
			if relationship.Type != "querymodel.RelationshipManyToOne" {
				continue
			}

			fieldNameConst := fmt.Sprintf(
				"RelColDisplayName%s%s",
				snakeToPascal(table.Name),
				snakeToPascal(relationship.ColumnName),
			)
			w.wtr.WriteLine("%s = %s", fieldNameConst, strLit(snakeToDisplayName(relationship.ExpansionColumnName)))
		}
	}
	w.wtr.WriteLine(")")
}

func (w *constantsWriter) writeMaxLengthConstants() {
	w.wtr.WriteLine("// Constants representing column max lengths")
	w.wtr.WriteLine("const(")
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
			w.wtr.WriteLine("%s = %d", constraintConst, column.MaxLength)
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

func msqlTypeToDataTypeString(columnData *builderrepo.ColumnMetadata) string {
	switch columnData.Type {
	case MsqlTypeNvarchar:
		return "DataTypeString"
	case MsqlTypeInt:
		return "DataTypeNumber"
	case MsqlTypeTinyInt:
		if columnData.EnumName != "" {
			return "DataTypeEnum"
		} else {
			return "DataTypeNumber"
		}
	case MsqlTypeDateTimeOffset:
		return "DataTypeTime"
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
			return "DataTypePoint"
		default:
			panic("unsupported geometry type: " + geometryType)
		}

	default:
		panic("unrecognised data type: " + columnData.Type)
	}
}
