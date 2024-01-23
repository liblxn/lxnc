package main

import (
	"fmt"
	"strings"

	"github.com/liblxn/lxnc/internal/cldr"
)

const (
	pluralTagBits  = 3
	relationIDBits = 13

	pluralChunkBits = 32

	operandBits    = 4
	modexpBits     = 4
	operatorBits   = 1
	rangeCountBits = 5
	connectiveBits = 2

	other = 0
	zero  = 1
	one   = 2
	two   = 3
	few   = 4
	many  = 5

	equal    = 0
	notEqual = 1

	absoluteValue       = 0 // n
	integerDigits       = 1 // i
	numFracDigit        = 2 // v
	numFracDigitNoZeros = 3 // w
	fracDigits          = 4 // f
	fracDigitsNoZeros   = 5 // t
	compactDecimalExp   = 5 // c, e

	noConnective = 0x0
	conjunction  = 0x1
	disjunction  = 0x2
)

// relation lookup
type relationLookup struct{}

func (l relationLookup) imports() []string {
	return nil
}

func (l relationLookup) generate(p *printer) {
	if operandBits+modexpBits+operatorBits+connectiveBits+rangeCountBits > pluralChunkBits {
		panic("invalid plural rule lookup configuration")
	}

	relIDBits := 64
	switch {
	case relationIDBits <= 8:
		relIDBits = 8
	case relationIDBits <= 16:
		relIDBits = 16
	case relationIDBits <= 32:
		relIDBits = 32
	}

	operandMask := fmt.Sprintf("%#x", (1<<operandBits)-1)
	modexpMask := fmt.Sprintf("%#x", (1<<modexpBits)-1)
	operatorMask := fmt.Sprintf("%#x", (1<<operatorBits)-1)
	connectiveMask := fmt.Sprintf("%#x", (1<<connectiveBits)-1)
	rangeCountMask := fmt.Sprintf("%#x", (1<<rangeCountBits)-1)

	p.Println(`type relation []uint`, pluralChunkBits)
	p.Println()
	p.Println(`func (r relation) operand() uint    { return uint((r[0] >> `, modexpBits+operatorBits+rangeCountBits+connectiveBits, `) & `, operandMask, `) }`)
	p.Println(`func (r relation) modexp() int      { return int((r[0] >> `, operatorBits+rangeCountBits+connectiveBits, `) & `, modexpMask, `) }`)
	p.Println(`func (r relation) operator() uint   { return uint((r[0] >> `, rangeCountBits+connectiveBits, `) & `, operatorMask, `) }`)
	p.Println(`func (r relation) rangeCount() int  { return int((r[0] >> `, connectiveBits, `) & `, rangeCountMask, `) }`)
	p.Println(`func (r relation) connective() uint { return uint((r[0]) & `, connectiveMask, `) }`)
	p.Println(`func (r relation) ranges() relation { return r[1 : 1+2*r.rangeCount()] }`)
	p.Println(`func (r relation) next() relation   { return r[1+2*r.rangeCount():] }`)
	p.Println()
	p.Println(`type relationID uint`, relIDBits)
	p.Println(`type relationLookup []uint`, pluralChunkBits)
	p.Println()
	p.Println(`func (l relationLookup) relation(id relationID) relation {`)
	p.Println(`	if id == 0 || int(id) > len(l) {`)
	p.Println(`		return nil`)
	p.Println(`	}`)
	p.Println(`	return relation(l[id-1:])`)
	p.Println(`}`)
}

func (l relationLookup) testImports() []string {
	return []string{"reflect"}
}

