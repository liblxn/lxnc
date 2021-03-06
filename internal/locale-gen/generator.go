package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/liblxn/lxnc/internal/cldr"
)

const lineLength = 60

type options struct {
	outputDir   string
	packageName string
}

type generator struct {
	outputDir   string
	mkOutputDir bool
	packageName string
}

func newGenerator(opt options) (*generator, error) {
	info, err := os.Stat(opt.outputDir)
	switch {
	case err != nil:
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory not found: %s", opt.outputDir)
		}
		return nil, err
	case !info.IsDir():
		return nil, fmt.Errorf("not a directory: %s", opt.outputDir)
	}

	return &generator{
		outputDir:   opt.outputDir,
		packageName: opt.packageName,
	}, nil
}

func (g *generator) Generate(data *cldr.Data) error {
	snippets := g.generateSnippets(data)
	for filename, snippet := range snippets {
		ext := filepath.Ext(filename)
		codeFilename := filepath.Join(g.outputDir, filename)
		testFilename := strings.TrimSuffix(codeFilename, ext) + "_test" + ext

		err := g.generateGoFile(snippet, codeFilename)
		if err != nil {
			return err
		}

		if testSnippet := testSnippetOf(snippet); testSnippet != nil {
			err = g.generateGoFile(testSnippet, testFilename)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) generateGoFile(s snippet, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	p := newPrinter(f)
	p.Println(`package `, g.packageName)
	p.Println()
	if imports := s.imports(); len(imports) != 0 {
		sort.Strings(imports)
		p.Println(`import (`)
		for _, imp := range imports {
			p.Println(`	"`, imp, `"`)
		}
		p.Println(`)`)
		p.Println()
	}
	s.generate(p)
	return p.Err()
}

// returns filename => snippet
func (g *generator) generateSnippets(data *cldr.Data) map[string]snippet {
	langs := newLangLookupVar("langTags")
	scripts := newScriptLookupVar("scriptTags")
	regions := newRegionLookupVar("regionTags")
	locales := newTagLookupVar("localeTags", langs, scripts, regions)
	regionContainment := newRegionContainmentLookupVar("regionContainment", regions)
	parentLocales := newParentTagLookupVar("parentLocaleTags", locales)

	iterateIdentities(data, func(id cldr.Identity) {
		langs.add(id.Language)
		if id.Script != "" {
			scripts.add(id.Script)
		}
		if id.Territory != "" {
			regions.add(id.Territory)
		}
		locales.add(id)
	})

	iterateRegionContainments(data, func(child string, parents []string) {
		regionContainment.add(child, parents)
	})

	iterateParentIdentities(data, func(child, parent cldr.Identity) {
		parentLocales.add(child, parent)
	})

	affixes := newAffixLookupVar("affixes")
	patterns := newPatternLookupVar("patterns", affixes)
	symbols := newSymbolsLookupVar("numberSymbols")
	zeros := newZeroLookupVar("zeros")
	decimalNumbers := newNumbersLookupVar("decimalNumbers", locales, patterns, symbols, zeros)
	moneyNumbers := newNumbersLookupVar("moneyNumbers", locales, patterns, symbols, zeros)
	percentNumbers := newNumbersLookupVar("percentNumbers", locales, patterns, symbols, zeros)

	iterateNumbers(data, func(id cldr.Identity, num cldr.Numbers, symb cldr.NumberSymbols, numsys cldr.NumberingSystem) {
		if decimal, has := num.DecimalFormats[numsys.ID]; has {
			affixes.add(decimal.PositivePrefix, decimal.PositiveSuffix)
			affixes.add(decimal.NegativePrefix, decimal.NegativeSuffix)
			patterns.add(decimal)
			decimalNumbers.add(id, decimal, symb, numsys)
		}
		if money, has := num.CurrencyFormats[numsys.ID]; has {
			affixes.add(money.PositivePrefix, money.PositiveSuffix)
			affixes.add(money.NegativePrefix, money.NegativeSuffix)
			patterns.add(money)
			moneyNumbers.add(id, money, symb, numsys)
		}
		if percent, has := num.PercentFormats[numsys.ID]; has {
			affixes.add(percent.PositivePrefix, percent.PositiveSuffix)
			affixes.add(percent.NegativePrefix, percent.NegativeSuffix)
			patterns.add(percent)
			percentNumbers.add(id, percent, symb, numsys)
		}

		symbols.add(symb)
		zeros.add(numsys.Digits[0])
	})

	relations := newRelationLookupVar("relations")

	iteratePluralRelations(data, func(rule cldr.PluralRule) {
		relations.add(rule)
	})

	cardinalRules := newPluralRuleLookupVar("cardinalRules", langs, relations)
	ordinalRules := newPluralRuleLookupVar("ordinalRules", langs, relations)

	iteratePluralRules(data, data.Plurals.Cardinal, func(lang string, rules map[string]cldr.PluralRule) {
		cardinalRules.add(lang, rules)
	})
	iteratePluralRules(data, data.Plurals.Ordinal, func(lang string, rules map[string]cldr.PluralRule) {
		ordinalRules.add(lang, rules)
	})

	return map[string]snippet{
		"error.go": snippets{
			errorString{},
		},
		"tags.go": snippets{
			langLookup{},
			scriptLookup{},
			regionLookup{},
			tagLookup{},
			regionContainmentLookup{},
			parentTagLookup{},
		},
		"numbers.go": snippets{
			affixLookup{},
			patternLookup{},
			symbolsLookup{},
			zeroLookup{},
			numbersLookup{},
		},
		"locale.go": snippets{
			newLocale(g.packageName, locales, parentLocales, regionContainment),
		},
		"number_format.go": snippets{
			newNumberFormat(decimalNumbers, moneyNumbers, percentNumbers, affixes),
		},
		"plural.go": snippets{
			newPlural(locales, relations, cardinalRules, ordinalRules),
			relationLookup{},
			pluralRuleLookup{},
		},
		"tables.go": snippets{
			langs,
			scripts,
			regions,
			regionContainment,
			zeros,
			affixes,
			patterns,
			symbols,
			locales,
			parentLocales,
			decimalNumbers,
			moneyNumbers,
			percentNumbers,
			relations,
			cardinalRules,
			ordinalRules,
		},
	}
}

func iterateIdentities(data *cldr.Data, iter func(cldr.Identity)) {
	for _, id := range data.Identities {
		if !skipIdentity(id) {
			iter(normalizeIdentity(id))
		}
	}
}

func iterateRegionContainments(data *cldr.Data, iter func(child string, parents []string)) {
	regionSet := make(map[string]struct{})
	iterateIdentities(data, func(id cldr.Identity) {
		if id.Territory != "" {
			regionSet[id.Territory] = struct{}{}
		}
	})

	regions := make([]string, 0, len(regionSet))
	for region := range regionSet {
		regions = append(regions, region)
	}
	sort.Strings(regions)

	containment := make(map[string][]string) // child => parents
	for _, parent := range regions {
		children := data.Regions.Territories(parent)
		if len(children) == 1 && children[0] == parent {
			continue
		}
		for _, child := range children {
			containment[child] = append(containment[child], parent)
		}
	}

	for child, parents := range containment {
		// Reverse the order of the parents to have the largest region at the end.
		for i, j := 0, len(parents)-1; i < j; {
			parents[i], parents[j] = parents[j], parents[i]
			i++
			j--
		}
		iter(child, parents)
	}
}

func iterateParentIdentities(data *cldr.Data, iter func(child, parent cldr.Identity)) {
	for childTag, parentTag := range data.ParentIdentities {
		child, hasChild := data.Identities[childTag]
		parent, hasParent := data.Identities[parentTag]
		if hasChild && hasParent {
			iter(normalizeIdentity(child), normalizeIdentity(parent))
		}
	}
}

func iterateNumbers(data *cldr.Data, iter func(id cldr.Identity, numbers cldr.Numbers, symbols cldr.NumberSymbols, numsys cldr.NumberingSystem)) {
	for locale, numbers := range data.Numbers {
		id, has := data.Identities[locale]
		switch {
		case !has:
			panic(fmt.Sprintf("cannot find locale identity: %s", locale))
		case skipIdentity(id):
			continue
		}

		numsysID := data.DefaultNumberingSystem(id)
		numsys, has := data.NumberingSystems[numsysID]
		switch {
		case !has:
			panic(fmt.Sprintf("numbering system not found for %s: %s", id.String(), numsysID))
		case !validDigits(numsys.Digits):
			panic(fmt.Sprintf("invalid digits for %s: %s", id.String(), numsys.Digits))
		}

		symbols := data.FullNumberSymbols(id, numsysID)
		iter(normalizeIdentity(id), numbers, symbols, numsys)
	}
}

func iteratePluralRelations(data *cldr.Data, iter func(cldr.PluralRule)) {
	langs := languages(data)
	process := func(r []cldr.PluralRules) {
		for _, rules := range r {
			has := false
			for _, lang := range rules.Locales {
				if _, has = langs[lang]; has {
					break
				}
			}
			if !has {
				continue
			}

			for tag, rule := range rules.Rules {
				if tag != "other" {
					iter(rule)
				}
			}
		}
	}

	process(data.Plurals.Cardinal)
	process(data.Plurals.Ordinal)
}

func iteratePluralRules(data *cldr.Data, rules []cldr.PluralRules, iter func(lang string, rules map[string]cldr.PluralRule)) {
	langs := languages(data)
	for _, r := range rules {
		for _, lang := range r.Locales {
			if _, has := langs[lang]; has {
				iter(lang, r.Rules)
			}
		}
	}
}

func skipIdentity(id cldr.Identity) bool {
	return id.Variant != ""
}

func normalizeIdentity(id cldr.Identity) cldr.Identity {
	if id.IsRoot() {
		return cldr.Identity{Language: "und"}
	}
	return id
}

func validDigits(digits []rune) bool {
	if len(digits) != 10 {
		return false
	}

	zero := digits[0]
	for i := 1; i < len(digits); i++ {
		if digits[i] != zero+rune(i) {
			return false
		}
	}
	return true
}

func languages(data *cldr.Data) map[string]struct{} {
	langs := make(map[string]struct{})
	for _, id := range data.Identities {
		langs[id.Language] = struct{}{}
	}
	return langs
}
