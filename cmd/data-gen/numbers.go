package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/liblxn/lxnc/internal/cldr"
)

const (
	digitsBits   = 4
	groupingBits = 4

	affixIDBits     = 8
	affixOffsetBits = 8

	patternIDBits = 8

	symbolsIDBits     = 8
	symbolsOffsetBits = 16

	zeroIDBits = 8
)

// affix lookup
type affixLookup struct {
	strings _multiString
}

func (l affixLookup) funcs() []string {
	return []string{"prefix", "suffix"}
}

func (l affixLookup) imports() []string {
	return l.strings.imports(affixOffsetBits)
}

func (l affixLookup) generate(p *printer) {
	l.strings.generate(p, "affix", affixIDBits, affixOffsetBits, l.funcs())
}

func (l affixLookup) testImports() []string {
	return l.strings.testImports()
}

func (l affixLookup) generateTest(p *printer) {
	l.strings.generateTest(p, "affix", affixOffsetBits, l.funcs())
}

type affixLookupVar struct {
	_multiStringLookupVar
}

func newAffixLookupVar(name string) *affixLookupVar {
	return &affixLookupVar{
		_multiStringLookupVar: _multiStringLookupVar{
			feature:    "affix",
			offsetBits: affixOffsetBits,
			idBits:     affixIDBits,
			name:       name,
		},
	}
}

func (v *affixLookupVar) newAffix(prefix, suffix string) (str string, n int) {
	var multiString _multiString
	return multiString.newString(prefix, suffix)
}

func (v *affixLookupVar) add(prefix, suffix string) {
	v._multiStringLookupVar.add(v.newAffix(prefix, suffix))

}

func (v *affixLookupVar) affixID(prefix, suffix string) uint {
	affix, _ := v.newAffix(prefix, suffix)
	return v._multiStringLookupVar.stringID(affix)
}

// pattern lookup
type patternLookup struct{}

func (l patternLookup) imports() []string {
	return nil
}

