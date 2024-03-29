package pgpkg

import (
	"fmt"
	"io"
)

type InfoWriter struct {
	w io.Writer
}

func NewInfoWriter(w io.Writer) InfoWriter {
	return InfoWriter{w: w}
}

func (i InfoWriter) Print(name string, value any) {
	if value == nil {
		value = "-"
	}

	_, _ = fmt.Fprintf(i.w, "%-20s %v\n", name+":", value)
}

func (i InfoWriter) Println(args ...any) {
	_, _ = fmt.Fprintln(i.w, args...)
}

func (i InfoWriter) Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(i.w, format, args...)
}
