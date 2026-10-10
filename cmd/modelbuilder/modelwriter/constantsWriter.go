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
	sb       *strings.Builder
}

func NewConstantsWriter(
	path string,
	metadata *builderrepo.DatabaseSchema,
) *constantsWriter {
	return &constantsWriter{
		path:     path,
		metadata: metadata,
		sb:       &strings.Builder{},
	}
}

func (w *constantsWriter) Write() {
	buildTableRelationshipData(w.metadata)
	w.writePackageDefinition()
	w.writeConstants()

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

func (w *constantsWriter) writeConstants() {

	w.writeColumnNameConstants()
	writeToBuilder(w.sb, "\n")

	// w.writeColumnTypeConstants()
	// writeToBuilder(w.sb, "\n")

	w.writeColumnDisplayNameConstants()
	writeToBuilder(w.sb, "\n")

	w.writeRelationshipNameConstants()
	writeToBuilder(w.sb, "\n")

	w.writeRelationshipDisplayNameConstants()
	writeToBuilder(w.sb, "\n")

	w.writeMaxLengthConstants()
}

func (w *constantsWriter) writePackageDefinition() {
	writeToBuilder(w.sb, `//GENERATED CODE
// see cmd/modelbuilder/
package model

type DataType string

const (
	DataTypeString DataType = "string"
	DataTypeNumber DataType = "number"
	DataTypeTime   DataType = "time"
	DataTypePoint  DataType = "point"
)
`)
}

func (w *constantsWriter) writeColumnNameConstants() {
	writeToBuilder(w.sb, "// Constants representing column names in the database\n")
	writeToBuilder(w.sb, "const(\n")
	for i, table := range w.metadata.Tables {
		if i != 0 {
			writeToBuilder(w.sb, "\n")
		}

		for _, column := range table.Columns {
			fieldNameConst := fmt.Sprintf(
				"Col%s%s",
				snakeToPascal(table.Name),
				snakeToPascal(column.Name),
			)
			writeToBuilder(w.sb, fmt.Sprintf("%s = \"%s\"\n", fieldNameConst, column.Name))
		}
	}
	writeToBuilder(w.sb, ")\n")
}

func (w *constantsWriter) writeColumnDisplayNameConstants() {
	writeToBuilder(w.sb, "// Constants representing column display names in the database\n")
	writeToBuilder(w.sb, "const(\n")
	for i, table := range w.metadata.Tables {
		if i != 0 {
			writeToBuilder(w.sb, "\n")
		}

		for _, column := range table.Columns {
			fieldNameConst := fmt.Sprintf(
				"ColDisplayName%s%s",
				snakeToPascal(table.Name),
				snakeToPascal(column.Name),
			)
			writeToBuilder(w.sb, fmt.Sprintf("%s = \"%s\"\n", fieldNameConst, snakeToDisplayName(column.Name)))
		}
	}
	writeToBuilder(w.sb, ")\n")
}

// func (w *constantsWriter) writeColumnTypeConstants() {
// 	writeToBuilder(w.sb, "// Constants representing column data types\n")
// 	writeToBuilder(w.sb, "const(\n")
// 	for i, table := range w.metadata.Tables {
// 		if i != 0 {
// 			writeToBuilder(w.sb, "\n")
// 		}

// 		for _, column := range table.Columns {
// 			dataType := msqlTypeToDataTypeString(column.Type, column.Checks)
// 			if dataType == "" {
// 				log.Fatal(fmt.Errorf("unable to convert type to go type: %s(%s)", column.Name, column.Type))
// 			}

// 			fieldNameConst := fmt.Sprintf(
// 				"ColType%s%s",
// 				snakeToPascal(table.Name),
// 				snakeToPascal(column.Name),
// 			)
// 			writeToBuilder(w.sb, fmt.Sprintf("%s = %s\n", fieldNameConst, dataType))
// 		}
// 	}
// 	writeToBuilder(w.sb, ")\n")
// }

func (w *constantsWriter) writeRelationshipNameConstants() {
	writeToBuilder(w.sb, "// Constants representing relationship columns names\n")
	writeToBuilder(w.sb, "const(\n")
	for i, table := range w.metadata.Tables {
		if i != 0 {
			writeToBuilder(w.sb, "\n")
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
			writeToBuilder(w.sb, fmt.Sprintf("%s = \"%s\"\n", fieldNameConst, relationship.ColumnName))
		}
	}
	writeToBuilder(w.sb, ")\n")
}

func (w *constantsWriter) writeRelationshipDisplayNameConstants() {
	writeToBuilder(w.sb, "// Constants representing relationship columns display names\n")
	writeToBuilder(w.sb, "const(\n")
	for i, table := range w.metadata.Tables {
		if i != 0 {
			writeToBuilder(w.sb, "\n")
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
			writeToBuilder(w.sb, fmt.Sprintf("%s = \"%s\"\n", fieldNameConst, snakeToDisplayName(relationship.ExpansionColumnName)))
		}
	}
	writeToBuilder(w.sb, ")\n")
}

func (w *constantsWriter) writeMaxLengthConstants() {
	writeToBuilder(w.sb, "// Constants representing column max lengths\n")
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
