package generate_cldr

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/liblxn/lxnc/internal/generator"
)

// Runestring lookup is used to generate a lookup for single runes.
type runeString struct {
	feature string
	idBits  uint
}

func (l *runeString) imports() []string {
	return []string{"unicode/utf8"}
}

func (l *runeString) generate(p *generator.Printer) {
	p.Println(`// The `, l.feature, ` is a concatenation of all possible rune values.`)
	p.Println(`// The lookup is a string of all existing `, l.feature, ` runes.`)
	p.Println(`type `, l.feature, `ID uint`, l.idBits)
	p.Println(`type `, l.feature, `Lookup string`)
	p.Println()
	p.Println(`func (l `, l.feature, `Lookup) `, l.feature, `(id `, l.feature, `ID) rune {`)
	p.Println(`	if id == 0 || int(id) > len(l) {`)
	p.Println(`		return utf8.RuneError`)
	p.Println(`	}`)
	p.Println(`	ch, _ := utf8.DecodeRuneInString(string(l[id-1:]))`)
	p.Println(`	return ch`)
	p.Println(`}`)
}

func (l *runeString) testImports() []string {
	return []string{"unicode/utf8"}
}

func (l *runeString) generateTest(p *generator.Printer) {
	p.Println(`func Test`, strings.Title(l.feature), `Lookup(t *testing.T) {`)
	p.Println(`	const lookup `, l.feature, `Lookup = "0az"`)
	p.Println()
	p.Println(`	if r := lookup.`, l.feature, `(0); r != utf8.RuneError {`)
	p.Println(`		t.Errorf("unexpected `, l.feature, ` for id 0: %U", r)`)
	p.Println(`	}`)
	p.Println(`	if r := lookup.`, l.feature, `(1); r != '0' {`)
	p.Println(`		t.Errorf("unexpected `, l.feature, ` for id 1: %U", r)`)
	p.Println(`	}`)
	p.Println(`	if r := lookup.`, l.feature, `(2); r != 'a' {`)
	p.Println(`		t.Errorf("unexpected `, l.feature, ` for id 2: %U", r)`)
	p.Println(`	}`)
	p.Println(`	if r := lookup.`, l.feature, `(3); r != 'z' {`)
	p.Println(`		t.Errorf("unexpected `, l.feature, ` for id 3: %U", r)`)
	p.Println(`	}`)
	p.Println(`	if r := lookup.`, l.feature, `(4); r != utf8.RuneError {`)
	p.Println(`		t.Errorf("unexpected `, l.feature, ` for id 4: %U", r)`)
	p.Println(`	}`)
	p.Println(`}`)
}

type runeStringVar struct {
	typ   *runeString
	name  string
	runes []rune
	bytes int
}

func newRuneStringVar(name string, typ *runeString) runeStringVar {
	return runeStringVar{
		typ:  typ,
		name: name,
	}
}

func (v *runeStringVar) runeID(rr rune) uint {
	off := 0
	for _, r := range v.runes {
		if r == rr {
			return uint(off + 1)
		}
		off += utf8.RuneLen(r)
	}
	panic(fmt.Sprintf("rune not found: %c", rr))
}

func (v *runeStringVar) add(r rune) {
	idx := 0
	for idx < len(v.runes) && v.runes[idx] < r {
		idx++
	}
	if idx == len(v.runes) || v.runes[idx] != r {
		v.runes = append(v.runes, 0)
		copy(v.runes[idx+1:], v.runes[idx:])
		v.runes[idx] = r

		v.bytes += utf8.RuneLen(r)
	}
}

func (v *runeStringVar) imports() []string {
	return nil
}

func (v *runeStringVar) generate(p *generator.Printer) {
	p.Println(`const `, v.name, ` `, v.typ.feature, `Lookup = "" + // `, len(v.runes), ` items, `, v.bytes, ` bytes`)

	count := 0
	p.Print(`	"`)
	for _, r := range v.runes {
		if count >= lineLength {
			p.Println(`" +`)
			p.Print(`	"`)
			count = 0
		}
		p.Print(string(r))
		count++
	}
	p.Println(`"`)
}

func (v *runeStringVar) testImports() []string {
	return nil
}

