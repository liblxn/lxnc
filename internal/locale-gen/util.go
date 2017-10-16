package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

// multi string
type _multiString struct{}

func (l _multiString) emptyString(funcs []string) string {
	return strings.Repeat(`\x00`, len(funcs))
}

func (l _multiString) newString(str ...string) (string, int) {
	offsets := ""
	off := 0
	for _, s := range str {
		if len(s) > math.MaxUint8 {
			panic(fmt.Sprintf("string exceeds the maximum length: %s", s))
		}
		off += len(s)
		offsets += fmt.Sprintf(`\x%02x`, off)
	}
	return offsets + strings.Join(str, ""), len(str) + off
}

func (l _multiString) imports(offsetBits uint) []string {
	if offsetBits <= 8 {
		return nil
	}
	return []string{"encoding/binary"}
}

func (l _multiString) generate(p *printer, feature string, idBits uint, offsetBits uint, funcs []string) {
	off := len(funcs)
	maxFunc := 0
	for _, fn := range funcs {
		if len(fn) > maxFunc {
			maxFunc = len(fn)
		}
	}

	featureID := feature + "ID"
	featureLookup := feature + "Lookup"

	// feature type
	p.Println(`// The `, feature, ` type is a concatenation of multiple strings. It starts with`)
	p.Println(`// the offsets for each string followed by the actual strings.`)
	p.Println(`// The `, feature, ` lookup consists of all `, feature, ` strings concatenated.`)
	p.Println(`// It is prefixed with an offset for each `, feature, ` block and its id is a`)
	p.Println(`// 1-based index which points to the offset.`)
	p.Println(`type `, feature, ` string`)
	p.Println()

	gap := strings.Repeat(" ", maxFunc-len(funcs[0]))
	p.Println(`func (s `, feature, `) `, funcs[0], `() string`, gap, ` { return string(s[`, off, ` : `, off, `+s[0]]) }`)
	for i := 1; i < len(funcs); i++ {
		gap = strings.Repeat(" ", maxFunc-len(funcs[i]))
		p.Println(`func (s `, feature, `) `, funcs[i], `() string`, gap, ` { return string(s[`, off, `+s[`, i-1, `] : `, off, `+s[`, i, `]]) }`)
	}

	// feature lookup
	p.Println()
	p.Println(`type `, featureID, ` uint`, idBits)
	p.Println(`type `, featureLookup, ` string`)
	p.Println()
	p.Println(`func (l `, featureLookup, `) `, feature, `(id `, featureID, `) `, feature, ` {`)
	p.Println(`	if id == 0 {`)
	p.Println(`		return "`, l.emptyString(funcs), `"`)
	p.Println(`	}`)
	if offsetBits <= 8 {
		p.Println(`	start, end := l[id-1], l[id]`)
	} else {
		p.Println(`	i := (id - 1) * `, offsetBits/8)
		p.Println(`	start := binary.BigEndian.Uint`, offsetBits, `([]byte(l[i : i+`, offsetBits/8, `]))`)
		p.Println(`	end := binary.BigEndian.Uint`, offsetBits, `([]byte(l[i+`, offsetBits/8, `:]))`)
	}
	p.Println(`	return `, feature, `(l[start:end])`)
	p.Println(`}`)
}

func (l _multiString) testImports() []string {
	return nil
}

