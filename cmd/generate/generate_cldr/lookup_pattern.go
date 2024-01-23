package generate_cldr

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*patternLookup)(nil)
	_ generator.TestSnippet = (*patternLookup)(nil)
)

type patternLookup struct {
	idBits       uint
	digitsBits   uint
	groupingBits uint
	affix        *affixLookup
}

func newPatternLookup(affix *affixLookup) *patternLookup {
	return &patternLookup{
		idBits:       8,
		digitsBits:   4,
		groupingBits: 4,
		affix:        affix,
	}
}

func (l *patternLookup) Imports() []string {
	return nil
}

func (l *patternLookup) Generate(p *generator.Printer) {
	affixIDMask := fmt.Sprintf("%#x", (1<<l.affix.idBits)-1)
	digitsMask := fmt.Sprintf("%#x", (1<<l.digitsBits)-1)
	groupingMask := fmt.Sprintf("%#x", (1<<l.groupingBits)-1)

	patternBits := 2*l.affix.idBits + 3*l.digitsBits + 3*l.groupingBits
	switch {
	case patternBits <= 16:
		patternBits = 16
	case patternBits <= 32:
		patternBits = 32
	case patternBits <= 64:
		patternBits = 64
	default:
		panic(fmt.Sprintf("pattern exceeds maximum bit size: %d", patternBits))
	}

	p.Println(`// A pattern is a tuple consisting of the positive and negative affixes, the integer and`)
	p.Println(`// fraction digits, and the grouping information. The lookup is a slice of patterns`)
	p.Println(`// where the pattern id is a 1-based index in this slice.`)
	p.Println(`type pattern uint`, patternBits)
	p.Println()
	p.Println(`func (p pattern) posAffixID() affixID     { return affixID((p >> `, l.affix.idBits+3*l.digitsBits+3*l.groupingBits, `) & `, affixIDMask, `) }`)
	p.Println(`func (p pattern) negAffixID() affixID     { return affixID((p >> `, 3*l.digitsBits+3*l.groupingBits, `) & `, affixIDMask, `) }`)
	p.Println(`func (p pattern) minIntDigits() int       { return int((p >> `, 2*l.digitsBits+3*l.groupingBits, `) & `, digitsMask, `) }`)
	p.Println(`func (p pattern) minFracDigits() int      { return int((p >> `, l.digitsBits+3*l.groupingBits, `) & `, digitsMask, `) }`)
	p.Println(`func (p pattern) maxFracDigits() int      { return int((p >> `, 3*l.groupingBits, `) & `, digitsMask, `) }`)
	p.Println(`func (p pattern) intGrouping() (int, int) { return int((p >> `, 2*l.groupingBits, `) & `, groupingMask, `), int((p >> `, l.groupingBits, `) & `, groupingMask, `) }`)
	p.Println(`func (p pattern) fracGrouping() int       { return int(p & `, groupingMask, `) }`)
	p.Println()
	p.Println(`type patternID uint`, l.idBits)
	p.Println()
	p.Println(`type patternLookup []pattern`)
	p.Println()
	p.Println(`func (l patternLookup) pattern(id patternID) pattern {`)
	p.Println(`	if 0 < id && int(id) <= len(l) {`)
	p.Println(`		return l[id-1]`)
	p.Println(`	}`)
	p.Println(`	return 0`)
	p.Println(`}`)
}

func (l *patternLookup) TestImports() []string {
	return nil
}

