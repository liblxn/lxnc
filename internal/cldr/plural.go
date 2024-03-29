package cldr

import (
	"bytes"
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/liblxn/lxnc/internal/errors"
)

// Available plural categories.
const (
	Zero  = "zero"
	One   = "one"
	Two   = "two"
	Few   = "few"
	Many  = "many"
	Other = "other"
)

// Plurals holds all the relevant data which is necessary to determine plurals
// in a specific locale.
type Plurals struct {
	Cardinal []PluralRules
	Ordinal  []PluralRules
}

func (p *Plurals) decode(d *xmlDecoder, elem xml.StartElement) {
	switch xmlAttrib(elem, "type") {
	case "cardinal":
		p.Cardinal = make([]PluralRules, 0, 64)
		d.DecodeElem("pluralRules", func(d *xmlDecoder, elem xml.StartElement) {
			p.Cardinal = append(p.Cardinal, PluralRules{})
			p.Cardinal[len(p.Cardinal)-1].decode(d, elem)
		})

	case "ordinal":
		p.Ordinal = make([]PluralRules, 0, 64)
		d.DecodeElem("pluralRules", func(d *xmlDecoder, elem xml.StartElement) {
			p.Ordinal = append(p.Ordinal, PluralRules{})
			p.Ordinal[len(p.Ordinal)-1].decode(d, elem)
		})

	default:
		d.SkipElem()
	}
}

// PluralRules holds the plural rules for a group of locales.
type PluralRules struct {
	Locales []string
	Rules   map[string]PluralRule // plural category => rule
}

func (p *PluralRules) decode(d *xmlDecoder, elem xml.StartElement) {
	if locales := xmlAttrib(elem, "locales"); locales != "" {
		p.Locales = strings.Split(locales, " ")
	}

	p.Rules = make(map[string]PluralRule)
	d.DecodeElem("pluralRule", func(d *xmlDecoder, elem xml.StartElement) {
		category := xmlAttrib(elem, "count")
		if category == "" {
			category = Other
		}

		parser := pluralRuleParser{}
		ruleStr := d.ReadString(elem)
		if rule, err := parser.parse(ruleStr); err != nil {
			d.ReportErr(errors.Newf("error in plural rule %s (%s): %v", category, ruleStr, err), elem)
		} else {
			p.Rules[category] = rule
		}
		d.SkipElem()
	})
}

// Operand represents an operand in a plural rule.
type Operand rune

// Available operands for the plural rules.
//
// https://unicode.org/reports/tr35/tr35-numbers.html#Operands
const (
	AbsoluteValue               Operand = 'n' // absolute value of the source number
	IntegerDigits               Operand = 'i' // integer digits of n
	FracDigitCountTrailingZeros Operand = 'v' // number of visible fraction digits in n, with trailing zeros
	FracDigitCount              Operand = 'w' // number of visible fraction digits in n, without trailing zeros
	FracDigitsTrailingZeros     Operand = 'f' // visible fraction digits in n, with trailing zeros
	FracDigits                  Operand = 't' // visible fraction digits in n, without trailing zeros
	CompactDecimalExponent      Operand = 'c' // compact decimal exponent value: exponent of the power of 10 used in compact decimal formatting
	CompactDecimalExponent2     Operand = 'e' // currently, synonym for ‘c’. however, may be redefined in the future.
)

// Operator represents an operator in a plural rule.
type Operator string

// Available operators for the plural rules.
const (
	Equal    Operator = "="
	NotEqual Operator = "!="
)

// IntRange represents an integer range, where the bounds are inclusive.
// If the lower and the upper bound are equal, the range will collapse
// to a single number.
type IntRange struct {
	LowerBound int
	UpperBound int
}

// String returns the string representation of the integer range.
func (r *IntRange) String() string {
	lower := strconv.FormatInt(int64(r.LowerBound), 10)
	if r.LowerBound == r.UpperBound {
		return lower
	}
	upper := strconv.FormatInt(int64(r.UpperBound), 10)
	return lower + ".." + upper
}

