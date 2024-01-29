package lxn

import "github.com/liblxn/lxnc/locale"

// NewCatalog returns a new catalog from the given locale data and the messages.
func NewCatalog(localeData locale.Locale, messages []Message) *Catalog {
	return &Catalog{
		Locale:   NewLocale(localeData),
		Messages: messages,
	}
}

// NewLocale returns a new locale from the given locale data.
func NewLocale(localeData locale.Locale) Locale {
	return Locale{
		ID:              localeData.String(),
		DecimalFormat:   newNumberFormat(locale.DecimalFormat(localeData)),
		MoneyFormat:     newNumberFormat(locale.MoneyFormat(localeData)),
		PercentFormat:   newNumberFormat(locale.PercentFormat(localeData)),
		CardinalPlurals: newPlurals(locale.CardinalPlural(localeData)),
		OrdinalPlurals:  newPlurals(locale.OrdinalPlural(localeData)),
	}
}

func newNumberFormat(nf locale.NumberFormat) NumberFormat {
	symbols := nf.Symbols()
	posAffixes := nf.PositiveAffixes()
	negAffixes := nf.NegativeAffixes()
	intGrouping := nf.IntegerGrouping()
	fracGrouping := nf.FractionGrouping()
	return NumberFormat{
		Symbols: Symbols{
			Decimal: symbols.Decimal,
			Group:   symbols.Group,
			Percent: symbols.Percent,
			Minus:   symbols.Minus,
			Inf:     symbols.Inf,
			Nan:     symbols.NaN,
			Zero:    uint32(symbols.Zero),
		},
		PositivePrefix:           posAffixes.Prefix,
		PositiveSuffix:           posAffixes.Suffix,
		NegativePrefix:           negAffixes.Prefix,
		NegativeSuffix:           negAffixes.Suffix,
		MinIntegerDigits:         nf.MinIntegerDigits(),
		MinFractionDigits:        nf.MinFractionDigits(),
		MaxFractionDigits:        nf.MaxFractionDigits(),
		PrimaryIntegerGrouping:   intGrouping.Primary,
		SecondaryIntegerGrouping: intGrouping.Secondary,
		FractionGrouping:         fracGrouping.Primary,
	}
}

func newPlurals(p locale.Plural) []Plural {
	var res []Plural
	for _, rules := range p.Rules() {
		plural := Plural{Category: PluralCategory(rules.Category())}
		rules.Iter(func(r locale.PluralRule) {
			plural.Rules = append(plural.Rules, newPluralRule(r))
		})
	}
	return res
}

func newPluralRule(r locale.PluralRule) PluralRule {
	nranges := r.Ranges.Len()
	ranges := make([]Range, nranges)
	for i := 0; i < nranges; i++ {
		ranges[i] = Range(r.Ranges.At(i))
	}

	mod := 0
	if r.ModuloExp > 0 {
		mod = 10
		for i := 1; i < r.ModuloExp; i++ {
			mod *= 10
		}
	}

	return PluralRule{
		Operand:    Operand(r.Operand),
		Modulo:     mod,
		Negate:     r.Operator == locale.NotEqual,
		Ranges:     ranges,
		Connective: Connective(r.Connective),
	}
}