func (l relationLookup) generateTest(p *printer) {
	p.Println(`func TestRelation(t *testing.T) {`)
	p.Println(`	rel := relation{`, newTestRelation(5, 4, 1, 2, 3), `, 0x0, 0x0, 0x0, 0x0, 0x77}`)
	p.Println()
	p.Println(`	if op := rel.operand(); op != 5 {`)
	p.Println(`		t.Errorf("unexpected operand: %d", op)`)
	p.Println(`	}`)
	p.Println(`	if modexp := rel.modexp(); modexp != 4 {`)
	p.Println(`		t.Errorf("unexpected modulo exponent: %d", modexp)`)
	p.Println(`	}`)
	p.Println(`	if op := rel.operator(); op != 1 {`)
	p.Println(`		t.Errorf("unexpected operator: %d", op)`)
	p.Println(`	}`)
	p.Println(`	if rc := rel.rangeCount(); rc != 2 {`)
	p.Println(`		t.Errorf("unexpected range count: %d", rc)`)
	p.Println(`	}`)
	p.Println(`	if c := rel.connective(); c != 3 {`)
	p.Println(`		t.Errorf("unexpected connective: %d", c)`)
	p.Println(`	}`)
	p.Println(`	if r := rel.ranges(); !reflect.DeepEqual(r, rel[1:5]) {`)
	p.Println(`		t.Errorf("unexpected ranges: %v", r)`)
	p.Println(`	}`)
	p.Println(`	if nxt := rel.next(); !reflect.DeepEqual(nxt, rel[5:]) {`)
	p.Println(`		t.Errorf("unexpected next value: %v", nxt)`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`func TestRelationLookup(t *testing.T) {`)
	p.Println(`	lookup := relationLookup{1, 2, 3}`)
	p.Println()
	p.Println(`	if r := lookup.relation(0); len(r) != 0 {`)
	p.Println(`		t.Errorf("unexpected relation: %v", r)`)
	p.Println(`	}`)
	p.Println(`	if r := lookup.relation(1); !reflect.DeepEqual(r, relation(lookup)) {`)
	p.Println(`		t.Errorf("unexpected relation: %v", r)`)
	p.Println(`	}`)
	p.Println(`	if r := lookup.relation(2); !reflect.DeepEqual(r, relation(lookup[1:])) {`)
	p.Println(`		t.Errorf("unexpected relation: %v", r)`)
	p.Println(`	}`)
	p.Println(`}`)
}

type pluralRule struct {
	key       string
	nchunks   int
	condition []cldr.Conjunction
}

type relationLookupVar struct {
	name    string
	rules   []pluralRule
	nchunks int
}

func newRelationLookupVar(name string) *relationLookupVar {
	return &relationLookupVar{
		name: name,
	}
}

func (v *relationLookupVar) newRelation(rel cldr.Relation, connective uint) uint64 {
	operand := 0
	switch rel.Operand {
	case cldr.AbsoluteValue:
		operand = absoluteValue
	case cldr.IntegerDigits:
		operand = integerDigits
	case cldr.FracDigitCountTrailingZeros:
		operand = numFracDigit
	case cldr.FracDigitCount:
		operand = numFracDigitNoZeros
	case cldr.FracDigitsTrailingZeros:
		operand = fracDigits
	case cldr.FracDigits:
		operand = fracDigitsNoZeros
	case cldr.CompactDecimalExponent, cldr.CompactDecimalExponent2:
		operand = compactDecimalExp
	}

	modexp := 0
	if mod := rel.Modulo; mod != 0 {
		for mod > 1 {
			modexp++
			mod /= 10
		}
	}

	operator := equal
	if rel.Operator == cldr.NotEqual {
		operator = notEqual
	}

	return uint64(operand)<<(modexpBits+operatorBits+rangeCountBits+connectiveBits) |
		uint64(modexp)<<(operatorBits+rangeCountBits+connectiveBits) |
		uint64(operator)<<(rangeCountBits+connectiveBits) |
		uint64(len(rel.Ranges))<<connectiveBits |
		uint64(connective)
}

func (v *relationLookupVar) keyOf(rule cldr.PluralRule) string {
	r := cldr.PluralRule{Condition: rule.Condition} // strip samples
	return r.String()
}

func (v *relationLookupVar) add(rule cldr.PluralRule) {
	nchunks := 0
	for _, conj := range rule.Condition {
		for _, rel := range conj {
			nchunks += 1 + 2*len(rel.Ranges)
			switch {
			case rel.Modulo != 0 && !isPowerOfTen(rel.Modulo):
				panic(fmt.Sprintf("relation modulo is not a power of 10, cannot add %q", rel.String()))
			case len(rel.Ranges) > (1<<rangeCountBits)-1:
				panic(fmt.Sprintf("number of ranges exceeds the maximum, cannot add %q", rel.String()))
			}
			for _, rng := range rel.Ranges {
				const maxBound = (1 << pluralChunkBits) - 1
				if rng.LowerBound > maxBound || rng.UpperBound > maxBound {
					panic(fmt.Sprintf("range values exceed the maximum, cannot add %q", rel.String()))
				}
			}
		}
	}

	switch {
	case v.nchunks+nchunks >= (1<<relationIDBits)-1:
		panic(fmt.Sprintf("number of relations exceeds the maximum, cannot add %s", rule.String()))
	case len(rule.Condition) == 0:
		return
	}

	key := v.keyOf(rule)
	idx := 0
	for idx < len(v.rules) && v.rules[idx].key < key {
		idx++
	}
	if idx == len(v.rules) || v.rules[idx].key != key {
		v.rules = append(v.rules, pluralRule{})
		copy(v.rules[idx+1:], v.rules[idx:])
		v.rules[idx] = pluralRule{key: key, condition: rule.Condition, nchunks: nchunks}
		v.nchunks += nchunks
	}
}