// FloatRange represents an floating-point range, where the bounds are inclusive.
// If the lower and the upper bound are equal, the range collapses to a single
// number.
type FloatRange struct {
	LowerBound float64
	UpperBound float64
	Decimals   int // number of decimal digits
}

// String returns the string representation of the floating-point range.
func (r *FloatRange) String() string {
	lower := strconv.FormatFloat(r.LowerBound, 'f', r.Decimals, 64)
	if r.LowerBound == r.UpperBound {
		return lower
	}
	upper := strconv.FormatFloat(r.UpperBound, 'f', r.Decimals, 64)
	return lower + "~" + upper
}

// Relation represents a relation in a plural rule, e.g. 'i % 10 = 1, 2'.
// If Modulo is zero, then there is no modulo operation in this relation.
type Relation struct {
	Operand  Operand
	Modulo   int
	Operator Operator
	Ranges   []IntRange
}

// String returns the string representation of the relation.
func (r *Relation) String() string {
	var buf bytes.Buffer
	r.writeTo(&buf)
	return buf.String()
}

func (r *Relation) writeTo(w io.Writer) {
	_, _ = io.WriteString(w, string(r.Operand))
	if r.Modulo != 0 {
		_, _ = io.WriteString(w, " % "+strconv.FormatInt(int64(r.Modulo), 10))
	}

	_, _ = io.WriteString(w, " "+string(r.Operator)+" ")

	needComma := false
	for _, rng := range r.Ranges {
		if needComma {
			_, _ = io.WriteString(w, ", ")
		}
		_, _ = io.WriteString(w, rng.String())
		needComma = true
	}
}

// Conjunction represents a group of relations which are connected with
// the 'and' operator, e.g. "i = 1 and n = 4".
type Conjunction []Relation

// String returns the string representation of the conjunction.
func (c Conjunction) String() string {
	var buf bytes.Buffer
	c.writeTo(&buf)
	return buf.String()
}

func (c Conjunction) writeTo(w io.Writer) {
	needAnd := false
	for _, rel := range c {
		if needAnd {
			_, _ = io.WriteString(w, " and ")
		}
		rel.writeTo(w)
		needAnd = true
	}
}

// PluralSample represents a sample for a plural rule. A sample consists
// of number ranges and the information if the sample is infinite.
type PluralSample struct {
	Ranges   []FloatRange
	Infinite bool
}

// String returns the string representation of the plural sample.
func (s *PluralSample) String() string {
	var buf bytes.Buffer
	s.writeTo(&buf)
	return buf.String()
}

func (s *PluralSample) writeTo(w io.Writer) {
	needComma := false
	for _, rng := range s.Ranges {
		if needComma {
			_, _ = io.WriteString(w, ", ")
		}
		_, _ = io.WriteString(w, rng.String())
		needComma = true
	}
	if s.Infinite {
		_, _ = io.WriteString(w, ", …")
	}
}

// PluralRule represents a single plural rule. A plural rule contains
// conditions and samples for integer and floating-point values. All
// conjunctions in the condition are connected with the 'or' operator.
type PluralRule struct {
	Condition     []Conjunction
	IntegerSample PluralSample
	DecimalSample PluralSample
}

// String returns the string representation of the plural rule.
func (r *PluralRule) String() string {
	var buf bytes.Buffer
	r.writeTo(&buf)
	return buf.String()
}

func (r *PluralRule) writeTo(w io.Writer) {
	needOr := false
	for _, conj := range r.Condition {
		if needOr {
			_, _ = io.WriteString(w, " or ")
		}
		conj.writeTo(w)
		needOr = true
	}

	if len(r.IntegerSample.Ranges) != 0 {
		_, _ = io.WriteString(w, " @integer ")
		r.IntegerSample.writeTo(w)
	}
	if len(r.DecimalSample.Ranges) != 0 {
		_, _ = io.WriteString(w, " @decimal ")
		r.DecimalSample.writeTo(w)
	}
}

