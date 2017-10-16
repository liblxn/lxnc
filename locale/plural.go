package locale

// Operand represents an operand in a plural rule.
type Operand uint8

// Available operands for the plural rules.
const (
	AbsoluteValue       Operand = 0 // n
	IntegerDigits       Operand = 1 // i
	NumFracDigit        Operand = 2 // v
	NumFracDigitNoZeros Operand = 3 // w
	FracDigits          Operand = 4 // f
	FracDigitsNoZeros   Operand = 5 // t
)

// Operator represents an operator in a plural rule.
type Operator uint8

// Available operators for the plural rules.
const (
	Equal    Operator = 0
	NotEqual Operator = 1
)

// PluralTag represents a tag for a specific plural form.
type PluralTag uint8

// Available plural tags.
const (
	Other PluralTag = 0
	Zero  PluralTag = 1
	One   PluralTag = 2
	Two   PluralTag = 3
	Few   PluralTag = 4
	Many  PluralTag = 5
)

// Connective represents a logical connective for two plural rules. A plural
// rule can be connected with another rule by a conjunction ('and' operator)
// or a disjunction ('or' operator). The conjunction binds more tightly.
type Connective uint

// Available connectives.
const (
	None        Connective = 0
	Conjunction Connective = 1
	Disjunction Connective = 2
)

// Range represents an integer range, where both bounds are inclusive.
// If the lower bound equals the upper bound, the range will collapse to a single value.
type Range struct {
	LowerBound int
	UpperBound int
}

// Ranges holds a collection of ranges.
type Ranges []uint16

// Len returns the number of ranges.
func (r Ranges) Len() int {
	return len(r) / 2
}

// At returns the range with the given index.
func (r Ranges) At(i int) Range {
	return Range{LowerBound: int(r[2*i]), UpperBound: int(r[2*i+1])}
}

// PluralRule holds the data for a single plural rule. The ModuloExp field defines the
// exponent for the modulo divisor to the base of 10, i.e. Operand % 10^ModuloExp.
// If ModuloExp is zero, no remainder has to be calculated.
//
// The plural rule could be connected with another rule. If so, the Connective field is
// set to the respective value (Conjunction or Disjunction). Otherwise the Connective
// field is set to None and there is no follow-up rule.
//
// Example for a plural rule: i%10=1..3
type PluralRule struct {
	Operand    Operand
	ModuloExp  int
	Operator   Operator
	Ranges     Ranges
	Connective Connective
}

// PluralRules holds a collection of plural rules for a specific plural tag.
// All rules in this collection are connected with each other (see PluralRule
// and Connective).
type PluralRules struct {
	Tag PluralTag
	rel relation
}

// Iter iterates over all plural rules in the collection. The iterator should
// return true, if the iteration should be continued. Otherwise false should be
// returned.
func (r PluralRules) Iter(iter func(PluralRule) bool) {
	rel := r.rel
	if len(rel) == 0 {
		return
	}

	for {
		rule := PluralRule{
			Operand:    Operand(rel.operand()),
			ModuloExp:  rel.modexp(),
			Operator:   Operator(rel.operator()),
			Ranges:     Ranges(rel.ranges()),
			Connective: Connective(rel.connective()),
		}
		if !iter(rule) || rule.Connective == None {
			return
		}
		rel = rel.next()
	}
}

// Plural holds the plural data for a language. There can at most be five plural rule
// collections for 'zero', 'one', 'two', 'few', and 'many'. If there are less than five
// rules, the rest will be filled with empty 'other' rules.
type Plural [5]PluralRules

// CardinalPlural returns the plural rules for cardinals in the given locale.
func CardinalPlural(loc Locale) Plural {
	return lookupPlural(loc, cardinalRules)
}

// OrdinalPlural returns the plural rules for ordinals in the given locale.
func OrdinalPlural(loc Locale) Plural {
	return lookupPlural(loc, ordinalRules)
}

func lookupPlural(loc Locale, lookup pluralRuleLookup) Plural {
	if loc == 0 {
		panic("invalid locale")
	}

	lang, _, _ := loc.tagIDs()
	rules, has := lookup[lang]
	if !has {
		return Plural{}
	}
	return Plural{
		{Tag: PluralTag(rules[0].tag()), rel: relations.relation(rules[0].relationID())},
		{Tag: PluralTag(rules[1].tag()), rel: relations.relation(rules[1].relationID())},
		{Tag: PluralTag(rules[2].tag()), rel: relations.relation(rules[2].relationID())},
		{Tag: PluralTag(rules[3].tag()), rel: relations.relation(rules[3].relationID())},
		{Tag: PluralTag(rules[4].tag()), rel: relations.relation(rules[4].relationID())},
	}
}

type relation []uint16

func (r relation) operand() uint    { return uint((r[0] >> 12) & 0xf) }
func (r relation) modexp() int      { return int((r[0] >> 8) & 0xf) }
func (r relation) operator() uint   { return uint((r[0] >> 7) & 0x1) }
func (r relation) rangeCount() int  { return int((r[0] >> 2) & 0x1f) }
func (r relation) connective() uint { return uint((r[0]) & 0x3) }
func (r relation) ranges() relation { return r[1 : 1+2*r.rangeCount()] }
func (r relation) next() relation   { return r[1+2*r.rangeCount():] }

type relationID uint16
type relationLookup []uint16

func (l relationLookup) relation(id relationID) relation {
	if id == 0 || int(id) > len(l) {
		return nil
	}
	return relation(l[id-1:])
}

type pluralRule uint16

func (r pluralRule) tag() uint              { return uint(r>>13) & 0x7 }
func (r pluralRule) relationID() relationID { return relationID(r & 0x1fff) }

type pluralRuleLookup map[langID][5]pluralRule
