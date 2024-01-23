package generate_cldr

import (
	"fmt"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*numberFormat)(nil)
	_ generator.TestSnippet = (*numberFormat)(nil)
)

type numberFormat struct {
	decimal *numbersLookupVar
	money   *numbersLookupVar
	percent *numbersLookupVar
	affixes *affixLookupVar
	tag     *tagLookup
}

func newNumberFormat(
	decimal *numbersLookupVar,
	money *numbersLookupVar,
	percent *numbersLookupVar,
	affixes *affixLookupVar,
	tag *tagLookup,
) *numberFormat {
	return &numberFormat{
		decimal: decimal,
		money:   money,
		percent: percent,
		affixes: affixes,
		tag:     tag,
	}
}

func (n *numberFormat) Imports() []string {
	return nil
}

func (n *numberFormat) Generate(p *generator.Printer) {
	patterns := n.decimal.patterns.name
	symbols := n.decimal.symbols.name
	zeros := n.decimal.zeros.name
	affixes := n.affixes.name

	p.Println(`// Grouping holds the sizes for number groups for a specific locale.`)
	p.Println(`type Grouping struct {`)
	p.Println(`	Primary   int`)
	p.Println(`	Secondary int`)
	p.Println(`}`)
	p.Println()
	p.Println(`// Affixes holds a prefix and suffix for locale specific number formatting.`)
	p.Println(`type Affixes struct {`)
	p.Println(`	Prefix string`)
	p.Println(`	Suffix string`)
	p.Println(`}`)
	p.Println()
	p.Println(`// Symbols holds all symbols that are used to format a number in a specific locale.`)
	p.Println(`type Symbols struct {`)
	p.Println(`	Decimal string`)
	p.Println(`	Group   string`)
	p.Println(`	Percent string`)
	p.Println(`	Minus   string`)
	p.Println(`	Inf     string`)
	p.Println(`	NaN     string`)
	p.Println(`	Zero    rune`)
	p.Println(`}`)
	p.Println()
	p.Println(`// NumberFormat holds all relevant information to format a number in a specific locale.`)
	p.Println(`type NumberFormat struct {`)
	p.Println(`	numbers  numbers`)
	p.Println(`	pattern  pattern`)
	p.Println(`	currency bool`)
	p.Println(`}`)
	p.Println()
	p.Println(`// DecimalFormat returns the data for formatting decimal numbers in the given locale.`)
	p.Println(`func DecimalFormat(loc Locale) NumberFormat {`)
	p.Println(`	return lookupNumberFormat(loc, `, n.decimal.name, `, false)`)
	p.Println(`}`)
	p.Println()
	p.Println(`// MoneyFormat returns the data for formatting currency values in the given locale.`)
	p.Println(`func MoneyFormat(loc Locale) NumberFormat {`)
	p.Println(`	return lookupNumberFormat(loc, `, n.money.name, `, true)`)
	p.Println(`}`)
	p.Println()
	p.Println(`// PercentFormat returns the data for formatting percent values in the given locale.`)
	p.Println(`func PercentFormat(loc Locale) NumberFormat {`)
	p.Println(`	return lookupNumberFormat(loc, `, n.percent.name, `, false)`)
	p.Println(`}`)
	p.Println()
	p.Println(`func lookupNumberFormat(loc Locale, lookup numbersLookup, currency bool) NumberFormat {`)
	p.Println(`	if loc == 0 {`)
	p.Println(`		panic("invalid locale")`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	for {`)
	p.Println(`		nums, has := lookup[tagID(loc)]`)
	p.Println(`		switch {`)
	p.Println(`		case has:`)
	p.Println(`			return NumberFormat{`)
	p.Println(`				numbers:  nums,`)
	p.Println(`				pattern:  `, patterns, `.pattern(nums.patternID()),`)
	p.Println(`				currency: currency,`)
	p.Println(`			}`)
	p.Println(`		case loc == root:`)
	p.Println(`			panic("number format not found for " + loc.String())`)
	p.Println(`		}`)
	p.Println(`		loc = loc.parent()`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`// Symbols returns the number symbols for the format.`)
	p.Println(`func (nf NumberFormat) Symbols() Symbols {`)
	p.Println(`	symbols := `, symbols, `.symbols(nf.numbers.symbolsID())`)
	p.Println(`	var decimal, group string`)
	p.Println(`	if nf.currency {`)
	p.Println(`		decimal, group = symbols.currDecimal(), symbols.currGroup()`)
	p.Println(`	} else {`)
	p.Println(`		decimal, group = symbols.decimal(), symbols.group()`)
	p.Println(`	}`)
	p.Println(`	return Symbols{`)
	p.Println(`		Decimal: decimal,`)
	p.Println(`		Group:   group,`)
	p.Println(`		Percent: symbols.percent(),`)
	p.Println(`		Minus:   symbols.minus(),`)
	p.Println(`		Inf:     symbols.inf(),`)
	p.Println(`		NaN:     symbols.nan(),`)
	p.Println(`		Zero:    `, zeros, `.zero(nf.numbers.zeroID()),`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`// PositiveAffixes returns the affixes for positive numbers.`)
	p.Println(`func (nf NumberFormat) PositiveAffixes() Affixes {`)
	p.Println(`	a := `, affixes, `.affix(nf.pattern.posAffixID())`)
	p.Println(`	return Affixes{Prefix: a.prefix(), Suffix: a.suffix()}`)
	p.Println(`}`)
	p.Println()
	p.Println(`// NegativeAffixes returns the affixes for negative numbers.`)
	p.Println(`func (nf NumberFormat) NegativeAffixes() Affixes {`)
	p.Println(`	a := `, affixes, `.affix(nf.pattern.negAffixID())`)
	p.Println(`	return Affixes{Prefix: a.prefix(), Suffix: a.suffix()}`)
	p.Println(`}`)
	p.Println()
	p.Println(`// MinIntegerDigits returns the minimum number of digits which should be displayed for the integer part.`)
	p.Println(`func (nf NumberFormat) MinIntegerDigits() int {`)
	p.Println(`	return nf.pattern.minIntDigits()`)
	p.Println(`}`)
	p.Println()
	p.Println(`// MinFractionDigits returns the minimum number of digits which should be displayed for the fraction part.`)
	p.Println(`func (nf NumberFormat) MinFractionDigits() int {`)
	p.Println(`	return nf.pattern.minFracDigits()`)
	p.Println(`}`)
	p.Println()
	p.Println(`// MaxFractionDigits returns the maximum number of digits which should be displayed for the fraction part.`)
	p.Println(`func (nf NumberFormat) MaxFractionDigits() int {`)
	p.Println(`	return nf.pattern.maxFracDigits()`)
	p.Println(`}`)
	p.Println()
	p.Println(`// IntegerGrouping returns the grouping information for the integer part.`)
	p.Println(`func (nf NumberFormat) IntegerGrouping() Grouping {`)
	p.Println(`	prim, sec := nf.pattern.intGrouping()`)
	p.Println(`	return Grouping{Primary: prim, Secondary: sec}`)
	p.Println(`}`)
	p.Println()
	p.Println(`// FractionGrouping returns the grouping information for the fraction part.`)
	p.Println(`func (nf NumberFormat) FractionGrouping() Grouping {`)
	p.Println(`	prim := nf.pattern.fracGrouping()`)
	p.Println(`	return Grouping{Primary: prim, Secondary: prim}`)
	p.Println(`}`)
}

func (n *numberFormat) TestImports() []string {
	return []string{"reflect"}
}

func (n *numberFormat) GenerateTest(p *generator.Printer) {
	newLocale := func(nums *numbersLookupVar, id cldr.Identity) string {
		return fmt.Sprintf("%#0[2]*[1]x", nums.tags.tagID(id), n.tag.idBits/4)
	}

	newNumbersData := func(data numbersData, money bool) string {
		decimal := data.symb.Decimal
		group := data.symb.Group
		if money {
			decimal = data.symb.CurrencyDecimal
			group = data.symb.CurrencyGroup
		}
		return fmt.Sprintf(`{Symbols{"%s", "%s", "%s", "%s", "%s", "%s", '%c'}, Affixes{"%s", "%s"}, Affixes{"%s", "%s"}, %d, %d, %d, Grouping{%d, %d}, Grouping{%d, %d}}`,
			decimal, group, data.symb.Percent, data.symb.Minus, data.symb.Infinity, data.symb.NaN, data.numsys.Digits[0],
			data.nf.PositivePrefix, data.nf.PositiveSuffix, data.nf.NegativePrefix, data.nf.NegativeSuffix,
			data.nf.MinIntegerDigits, data.nf.MinFractionDigits, data.nf.MaxFractionDigits,
			data.nf.IntegerGrouping.PrimarySize, data.nf.IntegerGrouping.SecondarySize,
			data.nf.FractionGrouping.PrimarySize, data.nf.FractionGrouping.SecondarySize,
		)
	}

	printNumbers := func(nums *numbersLookupVar, money bool) {
		for _, data := range nums.data {
			p.Println(`		`, newLocale(nums, data.id), `: `, newNumbersData(data, money), `,`)
		}
	}

	p.Println(`type numberFormatData struct {`)
	p.Println(`	symbols       Symbols`)
	p.Println(`	posAffixes    Affixes`)
	p.Println(`	negAffixes    Affixes`)
	p.Println(`	minIntDigits  int`)
	p.Println(`	minFracDigits int`)
	p.Println(`	maxFracDigits int`)
	p.Println(`	intGrouping   Grouping`)
	p.Println(`	fracGrouping  Grouping`)
	p.Println(`}`)
	p.Println()
	p.Println(`func TestLookupNumberFormat(t *testing.T) {`)
	p.Println(`	// decimal formats`)
	p.Println(`	testNumberFormatLookup(t, DecimalFormat, map[Locale]numberFormatData{`)
	printNumbers(n.decimal, false)
	p.Println(`	})`)
	p.Println()
	p.Println(`	// money formats`)
	p.Println(`	testNumberFormatLookup(t, MoneyFormat, map[Locale]numberFormatData{`)
	printNumbers(n.money, true)
	p.Println(`	})`)
	p.Println()
	p.Println(`	// percent formats`)
	p.Println(`	testNumberFormatLookup(t, PercentFormat, map[Locale]numberFormatData{`)
	printNumbers(n.percent, false)
	p.Println(`	})`)
	p.Println(`}`)
	p.Println()
	p.Println(`func testNumberFormatLookup(t *testing.T, lookup func(Locale) NumberFormat, expected map[Locale]numberFormatData) {`)
	p.Println(`	for loc, expectedData := range expected {`)
	p.Println(`		nf := lookup(loc)`)
	p.Println(`		data := numberFormatData{`)
	p.Println(`			symbols:       nf.Symbols(),`)
	p.Println(`			posAffixes:    nf.PositiveAffixes(),`)
	p.Println(`			negAffixes:    nf.NegativeAffixes(),`)
	p.Println(`			minIntDigits:  nf.MinIntegerDigits(),`)
	p.Println(`			minFracDigits: nf.MinFractionDigits(),`)
	p.Println(`			maxFracDigits: nf.MaxFractionDigits(),`)
	p.Println(`			intGrouping:   nf.IntegerGrouping(),`)
	p.Println(`			fracGrouping:  nf.FractionGrouping(),`)
	p.Println(`		}`)

	p.Println(`		if !reflect.DeepEqual(data, expectedData) {`)
	p.Println(`			t.Fatalf("unexpected number format for %s", loc.String())`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}