// rule            = condition samples
//
// condition       = and_condition ('or' and_condition)*
// and_condition   = relation ('and' relation)*
// relation        = is_relation | in_relation | within_relation
// is_relation     = expr 'is' ('not')? value
// in_relation     = expr (('not')? 'in' | '=' | '!=') range_list
// within_relation = expr ('not')? 'within' range_list
// expr            = operand (('mod' | '%') value)?
// operand         = 'n' | 'i' | 'f' | 't' | 'v' | 'w' | 'c' | 'e'
// range_list      = (range | value) (',' range_list)*
// range           = value'..'value
// value           = digit+
//
// samples         = ('@integer' sampleList)?
//
//	('@decimal' sampleList)?
//
// sampleList      = sampleRange (',' sampleRange)* (',' ('…'|'...'))?
// sampleRange     = sampleValue ('~' sampleValue)?
// sampleValue     = value ('.' value)? ([ce] digitPos digit+)?
//
// digit           = [0-9]
// digitPos        = [1-9]
type pluralRuleParser struct {
	s   string
	off int
	ch  rune
	err error
}

func (p *pluralRuleParser) parse(s string) (PluralRule, error) {
	p.s = s
	p.off = 0
	p.err = nil
	p.next() // initialize first character

	rule := p.parseRule()
	if p.err != io.EOF {
		p.seterr(errors.Newf("unexpected token: %q", p.ch))
	} else {
		p.err = nil
	}
	return rule, p.err
}

func (p *pluralRuleParser) parseRule() (rule PluralRule) {
	if p.ch != '@' {
		rule.Condition = p.parseCondition()
	}

	for p.ch == '@' {
		p.next()

		switch p.ch {
		case 'i':
			p.expect('i', 'n', 't', 'e', 'g', 'e', 'r')
			rule.IntegerSample = p.parseSampleList()
		case 'd':
			p.expect('d', 'e', 'c', 'i', 'm', 'a', 'l')
			rule.DecimalSample = p.parseSampleList()
		default:
			p.seterr(errors.New("invalid sample identifier"))
			return
		}
	}
	return
}

func (p *pluralRuleParser) parseCondition() (cond []Conjunction) {
	for {
		cond = append(cond, p.parseAndCondition())
		if p.ch != 'o' {
			return
		}
		p.expect('o', 'r')
	}
}

// and_condition
func (p *pluralRuleParser) parseAndCondition() (conj Conjunction) {
	for {
		conj = append(conj, p.parseRelation())
		if p.ch != 'a' {
			return
		}
		p.expect('a', 'n', 'd')
	}
}

// relation
func (p *pluralRuleParser) parseRelation() (rel Relation) {
	p.parseExpr(&rel)

	// For backward compatibility it is allowed to additionally
	// use 'not in/within'.
	// We only support the preferred syntax here ('=' and '!=')
	switch p.ch {
	case '=':
		p.next()
		rel.Operator = Equal
	case '!':
		p.expect('!', '=')
		rel.Operator = NotEqual
	default:
		p.seterr(errors.New("invalid operator"))
		return
	}

	rel.Ranges = p.parseRangeList()
	return
}

// range_list
func (p *pluralRuleParser) parseRangeList() (ranges []IntRange) {
	for {
		rng := p.parseRange()
		if n := len(ranges); n != 0 && (ranges[n-1].UpperBound == rng.LowerBound || ranges[n-1].UpperBound+1 == rng.LowerBound) {
			ranges[n-1].UpperBound = rng.UpperBound
		} else {
			ranges = append(ranges, rng)
		}

		if p.ch != ',' {
			return ranges
		}
		p.next() // skip ','
	}
}

