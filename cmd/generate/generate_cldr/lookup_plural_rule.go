package generate_cldr

import (
	"fmt"
	"sort"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*pluralRuleLookup)(nil)
	_ generator.TestSnippet = (*pluralRuleLookup)(nil)
)

type pluralRuleLookup struct {
	relation *relationLookup
	category *pluralCategory
}

func newPluralRuleLookup(relation *relationLookup, category *pluralCategory) *pluralRuleLookup {
	return &pluralRuleLookup{
		relation: relation,
		category: category,
	}
}

func (l *pluralRuleLookup) newPluralRule(category, relationID uint) uint {
	return (category << l.relation.idBits) | relationID
}

func (l *pluralRuleLookup) Imports() []string {
	return nil
}

func (l *pluralRuleLookup) Generate(p *generator.Printer) {
	ruleBits := l.category.bits + l.relation.idBits
	switch {
	case ruleBits <= 8:
		ruleBits = 8
	case ruleBits <= 16:
		ruleBits = 16
	case ruleBits <= 32:
		ruleBits = 32
	case ruleBits <= 64:
		ruleBits = 64
	default:
		panic("invalid plural rule bits")
	}

	categoryMask := fmt.Sprintf("%#x", (1<<l.category.bits)-1)
	relationIDMask := fmt.Sprintf("%#x", (1<<l.relation.idBits)-1)

	p.Println(`type pluralRule uint`, ruleBits)
	p.Println()
	p.Println(`func (r pluralRule) category() uint         { return uint(r>>`, l.relation.idBits, `) & `, categoryMask, ` }`)
	p.Println(`func (r pluralRule) relationID() relationID { return relationID(r & `, relationIDMask, `) }`)
	p.Println()
	p.Println(`type pluralRuleLookup map[langID][5]pluralRule`)
}

func (l *pluralRuleLookup) TestImports() []string {
	return nil
}

func (l *pluralRuleLookup) GenerateTest(p *generator.Printer) {
	rule := func(category, relationID uint) string {
		return fmt.Sprintf("%#x", (category<<l.relation.idBits)|relationID)
	}

	p.Println(`func TestPluralRule(t *testing.T) {`)
	p.Println(`	const rule pluralRule = `, rule(2, 7))
	p.Println()
	p.Println(`	if category := rule.category(); category != 2 {`)
	p.Println(`		t.Errorf("unexpected category: %d", category)`)
	p.Println(`	}`)
	p.Println(`	if rid := rule.relationID(); rid != 7 {`)
	p.Println(`		t.Errorf("unexpected relation id: %d", rid)`)
	p.Println(`	}`)
	p.Println(`}`)
}

var (
	_ generator.Snippet = (*pluralRuleLookupVar)(nil)
)

type pluralRuleLookupVar struct {
	name      string
	typ       *pluralRuleLookup
	category  *pluralCategory
	langs     *langLookupVar
	relations *relationLookupVar

	data []pluralRulesData
}

func newPluralRuleLookupVar(name string, typ *pluralRuleLookup, category *pluralCategory, langs *langLookupVar, relations *relationLookupVar, data *cldr.Data, pluralsType pluralRulesType) *pluralRuleLookupVar {
	rulesData := make([]pluralRulesData, 0, 8)
	forEachPluralRules(data, pluralsType, func(data pluralRulesData) {
		switch len(data.rules) {
		case 0:
			return
		case 1:
			if _, has := data.rules[cldr.Other]; has {
				return // we only have the "other" category
			}
		}

		rulesData = append(rulesData, data)
	})

	sort.Slice(rulesData, func(i, j int) bool {
		return rulesData[i].lang < rulesData[j].lang
	})

	return &pluralRuleLookupVar{
		name:      name,
		typ:       typ,
		category:  category,
		langs:     langs,
		relations: relations,
		data:      rulesData,
	}
}

func (v *pluralRuleLookupVar) forEachPluralRuleForLang(lang string, iter func(category uint, rule cldr.PluralRule)) {
	for _, data := range v.data {
		if data.lang == lang {
			v.forEachPluralRule(data, iter)
			return
		}
	}
}

func (v *pluralRuleLookupVar) forEachPluralRule(data pluralRulesData, iter func(category uint, rule cldr.PluralRule)) {
	for _, category := range [...]uint{v.category.zero, v.category.one, v.category.two, v.category.few, v.category.many} {
		if rule, has := data.rules[v.category.cldrConstantOf(category)]; has {
			iter(category, rule)
		}
	}
}

func (v *pluralRuleLookupVar) Imports() []string {
	return nil
}

func (v *pluralRuleLookupVar) Generate(p *generator.Printer) {
	pluralRuleBits := v.typ.category.bits + v.relations.typ.idBits
	switch {
	case pluralRuleBits <= 8:
		pluralRuleBits = 8
	case pluralRuleBits <= 16:
		pluralRuleBits = 16
	case pluralRuleBits <= 32:
		pluralRuleBits = 32
	case pluralRuleBits <= 64:
		pluralRuleBits = 64
	default:
		panic("invalid plural rule bits")
	}

	hex := func(v, bits uint) string {
		return fmt.Sprintf("%#0[2]*[1]x", v, (bits+3)/4)
	}

	p.Println(`var `, v.name, ` = pluralRuleLookup{ // `, len(v.data), ` items, `, 5*uint(len(v.data))*pluralRuleBits/4, ` bytes`)
	for _, data := range v.data {
		langID := v.langs.langID(data.lang)

		var pluralRules [5]uint
		idx := 0
		v.forEachPluralRule(data, func(category uint, rule cldr.PluralRule) {
			pluralRules[idx] = v.typ.newPluralRule(category, v.relations.relationID(rule))
			idx++
		})

		for ; idx < len(pluralRules); idx++ {
			pluralRules[idx] = v.typ.newPluralRule(v.category.other, 0) // "other" marks the end
		}

		p.Print(`	`, hex(langID, v.langs.typ.idBits), `: {`)
		for _, pluralRule := range pluralRules[:len(pluralRules)-1] {
			p.Print(hex(pluralRule, uint(pluralRuleBits)), `, `)
		}
		p.Print(hex(pluralRules[len(pluralRules)-1], uint(pluralRuleBits)))
		p.Println(`}, // `, data.lang)
	}

	p.Println(`}`)
}
