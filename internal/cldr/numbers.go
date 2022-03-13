package cldr

import (
	"encoding/xml"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/liblxn/lxnc/internal/errors"
)

// Placeholders for the number format prefixes and suffixes.
const (
	CurrencyPlaceholder rune = 0x00a4 // '¤'
	PercentPlaceholder  rune = '%'
	PermillePlaceholder rune = 0x2030 // '‰'
)

// Numbers holds all the relevant data which is necessary for formatting
// numbers in a specific locale.
type Numbers struct {
	DefaultSystem     string                   // default numbering system
	MinGroupingDigits int                      // minimum number of digits to enable grouping (zero: not set)
	Symbols           map[string]NumberSymbols // numbering system => symbols
	DecimalFormats    map[string]NumberFormat  // numbering system => number format
	ScientificFormats map[string]NumberFormat  // numbering system => number format
	PercentFormats    map[string]NumberFormat  // numbering system => number format
	CurrencyFormats   map[string]NumberFormat  // numbering system => number format
}

func (n *Numbers) empty() bool {
	return n.DefaultSystem == "" &&
		n.MinGroupingDigits == 0 &&
		len(n.Symbols) == 0 &&
		len(n.DecimalFormats) == 0 &&
		len(n.ScientificFormats) == 0 &&
		len(n.PercentFormats) == 0 &&
		len(n.CurrencyFormats) == 0
}

func (n *Numbers) decode(d *xmlDecoder, _ xml.StartElement) {
	n.Symbols = make(map[string]NumberSymbols)
	n.DecimalFormats = make(map[string]NumberFormat)
	n.ScientificFormats = make(map[string]NumberFormat)
	n.PercentFormats = make(map[string]NumberFormat)
	n.CurrencyFormats = make(map[string]NumberFormat)

	formatDecoder := func(formats map[string]NumberFormat) decodeFunc {
		return func(d *xmlDecoder, elem xml.StartElement) {
			if sys := xmlAttrib(elem, "numberSystem"); sys == "" {
				d.SkipElem()
			} else {
				var nf NumberFormat
				nf.decode(d, elem)
				if !nf.empty() {
					formats[sys] = nf
				}
			}
		}
	}

	d.DecodeElems(decoders{
		"defaultNumberingSystem": func(d *xmlDecoder, elem xml.StartElement) {
			n.DefaultSystem = d.ReadString(elem)
			d.SkipElem()
		},
		"minimumGroupingDigits": func(d *xmlDecoder, elem xml.StartElement) {
			n.MinGroupingDigits = d.ReadInt(elem)
			d.SkipElem()
		},
		"symbols": func(d *xmlDecoder, elem xml.StartElement) {
			sys := xmlAttrib(elem, "numberSystem")
			if sys == "" {
				sys = "latn"
			}
			var symbols NumberSymbols
			symbols.decode(d, elem)
			n.Symbols[sys] = symbols
		},
		"decimalFormats":    formatDecoder(n.DecimalFormats),
		"scientificFormats": formatDecoder(n.ScientificFormats),
		"percentFormats":    formatDecoder(n.PercentFormats),
		"currencyFormats":   formatDecoder(n.CurrencyFormats),
	})

	// normalize
	for sys, symbols := range n.Symbols {
		if symbols.Alias == "" {
			continue
		}
		if alias, has := n.Symbols[symbols.Alias]; has {
			alias.Alias = symbols.Alias
			n.Symbols[sys] = alias
		}
	}
}

// NumberSymbols holds all symbols which should be used to format a number
// in a specific locale. These symbols should replace the placeholders in
// the prefixes and suffixes.
type NumberSymbols struct {
	Alias               string // numbering system this data is copied from (empty for none)
	Decimal             string // decimal separator
	Group               string // group separator
	Percent             string // percentage sign
	Permille            string // permille sign
	Plus                string // plus sign for positive numbers
	Minus               string // minus sign for negative numbers
	Exponential         string // separator for mantissa and exponent
	SuperscriptExponent string // separator for the number and the exponential notation
	Infinity            string // sign for infinity
	NaN                 string // sign for not-a-number
	TimeSeparator       string // time separator
	CurrencyDecimal     string // decimal separator for currencies
	CurrencyGroup       string // group separator for currencies
}