func (v *runeStringVar) generateTest(p *generator.Printer) {
	newID := func(idx int) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.runeID(v.runes[idx]), v.typ.idBits/4)
	}

	newRune := func(r rune) string {
		return `"` + string(r) + `"`
	}

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[`, v.typ.feature, `ID]string{ // `, v.typ.feature, ` id => `, v.typ.feature)

	perLine := int(lineLength / (v.typ.idBits/4 + 2))
	for i := 0; i < len(v.runes); i += perLine {
		n := i + perLine
		if n > len(v.runes) {
			n = len(v.runes)
		}

		p.Print(`		`, newID(i), `: `, newRune(v.runes[i]))
		for k := i + 1; k < n; k++ {
			p.Print(`, `, newID(k), `: `, newRune(v.runes[k]))
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for id, expected := range expected {`)
	p.Println(`		if r := `, v.name, `.`, v.typ.feature, `(id); string(r) != expected {`)
	p.Println(`			t.Fatalf("unexpected `, v.typ.feature, ` for id %d: %s", uint(id), string(r))`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}

// Multistring lookup is used to generate a lookup for multiple strings with
// different lengths.
type multiString struct {
	feature    string
	idBits     uint
	offsetBits uint
	funcs      []string
}

func (l *multiString) emptyString(funcs []string) string {
	return strings.Repeat(`\x00`, len(funcs))
}

func (l *multiString) newString(str ...string) (string, int) {
	if len(str) != len(l.funcs) {
		panic("invalid number of strings")
	}

	offsets := ""
	concat := ""
	off := 0
	for _, s := range str {
		if len(s) > math.MaxUint8 {
			panic(fmt.Sprintf("string exceeds the maximum length: %s", s))
		}
		off += len(s)
		offsets += fmt.Sprintf(`\x%02x`, off)
		concat += s
	}
	return offsets + concat, len(str) + off
}

func (l *multiString) imports() []string {
	if l.offsetBits <= 8 {
		return nil
	}
	return []string{"encoding/binary"}
}

func (l *multiString) generate(p *generator.Printer) {
	off := len(l.funcs)
	maxFunc := 0
	for _, fn := range l.funcs {
		if len(fn) > maxFunc {
			maxFunc = len(fn)
		}
	}

	featureID := l.feature + "ID"
	featureLookup := l.feature + "Lookup"

	// feature type
	p.Println(`// The `, l.feature, ` type is a concatenation of multiple strings. It starts with`)
	p.Println(`// the offsets for each string followed by the actual strings.`)
	p.Println(`// The `, l.feature, ` lookup consists of all `, l.feature, ` strings concatenated.`)
	p.Println(`// It is prefixed with an offset for each `, l.feature, ` block and its id is a`)
	p.Println(`// 1-based index which points to the offset.`)
	p.Println(`type `, l.feature, ` string`)
	p.Println()

	gap := strings.Repeat(" ", maxFunc-len(l.funcs[0]))
	p.Println(`func (s `, l.feature, `) `, l.funcs[0], `() string`, gap, ` { return string(s[`, off, ` : `, off, `+s[0]]) }`)
	for i := 1; i < len(l.funcs); i++ {
		gap = strings.Repeat(" ", maxFunc-len(l.funcs[i]))
		p.Println(`func (s `, l.feature, `) `, l.funcs[i], `() string`, gap, ` { return string(s[`, off, `+s[`, i-1, `] : `, off, `+s[`, i, `]]) }`)
	}

	// feature lookup
	p.Println()
	p.Println(`type `, featureID, ` uint`, l.idBits)
	p.Println(`type `, featureLookup, ` string`)
	p.Println()
	p.Println(`func (l `, featureLookup, `) `, l.feature, `(id `, featureID, `) `, l.feature, ` {`)
	p.Println(`	if id == 0 {`)
	p.Println(`		return "`, l.emptyString(l.funcs), `"`)
	p.Println(`	}`)
	if l.offsetBits <= 8 {
		p.Println(`	start, end := l[id-1], l[id]`)
	} else {
		p.Println(`	i := (id - 1) * `, l.offsetBits/8)
		p.Println(`	start := binary.BigEndian.Uint`, l.offsetBits, `([]byte(l[i : i+`, l.offsetBits/8, `]))`)
		p.Println(`	end := binary.BigEndian.Uint`, l.offsetBits, `([]byte(l[i+`, l.offsetBits/8, `:]))`)
	}
	p.Println(`	return `, l.feature, `(l[start:end])`)
	p.Println(`}`)
}

func (l *multiString) testImports() []string {
	return nil
}

func (l *multiString) generateTest(p *generator.Printer) {
	// feature type
	const length = 2

	newString := func(idx int) string {
		letter := 'a' + rune(idx)
		return strings.Repeat(string(letter), length)
	}

	p.Println(`func Test`, strings.Title(l.feature), `(t *testing.T) {`)

	p.Print(`	const s `, l.feature, ` = "`)
	for i := range l.funcs {
		p.Print(fmt.Sprintf(`\x%02x`, (i+1)*length))
	}
	for i := range l.funcs {
		p.Print(newString(i))
	}
	p.Println(`"`)
	p.Println()

	expected := make([]string, len(l.funcs))
	for i := range l.funcs {
		expected[i] = `"` + newString(i) + `"`
	}
	p.Println(`	expected := [`, len(expected), `]string{`, strings.Join(expected, ", "), `}`)

	getters := make([]string, len(l.funcs))
	for i, fn := range l.funcs {
		getters[i] = fmt.Sprintf("s.%s", fn)
	}
	p.Println(`	get := [`, len(getters), `]func() string{`, strings.Join(getters, ", "), `}`)

	p.Println()
	p.Println(`	for i := 0; i < `, len(l.funcs), `; i++ {`)
	p.Println(`		if str := get[i](); str != expected[i] {`)
	p.Println(`			t.Errorf("unexpected `, l.feature, ` at %d: %s", i, str)`)
	p.Println(`		}`)
	p.Println(`	}`)

	p.Println(`}`)

	// feature lookup
	var offset func(off int) string
	switch {
	case l.offsetBits <= 8:
		offset = func(off int) string {
			return fmt.Sprintf(`\x%02x`, off)
		}
	case l.offsetBits <= 16:
		offset = func(off int) string {
			var buf [2]byte
			binary.BigEndian.PutUint16(buf[:], uint16(off))
			return fmt.Sprintf(`\x%02x\x%02x`, buf[0], buf[1])
		}
	case l.offsetBits <= 32:
		offset = func(off int) string {
			var buf [4]byte
			binary.BigEndian.PutUint32(buf[:], uint32(off))
			return fmt.Sprintf(`\x%02x\x%02x\x%02x\x%02x`, buf[0], buf[1], buf[2], buf[3])
		}
	case l.offsetBits <= 32:
		offset = func(off int) string {
			var buf [8]byte
			binary.BigEndian.PutUint64(buf[:], uint64(off))
			return fmt.Sprintf(`\x%02x\x%02x\x%02x\x%02x\x%02x\x%02x\x%02x\x%02x`, buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7])
		}
	default:
		panic("invalid offset bits")
	}

	newLookup := func(str ...string) string {
		n := (1 + len(str)) * int(l.offsetBits/8)
		offsets := offset(n)
		for _, s := range str {
			n += len(s)
			offsets += offset(n)
		}
		return offsets + strings.Join(str, "")
	}

	featureLookup := l.feature + "Lookup"
	testStrings := [...]string{"foo", "bar", "foobar"}

	p.Println()
	p.Println(`func Test`, strings.Title(featureLookup), `(t *testing.T) {`)
	p.Println(`	const lookup `, featureLookup, ` = "`, newLookup(testStrings[:]...), `"`)
	p.Println()
	for i, str := range testStrings {
		id := i + 1
		p.Println(`	if s := lookup.`, l.feature, `(`, id, `); s != "`, str, `" {`)
		p.Println(`		t.Errorf("unexpected `, l.feature, ` for id `, id, `: %q", s)`)
		p.Println(`	}`)
	}
	p.Println()
	p.Println(`	if s := lookup.`, l.feature, `(0); s != "`, l.emptyString(l.funcs), `" {`)
	p.Println(`		t.Errorf("unexpected symbols for id 0: %q", s)`)
	p.Println(`	}`)
	p.Println(`}`)
}

type multiStringVar struct {
	name    string
	typ     *multiString
	strings []string
	lengths []int
	bytes   int
}

func newMultiStringVar(name string, typ *multiString) multiStringVar {
	return multiStringVar{
		name: name,
		typ:  typ,
	}
}

func (v *multiStringVar) add(str string, length int) {
	switch {
	case v.bytes >= (1<<v.typ.offsetBits)-1:
		panic(fmt.Sprintf("%s length exceeds the maximum, cannot add %q", v.typ.feature, str))
	case len(v.strings) >= (1<<v.typ.idBits)-1:
		panic(fmt.Sprintf("%s count exceeds the maximum, cannot add %q", v.typ.feature, str))
	}

	idx := 0
	for idx < len(v.strings) && v.strings[idx] < str {
		idx++
	}
	if idx == len(v.strings) || v.strings[idx] != str {
		v.strings = append(v.strings, "")
		copy(v.strings[idx+1:], v.strings[idx:])
		v.strings[idx] = str

		v.lengths = append(v.lengths, 0)
		copy(v.lengths[idx+1:], v.lengths[idx:])
		v.lengths[idx] = length

		v.bytes += int(v.typ.offsetBits/8) + length
	}
}

func (v *multiStringVar) stringID(str string) uint {
	for i := 0; i < len(v.strings); i++ {
		if v.strings[i] == str {
			return uint(i + 1)
		}
	}
	panic(fmt.Sprintf("%s not found: %q", v.typ.feature, str))
}

func (v *multiStringVar) imports() []string {
	return nil
}

func (v *multiStringVar) generate(p *generator.Printer) {
	featureLookup := v.typ.feature + "Lookup"

	encodeHex := func(p []byte) string {
		s := ""
		for _, b := range p {
			s += fmt.Sprintf(`\x%02x`, int(b))
		}
		return s
	}

	var hexOffset func(off int) string
	offsetBits := 0
	switch {
	case v.typ.offsetBits <= 8:
		offsetBits = 8
		hexOffset = func(off int) string {
			buf := [1]byte{byte(off)}
			return encodeHex(buf[:])
		}
	case v.typ.offsetBits <= 16:
		offsetBits = 16
		hexOffset = func(off int) string {
			var buf [2]byte
			binary.BigEndian.PutUint16(buf[:], uint16(off))
			return encodeHex(buf[:])
		}
	case v.typ.offsetBits <= 32:
		offsetBits = 32
		hexOffset = func(off int) string {
			var buf [4]byte
			binary.BigEndian.PutUint32(buf[:], uint32(off))
			return encodeHex(buf[:])
		}
	case v.typ.offsetBits <= 64:
		offsetBits = 64
		hexOffset = func(off int) string {
			var buf [8]byte
			binary.BigEndian.PutUint64(buf[:], uint64(off))
			return encodeHex(buf[:])
		}
	default:
		panic("invalid offset bits")
	}

	p.Println(`const `, v.name, ` `, featureLookup, ` = "" + // `, len(v.strings), ` items, `, int(v.typ.offsetBits/8)+v.bytes, ` bytes`)

	count := 0
	offset := (1 + len(v.strings)) * offsetBits / 8

	// string offsets
	s := hexOffset(offset)
	p.Print(`	"`, s)
	count += len(s)
	for _, length := range v.lengths {
		if count > lineLength {
			p.Println(`" +`)
			p.Print(`	"`)
			count = 0
		}
		offset += length
		s = hexOffset(offset)
		p.Print(s)
		count += len(s)
	}
	p.Println(`" +`)

	// strings
	count = 0
	p.Print(`	"`)
	for _, str := range v.strings {
		if count > lineLength {
			p.Println(`" +`)
			p.Print(`	"`)
			count = 0
		}
		p.Print(str)
		count += utf8.RuneCountInString(str)
	}
	p.Println(`"`)
}

func (v *multiStringVar) testImports() []string {
	return nil
}

func (v *multiStringVar) generateTest(p *generator.Printer) {
	featureID := v.typ.feature + "ID"

	newID := func(idx int) string {
		return fmt.Sprintf("%#0[2]*[1]x", idx+1, v.typ.idBits/4)
	}

	newString := func(idx int) string {
		return `"` + v.strings[idx] + `"`
	}

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[`, featureID, `]`, v.typ.feature, `{ // `, v.typ.feature, ` id => `, v.typ.feature)

	perLine := lineLength / int(v.typ.idBits/4+16)
	for i := 0; i < len(v.strings); i += perLine {
		n := i + perLine
		if n > len(v.strings) {
			n = len(v.strings)
		}

		p.Print(`		`, newID(i), `: `, newString(i))
		for k := i + 1; k < n; k++ {
			p.Print(`, `, newID(k), `: `, newString(k))
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for expectedID, expectedStr := range expected {`)
	p.Println(`		if s := `, v.name, `.`, v.typ.feature, `(expectedID); s != expectedStr {`)
	p.Println(`			t.Fatalf("unexpected `, v.typ.feature, ` string for id %d: %s", uint(expectedID), s)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}

// String block lookup is used to generate a lookup for multiple strings where
// all strings have the same length. The string id is 1-based.
type stringBlock struct {
	feature   string
	idBits    uint
	blocksize int
}

func (l *stringBlock) imports() []string {
	return []string{"sort"}
}

func (l *stringBlock) generate(p *generator.Printer) {
	featureID := l.feature + "ID"
	featureLookup := l.feature + "Lookup"

	p.Println(`// A `, l.feature, ` id is an identifier of a specific fixed-width string and defines`)
	p.Println(`// a 1-based index into a lookup string. The lookup consists of concatenated`)
	p.Println(`// blocks of size `, l.blocksize, `, where each block contains a `, l.feature, ` string.`)
	p.Println(`type `, featureID, ` uint`, l.idBits)
	p.Println()
	p.Println(`type `, featureLookup, ` string`)
	p.Println()
	p.Println(`func (l `, featureLookup, `) `, l.feature, `(id `, featureID, `) string {`)
	p.Println(`	if id == 0 || `, l.blocksize, `*int(id) > len(l) {`)
	p.Println(`		return ""`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	code := l[int(id-1)*`, l.blocksize, ` : int(id)*`, l.blocksize, `]`)
	p.Println(`	end := `, l.blocksize)
	p.Println(`	for end > 0 && code[end-1] == ' ' {`)
	p.Println(`		end--`)
	p.Println(`	}`)
	p.Println(`	return string(code[:end])`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (l `, featureLookup, `) `, featureID, `(str []byte) `, featureID, ` {`)
	p.Println(`	idx := sort.Search(len(l)/`, l.blocksize, `, func(i int) bool {`)
	p.Println(`		return l[i*`, l.blocksize, `:(i+1)*`, l.blocksize, `] >= `, featureLookup, `(str)`)
	p.Println(`	})`)
	p.Println()
	p.Println(`	if idx*`, l.blocksize, ` < len(l) && l.`, l.feature, `(`, featureID, `(idx+1)) == string(str) {`)
	p.Println(`		return `, featureID, `(idx + 1)`)
	p.Println(`	}`)
	p.Println(`	return 0`)
	p.Println(`}`)
}

func (l *stringBlock) testImports() []string {
	return nil
}

func (l *stringBlock) generateTest(p *generator.Printer) {
	featureID := l.feature + "ID"
	featureLookup := l.feature + "Lookup"

	var (
		expected []string
		lookup   string
	)
	for i := 1; i <= l.blocksize; i++ {
		letters := strings.Repeat(string('a'+rune(i)-1), i)

		expected = append(expected, `"`+letters+`"`)
		lookup += letters + strings.Repeat(" ", l.blocksize-i)
	}

	p.Println(`func Test`, strings.Title(featureLookup), `(t *testing.T) {`)
	p.Println(`	expected := [`, l.blocksize, `]string{`, strings.Join(expected, ", "), `}`)
	p.Println(`	lookup := `, featureLookup, `("`, lookup, `")`)
	p.Println()
	p.Println(`	for i, expectedStr := range expected {`)
	p.Println(`		if id := lookup.`, featureID, `([]byte(expectedStr)); id != `, featureID, `(i+1) {`)
	p.Println(`			t.Errorf("unexpected `, l.feature, ` id for %q: %d", expectedStr, id)`)
	p.Println(`		}`)
	p.Println(`		if str := lookup.`, l.feature, `(`, featureID, `(i + 1)); str != expectedStr {`)
	p.Println(`			t.Errorf("unexpected string for `, l.feature, ` id %d: %s", i+1, str)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	if id := lookup.`, featureID, `([]byte{'1'}); id != 0 {`)
	p.Println(`		t.Errorf("unexpected `, l.feature, ` id: %d", id)`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	if str := lookup.`, l.feature, `(0); str != "" {`)
	p.Println(`		t.Errorf("unexpected string for id 0: %s", str)`)
	p.Println(`	}`)
	p.Println(`	if str := lookup.`, l.feature, `(`, featureID, `(len(lookup) + 1)); str != "" {`)
	p.Println(`		t.Errorf("unexpected string id %d: %s", len(lookup)+1, str)`)
	p.Println(`	}`)
	p.Println(`}`)
}

type stringBlockVar struct {
	typ     *stringBlock
	name    string
	strings []string
}

func newStringBlockVar(name string, typ *stringBlock) stringBlockVar {
	return stringBlockVar{
		typ:  typ,
		name: name,
	}
}

func (v *stringBlockVar) add(str string) {
	if str == "" {
		return
	}

	switch {
	case len(v.strings) == (1<<v.typ.idBits)-1:
		panic(fmt.Sprintf("number of %s strings exceeds the maximum (%d), cannot add %s", v.typ.feature, (1<<v.typ.idBits)-1, str))
	case len(str) > v.typ.blocksize:
		panic(fmt.Sprintf("%s string length exceeds the limit of %d: %s", v.typ.feature, v.typ.blocksize, str))
	}

	str += strings.Repeat(" ", v.typ.blocksize-len(str))
	idx := 0
	for idx < len(v.strings) && v.strings[idx] < str {
		idx++
	}
	if idx == len(v.strings) || v.strings[idx] != str {
		v.strings = append(v.strings, "")
		copy(v.strings[idx+1:], v.strings[idx:])
		v.strings[idx] = str
	}
}

func (v *stringBlockVar) stringID(str string) uint {
	str += strings.Repeat(" ", v.typ.blocksize-len(str))
	for i := 0; i < len(v.strings); i++ {
		if v.strings[i] == str {
			return uint(i + 1)
		}
	}
	panic(fmt.Sprintf("%s string not found: %s", v.typ.feature, str))
}

func (v *stringBlockVar) imports() []string {
	return nil
}

func (v *stringBlockVar) generate(p *generator.Printer) {
	featureLookup := v.typ.feature + "Lookup"

	p.Println(`const `, v.name, ` `, featureLookup, ` = "" + // `, len(v.strings), ` items, `, len(v.strings)*v.typ.blocksize, ` bytes`)

	perLine := lineLength / v.typ.blocksize
	strings := v.strings
	for len(strings) > perLine {
		p.Print(`	"`)
		for i := 0; i < perLine; i++ {
			p.Print(strings[i])
		}
		p.Println(`" +`)

		strings = strings[perLine:]
	}

	p.Print(`	"`)
	for _, str := range strings {
		p.Print(str)
	}
	p.Println(`"`)
}

func (v *stringBlockVar) testImports() []string {
	return []string{"strings"}
}

func (v *stringBlockVar) generateTest(p *generator.Printer) {
	newID := func(idx int) string {
		return fmt.Sprintf("%#0[2]*[1]x", idx+1, v.typ.idBits/4)
	}

	featureID := v.typ.feature + "ID"

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[`, featureID, `]string{ // `, v.typ.feature, ` id => string`)

	perLine := lineLength / (int(v.typ.idBits)/8 + v.typ.blocksize + 6)
	for i := 0; i < len(v.strings); i += perLine {
		n := i + perLine
		if n > len(v.strings) {
			n = len(v.strings)
		}

		p.Print(`		`, newID(i), `: "`, v.strings[i], `"`)
		for k := i + 1; k < n; k++ {
			p.Print(`, `, newID(k), `: "`, v.strings[k], `"`)
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for id, str := range expected {`)
	p.Println(`		if s := `, v.name, `.`, v.typ.feature, `(id); s != strings.TrimSpace(str) {`)
	p.Println(`			t.Fatalf("unexpected string for id %d: %q", uint(id), s)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}