func (l *patternLookup) GenerateTest(p *generator.Printer) {
	pattern := func(posAffixID, negAffixID, minIntDigits, minFracDigits, maxFracDigits, primIntGrouping, secIntGrouping, fracGrouping uint) string {
		p := (posAffixID << (l.affix.idBits + 3*l.digitsBits + 3*l.groupingBits)) |
			(negAffixID << (3*l.digitsBits + 3*l.groupingBits)) |
			(minIntDigits << (2*l.digitsBits + 3*l.groupingBits)) |
			(minFracDigits << (l.digitsBits + 3*l.groupingBits)) |
			(maxFracDigits << (3 * l.groupingBits)) |
			(primIntGrouping << (2 * l.groupingBits)) |
			(secIntGrouping << l.groupingBits) |
			fracGrouping

		return fmt.Sprintf("%#x", p)
	}

	p.Println(`func TestPattern(t *testing.T) {`)
	p.Println(`	const pattern pattern = `, pattern(1, 2, 3, 4, 5, 6, 7, 8))
	p.Println()
	p.Println(`	if id := pattern.posAffixID(); id != 1 {`)
	p.Println(`		t.Errorf("unexpected positive affix id: %d", id)`)
	p.Println(`	}`)
	p.Println(`	if id := pattern.negAffixID(); id != 2 {`)
	p.Println(`		t.Errorf("unexpected negative affix id: %d", id)`)
	p.Println(`	}`)
	p.Println(`	if n := pattern.minIntDigits(); n != 3 {`)
	p.Println(`		t.Errorf("unexpected minimum integer digits: %d", n)`)
	p.Println(`	}`)
	p.Println(`	if n := pattern.minFracDigits(); n != 4 {`)
	p.Println(`		t.Errorf("unexpected minimum fraction digits: %d", n)`)
	p.Println(`	}`)
	p.Println(`	if n := pattern.maxFracDigits(); n != 5 {`)
	p.Println(`		t.Errorf("unexpected minimum fraction digits: %d", n)`)
	p.Println(`	}`)
	p.Println(`	if m, n := pattern.intGrouping(); m != 6 || n != 7 {`)
	p.Println(`		t.Errorf("unexpected integer grouping: (%d, %d)", m, n)`)
	p.Println(`	}`)
	p.Println(`	if n := pattern.fracGrouping(); n != 8 {`)
	p.Println(`		t.Errorf("unexpected fraction grouping: %d", n)`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`func TestPatternLookup(t *testing.T) {`)
	p.Println(`	lookup := patternLookup{1, 2, 3}`)
	p.Println()
	p.Println(`	for i := 0; i < len(lookup); i++ {`)
	p.Println(`		if p := lookup.pattern(patternID(i + 1)); p != lookup[i] {`)
	p.Println(`			t.Errorf("unexpected pattern for id %d: %#x", i+1, p)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	if p := lookup.pattern(0); p != 0 {`)
	p.Println(`		t.Errorf("unexpected pattern for id 0: %#x", p)`)
	p.Println(`	}`)
	p.Println(`	if p := lookup.pattern(patternID(len(lookup) + 1)); p != 0 {`)
	p.Println(`		t.Errorf("unexpected pattern for id %d: %#x", len(lookup)+1, p)`)
	p.Println(`	}`)
	p.Println(`}`)
}

var (
	_ generator.Snippet     = (*patternLookupVar)(nil)
	_ generator.TestSnippet = (*patternLookupVar)(nil)
)

type patternLookupVar struct {
	name    string
	typ     *patternLookup
	affixes *affixLookupVar
	nfs     []cldr.NumberFormat
}

func newPatternLookupVar(name string, typ *patternLookup, affixes *affixLookupVar, data *cldr.Data) *patternLookupVar {
	nfs := make([]cldr.NumberFormat, 0, 16)
	patternMap := map[string]struct{}{}
	forEachNumbers(data, allFormats, func(data numbersData) {
		maxDigits := (1 << typ.digitsBits) - 1
		maxGrouping := (1 << typ.groupingBits) - 1
		switch {
		case data.nf.MinIntegerDigits > maxDigits:
			panic(fmt.Sprintf("minimum integer digits exceeds the limit for %q: %d", data.nf.Pattern, data.nf.MinIntegerDigits))
		case data.nf.MinFractionDigits > maxDigits:
			panic(fmt.Sprintf("minimum fraction digits exceeds the limit for %q: %d", data.nf.Pattern, data.nf.MinFractionDigits))
		case data.nf.MaxFractionDigits > maxDigits:
			panic(fmt.Sprintf("maximum fraction digits exceeds the limit for %q: %d", data.nf.Pattern, data.nf.MaxFractionDigits))
		case data.nf.IntegerGrouping.PrimarySize > maxGrouping:
			panic(fmt.Sprintf("primary integer grouping exceeds the limit for %q: %d", data.nf.Pattern, data.nf.IntegerGrouping.PrimarySize))
		case data.nf.IntegerGrouping.SecondarySize > maxGrouping:
			panic(fmt.Sprintf("secondary integer grouping exceeds the limit for %q: %d", data.nf.Pattern, data.nf.IntegerGrouping.SecondarySize))
		case data.nf.FractionGrouping.PrimarySize > maxGrouping:
			panic(fmt.Sprintf("primary fraction grouping exceeds the limit for %q: %d", data.nf.Pattern, data.nf.FractionGrouping.PrimarySize))
		}

		if _, has := patternMap[data.nf.Pattern]; !has {
			nfs = append(nfs, data.nf)
		}
	})

	if len(nfs) >= (1 << typ.idBits) {
		panic("number of patterns exceeds the maximum")
	}

	sort.Slice(nfs, func(i, j int) bool {
		return nfs[i].Pattern < nfs[j].Pattern
	})

	return &patternLookupVar{
		name:    name,
		typ:     typ,
		affixes: affixes,
		nfs:     nfs,
	}
}