func (n *NumberSymbols) merge(symb NumberSymbols) {
	if n.Decimal == "" {
		n.Decimal = symb.Decimal
	}
	if n.Group == "" {
		n.Group = symb.Group
	}
	if n.Percent == "" {
		n.Percent = symb.Percent
	}
	if n.Permille == "" {
		n.Permille = symb.Permille
	}
	if n.Plus == "" {
		n.Plus = symb.Plus
	}
	if n.Minus == "" {
		n.Minus = symb.Minus
	}
	if n.Exponential == "" {
		n.Exponential = symb.Exponential
	}
	if n.SuperscriptExponent == "" {
		n.SuperscriptExponent = symb.SuperscriptExponent
	}
	if n.Infinity == "" {
		n.Infinity = symb.Infinity
	}
	if n.NaN == "" {
		n.NaN = symb.NaN
	}
	if n.TimeSeparator == "" {
		n.TimeSeparator = symb.TimeSeparator
	}
	if n.CurrencyDecimal == "" {
		n.CurrencyDecimal = symb.CurrencyDecimal
	}
	if n.CurrencyGroup == "" {
		n.CurrencyGroup = symb.CurrencyGroup
	}
}

func (n *NumberSymbols) decode(d *xmlDecoder, elem xml.StartElement) {
	stringDecoder := func(s *string) decodeFunc {
		return func(d *xmlDecoder, elem xml.StartElement) {
			*s = d.ReadString(elem)
			d.SkipElem()
		}
	}

	d.DecodeElems(decoders{
		"decimal":                stringDecoder(&n.Decimal),
		"group":                  stringDecoder(&n.Group),
		"percentSign":            stringDecoder(&n.Percent),
		"perMille":               stringDecoder(&n.Permille),
		"plusSign":               stringDecoder(&n.Plus),
		"minusSign":              stringDecoder(&n.Minus),
		"exponential":            stringDecoder(&n.Exponential),
		"superscriptingExponent": stringDecoder(&n.SuperscriptExponent),
		"infinity":               stringDecoder(&n.Infinity),
		"nan":                    stringDecoder(&n.NaN),
		"timeSeparator":          stringDecoder(&n.TimeSeparator),
		"currencyDecimal":        stringDecoder(&n.CurrencyDecimal),
		"currencyGroup":          stringDecoder(&n.CurrencyGroup),
		"alias": func(d *xmlDecoder, elem xml.StartElement) {
			if xmlAttrib(elem, "source") != "locale" {
				return
			}

			// The attribute 'path' should have the following form:
			//	"../symbols[@numberSystem='<system>']"
			const prefix, suffix = "../symbols[@numberSystem='", "']"
			path := xmlAttrib(elem, "path")
			if strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix) {
				n.Alias = path[len(prefix) : len(path)-len(suffix)]
			}
			d.SkipElem()
		},
	})

	// normalize
	if n.CurrencyDecimal == "" {
		n.CurrencyDecimal = n.Decimal
	}
	if n.CurrencyGroup == "" {
		n.CurrencyGroup = n.Group
	}
}

// Grouping holds the sizes for grouping the number digits.
type Grouping struct {
	PrimarySize   int
	SecondarySize int
}

func (g *Grouping) normalize() {
	if g.SecondarySize == 0 {
		g.SecondarySize = g.PrimarySize
	}
}

// PaddingPos describes the position of the padding characters.
type PaddingPos int

// Valid padding positions.
const (
	BeforePrefix PaddingPos = iota
	AfterPrefix
	BeforeSuffix
	AfterSuffix
)

// Padding holds the padding data when a data when a number is formatted.
type Padding struct {
	Width int // number only (excluding affixes)
	Char  rune
	Pos   PaddingPos
}

func (p *Padding) normalize() {
	if p.Width == 0 {
		p.Char = 0
	} else if p.Char == 0 {
		p.Width = 0
	}
}

