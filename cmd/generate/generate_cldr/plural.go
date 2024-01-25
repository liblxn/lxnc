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
	tags          *tagLookupVar
	relations     *relationLookupVar
	cardinalRules *pluralRuleLookupVar
	ordinalRules  *pluralRuleLookupVar
}

func newPlural(opration *pluralOperation, connective *connective, tags *tagLookupVar, relations *relationLookupVar, cardinalRules, ordinalRules *pluralRuleLookupVar) *plural {
	return &plural{
		operation:     opration,
		connective:    connective,
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
	p.Println(`		return Plural{}`)
	p.Println(`	}`)
	p.Println(`	return Plural{`)
	p.Println(`		rules: [5]PluralRules{`)
	p.Println(`			{tag: PluralTag(rules[0].tag()), rel: `, pl.relations.name, `.relation(rules[0].relationID())},`)
	p.Println(`			{tag: PluralTag(rules[1].tag()), rel: `, pl.relations.name, `.relation(rules[1].relationID())},`)
	p.Println(`			{tag: PluralTag(rules[2].tag()), rel: `, pl.relations.name, `.relation(rules[2].relationID())},`)
	p.Println(`			{tag: PluralTag(rules[3].tag()), rel: `, pl.relations.name, `.relation(rules[3].relationID())},`)
	p.Println(`			{tag: PluralTag(rules[4].tag()), rel: `, pl.relations.name, `.relation(rules[4].relationID())},`)
	p.Println(`		},`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (r *Plural) Rules() []PluralRules {`)
	p.Println(`	for i := len(r.rules); i > 0; i-- {`)
	p.Println(`		rules := r.rules[i-1]`)
	p.Println(`		if rules.tag != 0 || len(rules.rel) != 0 {`)
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
	p.Println(`// PluralRules holds a collection of plural rules for a specific plural tag.`)
	p.Println(`// All rules in this collection are connected with each other (see PluralRule`)
	p.Println(`// and Connective).`)
	p.Println(`type PluralRules struct {`)
	p.Println(`	tag PluralTag`)
	p.Println(`	rel relation`)
	p.Println(`}`)
	p.Println()
	p.Println(`// Tag returns the plural tag for the plural rules.`)
	p.Println(`func (r PluralRules) Tag() PluralTag {`)
	p.Println(`	return r.tag`)
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
		rulesMap := pluralRules.rulesForLang(id.Language)
		if len(rulesMap) == 0 {
			return "{}"
		}

		var (
			rules [5]string
			idx   int
		)

		for _, tag := range [...]string{cldr.Zero, cldr.One, cldr.Two, cldr.Few, cldr.Many} {
			if rule, has := rulesMap[tag]; has {
				rules[idx] = fmt.Sprintf("{tag: %s, rel: "+pl.relations.name+".relation(%#x)}", strings.Title(tag), pl.relations.relationID(rule))
				idx++
			}
		}
		for ; idx < len(rules); idx++ {
			rules[idx] = "{}"
		}

		return fmt.Sprintf("{rules: [5]PluralRules{%s}}", strings.Join(rules[:], ", "))
	}

	printPlurals := func(rules *pluralRuleLookupVar) {
		for _, id := range pl.tags.ids {
			p.Println(`		`, newLocale(id), `: `, newPlural(id, rules), `,`)
		}
	}

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
