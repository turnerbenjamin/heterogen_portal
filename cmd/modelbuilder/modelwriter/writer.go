package modelwriter

import (
	"fmt"
	"strings"
)

type writer struct {
	sb *strings.Builder
}

func (w *writer) Write(statement string, args ...any) {
	if w.sb == nil {
		w.sb = &strings.Builder{}
	}
	fmt.Fprintf(w.sb, statement, args...)
}

func (w *writer) NewLine() {
	w.Write("\n")
}

func (w *writer) WriteLine(statement string, args ...any) {
	w.Write(statement, args...)
	w.NewLine()
}

func (w *writer) String() string {
	return w.sb.String()
}

func strLit(str string) string {
	return fmt.Sprintf("\"%s\"", str)
}

func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")

	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}

	return strings.Join(parts, "")
}