// NumberFormat represents the parsed number format pattern and is used
// to format a number.
type NumberFormat struct {
	Pattern                string
	PositivePrefix         string
	PositiveSuffix         string
	NegativePrefix         string
	NegativeSuffix         string
	MinIntegerDigits       int
	MaxIntegerDigits       int // zero for no maximum
	MinFractionDigits      int
	MaxFractionDigits      int  // zero for no maximum
	MinExponentDigits      int  // minimum number of exponent digits
	PrefixPositiveExponent bool // show a plus sign for positive expontents?
	IntegerGrouping        Grouping
	FractionGrouping       Grouping
	Padding                Padding
}

func (n *NumberFormat) empty() bool {
	return n.Pattern == ""
}

func (n *NumberFormat) decode(d *xmlDecoder, elem xml.StartElement) {
	// elem.Name.Local: *Formats (e.g. currencyFormats)
	formatKey := elem.Name.Local[:len(elem.Name.Local)-1] // *Format
	formatLengthKey := formatKey + "Length"               // *FormatLength

	d.DecodeElem(formatLengthKey, func(d *xmlDecoder, elem xml.StartElement) {
		if xmlAttrib(elem, "type") == "" {
			d.DecodeElem(formatKey, n.decodeFormat)
		} else {
			// Currently there is no support for compact number formats.
			d.SkipElem()
		}
	})
}

func (n *NumberFormat) decodeFormat(d *xmlDecoder, elem xml.StartElement) {
	typ := xmlAttrib(elem, "type")
	if typ != "" && typ != "standard" {
		d.SkipElem()
		return
	}

	var (
		parser numberFormatParser
		err    error
	)
	d.DecodeElem("pattern", func(d *xmlDecoder, elem xml.StartElement) {
		pattern := d.ReadString(elem)
		if *n, err = parser.parse(pattern); err != nil {
			d.ReportErr(errors.Newf("error parsing pattern %q: %v", pattern, err), elem)
		}
		d.SkipElem()
	})
}

type numberFormatParser struct {
	s   string
	off int
	ch  rune
	err error
	buf []rune
	nf  NumberFormat
}

func (p *numberFormatParser) parse(s string) (NumberFormat, error) {
	p.s = s
	p.off = 0
	p.err = nil
	p.buf = p.buf[:0]
	p.nf = NumberFormat{Pattern: s}
	p.next() // initialize first character

	// positive subpattern
	p.parsePadding(BeforePrefix)
	p.nf.PositivePrefix = p.parsePrefix(true)
	p.parsePadding(AfterPrefix)

	p.parseIntDigits()
	p.parseFracDigits()
	p.parseExponent()

	p.parsePadding(BeforeSuffix)
	p.nf.PositiveSuffix = p.parseSuffix(true)
	p.parsePadding(AfterSuffix)

	p.nf.Padding.Width -= utf8.RuneCountInString(p.nf.PositivePrefix)
	p.nf.Padding.Width -= utf8.RuneCountInString(p.nf.PositiveSuffix)

	// negative subpattern
	if p.ch == ';' {
		p.next()
		p.nf.NegativePrefix = p.parsePrefix(false)
		for isPatternNumberChar(p.ch) {
			p.next()
		}
		p.nf.NegativeSuffix = p.parseSuffix(false)
	} else {
		p.nf.NegativePrefix = "-" + p.nf.PositivePrefix
		p.nf.NegativeSuffix = p.nf.PositiveSuffix
	}

	p.nf.IntegerGrouping.normalize()
	p.nf.FractionGrouping.normalize()
	p.nf.Padding.normalize()

	if p.err == io.EOF {
		p.err = nil
	}
	return p.nf, p.err
}

func (p *numberFormatParser) parsePadding(pos PaddingPos) {
	if p.ch == '*' {
		p.next()
		if p.ch == utf8.RuneError {
			p.seterr(errors.New("expected padding character"))
			return
		}
		p.nf.Padding.Char = p.ch
		p.nf.Padding.Pos = pos
		p.next()
	}
}