func (l _multiString) generateTest(p *printer, feature string, bits uint, funcs []string) {
	// feature type
	const length = 2

	newString := func(idx int) string {
		letter := 'a' + rune(idx)
		return strings.Repeat(string(letter), length)
	}

	p.Println(`func Test`, strings.Title(feature), `(t *testing.T) {`)

	p.Print(`	const s `, feature, ` = "`)
	for i := range funcs {
		p.Print(fmt.Sprintf(`\x%02x`, (i+1)*length))
	}
	for i := range funcs {
		p.Print(newString(i))
	}
	p.Println(`"`)
	p.Println()

	expected := make([]string, len(funcs))
	for i := range funcs {
		expected[i] = `"` + newString(i) + `"`
	}
	p.Println(`	expected := [`, len(expected), `]string{`, strings.Join(expected, ", "), `}`)

	getters := make([]string, len(funcs))
	for i, fn := range funcs {
		getters[i] = fmt.Sprintf("s.%s", fn)
	}
	p.Println(`	get := [`, len(getters), `]func() string{`, strings.Join(getters, ", "), `}`)

	p.Println()
	p.Println(`	for i := 0; i < `, len(funcs), `; i++ {`)
	p.Println(`		if str := get[i](); str != expected[i] {`)
	p.Println(`			t.Errorf("unexpected `, feature, ` at %d: %s", i, str)`)
	p.Println(`		}`)
	p.Println(`	}`)

	p.Println(`}`)

	// feature lookup
	var offset func(off int) string
	switch {
	case bits <= 8:
		offset = func(off int) string {
			return fmt.Sprintf(`\x%02x`, off)
		}
	case bits <= 16:
		offset = func(off int) string {
			var buf [2]byte
			binary.BigEndian.PutUint16(buf[:], uint16(off))
			return fmt.Sprintf(`\x%02x\x%02x`, buf[0], buf[1])
		}
	case bits <= 32:
		offset = func(off int) string {
			var buf [4]byte
			binary.BigEndian.PutUint32(buf[:], uint32(off))
			return fmt.Sprintf(`\x%02x\x%02x\x%02x\x%02x`, buf[0], buf[1], buf[2], buf[3])
		}
	default:
		offset = func(off int) string {
			var buf [8]byte
			binary.BigEndian.PutUint64(buf[:], uint64(off))
			return fmt.Sprintf(`\x%02x\x%02x\x%02x\x%02x\x%02x\x%02x\x%02x\x%02x`, buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7])
		}
	}

	newLookup := func(str ...string) string {
		n := (1 + len(str)) * int(bits/8)
		offsets := offset(n)
		for _, s := range str {
			n += len(s)
			offsets += offset(n)
		}
		return offsets + strings.Join(str, "")
	}

	featureLookup := feature + "Lookup"
	testStrings := [...]string{"foo", "bar", "foobar"}

	p.Println()
	p.Println(`func Test`, strings.Title(featureLookup), `(t *testing.T) {`)
	p.Println(`	const lookup `, featureLookup, ` = "`, newLookup(testStrings[:]...), `"`)
	p.Println()
	for i, str := range testStrings {
		id := i + 1
		p.Println(`	if s := lookup.`, feature, `(`, id, `); s != "`, str, `" {`)
		p.Println(`		t.Errorf("unexpected `, feature, ` for id `, id, `: %q", s)`)
		p.Println(`	}`)
	}
	p.Println()
	p.Println(`	if s := lookup.`, feature, `(0); s != "`, l.emptyString(funcs), `" {`)
	p.Println(`		t.Errorf("unexpected symbols for id 0: %q", s)`)
	p.Println(`	}`)
	p.Println(`}`)
}

type _multiStringLookupVar struct {
	feature    string
	offsetBits uint
	idBits     uint

	name    string
	strings []string
	lengths []int
	bytes   int
}

func (v *_multiStringLookupVar) add(str string, length int) {
	switch {
	case v.bytes >= (1<<v.offsetBits)-1:
		panic(fmt.Sprintf("%s length exceeds the maximum, cannot add %q", v.feature, str))
	case len(v.strings) >= (1<<v.idBits)-1:
		panic(fmt.Sprintf("%s count exceeds the maximum, cannot add %q", v.feature, str))
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

		v.bytes += int(v.offsetBits/8) + length
	}
}

func (v *_multiStringLookupVar) stringID(str string) uint {
	for i := 0; i < len(v.strings); i++ {
		if v.strings[i] == str {
			return uint(i + 1)
		}
	}
	panic(fmt.Sprintf("%s not found: %q", v.feature, str))
}

func (v *_multiStringLookupVar) imports() []string {
	return nil
}

