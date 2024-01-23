package generate_cldr

import (
	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*scriptLookup)(nil)
	_ generator.TestSnippet = (*scriptLookup)(nil)
)

type scriptLookup struct {
	stringBlock
}

func newScriptLookup() *scriptLookup {
	return &scriptLookup{
		stringBlock: stringBlock{
			feature:   "script",
			idBits:    8,
			blocksize: 4,
		},
	}
}

func (l *scriptLookup) Imports() []string {
	return l.imports()
}

func (l *scriptLookup) Generate(p *generator.Printer) {
	l.generate(p)
}

func (l *scriptLookup) TestImports() []string {
	return l.testImports()
}

func (l *scriptLookup) GenerateTest(p *generator.Printer) {
	l.generateTest(p)
}

var (
	_ generator.Snippet     = (*scriptLookupVar)(nil)
	_ generator.TestSnippet = (*scriptLookupVar)(nil)
)

type scriptLookupVar struct {
	stringBlockVar
	typ *scriptLookup
}

func newScriptLookupVar(name string, typ *scriptLookup, data *cldr.Data) *scriptLookupVar {
	stringBlock := newStringBlockVar(name, &typ.stringBlock)
	forEachIdentity(data, func(id cldr.Identity) {
		if id.Script != "" {
			stringBlock.add(id.Script)
		}
	})

	return &scriptLookupVar{
		stringBlockVar: stringBlock,
		typ:            typ,
	}
}

func (v *scriptLookupVar) scriptID(code string) uint {
	return v.stringID(code)
}

func (v *scriptLookupVar) Imports() []string {
	return v.imports()
}

func (v *scriptLookupVar) Generate(p *generator.Printer) {
	v.generate(p)
}

func (v *scriptLookupVar) TestImports() []string {
	return v.testImports()
}

func (v *scriptLookupVar) GenerateTest(p *generator.Printer) {
	v.generateTest(p)
}
