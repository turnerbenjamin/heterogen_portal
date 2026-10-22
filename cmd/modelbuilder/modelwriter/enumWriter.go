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
	path  string
	wtr   writer
	enums []*builderrepo.EnumMetadata
}

func NewEnumWriter(
	path string,
	enums []*builderrepo.EnumMetadata,
) *enumWriter {
	return &enumWriter{
		path:  path,
		wtr:   writer{},
		enums: enums,
	}
}

func (w *enumWriter) Write() {
	w.writePackageDefinition()

	for i, enum := range w.enums {
		if i > 0 {
			w.wtr.NewLine()
		}
		w.writeEnum(enum)
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
}

func (w *enumWriter) writePackageDefinition() {
	w.wtr.Write(`//GENERATED CODE
// see cmd/modelbuilder/
package model

type EnumOption struct {
	DisplayName string
	Value       uint8
}

type Enum interface {
	String() (string, bool)
	Options() []EnumOption
}

`)
}

func (w *enumWriter) writeEnum(enum *builderrepo.EnumMetadata) {
	enumType := enumTypeName(enum.Name)

	// Type
	w.wtr.WriteLine("type %s uint8", enumType)
	w.wtr.NewLine()

	// Constants
	w.wtr.WriteLine("const (")
	for _, val := range enum.Values {
		valueName := enumValueName(enumType, val.DisplayName)
		w.wtr.WriteLine("%s %s = %d", valueName, enumType, val.Value)
	}
	w.wtr.WriteLine(")")
	w.wtr.NewLine()

	// String Getter
	w.wtr.WriteLine("func (e %s) String() (string, bool) {", enumType)
	w.wtr.WriteLine("switch e {")
	for _, val := range enum.Values {
		w.wtr.WriteLine("case %d:", val.Value)
		w.wtr.WriteLine("return \"%s\", true", val.DisplayName)
	}
	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return \"\", false")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// Options
	w.wtr.WriteLine("var %sOptions = []EnumOption {", enumType)
	for _, val := range enum.Values {
		w.wtr.WriteLine("{")
		w.wtr.WriteLine("DisplayName: %s,", strLit(val.DisplayName))
		w.wtr.WriteLine("Value: %d,", val.Value)
		w.wtr.WriteLine("},")
	}
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// Options Getter
	w.wtr.WriteLine("func (e %s) Options() []EnumOption{", enumType)
	w.wtr.WriteLine("  return %sOptions", enumType)
	w.wtr.WriteLine("}")
}

func enumTypeName(enumTableName string) string {
	return fmt.Sprintf("Enum%s", snakeToPascal(enumTableName))
}

func enumValueName(enumTypeName string, valueDisplayName string) string {
	valueSnakeName := strings.ReplaceAll(valueDisplayName, " ", "_")
	valueSnakeName = strings.ToLower(valueSnakeName)

	return fmt.Sprintf("%s%s", enumTypeName, snakeToPascal(valueSnakeName))
}
