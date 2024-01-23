package generate_cldr

import (
	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*symbolsLookup)(nil)
	_ generator.TestSnippet = (*symbolsLookup)(nil)
)

type symbolsLookup struct {
	multiString
}

func newSymbolsLookup() *symbolsLookup {
	return &symbolsLookup{
		multiString: multiString{
			feature:    "symbols",
			idBits:     8,
			offsetBits: 16,
			funcs:      []string{"decimal", "group", "percent", "minus", "inf", "nan", "currDecimal", "currGroup"},
		},
	}
}

func (l *symbolsLookup) newSymbols(symb cldr.NumberSymbols) (string, int) {
	symbols := [...]string{symb.Decimal, symb.Group, symb.Percent, symb.Minus, symb.Infinity, symb.NaN, symb.CurrencyDecimal, symb.CurrencyGroup}
	return l.newString(symbols[:]...)
}

func (l *symbolsLookup) Imports() []string {
	return l.imports()
}

func (l *symbolsLookup) Generate(p *generator.Printer) {
	l.generate(p)
}

func (l *symbolsLookup) TestImports() []string {
	return l.testImports()
}

func (l *symbolsLookup) GenerateTest(p *generator.Printer) {
	l.generateTest(p)
}

var (
	_ generator.Snippet     = (*symbolsLookupVar)(nil)
	_ generator.TestSnippet = (*symbolsLookupVar)(nil)
)

type symbolsLookupVar struct {
	multiStringVar
	typ *symbolsLookup
}

func newSymbolsLookupVar(name string, typ *symbolsLookup, data *cldr.Data) *symbolsLookupVar {
	multiString := newMultiStringVar("symbols", &typ.multiString)
	forEachNumbers(data, allFormats, func(data numbersData) {
		multiString.add(typ.newSymbols(data.symb))
	})

	return &symbolsLookupVar{
		multiStringVar: multiString,
		typ:            typ,
	}
}

func (v *symbolsLookupVar) symbolsID(symbols cldr.NumberSymbols) uint {
	s, _ := v.typ.newSymbols(symbols)
	return v.stringID(s)
}

func (v *symbolsLookupVar) Imports() []string {
	return v.imports()
}

func (v *symbolsLookupVar) Generate(p *generator.Printer) {
	v.generate(p)
}

func (v *symbolsLookupVar) TestImports() []string {
	return v.testImports()
}

func (v *symbolsLookupVar) GenerateTest(p *generator.Printer) {
	v.generateTest(p)
}