func (l patternLookup) generate(p *printer) {
	affixIDMask := fmt.Sprintf("%#x", (1<<affixIDBits)-1)
	digitsMask := fmt.Sprintf("%#x", (1<<digitsBits)-1)
	groupingMask := fmt.Sprintf("%#x", (1<<groupingBits)-1)

	patternBits := 2*affixIDBits + 3*digitsBits + 3*groupingBits
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
	p.Println(`func (p pattern) posAffixID() affixID     { return affixID((p >> `, affixIDBits+3*digitsBits+3*groupingBits, `) & `, affixIDMask, `) }`)
	p.Println(`func (p pattern) negAffixID() affixID     { return affixID((p >> `, 3*digitsBits+3*groupingBits, `) & `, affixIDMask, `) }`)
	p.Println(`func (p pattern) minIntDigits() int       { return int((p >> `, 2*digitsBits+3*groupingBits, `) & `, digitsMask, `) }`)
	p.Println(`func (p pattern) minFracDigits() int      { return int((p >> `, digitsBits+3*groupingBits, `) & `, digitsMask, `) }`)
	p.Println(`func (p pattern) maxFracDigits() int      { return int((p >> `, 3*groupingBits, `) & `, digitsMask, `) }`)
	p.Println(`func (p pattern) intGrouping() (int, int) { return int((p >> `, 2*groupingBits, `) & `, groupingMask, `), int((p >> `, groupingBits, `) & `, groupingMask, `) }`)
	p.Println(`func (p pattern) fracGrouping() int       { return int(p & `, groupingMask, `) }`)
	p.Println()
	p.Println(`type patternID uint`, patternIDBits)
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

func (l patternLookup) testImports() []string {
	return nil
}

func (l patternLookup) generateTest(p *printer) {
	pattern := func(posAffixID, negAffixID, minIntDigits, minFracDigits, maxFracDigits, primIntGrouping, secIntGrouping, fracGrouping uint) string {
		p := (posAffixID << (affixIDBits + 3*digitsBits + 3*groupingBits)) |
			(negAffixID << (3*digitsBits + 3*groupingBits)) |
			(minIntDigits << (2*digitsBits + 3*groupingBits)) |
			(minFracDigits << (digitsBits + 3*groupingBits)) |
			(maxFracDigits << (3 * groupingBits)) |
			(primIntGrouping << (2 * groupingBits)) |
			(secIntGrouping << groupingBits) |
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

type patternLookupVar struct {
	name    string
	affixes *affixLookupVar
	nfs     []cldr.NumberFormat
}

func newPatternLookupVar(name string, affixes *affixLookupVar) *patternLookupVar {
	return &patternLookupVar{
		name:    name,
		affixes: affixes,
	}
}

func (v *patternLookupVar) newPattern(nf cldr.NumberFormat) uint64 {
	const maxDigits = (1 << digitsBits) - 1
	const maxGrouping = (1 << groupingBits) - 1

	posAffixID := uint64(v.affixes.affixID(nf.PositivePrefix, nf.PositiveSuffix))
	negAffixID := uint64(v.affixes.affixID(nf.NegativePrefix, nf.NegativeSuffix))
	minIntDigits := minUint64(uint64(nf.MinIntegerDigits), maxDigits)
	minFracDigits := minUint64(uint64(nf.MinFractionDigits), maxDigits)
	maxFracDigits := minUint64(uint64(nf.MaxFractionDigits), maxDigits)
	primaryIntGrouping := minUint64(uint64(nf.IntegerGrouping.PrimarySize), maxGrouping)
	secondaryIntGrouping := minUint64(uint64(nf.IntegerGrouping.SecondarySize), maxGrouping)
	primaryFracGrouping := minUint64(uint64(nf.FractionGrouping.PrimarySize), maxGrouping)

	return (posAffixID << (affixIDBits + 3*digitsBits + 3*groupingBits)) |
		(negAffixID << (3*digitsBits + 3*groupingBits)) |
		(minIntDigits << (2*digitsBits + 3*groupingBits)) |
		(minFracDigits << (digitsBits + 3*groupingBits)) |
		(maxFracDigits << (3 * groupingBits)) |
		(primaryIntGrouping << (2 * groupingBits)) |
		(secondaryIntGrouping << groupingBits) |
		primaryFracGrouping
}

func (v *patternLookupVar) add(nf cldr.NumberFormat) {
	if len(v.nfs) == (1<<patternIDBits)-1 {
		panic(fmt.Sprintf("number of patterns exceeds the maximum, cannot add %s", nf.Pattern))
	}

	const maxDigits = (1 << digitsBits) - 1
	const maxGrouping = (1 << groupingBits) - 1
	switch {
	case nf.MinIntegerDigits > maxDigits:
		panic(fmt.Sprintf("minimum integer digits exceeds the limit for %q: %d", nf.Pattern, nf.MinIntegerDigits))
	case nf.MinFractionDigits > maxDigits:
		panic(fmt.Sprintf("minimum fraction digits exceeds the limit for %q: %d", nf.Pattern, nf.MinFractionDigits))
	case nf.MaxFractionDigits > maxDigits:
		panic(fmt.Sprintf("maximum fraction digits exceeds the limit for %q: %d", nf.Pattern, nf.MaxFractionDigits))
	case nf.IntegerGrouping.PrimarySize > maxGrouping:
		panic(fmt.Sprintf("primary integer grouping exceeds the limit for %q: %d", nf.Pattern, nf.IntegerGrouping.PrimarySize))
	case nf.IntegerGrouping.SecondarySize > maxGrouping:
		panic(fmt.Sprintf("secondary integer grouping exceeds the limit for %q: %d", nf.Pattern, nf.IntegerGrouping.SecondarySize))
	case nf.FractionGrouping.PrimarySize > maxGrouping:
		panic(fmt.Sprintf("primary fraction grouping exceeds the limit for %q: %d", nf.Pattern, nf.FractionGrouping.PrimarySize))
	}

	idx := 0
	for idx < len(v.nfs) && v.nfs[idx].Pattern < nf.Pattern {
		idx++
	}
	if idx == len(v.nfs) || v.nfs[idx] != nf {
		v.nfs = append(v.nfs, nf)
		copy(v.nfs[idx+1:], v.nfs[idx:])
		v.nfs[idx] = nf
	}
}

func (v *patternLookupVar) patternID(nf cldr.NumberFormat) uint {
	for i := 0; i < len(v.nfs); i++ {
		if v.nfs[i].Pattern == nf.Pattern {
			return uint(i + 1)
		}
	}
	panic(fmt.Sprintf("number format not found: %s", nf.Pattern))
}

func (v *patternLookupVar) imports() []string {
	return nil
}

func (v *patternLookupVar) generate(p *printer) {
	patternBits := 2*affixIDBits + 3*digitsBits + 3*groupingBits
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
	perLine := lineLength / (digitsPerPattern + 4) // additional "0x" and ", "

	hex := func(nf cldr.NumberFormat) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.newPattern(nf), digitsPerPattern)
	}

	p.Println(`var `, v.name, ` = patternLookup{ // `, len(v.nfs), ` items, `, len(v.nfs)*patternBits/8, ` bytes`)

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
		for k, nf := range v.nfs[i:n] {
			p.Print(strconv.Quote(nf.Pattern))
			if k+i < n-1 {
				p.Print(`, `)
			}
		}
		p.Println()
	}

	p.Println(`}`)
}

