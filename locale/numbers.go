package locale

import (
	"encoding/binary"
	"unicode/utf8"
)

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
	if id == 0 || int(id) > len(l) {
		return 0
	}
	return l[id-1]
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

// The zero is used to determine the digits. A digit n in [0,9] is determined by
// adding n to the zero. The lookup is a string of all existing zero runes.
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
