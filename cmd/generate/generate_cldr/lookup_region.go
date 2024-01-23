package generate_cldr

import (
	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*regionLookup)(nil)
	_ generator.TestSnippet = (*regionLookup)(nil)
)

type regionLookup struct {
	stringBlock
}

func newRegionLookup() *regionLookup {
	return &regionLookup{
		stringBlock: stringBlock{
			feature:   "region",
			idBits:    8,
			blocksize: 3,
		},
	}
}

func (l *regionLookup) Imports() []string {
	return l.imports()
}

func (l *regionLookup) Generate(p *generator.Printer) {
	l.generate(p)
}

func (l *regionLookup) TestImports() []string {
	return l.testImports()
}

func (l *regionLookup) GenerateTest(p *generator.Printer) {
	l.generateTest(p)
}

var (
	_ generator.Snippet     = (*regionLookupVar)(nil)
	_ generator.TestSnippet = (*regionLookupVar)(nil)
)

type regionLookupVar struct {
	stringBlockVar
	typ *regionLookup
}

func newRegionLookupVar(name string, typ *regionLookup, data *cldr.Data) *regionLookupVar {
	stringBlock := newStringBlockVar(name, &typ.stringBlock)
	forEachIdentity(data, func(id cldr.Identity) {
		if id.Territory != "" {
			stringBlock.add(id.Territory)
		}
	})

	return &regionLookupVar{
		stringBlockVar: stringBlock,
		typ:            typ,
	}
}

func (v *regionLookupVar) regionID(code string) uint {
	return v.stringID(code)
}

func (v *regionLookupVar) Imports() []string {
	return v.imports()
}

func (v *regionLookupVar) Generate(p *generator.Printer) {
	v.generate(p)
}

func (v *regionLookupVar) TestImports() []string {
	return v.testImports()
}

func (v *regionLookupVar) GenerateTest(p *generator.Printer) {
	v.generateTest(p)
}
