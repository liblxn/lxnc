package generate_cldr

import (
	"fmt"
	"sort"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*numbersLookup)(nil)
	_ generator.TestSnippet = (*numbersLookup)(nil)
)

type numbersLookup struct {
	pattern *patternLookup
	symbols *symbolsLookup
	zero    *zeroLookup
}

func newNumbersLookup(pattern *patternLookup, symbols *symbolsLookup, zero *zeroLookup) *numbersLookup {
	return &numbersLookup{
		pattern: pattern,
		symbols: symbols,
		zero:    zero,
	}
}

func (l *numbersLookup) Imports() []string {
	return nil
}

func (l *numbersLookup) Generate(p *generator.Printer) {
	patternIDMask := fmt.Sprintf("%#x", (1<<l.pattern.idBits)-1)
	symbolsIDMask := fmt.Sprintf("%#x", (1<<l.symbols.idBits)-1)
	zeroIDMask := fmt.Sprintf("%#x", (1<<l.zero.idBits)-1)

	numbersBits := l.pattern.idBits + l.symbols.idBits + l.zero.idBits
	switch {
	case numbersBits <= 8:
		numbersBits = 8
	case numbersBits <= 16:
		numbersBits = 16
	case numbersBits <= 32:
		numbersBits = 32
	case numbersBits <= 64:
		numbersBits = 64
	default:
		panic(fmt.Sprintf("numbers exceeds maximum bit size: %d", numbersBits))
	}

	p.Println(`// The numbers data is a tuple consisting of a pattern id, a symbols id, and`)
	p.Println(`// a zero id. The lookup maps a CLDR identity to a numbers data.`)
	p.Println(`type numbers uint`, numbersBits)
	p.Println()
	p.Println(`func (n numbers) patternID() patternID { return patternID((n >> `, l.symbols.idBits+l.zero.idBits, `) & `, patternIDMask, `) }`)
	p.Println(`func (n numbers) symbolsID() symbolsID { return symbolsID((n >> `, l.zero.idBits, `) & `, symbolsIDMask, `) }`)
	p.Println(`func (n numbers) zeroID() zeroID       { return zeroID(n & `, zeroIDMask, `) }`)
	p.Println()
	p.Println(`type numbersLookup map[tagID]numbers`)
}

func (l *numbersLookup) TestImports() []string {
	return nil
}

func (l *numbersLookup) GenerateTest(p *generator.Printer) {
	numbers := func(patternID, symbolsID, zeroID uint) string {
		return fmt.Sprintf("%#x", (patternID<<(l.symbols.idBits+l.zero.idBits))|(symbolsID<<l.zero.idBits)|zeroID)
	}

	p.Println(`func TestNumbers(t *testing.T) {`)
	p.Println(`	const numbers numbers = `, numbers(1, 2, 3))
	p.Println()
	p.Println(`	if id := numbers.patternID(); id != 1 {`)
	p.Println(`		t.Errorf("unexpected pattern id: %d", id)`)
	p.Println(`	}`)
	p.Println(`	if id := numbers.symbolsID(); id != 2 {`)
	p.Println(`		t.Errorf("unexpected symbols id: %d", id)`)
	p.Println(`	}`)
	p.Println(`	if id := numbers.zeroID(); id != 3 {`)
	p.Println(`		t.Errorf("unexpected zero id: %d", id)`)
	p.Println(`	}`)
	p.Println(`}`)
}

var (
	_ generator.Snippet = (*numbersLookupVar)(nil)
)

type numbersLookupVar struct {
	name     string
	typ      *numbersLookup
	tags     *tagLookupVar
	patterns *patternLookupVar
	symbols  *symbolsLookupVar
	zeros    *zeroLookupVar
	data     []numbersData
}

func newNumbersLookupVar(name string, typ *numbersLookup, tags *tagLookupVar, patterns *patternLookupVar, symbols *symbolsLookupVar, zeros *zeroLookupVar, data *cldr.Data, filter numbersFilter) *numbersLookupVar {
	idMap := map[cldr.Identity]struct{}{}
	numData := make([]numbersData, 0, 8)
	forEachNumbers(data, filter, func(data numbersData) {
		if _, has := idMap[data.id]; has {
			return
		}
		numData = append(numData, data)
	})

	sort.Slice(numData, func(i, j int) bool {
		return identityLess(numData[i].id, numData[j].id)
	})

	return &numbersLookupVar{
		name:     name,
		typ:      typ,
		tags:     tags,
		patterns: patterns,
		symbols:  symbols,
		zeros:    zeros,
		data:     numData,
	}
}

func (v *numbersLookupVar) Imports() []string {
	return nil
}

func (v *numbersLookupVar) Generate(p *generator.Printer) {
	numbersBits := v.patterns.typ.idBits + v.symbols.typ.idBits + v.zeros.typ.idBits
	numbersBytes := 8
	switch {
	case numbersBits <= 8:
		numbersBytes = 1
	case numbersBits <= 16:
		numbersBytes = 2
	case numbersBits <= 32:
		numbersBytes = 4
	}

	numbers := func(patternID, symbolsID, zeroID uint) uint {
		return (patternID << (v.symbols.typ.idBits + v.zeros.typ.idBits)) | (symbolsID << v.zeros.typ.idBits) | zeroID
	}

	hex := func(x, bits uint) string {
		return fmt.Sprintf("%#0[2]*[1]x", x, bits/4)
	}

	p.Println(`var `, v.name, ` = numbersLookup{ // `, len(v.data), ` items, `, len(v.data)*numbersBytes, ` bytes`)

	for _, data := range v.data {
		tagID := v.tags.tagID(data.id)
		patternID := v.patterns.patternID(data.nf)
		symbolsID := v.symbols.symbolsID(data.symb)
		zeroID := v.zeros.zeroID(data.numsys.Digits[0])

		num := numbers(patternID, symbolsID, zeroID)
		p.Println(`	`, hex(tagID, v.tags.typ.idBits), `: `, hex(num, numbersBits), `, // `, data.id.String())
	}

	p.Println(`}`)
}
