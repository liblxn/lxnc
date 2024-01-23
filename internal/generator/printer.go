package generator

import (
	"fmt"
	"io"
)

type Printer struct {
	w   io.Writer
	err error
}

func newPrinter(w io.Writer) *Printer {
	return &Printer{w: w}
}

func (p *Printer) Err() error {
	return p.err
}

func (p *Printer) Print(args ...interface{}) {
	if p.err == nil {
		_, p.err = fmt.Fprint(p.w, args...)
	}
}

func (p *Printer) Println(args ...interface{}) {
	if p.err == nil {
		p.Print(append(args, "\n")...)
	}
}