// range
func (p *pluralRuleParser) parseRange() (rng IntRange) {
	var ok bool
	if rng.LowerBound, ok = p.parseValue(); !ok {
		p.seterr(errors.New("lower bound value expected"))
		return
	}
	if p.ch != '.' {
		rng.UpperBound = rng.LowerBound
	} else {
		p.expect('.', '.')
		if rng.UpperBound, ok = p.parseValue(); !ok {
			p.seterr(errors.New("upper bound value expected"))
		}
	}
	return
}

// expr
func (p *pluralRuleParser) parseExpr(rel *Relation) {
	switch op := Operand(p.ch); op {
	case AbsoluteValue, IntegerDigits, FracDigitCountTrailingZeros, FracDigitCount, FracDigitsTrailingZeros, FracDigits, CompactDecimalExponent, CompactDecimalExponent2:
		rel.Operand = op
	default:
		p.seterr(errors.Newf("invalid operand %q", op))
		return
	}

	// For backward compatibility it is allowed to use '%' and 'mod'.
	// We only support the preferred syntax here ('%')

	p.next()
	if p.ch == '%' {
		p.next()
		if mod, ok := p.parseValue(); ok {
			rel.Modulo = mod
		} else {
			p.seterr(errors.New("module value expected"))
		}
	}
}

// value
func (p *pluralRuleParser) parseValue() (val int, ok bool) {
	for {
		switch p.ch {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			val *= 10
			val += int(p.ch - '0')
			ok = true
			p.next()
		default:
			return val, ok
		}
	}
}

// sampleList
func (p *pluralRuleParser) parseSampleList() (s PluralSample) {
	for {
		s.Ranges = append(s.Ranges, p.parseSampleRange())
		if p.ch != ',' {
			return
		}
		p.next()
		switch p.ch {
		case '…':
			p.next()
			s.Infinite = true
			return
		case '.':
			p.expect('.', '.', '.')
			s.Infinite = true
			return
		}
	}
}

// sampleRange
func (p *pluralRuleParser) parseSampleRange() (rng FloatRange) {
	var ok bool
	if rng.LowerBound, rng.Decimals, ok = p.parseSampleValue(); ok {
		if p.ch != '~' {
			rng.UpperBound = rng.LowerBound
		} else {
			p.next()
			var decimals int
			if rng.UpperBound, decimals, ok = p.parseSampleValue(); !ok {
				p.seterr(errors.New("decimal value expected"))
			} else if rng.Decimals < decimals {
				rng.Decimals = decimals
			}
		}
	}
	return
}

// sampleValue
func (p *pluralRuleParser) parseSampleValue() (val float64, decimals int, ok bool) {
	var buf bytes.Buffer

	_ = p.readDigits(&buf, '0')
	if p.ch == '.' {
		buf.WriteRune(p.ch)
		p.next()
		decimals = p.readDigits(&buf, '0')
	}

	if p.ch == 'c' || p.ch == 'e' {
		buf.WriteByte('e')
		p.next()
		_ = p.readDigits(&buf, '1') // first digits must not be '0'
		_ = p.readDigits(&buf, '0') // read the rest
	}

	val, err := strconv.ParseFloat(buf.String(), 64)
	return val, decimals, err == nil && buf.Len() != 0
}

func (p *pluralRuleParser) readDigits(buf *bytes.Buffer, lowerDigit rune) (count int) {
	for lowerDigit <= p.ch && p.ch <= '9' {
		buf.WriteRune(p.ch)
		count++
		p.next()
	}
	return
}

func (p *pluralRuleParser) expect(chars ...rune) {
	for _, ch := range chars {
		if p.ch != ch {
			p.seterr(errors.Newf("unexpected token %q", p.ch))
			return
		}
		p.next()
	}
}

func (p *pluralRuleParser) next() {
	if p.err != nil {
		p.ch = utf8.RuneError
		return
	}

	var n int
	for {
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
		if !unicode.IsSpace(p.ch) {
			return
		}
	}
}

func (p *pluralRuleParser) seterr(err error) {
	if p.err == nil || p.err == io.EOF {
		p.err = err
	}
}
