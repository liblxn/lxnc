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
	tagBits  uint
}

func newPluralRuleLookup(relation *relationLookup) *pluralRuleLookup {
	return &pluralRuleLookup{
		relation: relation,
		tagBits:  3,
	}
}

func (l *pluralRuleLookup) newPluralRule(tag, relationID uint) uint {
	return (tag << l.relation.idBits) | relationID
}

func (l *pluralRuleLookup) Imports() []string {
	return nil
}

func (l *pluralRuleLookup) Generate(p *generator.Printer) {
	ruleBits := l.tagBits + l.relation.idBits
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

	tagMask := fmt.Sprintf("%#x", (1<<l.tagBits)-1)
	relationIDMask := fmt.Sprintf("%#x", (1<<l.relation.idBits)-1)

	p.Println(`type pluralRule uint`, ruleBits)
	p.Println()
	p.Println(`func (r pluralRule) tag() uint              { return uint(r>>`, l.relation.idBits, `) & `, tagMask, ` }`)
	p.Println(`func (r pluralRule) relationID() relationID { return relationID(r & `, relationIDMask, `) }`)
	p.Println()
	p.Println(`type pluralRuleLookup map[langID][5]pluralRule`)
}

func (l *pluralRuleLookup) TestImports() []string {
	return nil
}

func (l *pluralRuleLookup) GenerateTest(p *generator.Printer) {
	rule := func(tag, relationID uint) string {
		return fmt.Sprintf("%#x", (tag<<l.relation.idBits)|relationID)
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

var (
	_ generator.Snippet = (*pluralRuleLookupVar)(nil)
)

type pluralRuleLookupVar struct {
	name      string
	typ       *pluralRuleLookup
	tag       *pluralTag
	langs     *langLookupVar
	relations *relationLookupVar

	data []pluralRulesData
}

func newPluralRuleLookupVar(name string, typ *pluralRuleLookup, tag *pluralTag, langs *langLookupVar, relations *relationLookupVar, data *cldr.Data, pluralsType pluralRulesType) *pluralRuleLookupVar {
	rulesData := make([]pluralRulesData, 0, 8)
	forEachPluralRules(data, pluralsType, func(data pluralRulesData) {
		switch len(data.rules) {
		case 0:
			return
		case 1:
			if _, has := data.rules[cldr.Other]; has {
				return // we only have the "other" tag
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
		tag:       tag,
		langs:     langs,
		relations: relations,
		data:      rulesData,
	}
}

func (v *pluralRuleLookupVar) rulesForLang(lang string) map[string]cldr.PluralRule {
	for _, data := range v.data {
		if data.lang == lang {
			return data.rules
		}
	}
	return nil
}

func (v *pluralRuleLookupVar) Imports() []string {
	return nil
}

func (v *pluralRuleLookupVar) Generate(p *generator.Printer) {
	pluralRuleBits := v.typ.tagBits + v.relations.typ.idBits
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
		panic("invalud plural rule bits")
	}

	pluralTags := map[string]uint{
		cldr.Other: v.tag.other,
		cldr.Zero:  v.tag.zero,
		cldr.One:   v.tag.one,
		cldr.Two:   v.tag.two,
		cldr.Few:   v.tag.few,
		cldr.Many:  v.tag.many,
	}

	p.Println(`var `, v.name, ` = pluralRuleLookup{ // `, len(v.data), ` items, `, 5*uint(len(v.data))*pluralRuleBits/4, ` bytes`)

	hex := func(v, bits uint) string {
		return fmt.Sprintf("%#0[2]*[1]x", v, (bits+3)/4)
	}

	for i := 0; i < len(v.data); i++ {
		data := v.data[i]
		langID := v.langs.langID(data.lang)

		var pluralRules [5]uint
		idx := 0
		for _, tag := range [...]string{cldr.Zero, cldr.One, cldr.Two, cldr.Few, cldr.Many} {
			if rule, has := data.rules[tag]; has {
				pluralRules[idx] = v.typ.newPluralRule(pluralTags[tag], v.relations.relationID(rule))
				idx++
			}
		}
		for ; idx < len(pluralRules); idx++ {
			pluralRules[idx] = v.typ.newPluralRule(v.tag.other, 0) // "other" marks the end
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
