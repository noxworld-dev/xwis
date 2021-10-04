package xwis

import (
	"bufio"
	"fmt"
	"io"
)

const (
	eol = "\n"
)

type writer struct {
	bw *bufio.Writer
}

func newWriter(w io.Writer) *writer {
	return &writer{bw: bufio.NewWriter(w)}
}

func (w *writer) Flush() error {
	return w.bw.Flush()
}

func (w *writer) WriteLine(line string) error {
	if DebugLog != nil {
		DebugLog.Println(line)
	}
	_, err := w.bw.WriteString(line + eol)
	return err
}

func (w *writer) WriteLinef(format string, args ...interface{}) error {
	if DebugLog != nil {
		DebugLog.Printf(format+"\n", args...)
	}
	_, err := fmt.Fprintf(w.bw, format+eol, args...)
	return err
}
