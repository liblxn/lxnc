package lxn

import (
	"fmt"
	"unicode/utf8"
)

// Pos describes a position in the lxn file.
type Pos struct {
	File   string
	Line   int
	Column int
	Offset int
}

// String returns a string representation of the position.
func (p Pos) String() string {
	prefix := ""
	if p.File != "" {
		prefix = p.File + ":"
	}
	return prefix + fmt.Sprintf("%d:%d", p.Line, p.Column)
}

func (p *Pos) advance(ch rune) {
	if ch == '\n' {
		p.Line++
		p.Column = 0
	} else {
		p.Column++
	}
	p.Offset += utf8.RuneLen(ch)
}
