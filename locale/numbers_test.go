package locale

import (
	"testing"
	"unicode/utf8"
)

func TestAffix(t *testing.T) {
	const s affix = "\x02\x04aabb"

	expected := [2]string{"aa", "bb"}
	get := [2]func() string{s.prefix, s.suffix}

	for i := 0; i < 2; i++ {
		if str := get[i](); str != expected[i] {
			t.Errorf("unexpected affix at %d: %s", i, str)
		}
	}
}

func TestAffixLookup(t *testing.T) {
	const lookup affixLookup = "\x04\x07\x0a\x10foobarfoobar"

	if s := lookup.affix(1); s != "foo" {
		t.Errorf("unexpected affix for id 1: %q", s)
	}
	if s := lookup.affix(2); s != "bar" {
		t.Errorf("unexpected affix for id 2: %q", s)
	}
	if s := lookup.affix(3); s != "foobar" {
		t.Errorf("unexpected affix for id 3: %q", s)
	}

	if s := lookup.affix(0); s != "\x00\x00" {
		t.Errorf("unexpected symbols for id 0: %q", s)
	}
}

func TestPattern(t *testing.T) {
	const pattern pattern = 0x102345678

	if id := pattern.posAffixID(); id != 1 {
		t.Errorf("unexpected positive affix id: %d", id)
	}
	if id := pattern.negAffixID(); id != 2 {
		t.Errorf("unexpected negative affix id: %d", id)
	}
	if n := pattern.minIntDigits(); n != 3 {
		t.Errorf("unexpected minimum integer digits: %d", n)
	}
	if n := pattern.minFracDigits(); n != 4 {
		t.Errorf("unexpected minimum fraction digits: %d", n)
	}
	if n := pattern.maxFracDigits(); n != 5 {
		t.Errorf("unexpected minimum fraction digits: %d", n)
	}
	if m, n := pattern.intGrouping(); m != 6 || n != 7 {
		t.Errorf("unexpected integer grouping: (%d, %d)", m, n)
	}
	if n := pattern.fracGrouping(); n != 8 {
		t.Errorf("unexpected fraction grouping: %d", n)
	}
}

func TestPatternLookup(t *testing.T) {
	lookup := patternLookup{1, 2, 3}

	for i := 0; i < len(lookup); i++ {
		if p := lookup.pattern(patternID(i + 1)); p != lookup[i] {
			t.Errorf("unexpected pattern for id %d: %#x", i+1, p)
		}
	}

	if p := lookup.pattern(0); p != 0 {
		t.Errorf("unexpected pattern for id 0: %#x", p)
	}
	if p := lookup.pattern(patternID(len(lookup) + 1)); p != 0 {
		t.Errorf("unexpected pattern for id %d: %#x", len(lookup)+1, p)
	}
}

func TestSymbols(t *testing.T) {
	const s symbols = "\x02\x04\x06\x08\x0a\x0c\x0e\x10aabbccddeeffgghh"

	expected := [8]string{"aa", "bb", "cc", "dd", "ee", "ff", "gg", "hh"}
	get := [8]func() string{s.decimal, s.group, s.percent, s.minus, s.inf, s.nan, s.currDecimal, s.currGroup}

	for i := 0; i < 8; i++ {
		if str := get[i](); str != expected[i] {
			t.Errorf("unexpected symbols at %d: %s", i, str)
		}
	}
}

func TestSymbolsLookup(t *testing.T) {
	const lookup symbolsLookup = "\x00\x08\x00\x0b\x00\x0e\x00\x14foobarfoobar"

	if s := lookup.symbols(1); s != "foo" {
		t.Errorf("unexpected symbols for id 1: %q", s)
	}
	if s := lookup.symbols(2); s != "bar" {
		t.Errorf("unexpected symbols for id 2: %q", s)
	}
	if s := lookup.symbols(3); s != "foobar" {
		t.Errorf("unexpected symbols for id 3: %q", s)
	}

	if s := lookup.symbols(0); s != "\x00\x00\x00\x00\x00\x00\x00\x00" {
		t.Errorf("unexpected symbols for id 0: %q", s)
	}
}

func TestZeroLookup(t *testing.T) {
	const lookup zeroLookup = "0az"

	if z := lookup.zero(0); z != utf8.RuneError {
		t.Errorf("unexpected zero for id 0: %U", z)
	}
	if z := lookup.zero(1); z != '0' {
		t.Errorf("unexpected zero for id 1: %U", z)
	}
	if z := lookup.zero(2); z != 'a' {
		t.Errorf("unexpected zero for id 2: %U", z)
	}
	if z := lookup.zero(3); z != 'z' {
		t.Errorf("unexpected zero for id 3: %U", z)
	}
	if z := lookup.zero(4); z != utf8.RuneError {
		t.Errorf("unexpected zero for id 4: %U", z)
	}
}

func TestNumbers(t *testing.T) {
	const numbers numbers = 0x10203

	if id := numbers.patternID(); id != 1 {
		t.Errorf("unexpected pattern id: %d", id)
	}
	if id := numbers.symbolsID(); id != 2 {
		t.Errorf("unexpected symbols id: %d", id)
	}
	if id := numbers.zeroID(); id != 3 {
		t.Errorf("unexpected zero id: %d", id)
	}
}
