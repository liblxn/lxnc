package generate_cldr

import (
	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

const lineLength = 60

func FileSnippets(packageName string, data *cldr.Data) map[string]generator.Snippet {
	// locale
	langLookup := newLangLookup()
	scriptLookup := newScriptLookup()
	regionLookup := newRegionLookup()
	regionContainmentLookup := newRegionContainmentLookup(regionLookup)
	tagLookup := newTagLookup(langLookup, scriptLookup, regionLookup)
	parentTagLookup := newParentTagLookup(tagLookup)

	langLookupVar := newLangLookupVar("langTags", langLookup, data)
	scriptLookupVar := newScriptLookupVar("scriptTags", scriptLookup, data)
	regionLookupVar := newRegionLookupVar("regionTags", regionLookup, data)
	regionContainmentLookupVar := newRegionContainmentLookupVar("regionContainments", regionContainmentLookup, regionLookupVar, data)
	tagLookupVar := newTagLookupVar("localeTags", tagLookup, langLookupVar, scriptLookupVar, regionLookupVar, data)
	parentTagLookupVar := newParentTagLookupVar("parentLocaleTags", parentTagLookup, tagLookupVar, data)

	// number format
	affixLookup := newAffixLookup()
	patternLookup := newPatternLookup(affixLookup)
	symbolsLookup := newSymbolsLookup()
	zeroLookup := newZeroLookup()
	numbersLookup := newNumbersLookup(patternLookup, symbolsLookup, zeroLookup)

	affixLookupVar := newAffixLookupVar("affixes", affixLookup, data)
	patternLookupVar := newPatternLookupVar("patterns", patternLookup, affixLookupVar, data)
	symbolsLookupVar := newSymbolsLookupVar("numberSymbols", symbolsLookup, data)
	zeroLookupVar := newZeroLookupVar("zeros", zeroLookup, data)
	decimalNumbersLookupVar := newNumbersLookupVar("decimalNumbers", numbersLookup, tagLookupVar, patternLookupVar, symbolsLookupVar, zeroLookupVar, data, decimalNumbers)
	moneyNumbersLookupVar := newNumbersLookupVar("moneyNumbers", numbersLookup, tagLookupVar, patternLookupVar, symbolsLookupVar, zeroLookupVar, data, currencyNumbers)
	percentNumbersLookupVar := newNumbersLookupVar("percentNumbers", numbersLookup, tagLookupVar, patternLookupVar, symbolsLookupVar, zeroLookupVar, data, percentNumbers)

	// plural
	connective := newConnective()
	pluralOperation := newPluralOperation()
	relationLookup := newRelationLookup(pluralOperation, connective)
	pluralCategory := newPluralCategory()
	pluralRuleLookup := newPluralRuleLookup(relationLookup, pluralCategory)

	relationLookupVar := newRelationLookupVar("relations", relationLookup, pluralOperation, data)
	cardinalPluralRulesLookupVar := newPluralRuleLookupVar("cardinalRules", pluralRuleLookup, pluralCategory, langLookupVar, relationLookupVar, data, cardinalPluralRules)
	ordinalPluralRulesLookupVar := newPluralRuleLookupVar("ordinalRules", pluralRuleLookup, pluralCategory, langLookupVar, relationLookupVar, data, ordinalPluralRules)

	return map[string]generator.Snippet{
		"locale.go": generator.Snippets{
			newLocale(packageName, tagLookupVar, parentTagLookupVar, regionContainmentLookupVar),
			tagLookup,
			langLookup,
			scriptLookup,
			regionLookup,
			regionContainmentLookup,
			parentTagLookup,
		},
		"number_format.go": generator.Snippets{
			newNumberFormat(decimalNumbersLookupVar, moneyNumbersLookupVar, percentNumbersLookupVar, affixLookupVar, tagLookup),
			affixLookup,
			patternLookup,
			symbolsLookup,
			zeroLookup,
			numbersLookup,
		},
		"plural.go": generator.Snippets{
			newPlural(pluralOperation, connective, pluralCategory, tagLookupVar, relationLookupVar, cardinalPluralRulesLookupVar, ordinalPluralRulesLookupVar),
			connective,
			pluralCategory,
			pluralOperation,
			relationLookup,
			pluralRuleLookup,
		},
		"tables.go": generator.Snippets{
			langLookupVar,
			scriptLookupVar,
			regionLookupVar,
			regionContainmentLookupVar,
			tagLookupVar,
			parentTagLookupVar,

			affixLookupVar,
			patternLookupVar,
			symbolsLookupVar,
			zeroLookupVar,
			decimalNumbersLookupVar,
			moneyNumbersLookupVar,
			percentNumbersLookupVar,

			relationLookupVar,
			cardinalPluralRulesLookupVar,
			ordinalPluralRulesLookupVar,
		},
	}
}