func (p *numberFormatParser) parsePrefix(positive bool) string {
	p.buf = p.buf[:0]
	for (!positive || p.ch != '*') && !isPatternDigit(p.ch) {
		if p.ch == '\'' {
			p.parseQuote()
		} else {
			p.buf = append(p.buf, p.ch)
			p.next()
		}
	}
	if positive {
		p.nf.Padding.Width += len(p.buf)
	}
	return string(p.buf)
}

func (p *numberFormatParser) parseSuffix(positive bool) string {
	p.buf = p.buf[:0]
	for (!positive || p.ch != '*') && p.ch != ';' && p.ch != utf8.RuneError {
		if p.ch == '\'' {
			p.parseQuote()
		} else {
			p.buf = append(p.buf, p.ch)
			p.next()
		}
	}
	if positive {
		p.nf.Padding.Width += len(p.buf)
	}
	return string(p.buf)
}

func (p *numberFormatParser) parseIntDigits() {
	grouping := false
	groupsize := 0
	for {
		if p.ch == ',' {
			if grouping {
				p.nf.IntegerGrouping.SecondarySize = groupsize
			}
			groupsize = 0
			grouping = true
		} else if isPatternDigit(p.ch) {
			if p.ch == '0' || p.ch == '@' {
				p.nf.MinIntegerDigits++
			} else {
				p.nf.MinIntegerDigits = 0
			}
			p.nf.MaxIntegerDigits++
			groupsize++
		} else {
			break
		}
		p.nf.Padding.Width++
		p.next()
	}

	if grouping {
		p.nf.IntegerGrouping.PrimarySize = groupsize
	}
}

func (p *numberFormatParser) parseFracDigits() {
	if p.ch != '.' {
		return
	}
	p.next()
	p.nf.Padding.Width++

	groupsize := 0
	countMin := true
	for {
		if p.ch == ',' {
			if p.nf.FractionGrouping.SecondarySize == 0 {
				p.nf.FractionGrouping.SecondarySize = p.nf.FractionGrouping.PrimarySize
				p.nf.FractionGrouping.PrimarySize = groupsize
			}
			groupsize = 0
		} else if isPatternDigit(p.ch) {
			if countMin && (p.ch == '0' || p.ch == '@') {
				p.nf.MinFractionDigits++
			} else {
				countMin = false
			}
			p.nf.MaxFractionDigits++
			groupsize++
		} else {
			break
		}
		p.nf.Padding.Width++
		p.next()
	}
}

func (p *numberFormatParser) parseExponent() {
	if p.ch == '+' {
		p.nf.PrefixPositiveExponent = true
		p.nf.Padding.Width++
		p.next()
	}

	if p.ch != 'E' {
		p.nf.PrefixPositiveExponent = false
		p.nf.MaxIntegerDigits = 0
		return
	}

	p.next() // skip 'E'
	p.nf.Padding.Width++
	for isPatternDigit(p.ch) {
		if p.ch == '0' {
			p.nf.MinExponentDigits++
		}
		p.next()
	}
}

func (p *numberFormatParser) parseQuote() {
	p.next() // skip '\''
	if p.ch == '\'' {
		p.buf = append(p.buf, p.ch)
	} else {
		for p.ch != '\'' && p.ch != utf8.RuneError {
			p.buf = append(p.buf, p.ch)
			p.next()
		}
	}
	p.next() // skip '\''
}

func (p *numberFormatParser) next() {
	if p.err != nil {
		p.ch = utf8.RuneError
		return
	}

	var n int
	p.ch, n = utf8.DecodeRuneInString(p.s[p.off:])
	if p.ch == utf8.RuneError {
		if n == 0 {
			p.seterr(io.EOF)
		} else {
			p.seterr(errors.New("invalid utf-8 encoding"))
		}
		return
	}
	p.off += n
}

func (p *numberFormatParser) seterr(err error) {
	if p.err == nil || p.err == io.EOF {
		p.err = err
	}
}

func isPatternDigit(ch rune) bool {
	return ch == '#' || ch == '@' || ('0' <= ch && ch <= '9')
}

func isPatternNumberChar(ch rune) bool {
	return isPatternDigit(ch) || ch == '.' || ch == '+' || ch == '-' || ch == ',' || ch == 'E'
}
