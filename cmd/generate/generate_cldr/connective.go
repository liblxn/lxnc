package generate_cldr

import (
	"github.com/liblxn/lxnc/internal/generator"
	"github.com/liblxn/lxnc/lxn"
)

var (
	_ generator.Snippet = (*connective)(nil)
)

type connective struct {
	bits uint

	none        uint
	conjunction uint
	disjunction uint
}

func newConnective() *connective {
	c := &connective{
		bits: 2,

		none:        uint(lxn.None),
		conjunction: uint(lxn.Conjunction),
		disjunction: uint(lxn.Disjunction),
	}

	max := uint(1<<c.bits) - 1
	for _, v := range []uint{c.none, c.conjunction, c.disjunction} {
		if v > max {
			panic("connective out of range")
		}
	}

	return c
}

func (c *connective) Imports() []string {
	return nil
}

func (c *connective) Generate(p *generator.Printer) {
	p.Println(`// Connective represents a logical connective for two plural rules. A plural`)
	p.Println(`// rule can be connected with another rule by a conjunction ('and' operator)`)
	p.Println(`// or a disjunction ('or' operator). The conjunction binds more tightly.`)
	p.Println(`type Connective uint`)
	p.Println()
	p.Println(`// Available connectives.`)
	p.Println(`const (`)
	p.Println(`	None        Connective = `, c.none)
	p.Println(`	Conjunction Connective = `, c.conjunction)
	p.Println(`	Disjunction Connective = `, c.disjunction)
	p.Println(`)`)
}
