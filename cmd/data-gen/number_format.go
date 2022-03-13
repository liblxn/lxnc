package main

import (
	"fmt"

	"github.com/liblxn/lxnc/internal/cldr"
)

type numberFormat struct {
	decimalNumbers *numbersLookupVar
	moneyNumbers   *numbersLookupVar
	percentNumbers *numbersLookupVar
	affixes        *affixLookupVar
}

func newNumberFormat(decimalNumbers, moneyNumbers, percentNumbers *numbersLookupVar, affixes *affixLookupVar) *numberFormat {
	return &numberFormat{
		decimalNumbers: decimalNumbers,
		moneyNumbers:   moneyNumbers,
		percentNumbers: percentNumbers,
		affixes:        affixes,
	}
}

func (n *numberFormat) imports() []string {
	return nil
}

func (n *numberFormat) generate(p *printer) {
	// There should be no difference which numbers variable we use here, since
	// every numbers variable should hold the same reference to patterns, symbols,
	// and zeros.
	patterns := n.decimalNumbers.patterns.name
	symbols := n.decimalNumbers.symbols.name
	zeros := n.decimalNumbers.zeros.name
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
	p.Println()
	p.Println(`// DecimalFormat returns the data for formatting decimal numbers in the given locale.`)
	p.Println(`func DecimalFormat(loc Locale) NumberFormat {`)
	p.Println(`	return lookupNumberFormat(loc, `, n.decimalNumbers.name, `, false)`)
	p.Println(`}`)
	p.Println()
	p.Println(`// MoneyFormat returns the data for formatting currency values in the given locale.`)
	p.Println(`func MoneyFormat(loc Locale) NumberFormat {`)
	p.Println(`	return lookupNumberFormat(loc, `, n.moneyNumbers.name, `, true)`)
	p.Println(`}`)
	p.Println()
	p.Println(`// PercentFormat returns the data for formatting percent values in the given locale.`)
	p.Println(`func PercentFormat(loc Locale) NumberFormat {`)
	p.Println(`	return lookupNumberFormat(loc, `, n.percentNumbers.name, `, false)`)
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
}

func (n *numberFormat) testImports() []string {
	return []string{"reflect"}
}

func (n *numberFormat) generateTest(p *printer) {
	newLocale := func(nums *numbersLookupVar, id cldr.Identity) string {
		return fmt.Sprintf("%#0[2]*[1]x", nums.tags.tagID(id), tagIDBits/4)
	}

	newNumberFormatData := func(nums *numbersLookupVar, nf cldr.NumberFormat, symb cldr.NumberSymbols, numsys cldr.NumberingSystem, money bool) string {
		decimal := symb.Decimal
		group := symb.Group
		if money {
			decimal = symb.CurrencyDecimal
			group = symb.CurrencyGroup
		}
		return fmt.Sprintf(`{Symbols{"%s", "%s", "%s", "%s", "%s", "%s", '%c'}, Affixes{"%s", "%s"}, Affixes{"%s", "%s"}, %d, %d, %d, Grouping{%d, %d}, Grouping{%d, %d}}`,
			decimal, group, symb.Percent, symb.Minus, symb.Infinity, symb.NaN, numsys.Digits[0],
			nf.PositivePrefix, nf.PositiveSuffix, nf.NegativePrefix, nf.NegativeSuffix,
			nf.MinIntegerDigits, nf.MinFractionDigits, nf.MaxFractionDigits,
			nf.IntegerGrouping.PrimarySize, nf.IntegerGrouping.SecondarySize,
			nf.FractionGrouping.PrimarySize, nf.FractionGrouping.SecondarySize,
		)
	}

	printNumbers := func(nums *numbersLookupVar, money bool) {
		nums.iterateNumbers(func(id cldr.Identity, nf cldr.NumberFormat, symb cldr.NumberSymbols, numsys cldr.NumberingSystem) {
			p.Println(`		`, newLocale(nums, id), `: `, newNumberFormatData(nums, nf, symb, numsys, money), `,`)
		})
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
	printNumbers(n.decimalNumbers, false)
	p.Println(`	})`)
	p.Println()
	p.Println(`	// money formats`)
	p.Println(`	testNumberFormatLookup(t, MoneyFormat, map[Locale]numberFormatData{`)
	printNumbers(n.moneyNumbers, true)
	p.Println(`	})`)
	p.Println()
	p.Println(`	// percent formats`)
	p.Println(`	testNumberFormatLookup(t, PercentFormat, map[Locale]numberFormatData{`)
	printNumbers(n.percentNumbers, false)
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
