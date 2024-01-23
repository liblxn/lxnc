package generate_cldr

import "github.com/liblxn/lxnc/internal/generator"

var (
	_ generator.Snippet = (*pluralTag)(nil)
)

type pluralTag struct {
	other uint
	zero  uint
	one   uint
	two   uint
	few   uint
	many  uint
}

func newPluralTag() *pluralTag {
	return &pluralTag{
		other: 0,
		zero:  1,
		one:   2,
		two:   3,
		few:   4,
		many:  5,
	}
}

func (pr *pluralTag) Imports() []string {
	return nil
}

func (pt *pluralTag) Generate(p *generator.Printer) {
	p.Println(`// PluralTag represents a tag for a specific plural form.`)
	p.Println(`type PluralTag uint8`)
	p.Println()
	p.Println(`// Available plural tags.`)
	p.Println(`const (`)
	p.Println(`	Other PluralTag = `, pt.other)
	p.Println(`	Zero  PluralTag = `, pt.zero)
	p.Println(`	One   PluralTag = `, pt.one)
	p.Println(`	Two   PluralTag = `, pt.two)
	p.Println(`	Few   PluralTag = `, pt.few)
	p.Println(`	Many  PluralTag = `, pt.many)
	p.Println(`)`)
}
