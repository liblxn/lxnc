// This file was generated by the 'generate' command. Do not edit.
// CLDR version: 44

package locale

import (
	"encoding/binary"
	"unicode/utf8"
)

// Grouping holds the sizes for number groups for a specific locale.
type Grouping struct {
	Primary   int
	Secondary int
}

// Affixes holds a prefix and suffix for locale specific number formatting.
type Affixes struct {
	Prefix string
	Suffix string
}

// Symbols holds all symbols that are used to format a number in a specific locale.
type Symbols struct {
	Decimal string
	Group   string
	Percent string
	Minus   string
	Inf     string
	NaN     string
	Zero    rune
}

// NumberFormat holds all relevant information to format a number in a specific locale.
type NumberFormat struct {
	numbers  numbers
	pattern  pattern
	currency bool
}

// DecimalFormat returns the data for formatting decimal numbers in the given locale.
func DecimalFormat(loc Locale) NumberFormat {
	return lookupNumberFormat(loc, decimalNumbers, false)
}

// MoneyFormat returns the data for formatting currency values in the given locale.
func MoneyFormat(loc Locale) NumberFormat {
	return lookupNumberFormat(loc, moneyNumbers, true)
}

// PercentFormat returns the data for formatting percent values in the given locale.
func PercentFormat(loc Locale) NumberFormat {
	return lookupNumberFormat(loc, percentNumbers, false)
}

func lookupNumberFormat(loc Locale, lookup numbersLookup, currency bool) NumberFormat {
	if loc == 0 {
		panic("invalid locale")
	}

	for {
		nums, has := lookup[tagID(loc)]
		switch {
		case has:
			return NumberFormat{
				numbers:  nums,
				pattern:  patterns.pattern(nums.patternID()),
				currency: currency,
			}
		case loc == root:
			panic("number format not found for " + loc.String())
		}
		loc = loc.parent()
	}
}

// Symbols returns the number symbols for the format.
func (nf NumberFormat) Symbols() Symbols {
	symbols := numberSymbols.symbols(nf.numbers.symbolsID())
	var decimal, group string
	if nf.currency {
		decimal, group = symbols.currDecimal(), symbols.currGroup()
	} else {
		decimal, group = symbols.decimal(), symbols.group()
	}
	return Symbols{
		Decimal: decimal,
		Group:   group,
		Percent: symbols.percent(),
		Minus:   symbols.minus(),
		Inf:     symbols.inf(),
		NaN:     symbols.nan(),
		Zero:    zeros.zero(nf.numbers.zeroID()),
	}
}

// PositiveAffixes returns the affixes for positive numbers.
func (nf NumberFormat) PositiveAffixes() Affixes {
	a := affixes.affix(nf.pattern.posAffixID())
	return Affixes{Prefix: a.prefix(), Suffix: a.suffix()}
}

// NegativeAffixes returns the affixes for negative numbers.
func (nf NumberFormat) NegativeAffixes() Affixes {
	a := affixes.affix(nf.pattern.negAffixID())
	return Affixes{Prefix: a.prefix(), Suffix: a.suffix()}
}

// MinIntegerDigits returns the minimum number of digits which should be displayed for the integer part.
func (nf NumberFormat) MinIntegerDigits() int {
	return nf.pattern.minIntDigits()
}

// MinFractionDigits returns the minimum number of digits which should be displayed for the fraction part.
func (nf NumberFormat) MinFractionDigits() int {
	return nf.pattern.minFracDigits()
}

// MaxFractionDigits returns the maximum number of digits which should be displayed for the fraction part.
func (nf NumberFormat) MaxFractionDigits() int {
	return nf.pattern.maxFracDigits()
}

// IntegerGrouping returns the grouping information for the integer part.
func (nf NumberFormat) IntegerGrouping() Grouping {
	prim, sec := nf.pattern.intGrouping()
	return Grouping{Primary: prim, Secondary: sec}
}

// FractionGrouping returns the grouping information for the fraction part.
func (nf NumberFormat) FractionGrouping() Grouping {
	prim := nf.pattern.fracGrouping()
	return Grouping{Primary: prim, Secondary: prim}
}

