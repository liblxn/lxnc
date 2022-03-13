package lxn

import (
	"io/ioutil"

	schema "github.com/liblxn/lxn/schema/golang"

	"github.com/liblxn/lxnc/internal/locale"
)

type input struct {
	filename string
	bytes    []byte
}

// Compile parses the given input and determines the locale information which is need
// for formatting data.
func Compile(loc locale.Locale, data ...[]byte) (schema.Catalog, error) {
	inputs := make([]input, 0, len(data))
	for _, bytes := range data {
		inputs = append(inputs, input{bytes: bytes})
	}
	return compile(loc, inputs)
}

// CompileFile parses the given file and determines the locale information which is need
// for formatting data.
func CompileFile(loc locale.Locale, filenames ...string) (schema.Catalog, error) {
	inputs := make([]input, 0, len(filenames))
	for _, filename := range filenames {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			return schema.Catalog{}, err
		}
		inputs = append(inputs, input{filename: filename, bytes: bytes})
	}
	return compile(loc, inputs)
}

func compile(loc locale.Locale, inputs []input) (schema.Catalog, error) {
	var (
		p   parser
		msg []schema.Message
	)
	for _, input := range inputs {
		m, err := p.Parse(input.filename, input.bytes)
		if err != nil {
			return schema.Catalog{}, err
		}
		msg = append(msg, m...)
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
		Messages: msg,
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
	for i := 0; i < len(p) && p[i].Tag != locale.Other; i++ {
		rules := schema.Plural{Tag: schema.PluralTag(p[i].Tag)}
		p[i].Iter(func(r locale.PluralRule) bool {
			rules.Rules = append(rules.Rules, newPluralRule(r))
			return true
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
