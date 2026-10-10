package iconbuilderutils

import (
	"go/format"
	"log"
	"os"
	"slices"
	"strings"
)

func WriteIcons(path string, icons []svgData) {
	slices.SortFunc(icons, func(a svgData, b svgData) int {
		return strings.Compare(a.name, b.name)
	})

	w := &writer{}
	writePackageDefinition(w)
	w.NewLine()

	writeIconSetType(w, icons)
	w.NewLine()

	writeIconSet(w, icons)
	w.NewLine()

	raw := w.String()
	formatted, err := format.Source([]byte(raw))
	if err != nil {
		_ = os.WriteFile(strings.Replace(path, ".go", ".raw.go", 1), []byte(raw), 0644)
		log.Fatal(err)
	}

	err = os.WriteFile(path, formatted, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func writePackageDefinition(w *writer) {
	w.WriteLine("// GENERATED CODE")
	w.WriteLine("// see cmd/iconbuilder")
	w.WriteLine("package icons")
	w.WriteLine("import(")
	w.WriteLine("\"html/template\"")
	w.WriteLine(")")
}

func writeIconSetType(w *writer, icons []svgData) {
	w.WriteLine("type IconSet struct{")
	for _, icon := range icons {
		w.WriteLine("%s template.HTML", icon.name)
	}
	w.WriteLine("}")
}

func writeIconSet(w *writer, icons []svgData) {
	w.WriteLine("var Set = &IconSet{")
	for _, icon := range icons {
		w.WriteLine("%s: template.HTML(`%s`),", icon.name, icon.contents)
	}
	w.WriteLine("}")
}