// The affix type is a concatenation of multiple strings. It starts with
// the offsets for each string followed by the actual strings.
// The affix lookup consists of all affix strings concatenated.
// It is prefixed with an offset for each affix block and its id is a
// 1-based index which points to the offset.
type affix string

func (s affix) prefix() string { return string(s[2 : 2+s[0]]) }
func (s affix) suffix() string { return string(s[2+s[0] : 2+s[1]]) }

type affixID uint8
type affixLookup string

func (l affixLookup) affix(id affixID) affix {
	if id == 0 {
		return "\x00\x00"
	}
	start, end := l[id-1], l[id]
	return affix(l[start:end])
}

// A pattern is a tuple consisting of the positive and negative affixes, the integer and
// fraction digits, and the grouping information. The lookup is a slice of patterns
// where the pattern id is a 1-based index in this slice.
type pattern uint64

func (p pattern) posAffixID() affixID     { return affixID((p >> 32) & 0xff) }
func (p pattern) negAffixID() affixID     { return affixID((p >> 24) & 0xff) }
func (p pattern) minIntDigits() int       { return int((p >> 20) & 0xf) }
func (p pattern) minFracDigits() int      { return int((p >> 16) & 0xf) }
func (p pattern) maxFracDigits() int      { return int((p >> 12) & 0xf) }
func (p pattern) intGrouping() (int, int) { return int((p >> 8) & 0xf), int((p >> 4) & 0xf) }
func (p pattern) fracGrouping() int       { return int(p & 0xf) }

type patternID uint8

type patternLookup []pattern

func (l patternLookup) pattern(id patternID) pattern {
	if 0 < id && int(id) <= len(l) {
		return l[id-1]
	}
	return 0
}

// The symbols type is a concatenation of multiple strings. It starts with
// the offsets for each string followed by the actual strings.
// The symbols lookup consists of all symbols strings concatenated.
// It is prefixed with an offset for each symbols block and its id is a
// 1-based index which points to the offset.
type symbols string

func (s symbols) decimal() string     { return string(s[8 : 8+s[0]]) }
func (s symbols) group() string       { return string(s[8+s[0] : 8+s[1]]) }
func (s symbols) percent() string     { return string(s[8+s[1] : 8+s[2]]) }
func (s symbols) minus() string       { return string(s[8+s[2] : 8+s[3]]) }
func (s symbols) inf() string         { return string(s[8+s[3] : 8+s[4]]) }
func (s symbols) nan() string         { return string(s[8+s[4] : 8+s[5]]) }
func (s symbols) currDecimal() string { return string(s[8+s[5] : 8+s[6]]) }
func (s symbols) currGroup() string   { return string(s[8+s[6] : 8+s[7]]) }

type symbolsID uint8
type symbolsLookup string

func (l symbolsLookup) symbols(id symbolsID) symbols {
	if id == 0 {
		return "\x00\x00\x00\x00\x00\x00\x00\x00"
	}
	i := (id - 1) * 2
	start := binary.BigEndian.Uint16([]byte(l[i : i+2]))
	end := binary.BigEndian.Uint16([]byte(l[i+2:]))
	return symbols(l[start:end])
}

// The zero is a concatenation of all possible rune values.
// The lookup is a string of all existing zero runes.
type zeroID uint8
type zeroLookup string

func (l zeroLookup) zero(id zeroID) rune {
	if id == 0 || int(id) > len(l) {
		return utf8.RuneError
	}
	ch, _ := utf8.DecodeRuneInString(string(l[id-1:]))
	return ch
}

// The numbers data is a tuple consisting of a pattern id, a symbols id, and
// a zero id. The lookup maps a CLDR identity to a numbers data.
type numbers uint32

func (n numbers) patternID() patternID { return patternID((n >> 16) & 0xff) }
func (n numbers) symbolsID() symbolsID { return symbolsID((n >> 8) & 0xff) }
func (n numbers) zeroID() zeroID       { return zeroID(n & 0xff) }

type numbersLookup map[tagID]numbers
