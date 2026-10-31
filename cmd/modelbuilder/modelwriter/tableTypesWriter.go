package modelwriter

import (
	"go/format"
	"log"
	"os"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderrepo"
)

type tableTypesWriter struct {
	path     string
	metadata *builderrepo.DatabaseSchema
	wtr      *writer
}

func NewTableTypesWriter(
	path string,
	metadata *builderrepo.DatabaseSchema,
) *tableTypesWriter {
	return &tableTypesWriter{
		path:     path,
		metadata: metadata,
		wtr:      &writer{&strings.Builder{}},
	}
}

func (w *tableTypesWriter) Write() {
	w.writePackageDefinition()
	w.wtr.NewLine()

	w.writeGuidType()
	w.wtr.NewLine()

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

func (w *tableTypesWriter) writePackageDefinition() {
	w.wtr.WriteLine(`//GENERATED CODE
// see cmd/modelbuilder/
package model
`)
}

func (w *tableTypesWriter) writeGuidType() {
	w.wtr.WriteLine("type Guid string")
	w.wtr.NewLine()

	w.wtr.WriteLine("func (g Guid) String() string {")
	w.wtr.WriteLine(" return string(g)")
	w.wtr.WriteLine("}")
}
