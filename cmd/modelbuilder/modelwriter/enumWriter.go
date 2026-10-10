package modelwriter

import (
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderrepo"
)

type enumWriter struct {
	path   string
	writer writer
	enums  []*builderrepo.EnumMetadata
}

func NewEnumWriter(
	path string,
	enums []*builderrepo.EnumMetadata,
) *enumWriter {
	return &enumWriter{
		path:   path,
		writer: writer{},
		enums:  enums,
	}
}

func (w *enumWriter) Write() {
	w.writePackageDefinition()

	for i, enum := range w.enums {
		if i > 0 {
			w.writer.NewLine()
		}
		w.writeEnum(enum)
	}

	raw := w.writer.String()
	formatted, err := format.Source([]byte(raw))
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(w.path, formatted, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func (w *enumWriter) writePackageDefinition() {
	w.writer.Write(`//GENERATED CODE
// see cmd/modelbuilder/
package model

`)
}

func (w *enumWriter) writeEnum(enum *builderrepo.EnumMetadata) {
	enumType := enumTypeName(enum.Name)

	// Type
	w.writer.WriteLine("type %s uint8", enumType)
	w.writer.NewLine()

	// Constants
	w.writer.WriteLine("const (")
	for _, val := range enum.Values {
		valueName := enumValueName(enumType, val.DisplayName)
		w.writer.WriteLine("%s %s = %d", valueName, enumType, val.Value)
	}
	w.writer.WriteLine(")")
	w.writer.NewLine()

	// String Getter
	w.writer.WriteLine("func (e %s) String() (string, bool) {", enumType)
	w.writer.WriteLine("switch e {")
	for _, val := range enum.Values {
		w.writer.WriteLine("case %d:", val.Value)
		w.writer.WriteLine("return \"%s\", true", val.DisplayName)
	}
	w.writer.WriteLine("default:")
	w.writer.WriteLine("return \"\", false")
	w.writer.WriteLine("}")
	w.writer.WriteLine("}")
}

func enumTypeName(enumTableName string) string {
	return fmt.Sprintf("Enum%s", snakeToPascal(enumTableName))
}

func enumValueName(enumTypeName string, valueDisplayName string) string {
	valueSnakeName := strings.ReplaceAll(valueDisplayName, " ", "_")
	valueSnakeName = strings.ToLower(valueSnakeName)

	return fmt.Sprintf("%s%s", enumTypeName, snakeToPascal(valueSnakeName))
}
