package generate_cldr

import (
	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*zeroLookup)(nil)
	_ generator.TestSnippet = (*zeroLookup)(nil)
)

type zeroLookup struct {
	runeString
}

func newZeroLookup() *zeroLookup {
	return &zeroLookup{
		runeString: runeString{
			feature: "zero",
			idBits:  8,
		},
	}
}

func (l *zeroLookup) Imports() []string {
	return l.imports()
}

func (l *zeroLookup) Generate(p *generator.Printer) {
	l.generate(p)
}

func (l *zeroLookup) TestImports() []string {
	return l.testImports()
}

func (l *zeroLookup) GenerateTest(p *generator.Printer) {
	l.generateTest(p)
}

var (
	_ generator.Snippet     = (*zeroLookupVar)(nil)
	_ generator.TestSnippet = (*zeroLookupVar)(nil)
)

type zeroLookupVar struct {
	runeStringVar
	typ *zeroLookup
}

func newZeroLookupVar(name string, typ *zeroLookup, data *cldr.Data) *zeroLookupVar {
	runeString := newRuneStringVar(name, &typ.runeString)
	forEachNumbers(data, allFormats, func(data numbersData) {
		runeString.add(data.numsys.Digits[0])
	})

	return &zeroLookupVar{
		runeStringVar: runeString,
		typ:           typ,
	}
}

func (v *zeroLookupVar) zeroID(zero rune) uint {
	return v.runeID(zero)
}

func (v *zeroLookupVar) Imports() []string {
	return v.imports()
}

func (v *zeroLookupVar) Generate(p *generator.Printer) {
	v.generate(p)
}

func (v *zeroLookupVar) TestImports() []string {
	return v.testImports()
}

func (v *zeroLookupVar) GenerateTest(p *generator.Printer) {
	v.generateTest(p)
}