func (v *patternLookupVar) testImports() []string {
	return nil
}

func (v *patternLookupVar) generateTest(p *printer) {
	newID := func(idx int) string {
		return fmt.Sprintf("%#0[2]*[1]x", idx+1, patternIDBits/4)
	}

	newPattern := func(nf cldr.NumberFormat) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.newPattern(nf), (2*affixIDBits+3*digitsBits+3*groupingBits)/4)
	}

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[patternID]pattern{`)

	perLine := lineLength / ((patternIDBits + 2*affixIDBits + 3*digitsBits + 3*groupingBits) / 4)
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

// symbols lookup
type symbolsLookup struct {
	strings _multiString
}

func (l symbolsLookup) funcs() []string {
	return []string{"decimal", "group", "percent", "minus", "inf", "nan", "currDecimal", "currGroup"}
}

func (l symbolsLookup) imports() []string {
	return l.strings.imports(symbolsOffsetBits)
}

func (l symbolsLookup) generate(p *printer) {
	l.strings.generate(p, "symbols", symbolsIDBits, symbolsOffsetBits, l.funcs())
}

func (l symbolsLookup) testImports() []string {
	return l.strings.testImports()
}

func (l symbolsLookup) generateTest(p *printer) {
	l.strings.generateTest(p, "symbols", symbolsOffsetBits, l.funcs())
}

type symbolsLookupVar struct {
	_multiStringLookupVar
}

func newSymbolsLookupVar(name string) *symbolsLookupVar {
	return &symbolsLookupVar{
		_multiStringLookupVar: _multiStringLookupVar{
			feature:    "symbols",
			offsetBits: symbolsOffsetBits,
			idBits:     symbolsIDBits,
			name:       name,
		},
	}
}

func (v *symbolsLookupVar) newSymbols(symb cldr.NumberSymbols) (string, int) {
	symbols := [...]string{symb.Decimal, symb.Group, symb.Percent, symb.Minus, symb.Infinity, symb.NaN, symb.CurrencyDecimal, symb.CurrencyGroup}
	off := 0
	s := ""
	n := 0
	for _, symbol := range symbols {
		off += len(symbol)
		if off > math.MaxUint8 {
			panic(fmt.Sprintf("symbols exceeds the maximum length: %v", symbols))
		}
		s += fmt.Sprintf(`\x%02x`, off)
		n++
	}
	for _, symbol := range symbols {
		s += symbol
		n += len(symbol)
	}
	return s, n
}

func (v *symbolsLookupVar) add(symb cldr.NumberSymbols) {
	v._multiStringLookupVar.add(v.newSymbols(symb))
}

func (v *symbolsLookupVar) symbolsID(symbols string) uint {
	return v._multiStringLookupVar.stringID(symbols)
}

// zero lookup
type zeroLookup struct{}

func (l zeroLookup) imports() []string {
	return []string{"unicode/utf8"}
}

func (l zeroLookup) generate(p *printer) {
	p.Println(`// The zero is used to determine the digits. A digit n in [0,9] is determined by`)
	p.Println(`// adding n to the zero. The lookup is a string of all existing zero runes.`)
	p.Println(`type zeroID uint`, zeroIDBits)
	p.Println(`type zeroLookup string`)
	p.Println()
	p.Println(`func (l zeroLookup) zero(id zeroID) rune {`)
	p.Println(`	if id == 0 || int(id) > len(l) {`)
	p.Println(`		return utf8.RuneError`)
	p.Println(`	}`)
	p.Println(`	ch, _ := utf8.DecodeRuneInString(string(l[id-1:]))`)
	p.Println(`	return ch`)
	p.Println(`}`)
}

