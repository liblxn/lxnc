package main

import (
	"fmt"
	"io"
)

type printer struct {
	w   io.Writer
	err error
}

func newPrinter(w io.Writer) *printer {
	return &printer{w: w}
}

func (p *printer) Err() error {
	return p.err
}

func (p *printer) Print(args ...interface{}) {
	if p.err == nil {
		_, p.err = fmt.Fprint(p.w, args...)
	}
}

func (p *printer) Println(args ...interface{}) {
	p.Print(append(args, "\n")...)
}
