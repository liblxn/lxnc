package generate_cldr

import (
	"fmt"
	"sort"
	"strings"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
	"github.com/liblxn/lxnc/lxn"
)

var (
	_ generator.Snippet = (*pluralCategory)(nil)
)

type pluralCategory struct {
	bits uint

	zero  uint
	one   uint
	two   uint
	few   uint
	many  uint
	other uint
}

func newPluralCategory() *pluralCategory {
	c := &pluralCategory{
		bits:  3,
		zero:  uint(lxn.Zero),
		one:   uint(lxn.One),
		two:   uint(lxn.Two),
		few:   uint(lxn.Few),
		many:  uint(lxn.Many),
		other: uint(lxn.Other),
	}

	max := uint(1<<c.bits) - 1
	for _, v := range []uint{c.other, c.zero, c.one, c.two, c.few, c.many} {
		if v > max {
			panic("plural category out of range")
		}
	}

	return c
}

func (pr *pluralCategory) enumeratorOf(category uint) string {
	switch category {
	case pr.other:
		return "Other"
	case pr.zero:
		return "Zero"
	case pr.one:
		return "One"
	case pr.two:
		return "Two"
	case pr.few:
		return "Few"
	case pr.many:
		return "Many"
	default:
		panic(fmt.Sprintf("invalid plural category %d", category))
	}
}

func (pr *pluralCategory) cldrConstantOf(category uint) string {
	switch category {
	case pr.other:
		return cldr.Other
	case pr.zero:
		return cldr.Zero
	case pr.one:
		return cldr.One
	case pr.two:
		return cldr.Two
	case pr.few:
		return cldr.Few
	case pr.many:
		return cldr.Many
	default:
		panic(fmt.Sprintf("invalid plural category %d", category))
	}
}

func (pr *pluralCategory) Imports() []string {
	return nil
}

func (pt *pluralCategory) Generate(p *generator.Printer) {
	values := []uint{pt.zero, pt.one, pt.two, pt.few, pt.many, pt.other}
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	names := make([]string, len(values))
	maxNameLen := 0
	for i, v := range values {
		name := pt.enumeratorOf(v)
		names[i] = name
		if len(name) > maxNameLen {
			maxNameLen = len(name)
		}
	}

	padName := func(name string) string {
		return name + strings.Repeat(" ", maxNameLen-len(name))
	}

	p.Println(`// PluralCategory represents a category for a specific plural form.`)
	p.Println(`type PluralCategory uint8`)
	p.Println()
	p.Println(`// Available plural categories.`)
	p.Println(`const (`)
	for i, name := range names {
		p.Println(`	`, padName(name), ` PluralCategory = `, values[i])
	}
	p.Println(`)`)
}
