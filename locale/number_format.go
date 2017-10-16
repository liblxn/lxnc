package locale

// Grouping holds the sizes for number groups for a specific locale.
type Grouping struct {
	Primary   int
	Secondary int
}

// Affixes holds a prefix and suffix for locale specific number formatting.
type Affixes struct {
	Prefix string
	Suffix string
}

// Symbols holds all symbols that are used to format a number in a specific locale.
type Symbols struct {
	Decimal string
	Group   string
	Percent string
	Minus   string
	Inf     string
	NaN     string
	Zero    rune
}

// NumberFormat holds all relevant information to format a number in a specific locale.
type NumberFormat struct {
	numbers  numbers
	pattern  pattern
	currency bool
}

// Symbols returns the number symbols for the format.
func (nf NumberFormat) Symbols() Symbols {
	symbols := numberSymbols.symbols(nf.numbers.symbolsID())
	var decimal, group string
	if nf.currency {
		decimal, group = symbols.currDecimal(), symbols.currGroup()
	} else {
		decimal, group = symbols.decimal(), symbols.group()
	}
	return Symbols{
		Decimal: decimal,
		Group:   group,
		Percent: symbols.percent(),
		Minus:   symbols.minus(),
		Inf:     symbols.inf(),
		NaN:     symbols.nan(),
		Zero:    zeros.zero(nf.numbers.zeroID()),
	}
}

// PositiveAffixes returns the affixes for positive numbers.
func (nf NumberFormat) PositiveAffixes() Affixes {
	a := affixes.affix(nf.pattern.posAffixID())
	return Affixes{Prefix: a.prefix(), Suffix: a.suffix() }
}

// NegativeAffixes returns the affixes for negative numbers.
func (nf NumberFormat) NegativeAffixes() Affixes {
	a := affixes.affix(nf.pattern.negAffixID())
	return Affixes{Prefix: a.prefix(), Suffix: a.suffix() }
}

// MinIntegerDigits returns the minimum number of digits which should be displayed for the integer part.
func (nf NumberFormat) MinIntegerDigits() int {
	return nf.pattern.minIntDigits()
}

// MinFractionDigits returns the minimum number of digits which should be displayed for the fraction part.
func (nf NumberFormat) MinFractionDigits() int {
	return nf.pattern.minFracDigits()
}

// MaxFractionDigits returns the maximum number of digits which should be displayed for the fraction part.
func (nf NumberFormat) MaxFractionDigits() int {
	return nf.pattern.maxFracDigits()
}

// IntegerGrouping returns the grouping information for the integer part.
func (nf NumberFormat) IntegerGrouping() Grouping {
	prim, sec := nf.pattern.intGrouping()
	return Grouping{Primary: prim, Secondary: sec}
}

// FractionGrouping returns the grouping information for the fraction part.
func (nf NumberFormat) FractionGrouping() Grouping {
	prim := nf.pattern.fracGrouping()
	return Grouping{Primary: prim, Secondary: prim}
}

// DecimalFormat returns the data for formatting decimal numbers in the given locale.
func DecimalFormat(loc Locale) NumberFormat {
	return lookupNumberFormat(loc, decimalNumbers, false)
}

// MoneyFormat returns the data for formatting currency values in the given locale.
func MoneyFormat(loc Locale) NumberFormat {
	return lookupNumberFormat(loc, moneyNumbers, true)
}

// PercentFormat returns the data for formatting percent values in the given locale.
func PercentFormat(loc Locale) NumberFormat {
	return lookupNumberFormat(loc, percentNumbers, false)
}

func lookupNumberFormat(loc Locale, lookup numbersLookup, currency bool) NumberFormat {
	if loc == 0 {
		panic("invalid locale")
	}

	for {
		nums, has := lookup[tagID(loc)]
		switch {
		case has:
			return NumberFormat{
				numbers:  nums,
				pattern:  patterns.pattern(nums.patternID()),
				currency: currency,
			}
		case loc == root:
			panic("number format not found for " + loc.String())
		}
		loc = loc.parent()
	}
}