func (v *_multiStringLookupVar) generate(p *printer) {
	featureLookup := v.feature + "Lookup"

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
	case v.offsetBits <= 8:
		offsetBits = 8
		hexOffset = func(off int) string {
			buf := [1]byte{byte(off)}
			return encodeHex(buf[:])
		}
	case v.offsetBits <= 16:
		offsetBits = 16
		hexOffset = func(off int) string {
			var buf [2]byte
			binary.BigEndian.PutUint16(buf[:], uint16(off))
			return encodeHex(buf[:])
		}
	case v.offsetBits <= 32:
		offsetBits = 32
		hexOffset = func(off int) string {
			var buf [4]byte
			binary.BigEndian.PutUint32(buf[:], uint32(off))
			return encodeHex(buf[:])
		}
	default:
		offsetBits = 64
		hexOffset = func(off int) string {
			var buf [8]byte
			binary.BigEndian.PutUint64(buf[:], uint64(off))
			return encodeHex(buf[:])
		}
	}

	p.Println(`const `, v.name, ` `, featureLookup, ` = "" + // `, len(v.strings), ` items, `, int(v.offsetBits/8)+v.bytes, ` bytes`)

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

func (v *_multiStringLookupVar) testImports() []string {
	return nil
}

func (v *_multiStringLookupVar) generateTest(p *printer) {
	featureID := v.feature + "ID"

	newID := func(idx int) string {
		return fmt.Sprintf("%#0[2]*[1]x", idx+1, v.idBits/4)
	}

	newString := func(idx int) string {
		return `"` + v.strings[idx] + `"`
	}

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[`, featureID, `]`, v.feature, `{ // `, v.feature, ` id => `, v.feature)

	perLine := lineLength / int(v.idBits/4+16)
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
	p.Println(`		if s := `, v.name, `.`, v.feature, `(expectedID); s != expectedStr {`)
	p.Println(`			t.Fatalf("unexpected `, v.feature, ` string for id %d: %s", uint(expectedID), s)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}

// string block lookup
type _stringBlockLookup struct{}

func (l _stringBlockLookup) imports() []string {
	return []string{"sort"}
}

func (l _stringBlockLookup) generate(p *printer, feature string, idBits int, blocksize int) {
	featureID := feature + "ID"
	featureLookup := feature + "Lookup"

	p.Println(`// A `, feature, ` id is an identifier of a specific fixed-width string and defines`)
	p.Println(`// a 1-based index into a lookup string. The lookup consists of concatenated`)
	p.Println(`// blocks of size `, blocksize, `, where each block contains a `, feature, ` string.`)
	p.Println(`type `, featureID, ` uint`, idBits)
	p.Println()
	p.Println(`type `, featureLookup, ` string`)
	p.Println()
	p.Println(`func (l `, featureLookup, `) `, feature, `(id `, featureID, `) string {`)
	p.Println(`	if id == 0 || `, blocksize, `*int(id) > len(l) {`)
	p.Println(`		return ""`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	code := l[int(id-1)*`, blocksize, ` : int(id)*`, blocksize, `]`)
	p.Println(`	end := `, blocksize)
	p.Println(`	for end > 0 && code[end-1] == ' ' {`)
	p.Println(`		end--`)
	p.Println(`	}`)
	p.Println(`	return string(code[:end])`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (l `, featureLookup, `) `, featureID, `(str []byte) `, featureID, ` {`)
	p.Println(`	idx := sort.Search(len(l)/`, blocksize, `, func(i int) bool {`)
	p.Println(`		return l[i*`, blocksize, `:(i+1)*`, blocksize, `] >= `, featureLookup, `(str)`)
	p.Println(`	})`)
	p.Println()
	p.Println(`	if idx*`, blocksize, ` < len(l) && l.`, feature, `(`, featureID, `(idx+1)) == string(str) {`)
	p.Println(`		return `, featureID, `(idx + 1)`)
	p.Println(`	}`)
	p.Println(`	return 0`)
	p.Println(`}`)
}

func (l _stringBlockLookup) testImports() []string {
	return nil
}

func (l _stringBlockLookup) generateTest(p *printer, feature string, blocksize int) {
	featureID := feature + "ID"
	featureLookup := feature + "Lookup"

	var (
		expected []string
		lookup   string
	)
	for i := 1; i <= blocksize; i++ {
		letters := strings.Repeat(string('a'+rune(i)-1), i)

		expected = append(expected, `"`+letters+`"`)
		lookup += letters + strings.Repeat(" ", blocksize-i)
	}

	p.Println(`func Test`, strings.Title(featureLookup), `(t *testing.T) {`)
	p.Println(`	expected := [`, blocksize, `]string{`, strings.Join(expected, ", "), `}`)
	p.Println(`	lookup := `, featureLookup, `("`, lookup, `")`)
	p.Println()
	p.Println(`	for i, expectedStr := range expected {`)
	p.Println(`		if id := lookup.`, featureID, `([]byte(expectedStr)); id != `, featureID, `(i+1) {`)
	p.Println(`			t.Errorf("unexpected `, feature, ` id for %q: %d", expectedStr, id)`)
	p.Println(`		}`)
	p.Println(`		if str := lookup.`, feature, `(`, featureID, `(i + 1)); str != expectedStr {`)
	p.Println(`			t.Errorf("unexpected string for `, feature, ` id %d: %s", i+1, str)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	if id := lookup.`, featureID, `([]byte{'1'}); id != 0 {`)
	p.Println(`		t.Errorf("unexpected `, feature, ` id: %d", id)`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	if str := lookup.`, feature, `(0); str != "" {`)
	p.Println(`		t.Errorf("unexpected string for id 0: %s", str)`)
	p.Println(`	}`)
	p.Println(`	if str := lookup.`, feature, `(`, featureID, `(len(lookup) + 1)); str != "" {`)
	p.Println(`		t.Errorf("unexpected string id %d: %s", len(lookup)+1, str)`)
	p.Println(`	}`)
	p.Println(`}`)
}

type _stringBlockLookupVar struct {
	feature   string
	idBits    uint
	blocksize int

	name    string
	strings []string
}

func (v *_stringBlockLookupVar) add(str string) {
	switch {
	case len(v.strings) == (1<<v.idBits)-1:
		panic(fmt.Sprintf("number of %s strings exceeds the maximum, cannot add %s", v.feature, str))
	case len(str) > v.blocksize:
		panic(fmt.Sprintf("%s string length exceeds the limit of %d: %s", v.feature, v.blocksize, str))
	}

	str += strings.Repeat(" ", v.blocksize-len(str))
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

func (v *_stringBlockLookupVar) stringID(str string) uint {
	str += strings.Repeat(" ", v.blocksize-len(str))
	for i := 0; i < len(v.strings); i++ {
		if v.strings[i] == str {
			return uint(i + 1)
		}
	}
	panic(fmt.Sprintf("%s string not found: %s", v.feature, str))
}

func (v *_stringBlockLookupVar) imports() []string {
	return nil
}

func (v *_stringBlockLookupVar) generate(p *printer) {
	featureLookup := v.feature + "Lookup"

	p.Println(`const `, v.name, ` `, featureLookup, ` = "" + // `, len(v.strings), ` items, `, len(v.strings)*v.blocksize, ` bytes`)

	perLine := lineLength / v.blocksize
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

func (v *_stringBlockLookupVar) testImports() []string {
	return []string{"strings"}
}

func (v *_stringBlockLookupVar) generateTest(p *printer) {
	newID := func(idx int) string {
		return fmt.Sprintf("%#0[2]*[1]x", idx+1, v.idBits/4)
	}

	featureID := v.feature + "ID"

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[`, featureID, `]string{ // `, v.feature, ` id => string`)

	perLine := lineLength / (int(v.idBits)/8 + v.blocksize + 6)
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
	p.Println(`		if s := `, v.name, `.`, v.feature, `(id); s != strings.TrimSpace(str) {`)
	p.Println(`			t.Fatalf("unexpected string for id %d: %q", uint(id), s)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}
