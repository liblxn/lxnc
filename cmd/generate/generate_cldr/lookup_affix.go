package generate_cldr

import (
	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*affixLookup)(nil)
	_ generator.TestSnippet = (*affixLookup)(nil)
)

type affixLookup struct {
	multiString
}

func newAffixLookup() *affixLookup {
	return &affixLookup{
		multiString: multiString{
			feature:    "affix",
			idBits:     8,
			offsetBits: 8,
			funcs:      []string{"prefix", "suffix"},
		},
	}
}

func (l *affixLookup) newAffix(prefix, suffix string) (string, int) {
	return l.newString(prefix, suffix)
}

func (l *affixLookup) Imports() []string {
	return l.imports()
}

func (l *affixLookup) Generate(p *generator.Printer) {
	l.generate(p)
}

func (l *affixLookup) TestImports() []string {
	return l.testImports()
}

func (l *affixLookup) GenerateTest(p *generator.Printer) {
	l.generateTest(p)
}

var (
	_ generator.Snippet     = (*affixLookupVar)(nil)
	_ generator.TestSnippet = (*affixLookupVar)(nil)
)

type affixLookupVar struct {
	multiStringVar
	typ *affixLookup
}

func newAffixLookupVar(name string, typ *affixLookup, data *cldr.Data) *affixLookupVar {
	multiString := newMultiStringVar(name, &typ.multiString)
	forEachNumbers(data, allFormats, func(data numbersData) {
		multiString.add(multiString.typ.newString(data.nf.PositivePrefix, data.nf.PositiveSuffix))
		multiString.add(multiString.typ.newString(data.nf.NegativePrefix, data.nf.NegativeSuffix))
	})

	return &affixLookupVar{
		multiStringVar: multiString,
		typ:            typ,
	}
}

func (v *affixLookupVar) affixID(prefix, suffix string) uint {
	affix, _ := v.typ.newAffix(prefix, suffix)
	return v.stringID(affix)
}

func (v *affixLookupVar) Imports() []string {
	return v.imports()
}

func (v *affixLookupVar) Generate(p *generator.Printer) {
	v.generate(p)
}

func (v *affixLookupVar) TestImports() []string {
	return v.testImports()
}

func (v *affixLookupVar) GenerateTest(p *generator.Printer) {
	v.generateTest(p)
}