func (v *relationLookupVar) relationID(rule cldr.PluralRule) uint {
	key := v.keyOf(rule)
	id := 1
	for _, r := range v.rules {
		if r.key == key {
			return uint(id)
		}
		id += r.nchunks
	}
	panic(fmt.Sprintf("plural rule not found: %s", key))
}

func (v *relationLookupVar) imports() []string {
	return nil
}

func (v *relationLookupVar) generate(p *printer) {
	nchunks := 0
	for _, rule := range v.rules {
		nchunks += rule.nchunks
	}

	hex := func(chunk uint64) string {
		return fmt.Sprintf("%#0[2]*[1]x", chunk, pluralChunkBits/4)
	}

	p.Println(`var `, v.name, ` = relationLookup{ // `, nchunks, ` items, `, nchunks*pluralChunkBits/8, ` bytes`)

	for _, rule := range v.rules {
		p.Print(`	`)
		for d := 0; d < len(rule.condition); d++ {
			conj := rule.condition[d]

			or := uint(disjunction)
			if d == len(conj)-1 {
				or = noConnective
			}

			for r := 0; r < len(conj); r++ {
				rel := conj[r]
				and := uint(conjunction)
				if r == len(conj)-1 {
					and = or
				}

				p.Print(hex(v.newRelation(rel, and)), `, `)
				for _, rng := range rel.Ranges {
					p.Print(hex(uint64(rng.LowerBound)), `, `, hex(uint64(rng.UpperBound)), `, `)
				}
			}
		}

		pr := cldr.PluralRule{Condition: rule.condition}
		p.Println(`// `, pr.String())
	}
	p.Println(`}`)
}

// plural rule lookup
type pluralRuleLookup struct{}

func (l pluralRuleLookup) imports() []string {
	return nil
}

func (l pluralRuleLookup) generate(p *printer) {
	ruleBits := pluralTagBits + relationIDBits
	switch {
	case ruleBits <= 8:
		ruleBits = 8
	case ruleBits <= 16:
		ruleBits = 16
	case ruleBits <= 32:
		ruleBits = 32
	default:
		ruleBits = 64
	}

	tagMask := fmt.Sprintf("%#x", (1<<pluralTagBits)-1)
	relationIDMask := fmt.Sprintf("%#x", (1<<relationIDBits)-1)

	p.Println(`type pluralRule uint`, ruleBits)
	p.Println()
	p.Println(`func (r pluralRule) tag() uint              { return uint(r>>`, relationIDBits, `) & `, tagMask, ` }`)
	p.Println(`func (r pluralRule) relationID() relationID { return relationID(r & `, relationIDMask, `) }`)
	p.Println()
	p.Println(`type pluralRuleLookup map[langID][5]pluralRule`)
}

func (l pluralRuleLookup) testImports() []string {
	return nil
}

func (l pluralRuleLookup) generateTest(p *printer) {
	rule := func(tag, relationID uint) string {
		return fmt.Sprintf("%#x", (tag<<relationIDBits)|relationID)
	}

	p.Println(`func TestPluralRule(t *testing.T) {`)
	p.Println(`	const rule pluralRule = `, rule(2, 7))
	p.Println()
	p.Println(`	if tag := rule.tag(); tag != 2 {`)
	p.Println(`		t.Errorf("unexpected tag: %d", tag)`)
	p.Println(`	}`)
	p.Println(`	if rid := rule.relationID(); rid != 7 {`)
	p.Println(`		t.Errorf("unexpected relation id: %d", rid)`)
	p.Println(`	}`)
	p.Println(`}`)
}

type pluralRuleLookupVar struct {
	name      string
	langs     *langLookupVar
	relations *relationLookupVar

	langTags []string
	rules    []map[string]cldr.PluralRule
}