func (v *patternLookupVar) newPattern(nf cldr.NumberFormat) uint64 {
	min := func(x, y uint64) uint64 {
		if x < y {
			return x
		}
		return y
	}

	maxDigits := uint64(1<<v.typ.digitsBits) - 1
	maxGrouping := uint64(1<<v.typ.groupingBits) - 1

	posAffixID := uint64(v.affixes.affixID(nf.PositivePrefix, nf.PositiveSuffix))
	negAffixID := uint64(v.affixes.affixID(nf.NegativePrefix, nf.NegativeSuffix))
	minIntDigits := min(uint64(nf.MinIntegerDigits), maxDigits)
	minFracDigits := min(uint64(nf.MinFractionDigits), maxDigits)
	maxFracDigits := min(uint64(nf.MaxFractionDigits), maxDigits)
	primaryIntGrouping := min(uint64(nf.IntegerGrouping.PrimarySize), maxGrouping)
	secondaryIntGrouping := min(uint64(nf.IntegerGrouping.SecondarySize), maxGrouping)
	primaryFracGrouping := min(uint64(nf.FractionGrouping.PrimarySize), maxGrouping)

	return (posAffixID << (v.affixes.typ.idBits + 3*v.typ.digitsBits + 3*v.typ.groupingBits)) |
		(negAffixID << (3*v.typ.digitsBits + 3*v.typ.groupingBits)) |
		(minIntDigits << (2*v.typ.digitsBits + 3*v.typ.groupingBits)) |
		(minFracDigits << (v.typ.digitsBits + 3*v.typ.groupingBits)) |
		(maxFracDigits << (3 * v.typ.groupingBits)) |
		(primaryIntGrouping << (2 * v.typ.groupingBits)) |
		(secondaryIntGrouping << v.typ.groupingBits) |
		primaryFracGrouping
}

func (v *patternLookupVar) patternID(nf cldr.NumberFormat) uint {
	for i := 0; i < len(v.nfs); i++ {
		if v.nfs[i].Pattern == nf.Pattern {
			return uint(i + 1)
		}
	}
	panic(fmt.Sprintf("number format not found: %s", nf.Pattern))
}

func (v *patternLookupVar) Imports() []string {
	return nil
}

func (v *patternLookupVar) Generate(p *generator.Printer) {
	patternBits := 2*v.affixes.typ.idBits + 3*v.typ.digitsBits + 3*v.typ.groupingBits
	switch {
	case patternBits <= 16:
		patternBits = 16
	case patternBits <= 32:
		patternBits = 32
	case patternBits <= 64:
		patternBits = 64
	default:
		panic("pattern bits exceeded")
	}

	digitsPerPattern := patternBits / 4
	perLine := int(lineLength / (digitsPerPattern + 4)) // additional "0x" and ", "

	hex := func(nf cldr.NumberFormat) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.newPattern(nf), digitsPerPattern)
	}

	p.Println(`var `, v.name, ` = patternLookup{ // `, len(v.nfs), ` items, `, uint(len(v.nfs))*patternBits/8, ` bytes`)

	for i := 0; i < len(v.nfs); i += perLine {
		n := i + perLine
		if n > len(v.nfs) {
			n = len(v.nfs)
		}

		p.Print(`	`)
		for _, nf := range v.nfs[i:n] {
			p.Print(hex(nf), `, `)
		}
		p.Print(`// `)
		for k := i; k < n; k++ {
			p.Print(strconv.Quote(v.nfs[k].Pattern))
			if k+i < n-1 {
				p.Print(`, `)
			}
		}
		p.Println()
	}

	p.Println(`}`)
}

func (v *patternLookupVar) TestImports() []string {
	return nil
}

func (v *patternLookupVar) GenerateTest(p *generator.Printer) {
	newID := func(idx int) string {
		return fmt.Sprintf("%#0[2]*[1]x", idx+1, v.typ.idBits/4)
	}

	newPattern := func(nf cldr.NumberFormat) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.newPattern(nf), (2*v.affixes.typ.idBits+3*v.typ.digitsBits+3*v.typ.groupingBits)/4)
	}

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[patternID]pattern{`)

	perLine := int(lineLength / ((v.typ.idBits + 2*v.affixes.typ.idBits + 3*v.typ.digitsBits + 3*v.typ.groupingBits) / 4))
	for i := 0; i < len(v.nfs); i += perLine {
		n := i + perLine
		if n > len(v.nfs) {
			n = len(v.nfs)
		}

		p.Print(`		`, newID(i), `: `, newPattern(v.nfs[i]))
		for k := i + 1; k < n; k++ {
			p.Print(`, `, newID(k), `: `, newPattern(v.nfs[k]))
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for id, expectedPattern := range expected {`)
	p.Println(`		if pattern := `, v.name, `.pattern(id); pattern != expectedPattern {`)
	p.Println(`			t.Fatalf("unexpected pattern for id %d: %#x", uint(id), pattern)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}
