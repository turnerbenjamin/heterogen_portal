package modelwriter

import (
	"go/format"
	"log"
	"os"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/schemamodel"
)

type enumWriter struct {
	path     string
	wtr      writer
	enumData schemamodel.EnumData
}

func NewEnumWriter(
	path string,
	enumData schemamodel.EnumData,
) *enumWriter {
	return &enumWriter{
		path:     path,
		wtr:      writer{},
		enumData: enumData,
	}
}

func (w *enumWriter) Write() {
	w.writePackageDefinition()

	for i, enum := range w.enumData.Ordered {
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

func (w *enumWriter) writeEnum(enum *schemamodel.Enum) {
	// Type
	w.wtr.WriteLine("type %s uint8", enum.TypeName)
	w.wtr.NewLine()

	// Constants
	w.wtr.WriteLine("const (")
	for _, val := range enum.Options {
		w.wtr.WriteLine("%s %s = %d", val.ConstantName, enum.TypeName, val.Value)
	}
	w.wtr.WriteLine(")")
	w.wtr.NewLine()

	// String Getter
	w.wtr.WriteLine("func (e %s) String() (string, bool) {", enum.TypeName)
	w.wtr.WriteLine("switch e {")
	for _, val := range enum.Options {
		w.wtr.WriteLine("case %d:", val.Value)
		w.wtr.WriteLine("return \"%s\", true", val.Label)
	}
	w.wtr.WriteLine("default:")
	w.wtr.WriteLine("return \"\", false")
	w.wtr.WriteLine("}")
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// Options
	w.wtr.WriteLine("var %s = []EnumOption {", enum.OptionSetName)
	for _, val := range enum.Options {
		w.wtr.WriteLine("{")
		w.wtr.WriteLine("DisplayName: %s,", strLit(val.Label))
		w.wtr.WriteLine("Value: %d,", val.Value)
		w.wtr.WriteLine("},")
	}
	w.wtr.WriteLine("}")
	w.wtr.NewLine()

	// Options Getter
	w.wtr.WriteLine("func (e %s) Options() []EnumOption{", enum.TypeName)
	w.wtr.WriteLine("  return %s", enum.OptionSetName)
	w.wtr.WriteLine("}")
}