func (l zeroLookup) testImports() []string {
	return []string{"unicode/utf8"}
}

func (l zeroLookup) generateTest(p *printer) {
	p.Println(`func TestZeroLookup(t *testing.T) {`)
	p.Println(`	const lookup zeroLookup = "0az"`)
	p.Println()
	p.Println(`	if z := lookup.zero(0); z != utf8.RuneError {`)
	p.Println(`		t.Errorf("unexpected zero for id 0: %U", z)`)
	p.Println(`	}`)
	p.Println(`	if z := lookup.zero(1); z != '0' {`)
	p.Println(`		t.Errorf("unexpected zero for id 1: %U", z)`)
	p.Println(`	}`)
	p.Println(`	if z := lookup.zero(2); z != 'a' {`)
	p.Println(`		t.Errorf("unexpected zero for id 2: %U", z)`)
	p.Println(`	}`)
	p.Println(`	if z := lookup.zero(3); z != 'z' {`)
	p.Println(`		t.Errorf("unexpected zero for id 3: %U", z)`)
	p.Println(`	}`)
	p.Println(`	if z := lookup.zero(4); z != utf8.RuneError {`)
	p.Println(`		t.Errorf("unexpected zero for id 4: %U", z)`)
	p.Println(`	}`)
	p.Println(`}`)
}

type zeroLookupVar struct {
	name  string
	zeros []rune
	bytes int
}

func newZeroLookupVar(name string) *zeroLookupVar {
	return &zeroLookupVar{
		name: name,
	}
}

func (v *zeroLookupVar) zeroID(zero rune) uint {
	off := 0
	for _, z := range v.zeros {
		if z == zero {
			return uint(off + 1)
		}
		off += utf8.RuneLen(z)
	}
	panic(fmt.Sprintf("zero not found: %c", zero))
}

func (v *zeroLookupVar) add(zero rune) {
	idx := 0
	for idx < len(v.zeros) && v.zeros[idx] < zero {
		idx++
	}
	if idx == len(v.zeros) || v.zeros[idx] != zero {
		v.zeros = append(v.zeros, 0)
		copy(v.zeros[idx+1:], v.zeros[idx:])
		v.zeros[idx] = zero

		v.bytes += utf8.RuneLen(zero)
	}
}

func (v *zeroLookupVar) imports() []string {
	return nil
}

func (v *zeroLookupVar) generate(p *printer) {
	p.Println(`const `, v.name, ` zeroLookup = "" + // `, len(v.zeros), ` items, `, v.bytes, ` bytes`)

	count := 0
	p.Print(`	"`)
	for _, zero := range v.zeros {
		if count >= lineLength {
			p.Println(`" +`)
			p.Print(`	"`)
			count = 0
		}
		p.Print(string(zero))
		count++
	}
	p.Println(`"`)
}

func (v *zeroLookupVar) testImports() []string {
	return nil
}