func newPluralRuleLookupVar(name string, langs *langLookupVar, relations *relationLookupVar) *pluralRuleLookupVar {
	return &pluralRuleLookupVar{
		name:      name,
		langs:     langs,
		relations: relations,
	}
}

func (v *pluralRuleLookupVar) add(langTag string, rules map[string]cldr.PluralRule) {
	switch len(rules) {
	case 0:
		return
	case 1:
		if _, has := rules[cldr.Other]; has {
			return // we only have the "other" tag
		}
	}

	idx := 0
	for idx < len(v.langTags) && v.langTags[idx] < langTag {
		idx++
	}
	if idx == len(v.langTags) || v.langTags[idx] != langTag {
		v.langTags = append(v.langTags, "")
		copy(v.langTags[idx+1:], v.langTags[idx:])
		v.langTags[idx] = langTag

		v.rules = append(v.rules, nil)
		copy(v.rules[idx+1:], v.rules[idx:])
		v.rules[idx] = rules
	}
}

func (v *pluralRuleLookupVar) pluralRules(lang string) map[string]cldr.PluralRule {
	for i := 0; i < len(v.langTags); i++ {
		if v.langTags[i] == lang {
			return v.rules[i]
		}
	}
	return nil
}

func (v *pluralRuleLookupVar) newPluralRule(tag, relationID uint) uint {
	return (tag << relationIDBits) | relationID
}

func (v *pluralRuleLookupVar) imports() []string {
	return nil
}

func (v *pluralRuleLookupVar) generate(p *printer) {
	pluralRuleBits := pluralTagBits + relationIDBits
	switch {
	case pluralRuleBits <= 8:
		pluralRuleBits = 8
	case pluralRuleBits <= 16:
		pluralRuleBits = 16
	case pluralRuleBits <= 32:
		pluralRuleBits = 32
	default:
		pluralRuleBits = 64
	}

	pluralTags := map[string]uint{
		cldr.Other: other, cldr.Zero: zero, cldr.One: one, cldr.Two: two, cldr.Few: few, cldr.Many: many,
	}

	p.Println(`var `, v.name, ` = pluralRuleLookup{ // `, len(v.langTags), ` items, `, 5*len(v.langTags)*pluralRuleBits/4, ` bytes`)

	hex := func(v, bits uint) string {
		return fmt.Sprintf("%#0[2]*[1]x", v, (bits+3)/4)
	}

	for i := 0; i < len(v.langTags); i++ {
		langTag := v.langTags[i]
		rules := v.rules[i]
		langID := v.langs.langID(langTag)

		var pluralRules [5]uint
		idx := 0
		for _, tag := range [...]string{cldr.Zero, cldr.One, cldr.Two, cldr.Few, cldr.Many} {
			if rule, has := rules[tag]; has {
				pluralRules[idx] = v.newPluralRule(pluralTags[tag], v.relations.relationID(rule))
				idx++
			}
		}
		for ; idx < len(pluralRules); idx++ {
			pluralRules[idx] = v.newPluralRule(other, 0) // "other" marks the end
		}

		p.Print(`	`, hex(langID, langIDBits), `: {`)
		for _, pluralRule := range pluralRules[:len(pluralRules)-1] {
			p.Print(hex(pluralRule, uint(pluralRuleBits)), `, `)
		}
		p.Print(hex(pluralRules[len(pluralRules)-1], uint(pluralRuleBits)))
		p.Println(`}, // `, langTag)
	}

	p.Println(`}`)
}

// plural
type plural struct {
	tags          *tagLookupVar
	relations     *relationLookupVar
	cardinalRules *pluralRuleLookupVar
	ordinalRules  *pluralRuleLookupVar
}

func newPlural(tags *tagLookupVar, relations *relationLookupVar, cardinalRules, ordinalRules *pluralRuleLookupVar) *plural {
	return &plural{
		tags:          tags,
		relations:     relations,
		cardinalRules: cardinalRules,
		ordinalRules:  ordinalRules,
	}
}

func (pl *plural) imports() []string {
	return nil
}

