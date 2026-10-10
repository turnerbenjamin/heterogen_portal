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
