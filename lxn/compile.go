package lxn

import (
	"os"

	"github.com/liblxn/lxnc/internal/locale"
	"github.com/liblxn/lxnc/schema"
)

type input struct {
	filename string
	bytes    []byte
}

// CompileFiles parses the given files and determines the locale information which is need
// for formatting data.
func CompileFiles(loc locale.Locale, filenames ...string) (schema.Catalog, error) {
	inputs := make([]input, 0, len(filenames))
	for _, filename := range filenames {
		bytes, err := os.ReadFile(filename)
		if err != nil {
			return schema.Catalog{}, err
		}
		inputs = append(inputs, input{filename: filename, bytes: bytes})
	}
	return compile(loc, inputs)
}

func compile(loc locale.Locale, inputs []input) (schema.Catalog, error) {
	var (
		p    parser
		msgs []schema.Message
	)
	for _, input := range inputs {
		m, err := p.Parse(input.filename, input.bytes)
		if err != nil {
			return schema.Catalog{}, err
		}
		msgs = append(msgs, m...)
	}

	return schema.Catalog{
		Locale: schema.Locale{
			ID:              loc.String(),
			DecimalFormat:   newNumberFormat(locale.DecimalFormat(loc)),
			MoneyFormat:     newNumberFormat(locale.MoneyFormat(loc)),
			PercentFormat:   newNumberFormat(locale.PercentFormat(loc)),
			CardinalPlurals: newPlurals(locale.CardinalPlural(loc)),
			OrdinalPlurals:  newPlurals(locale.OrdinalPlural(loc)),
		},
		Messages: msgs,
	}, nil
}

func newNumberFormat(nf locale.NumberFormat) schema.NumberFormat {
	symbols := nf.Symbols()
	posAffixes := nf.PositiveAffixes()
	negAffixes := nf.NegativeAffixes()
	intGrouping := nf.IntegerGrouping()
	fracGrouping := nf.FractionGrouping()
	return schema.NumberFormat{
		Symbols: schema.Symbols{
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

func newPlurals(p locale.Plural) []schema.Plural {
	var res []schema.Plural
	for _, rules := range p.Rules() {
		plural := schema.Plural{Tag: schema.PluralTag(rules.Tag())}
		rules.Iter(func(r locale.PluralRule) {
			plural.Rules = append(plural.Rules, newPluralRule(r))
		})
	}
	return res
}

func newPluralRule(r locale.PluralRule) schema.PluralRule {
	nranges := r.Ranges.Len()
	ranges := make([]schema.Range, nranges)
	for i := 0; i < nranges; i++ {
		ranges[i] = schema.Range(r.Ranges.At(i))
	}

	mod := 0
	if r.ModuloExp > 0 {
		mod = 10
		for i := 1; i < r.ModuloExp; i++ {
			mod *= 10
		}
	}

	return schema.PluralRule{
		Operand:    schema.Operand(r.Operand),
		Modulo:     mod,
		Negate:     r.Operator == locale.NotEqual,
		Ranges:     ranges,
		Connective: schema.Connective(r.Connective),
	}
}