func (pl *plural) generate(p *printer) {
	p.Println(`// Operand represents an operand in a plural rule.`)
	p.Println(`type Operand uint8`)
	p.Println()
	p.Println(`// Available operands for the plural rules.`)
	p.Println(`const (`)
	p.Println(`	AbsoluteValue       Operand = `, absoluteValue, ` // n`)
	p.Println(`	IntegerDigits       Operand = `, integerDigits, ` // i`)
	p.Println(`	NumFracDigit        Operand = `, numFracDigit, ` // v`)
	p.Println(`	NumFracDigitNoZeros Operand = `, numFracDigitNoZeros, ` // w`)
	p.Println(`	FracDigits          Operand = `, fracDigits, ` // f`)
	p.Println(`	FracDigitsNoZeros   Operand = `, fracDigitsNoZeros, ` // t`)
	p.Println(`	CompactDecimalExp   Operand = `, compactDecimalExp, ` // c, e`)
	p.Println(`)`)
	p.Println()
	p.Println(`// Operator represents an operator in a plural rule.`)
	p.Println(`type Operator uint8`)
	p.Println()
	p.Println(`// Available operators for the plural rules.`)
	p.Println(`const (`)
	p.Println(`	Equal    Operator = `, equal)
	p.Println(`	NotEqual Operator = `, notEqual)
	p.Println(`)`)
	p.Println()
	p.Println(`// PluralTag represents a tag for a specific plural form.`)
	p.Println(`type PluralTag uint8`)
	p.Println()
	p.Println(`// Available plural tags.`)
	p.Println(`const (`)
	p.Println(`	Other PluralTag = `, other)
	p.Println(`	Zero  PluralTag = `, zero)
	p.Println(`	One   PluralTag = `, one)
	p.Println(`	Two   PluralTag = `, two)
	p.Println(`	Few   PluralTag = `, few)
	p.Println(`	Many  PluralTag = `, many)
	p.Println(`)`)
	p.Println()
	p.Println(`// Connective represents a logical connective for two plural rules. A plural`)
	p.Println(`// rule can be connected with another rule by a conjunction ('and' operator)`)
	p.Println(`// or a disjunction ('or' operator). The conjunction binds more tightly.`)
	p.Println(`type Connective uint`)
	p.Println()
	p.Println(`// Available connectives.`)
	p.Println(`const (`)
	p.Println(`	None        Connective = `, noConnective)
	p.Println(`	Conjunction Connective = `, conjunction)
	p.Println(`	Disjunction Connective = `, disjunction)
	p.Println(`)`)
	p.Println()
	p.Println(`// Range represents an integer range, where both bounds are inclusive.`)
	p.Println(`// If the lower bound equals the upper bound, the range will collapse to a single value.`)
	p.Println(`type Range struct {`)
	p.Println(`	LowerBound int`)
	p.Println(`	UpperBound int`)
	p.Println(`}`)
	p.Println()
	p.Println(`// Ranges holds a collection of ranges.`)
	p.Println(`type Ranges struct {`)
	p.Println(`	r []uint`, pluralChunkBits)
	p.Println(`}`)
	p.Println()
	p.Println(`// Len returns the number of ranges.`)
	p.Println(`func (r Ranges) Len() int {`)
	p.Println(`	return len(r.r) / 2`)
	p.Println(`}`)
	p.Println()
	p.Println(`// At returns the range with the given index.`)
	p.Println(`func (r Ranges) At(i int) Range {`)
	p.Println(`	return Range{LowerBound: int(r.r[2*i]), UpperBound: int(r.r[2*i+1])}`)
	p.Println(`}`)
	p.Println()
	p.Println(`// PluralRule holds the data for a single plural rule. The ModuloExp field defines the`)
	p.Println(`// exponent for the modulo divisor to the base of 10, i.e. Operand % 10^ModuloExp.`)
	p.Println(`// If ModuloExp is zero, no remainder has to be calculated.`)
	p.Println(`//`)
	p.Println(`// The plural rule could be connected with another rule. If so, the Connective field is`)
	p.Println(`// set to the respective value (Conjunction or Disjunction). Otherwise the Connective`)
	p.Println(`// field is set to None and there is no follow-up rule.`)
	p.Println(`//`)
	p.Println(`// Example for a plural rule: i%10=1..3`)
	p.Println(`type PluralRule struct {`)
	p.Println(`	Operand    Operand`)
	p.Println(`	ModuloExp  int`)
	p.Println(`	Operator   Operator`)
	p.Println(`	Ranges     Ranges`)
	p.Println(`	Connective Connective`)
	p.Println(`}`)
	p.Println()
	p.Println(`// PluralRules holds a collection of plural rules for a specific plural tag.`)
	p.Println(`// All rules in this collection are connected with each other (see PluralRule`)
	p.Println(`// and Connective).`)
	p.Println(`type PluralRules struct {`)
	p.Println(`	Tag PluralTag`)
	p.Println(`	rel relation`)
	p.Println(`}`)
	p.Println()
	p.Println(`// Iter iterates over all plural rules in the collection. The iterator should`)
	p.Println(`// return true, if the iteration should be continued. Otherwise false should be`)
	p.Println(`// returned.`)
	p.Println(`func (r PluralRules) Iter(iter func(PluralRule) bool) {`)
	p.Println(`	rel := r.rel`)
	p.Println(`	if len(rel) == 0 {`)
	p.Println(`		return`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	for {`)
	p.Println(`		rule := PluralRule{`)
	p.Println(`			Operand:    Operand(rel.operand()),`)
	p.Println(`			ModuloExp:  rel.modexp(),`)
	p.Println(`			Operator:   Operator(rel.operator()),`)
	p.Println(`			Ranges:     Ranges{r: rel.ranges()},`)
	p.Println(`			Connective: Connective(rel.connective()),`)
	p.Println(`		}`)
	p.Println(`		if !iter(rule) || rule.Connective == None {`)
	p.Println(`			return`)
	p.Println(`		}`)
	p.Println(`		rel = rel.next()`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`// Plural holds the plural data for a language. There can at most be five plural rule`)
	p.Println(`// collections for 'zero', 'one', 'two', 'few', and 'many'. If there are less than five`)
	p.Println(`// rules, the rest will be filled with empty 'other' rules.`)
	p.Println(`type Plural [5]PluralRules`)
	p.Println()
	p.Println(`// CardinalPlural returns the plural rules for cardinals in the given locale.`)
	p.Println(`func CardinalPlural(loc Locale) Plural {`)
	p.Println(`	return lookupPlural(loc, `, pl.cardinalRules.name, `)`)
	p.Println(`}`)
	p.Println()
	p.Println(`// OrdinalPlural returns the plural rules for ordinals in the given locale.`)
	p.Println(`func OrdinalPlural(loc Locale) Plural {`)
	p.Println(`	return lookupPlural(loc, `, pl.ordinalRules.name, `)`)
	p.Println(`}`)
	p.Println()
	p.Println(`func lookupPlural(loc Locale, lookup pluralRuleLookup) Plural {`)
	p.Println(`	if loc == 0 {`)
	p.Println(`		panic("invalid locale")`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	lang, _, _ := loc.tagIDs()`)
	p.Println(`	rules, has := lookup[lang]`)
	p.Println(`	if !has {`)
	p.Println(`		return Plural{}`)
	p.Println(`	}`)
	p.Println(`	return Plural{`)
	p.Println(`		{Tag: PluralTag(rules[0].tag()), rel: `, pl.relations.name, `.relation(rules[0].relationID())},`)
	p.Println(`		{Tag: PluralTag(rules[1].tag()), rel: `, pl.relations.name, `.relation(rules[1].relationID())},`)
	p.Println(`		{Tag: PluralTag(rules[2].tag()), rel: `, pl.relations.name, `.relation(rules[2].relationID())},`)
	p.Println(`		{Tag: PluralTag(rules[3].tag()), rel: `, pl.relations.name, `.relation(rules[3].relationID())},`)
	p.Println(`		{Tag: PluralTag(rules[4].tag()), rel: `, pl.relations.name, `.relation(rules[4].relationID())},`)
	p.Println(`	}`)
	p.Println(`}`)
}

func (pl *plural) testImports() []string {
	return []string{"reflect"}
}

func (pl *plural) generateTest(p *printer) {
	newLocale := func(id cldr.Identity) string {
		return fmt.Sprintf("%#0[2]*[1]x", pl.tags.tagID(id), tagIDBits/4)
	}

	newPlural := func(id cldr.Identity, pluralRules *pluralRuleLookupVar) string {
		rulesMap := pluralRules.pluralRules(id.Language)
		if len(rulesMap) == 0 {
			return "{}"
		}

		var (
			rules [5]string
			idx   int
		)

		for _, tag := range [...]string{cldr.Zero, cldr.One, cldr.Two, cldr.Few, cldr.Many} {
			if rule, has := rulesMap[tag]; has {
				rules[idx] = fmt.Sprintf("{Tag: %s, rel: "+pl.relations.name+".relation(%#x)}", strings.Title(tag), pl.relations.relationID(rule))
				idx++
			}
		}
		for ; idx < len(rules); idx++ {
			rules[idx] = "{}"
		}

		return fmt.Sprintf("{%s}", strings.Join(rules[:], ", "))
	}

	printPlurals := func(rules *pluralRuleLookupVar) {
		pl.tags.iterate(func(id cldr.Identity) {
			p.Println(`		`, newLocale(id), `: `, newPlural(id, rules), `,`)
		})
	}

	p.Println(`func TestRanges(t *testing.T) {`)
	p.Println(`	ranges := Ranges{1, 2, 3, 4}`)
	p.Println()
	p.Println(`	n := ranges.Len()`)
	p.Println(`	if n != 2 {`)
	p.Println(`		t.Errorf("unexpected length: %d", n)`)
	p.Println(`	}`)
	p.Println(`	for i := 0; i < n; i++ {`)
	p.Println(`		rng := ranges.At(i)`)
	p.Println(`		if rng.LowerBound != int(ranges[2*i]) || rng.UpperBound != int(ranges[2*i+1]) {`)
	p.Println(`			t.Errorf("unexpected range for index %d: %v", i, rng)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`func TestPluralRules(t *testing.T) {`)
	p.Println(`	collectRules := func(r PluralRules) []PluralRule {`)
	p.Println(`		var res []PluralRule`)
	p.Println(`		r.Iter(func(rule PluralRule) bool {`)
	p.Println(`			res = append(res, rule)`)
	p.Println(`			return true`)
	p.Println(`		})`)
	p.Println(`		return res`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	rules := PluralRules{}`)
	p.Println(`	collected := collectRules(rules)`)
	p.Println(`	if len(collected) != 0 {`)
	p.Println(`		t.Errorf("unexpected plural rules: %+v", collected)`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	rules = PluralRules{`)
	p.Println(`		rel: relation{`, newTestRelation(integerDigits, 2, notEqual, 1, conjunction), `, 0x1, 0x1, `, newTestRelation(integerDigits, 0, equal, 0, 0), `},`)
	p.Println(`	}`)
	p.Println(`	collected = collectRules(rules)`)
	p.Println(`	expected := []PluralRule{`)
	p.Println(`		{Operand: IntegerDigits, ModuloExp: 2, Operator: NotEqual, Ranges: Ranges{0x1, 0x1}, Connective: Conjunction},`)
	p.Println(`		{Operand: IntegerDigits, ModuloExp: 0, Operator: Equal, Ranges: Ranges{}, Connective: None},`)
	p.Println(`	}`)
	p.Println(`	if !reflect.DeepEqual(collected, expected) {`)
	p.Println(`		t.Errorf("unexpected plural rules: %+v", collected)`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`func TestLookupPlural(t *testing.T) {`)
	p.Println(`	// cardinal plurals`)
	p.Println(`	testLookupPlural(t, "cardinal", CardinalPlural, map[Locale]Plural{`)
	printPlurals(pl.cardinalRules)
	p.Println(`	})`)
	p.Println()
	p.Println(`	// ordinal plurals`)
	p.Println(`	testLookupPlural(t, "ordinal", OrdinalPlural, map[Locale]Plural{`)
	printPlurals(pl.ordinalRules)
	p.Println(`	})`)
	p.Println(`}`)
	p.Println()
	p.Println(`func testLookupPlural(t *testing.T, typ string, lookup func(Locale) Plural, expected map[Locale]Plural) {`)
	p.Println(`	for loc, expectedPlural := range expected {`)
	p.Println(`		plural := lookup(loc)`)
	p.Println(`		if !reflect.DeepEqual(plural, expectedPlural) {`)
	p.Println(`			t.Fatalf("unexpected plural for %s", loc.String())`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}

func isPowerOfTen(n int) bool {
	switch n {
	case 1, 10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000:
		return true
	default:
		return false
	}
}

func newTestRelation(operand, modexp, operator, rangeCount, connective uint) string {
	return fmt.Sprintf("%#x",
		(operand<<(modexpBits+operatorBits+rangeCountBits+connectiveBits))|
			(modexp<<(operatorBits+rangeCountBits+connectiveBits))|
			(operator<<(rangeCountBits+connectiveBits))|
			(rangeCount<<connectiveBits)|
			connective,
	)
}