func (v *zeroLookupVar) generateTest(p *printer) {
	newID := func(idx int) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.zeroID(v.zeros[idx]), zeroIDBits/4)
	}

	newZero := func(zero rune) string {
		return `"` + string(zero) + `"`
	}

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[zeroID]string{ // zero id => zero`)

	perLine := lineLength / (zeroIDBits/4 + 2)
	for i := 0; i < len(v.zeros); i += perLine {
		n := i + perLine
		if n > len(v.zeros) {
			n = len(v.zeros)
		}

		p.Print(`		`, newID(i), `: `, newZero(v.zeros[i]))
		for k := i + 1; k < n; k++ {
			p.Print(`, `, newID(k), `: `, newZero(v.zeros[k]))
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for id, expectedZero := range expected {`)
	p.Println(`		if zero := `, v.name, `.zero(id); string(zero) != expectedZero {`)
	p.Println(`			t.Fatalf("unexpected zero for id %d: %s", uint(id), string(zero))`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}

// numbers lookup
type numbersLookup struct{}

func (l numbersLookup) imports() []string {
	return nil
}

func (l numbersLookup) generate(p *printer) {
	patternIDMask := fmt.Sprintf("%#x", (1<<patternIDBits)-1)
	symbolsIDMask := fmt.Sprintf("%#x", (1<<symbolsIDBits)-1)
	zeroIDMask := fmt.Sprintf("%#x", (1<<zeroIDBits)-1)

	numbersBits := patternIDBits + symbolsIDBits + zeroIDBits
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
	p.Println(`func (n numbers) patternID() patternID { return patternID((n >> `, symbolsIDBits+zeroIDBits, `) & `, patternIDMask, `) }`)
	p.Println(`func (n numbers) symbolsID() symbolsID { return symbolsID((n >> `, zeroIDBits, `) & `, symbolsIDMask, `) }`)
	p.Println(`func (n numbers) zeroID() zeroID       { return zeroID(n & `, zeroIDMask, `) }`)
	p.Println()
	p.Println(`type numbersLookup map[tagID]numbers`)
}

func (l numbersLookup) testImports() []string {
	return nil
}

func (l numbersLookup) generateTest(p *printer) {
	numbers := func(patternID, symbolsID, zeroID uint) string {
		return fmt.Sprintf("%#x", (patternID<<(symbolsIDBits+zeroIDBits))|(symbolsID<<zeroIDBits)|zeroID)
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

type numbersLookupVar struct {
	name     string
	tags     *tagLookupVar
	patterns *patternLookupVar
	symbols  *symbolsLookupVar
	zeros    *zeroLookupVar
	ids      []cldr.Identity
	nfs      []cldr.NumberFormat
	symb     []cldr.NumberSymbols
	numsys   []cldr.NumberingSystem
}

func newNumbersLookupVar(name string, tags *tagLookupVar, patterns *patternLookupVar, symbols *symbolsLookupVar, zeros *zeroLookupVar) *numbersLookupVar {
	return &numbersLookupVar{
		name:     name,
		tags:     tags,
		patterns: patterns,
		symbols:  symbols,
		zeros:    zeros,
	}
}

func (v *numbersLookupVar) iterateNumbers(iter func(cldr.Identity, cldr.NumberFormat, cldr.NumberSymbols, cldr.NumberingSystem)) {
	for i := 0; i < len(v.ids); i++ {
		iter(v.ids[i], v.nfs[i], v.symb[i], v.numsys[i])
	}
}

func (v *numbersLookupVar) add(id cldr.Identity, nf cldr.NumberFormat, symb cldr.NumberSymbols, numsys cldr.NumberingSystem) {
	idx := 0
	for idx < len(v.ids) && identityLess(v.ids[idx], id) {
		idx++
	}
	if idx == len(v.ids) || v.ids[idx] != id {
		v.ids = append(v.ids, id)
		copy(v.ids[idx+1:], v.ids[idx:])
		v.ids[idx] = id

		v.nfs = append(v.nfs, nf)
		copy(v.nfs[idx+1:], v.nfs[idx:])
		v.nfs[idx] = nf

		v.symb = append(v.symb, symb)
		copy(v.symb[idx+1:], v.symb[idx:])
		v.symb[idx] = symb

		v.numsys = append(v.numsys, numsys)
		copy(v.numsys[idx+1:], v.numsys[idx:])
		v.numsys[idx] = numsys
	}
}

func (v *numbersLookupVar) imports() []string {
	return nil
}

func (v *numbersLookupVar) generate(p *printer) {
	numbersBits := uint(patternIDBits + symbolsIDBits + zeroIDBits)
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
		return (patternID << (symbolsIDBits + zeroIDBits)) | (symbolsID << zeroIDBits) | zeroID
	}

	hex := func(x, bits uint) string {
		return fmt.Sprintf("%#0[2]*[1]x", x, bits/4)
	}

	p.Println(`var `, v.name, ` = numbersLookup{ // `, len(v.ids), ` items, `, len(v.ids)*numbersBytes, ` bytes`)

	v.iterateNumbers(func(id cldr.Identity, nf cldr.NumberFormat, symb cldr.NumberSymbols, numsys cldr.NumberingSystem) {
		tagID := v.tags.tagID(id)
		patternID := v.patterns.patternID(nf)
		symbols, _ := v.symbols.newSymbols(symb)
		symbolsID := v.symbols.symbolsID(symbols)
		zeroID := v.zeros.zeroID(numsys.Digits[0])

		num := numbers(patternID, symbolsID, zeroID)
		p.Println(`	`, hex(tagID, tagIDBits), `: `, hex(num, numbersBits), `, // `, id.String())
	})

	p.Println(`}`)
}

// utility functions
func minUint64(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}
