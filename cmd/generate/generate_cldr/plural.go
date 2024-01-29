package generate_cldr

import (
	"fmt"
	"strings"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*plural)(nil)
	_ generator.TestSnippet = (*plural)(nil)
)

type plural struct {
	operation     *pluralOperation
	connective    *connective
	category      *pluralCategory
	tags          *tagLookupVar
	relations     *relationLookupVar
	cardinalRules *pluralRuleLookupVar
	ordinalRules  *pluralRuleLookupVar
}

func newPlural(opration *pluralOperation, connective *connective, category *pluralCategory, tags *tagLookupVar, relations *relationLookupVar, cardinalRules, ordinalRules *pluralRuleLookupVar) *plural {
	return &plural{
		operation:     opration,
		connective:    connective,
		category:      category,
		tags:          tags,
		relations:     relations,
		cardinalRules: cardinalRules,
		ordinalRules:  ordinalRules,
	}
}

func (pl *plural) Imports() []string {
	return nil
}

func (pl *plural) Generate(p *generator.Printer) {
	other := pl.category.enumeratorOf(pl.category.other)

	p.Println(`// Plural holds the plural data for a language. There can at most be five plural rule`)
	p.Println(`// collections for 'zero', 'one', 'two', 'few', and 'many'. If there are less than five`)
	p.Println(`// rules, the rest will be filled with empty 'other' rules.`)
	p.Println(`type Plural struct{`)
	p.Println(`	rules [5]PluralRules`)
	p.Println(`}`)
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
	p.Println(`		return Plural{`)
	p.Println(`			rules: [5]PluralRules{{cat: `, other, `}, {cat: `, other, `}, {cat: `, other, `}, {cat: `, other, `}, {cat: `, other, `}},`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`	return Plural{`)
	p.Println(`		rules: [5]PluralRules{`)
	p.Println(`			{cat: PluralCategory(rules[0].category()), rel: `, pl.relations.name, `.relation(rules[0].relationID())},`)
	p.Println(`			{cat: PluralCategory(rules[1].category()), rel: `, pl.relations.name, `.relation(rules[1].relationID())},`)
	p.Println(`			{cat: PluralCategory(rules[2].category()), rel: `, pl.relations.name, `.relation(rules[2].relationID())},`)
	p.Println(`			{cat: PluralCategory(rules[3].category()), rel: `, pl.relations.name, `.relation(rules[3].relationID())},`)
	p.Println(`			{cat: PluralCategory(rules[4].category()), rel: `, pl.relations.name, `.relation(rules[4].relationID())},`)
	p.Println(`		},`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (r *Plural) Rules() []PluralRules {`)
	p.Println(`	for i := len(r.rules); i > 0; i-- {`)
	p.Println(`		rules := r.rules[i-1]`)
	p.Println(`		if rules.cat != `, other, ` || len(rules.rel) != 0 {`)
	p.Println(`			return r.rules[:i]`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`	return nil`)
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
	p.Println(`// PluralRules holds a collection of plural rules for a specific plural category.`)
	p.Println(`// All rules in this collection are connected with each other (see PluralRule`)
	p.Println(`// and Connective).`)
	p.Println(`type PluralRules struct {`)
	p.Println(`	cat PluralCategory`)
	p.Println(`	rel relation`)
	p.Println(`}`)
	p.Println()
	p.Println(`// Category returns the plural category for the plural rules.`)
	p.Println(`func (r PluralRules) Category() PluralCategory {`)
	p.Println(`	return r.cat`)
	p.Println(`}`)
	p.Println()
	p.Println(`// Iter iterates over all plural rules in the collection. The iterator should`)
	p.Println(`// return true, if the iteration should be continued. Otherwise false should be`)
	p.Println(`// returned.`)
	p.Println(`func (r PluralRules) Iter(iter func(PluralRule)) {`)
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
	p.Println(`		iter(rule)`)
	p.Println(`		if rule.Connective == None {`)
	p.Println(`			return`)
	p.Println(`		}`)
	p.Println(`		rel = rel.next()`)
	p.Println(`	}`)
	p.Println(`}`)
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
	p.Println(`	r []uint`, pl.relations.typ.pluralChunkBits)
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
}

func (pl *plural) TestImports() []string {
	return []string{"reflect"}
}

func (pl *plural) GenerateTest(p *generator.Printer) {
	newLocale := func(id cldr.Identity) string {
		return fmt.Sprintf("%#0[2]*[1]x", pl.tags.tagID(id), pl.tags.typ.idBits/4)
	}

	newPlural := func(id cldr.Identity, pluralRules *pluralRuleLookupVar) string {
		rules := [5]string{}
		idx := 0

		pluralRules.forEachPluralRuleForLang(id.Language, func(category uint, rule cldr.PluralRule) {
			rules[idx] = fmt.Sprintf("{cat: %s, rel: "+pl.relations.name+".relation(%#x)}", pl.category.enumeratorOf(category), pl.relations.relationID(rule))
			idx++
		})

		if idx == 0 {
			return "nil"
		}
		return fmt.Sprintf("{%s}", strings.Join(rules[:idx], ", "))
	}

	printPlurals := func(rules *pluralRuleLookupVar) {
		maxIDLen := 0
		for _, id := range pl.tags.ids {
			if n := len(id.String()); n > maxIDLen {
				maxIDLen = n
			}
		}

		for _, id := range pl.tags.ids {
			ident := id.String()
			ident += strings.Repeat(" ", maxIDLen-len(ident))
			p.Println(`		/* `, ident, ` */ `, newLocale(id), `: `, newPlural(id, rules), `,`)
		}

	}

	p.Println(`func TestLookupPlural(t *testing.T) {`)
	p.Println(`	// cardinal plurals`)
	p.Println(`	testLookupPlural(t, "cardinal", CardinalPlural, map[Locale][]PluralRules{`)
	printPlurals(pl.cardinalRules)
	p.Println(`	})`)
	p.Println()
	p.Println(`	// ordinal plurals`)
	p.Println(`	testLookupPlural(t, "ordinal", OrdinalPlural, map[Locale][]PluralRules{`)
	printPlurals(pl.ordinalRules)
	p.Println(`	})`)
	p.Println(`}`)
	p.Println()
	p.Println(`func TestRanges(t *testing.T) {`)
	p.Println(`	ranges := Ranges{r: []uint`, pl.relations.typ.pluralChunkBits, `{1, 2, 3, 4}}`)
	p.Println()
	p.Println(`	n := ranges.Len()`)
	p.Println(`	if n != 2 {`)
	p.Println(`		t.Errorf("unexpected length: %d", n)`)
	p.Println(`	}`)
	p.Println(`	for i := 0; i < n; i++ {`)
	p.Println(`		rng := ranges.At(i)`)
	p.Println(`		if rng.LowerBound != int(ranges.r[2*i]) || rng.UpperBound != int(ranges.r[2*i+1]) {`)
	p.Println(`			t.Errorf("unexpected range for index %d: %v", i, rng)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`func TestPluralRules(t *testing.T) {`)
	p.Println(`	collectRules := func(r PluralRules) []PluralRule {`)
	p.Println(`		var res []PluralRule`)
	p.Println(`		r.Iter(func(rule PluralRule) {`)
	p.Println(`			res = append(res, rule)`)
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
	p.Println(`		rel: relation{`, pl.relations.typ.newTestRelation(uint(pl.operation.intDigits), 2, uint(pl.operation.neq), 1, pl.connective.conjunction), `, 0x1, 0x1, `, pl.relations.typ.newTestRelation(uint(pl.operation.intDigits), 0, uint(pl.operation.eq), 0, 0), `},`)
	p.Println(`	}`)
	p.Println(`	collected = collectRules(rules)`)
	p.Println(`	expected := []PluralRule{`)
	p.Println(`		{Operand: IntegerDigits, ModuloExp: 2, Operator: NotEqual, Ranges: Ranges{r: []uint`, pl.relations.typ.pluralChunkBits, `{0x1, 0x1}}, Connective: Conjunction},`)
	p.Println(`		{Operand: IntegerDigits, ModuloExp: 0, Operator: Equal, Ranges: Ranges{r: []uint`, pl.relations.typ.pluralChunkBits, `{}}, Connective: None},`)
	p.Println(`	}`)
	p.Println(`	if !reflect.DeepEqual(collected, expected) {`)
	p.Println(`		t.Errorf("unexpected plural rules: %+v", collected)`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`func testLookupPlural(t *testing.T, typ string, lookup func(Locale) Plural, expected map[Locale][]PluralRules) {`)
	p.Println(`	for loc, expectedRules := range expected {`)
	p.Println(`		plural := lookup(loc)`)
	p.Println(`		rules := plural.Rules()`)
	p.Println(`		if !reflect.DeepEqual(rules, expectedRules) {`)
	p.Println(`			t.Fatalf("unexpected %s plural for %s", typ, loc.String())`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}
