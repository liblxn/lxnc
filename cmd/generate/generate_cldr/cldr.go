package generate_cldr

import (
	"fmt"
	"sort"

	"github.com/liblxn/lxnc/internal/cldr"
)

func forEachIdentity(data *cldr.Data, iter func(cldr.Identity)) {
	for _, id := range data.Identities {
		if skipIdentity(id) {
			continue
		}
		iter(normalizeIdentity(id))
	}
}

type parentTagData struct {
	child  cldr.Identity
	parent cldr.Identity
}

func forEachParentIdentity(data *cldr.Data, iter func(data parentTagData)) {
	for childTag, parentTag := range data.ParentIdentities {
		child, hasChild := data.Identities[childTag]
		parent, hasParent := data.Identities[parentTag]
		if hasChild && hasParent {
			iter(parentTagData{
				child:  normalizeIdentity(child),
				parent: normalizeIdentity(parent),
			})
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

func identityLess(x, y cldr.Identity) bool {
	switch {
	case x.Language != y.Language:
		return x.Language < y.Language
	case x.Script != y.Script:
		return x.Script < y.Script
	default:
		return x.Territory < y.Territory
	}
}

type regionContainmentData struct {
	childRegion   string
	parentRegions []string
}

func forEachRegionContainment(data *cldr.Data, iter func(regionContainmentData)) {
	regionSet := make(map[string]struct{})
	forEachIdentity(data, func(id cldr.Identity) {
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
		iter(regionContainmentData{
			childRegion:   child,
			parentRegions: parents,
		})
	}
}

type numbersData struct {
	id     cldr.Identity
	nf     cldr.NumberFormat
	symb   cldr.NumberSymbols
	numsys cldr.NumberingSystem
}

type numbersFilter uint

const (
	decimalNumbers numbersFilter = 1 << iota
	currencyNumbers
	percentNumbers
	allFormats = decimalNumbers | currencyNumbers | percentNumbers
)

func forEachNumbers(data *cldr.Data, filter numbersFilter, iter func(numbersData)) {
	needDecimal := (filter & decimalNumbers) != 0
	needCurrency := (filter & currencyNumbers) != 0
	needPercent := (filter & percentNumbers) != 0

	iterateNumbers(data, func(id cldr.Identity, nums cldr.Numbers, symbols cldr.NumberSymbols, numsys cldr.NumberingSystem) {
		if decimal, has := nums.DecimalFormats[numsys.ID]; needDecimal && has {
			iter(numbersData{id: id, nf: decimal, symb: symbols, numsys: numsys})
		}
		if money, has := nums.CurrencyFormats[numsys.ID]; needCurrency && has {
			iter(numbersData{id: id, nf: money, symb: symbols, numsys: numsys})
		}
		if percent, has := nums.PercentFormats[numsys.ID]; needPercent && has {
			iter(numbersData{id: id, nf: percent, symb: symbols, numsys: numsys})
		}
	})
}

func iterateNumbers(data *cldr.Data, iter func(id cldr.Identity, numbers cldr.Numbers, symbols cldr.NumberSymbols, numsys cldr.NumberingSystem)) {
	validDigits := func(digits []rune) bool {
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
			panic(fmt.Sprintf("invalid digits for %s: %s", id.String(), string(numsys.Digits)))
		}

		symbols := data.NumberSymbols(id, numsysID)
		iter(normalizeIdentity(id), numbers, symbols, numsys)
	}
}

func forEachPluralRelation(data *cldr.Data, iter func(cldr.PluralRule)) {
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

type pluralRulesType int

const (
	cardinalPluralRules pluralRulesType = 1
	ordinalPluralRules  pluralRulesType = 2
)

type pluralRulesData struct {
	lang  string
	rules map[string]cldr.PluralRule
}

func forEachPluralRules(data *cldr.Data, typ pluralRulesType, iter func(data pluralRulesData)) {
	var rules []cldr.PluralRules
	switch typ {
	case cardinalPluralRules:
		rules = data.Plurals.Cardinal
	case ordinalPluralRules:
		rules = data.Plurals.Ordinal
	default:
		panic("invalid filter")
	}

	langs := languages(data)
	for _, r := range rules {
		for _, lang := range r.Locales {
			if _, has := langs[lang]; has {
				iter(pluralRulesData{
					lang:  lang,
					rules: r.Rules,
				})
			}
		}
	}
}

func languages(data *cldr.Data) map[string]struct{} {
	langs := make(map[string]struct{})
	for _, id := range data.Identities {
		langs[id.Language] = struct{}{}
	}
	return langs
}
