package generate_cldr

import (
	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*langLookup)(nil)
	_ generator.TestSnippet = (*langLookup)(nil)
)

type langLookup struct {
	stringBlock
}

func newLangLookup() *langLookup {
	return &langLookup{
		stringBlock: stringBlock{
			feature:   "lang",
			idBits:    16,
			blocksize: 3,
		},
	}
}

func (l *langLookup) Imports() []string {
	return l.imports()
}

func (l *langLookup) Generate(p *generator.Printer) {
	l.generate(p)
}

func (l *langLookup) TestImports() []string {
	return l.testImports()
}

func (l *langLookup) GenerateTest(p *generator.Printer) {
	l.generateTest(p)
}

var (
	_ generator.Snippet     = (*langLookupVar)(nil)
	_ generator.TestSnippet = (*langLookupVar)(nil)
)

type langLookupVar struct {
	stringBlockVar
	typ *langLookup
}

func newLangLookupVar(name string, typ *langLookup, data *cldr.Data) *langLookupVar {
	stringBlock := newStringBlockVar(name, &typ.stringBlock)
	forEachIdentity(data, func(id cldr.Identity) {
		stringBlock.add(id.Language)
	})

	return &langLookupVar{
		stringBlockVar: stringBlock,
		typ:            typ,
	}
}

func (v *langLookupVar) langID(code string) uint {
	return v.stringID(code)
}

func (v *langLookupVar) Imports() []string {
	return v.imports()
}

func (v *langLookupVar) Generate(p *generator.Printer) {
	v.generate(p)
}

func (v *langLookupVar) TestImports() []string {
	return v.testImports()
}

func (v *langLookupVar) GenerateTest(p *generator.Printer) {
	v.generateTest(p)
}
