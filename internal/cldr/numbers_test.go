package cldr

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
)

func TestNumbersDecode(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<numbers>
			<defaultNumberingSystem draft="contributed">latn</defaultNumberingSystem>
			<minimumGroupingDigits draft="contributed">1</minimumGroupingDigits>
			<symbols numberSystem="latn">
				<decimal>,</decimal>
			</symbols>
			<symbols numberSystem="bali">
				<alias source="locale" path="../symbols[@numberSystem='latn']"/>
			</symbols>
			<decimalFormats numberSystem="latn">
				<decimalFormatLength>
					<decimalFormat>
						<pattern>#,##0.###</pattern>
					</decimalFormat>
				</decimalFormatLength>
			</decimalFormats>
			<scientificFormats numberSystem="latn">
				<scientificFormatLength>
					<scientificFormat>
						<pattern>#E0</pattern>
					</scientificFormat>
				</scientificFormatLength>
			</scientificFormats>
			<percentFormats numberSystem="latn">
				<percentFormatLength>
					<percentFormat>
						<pattern>#,##0 %</pattern>
					</percentFormat>
				</percentFormatLength>
			</percentFormats>
			<currencyFormats numberSystem="latn">
				<currencyFormatLength>
					<currencyFormat>
						<pattern>#,##0.00 ¤</pattern>
					</currencyFormat>
				</currencyFormatLength>
			</currencyFormats>
		</numbers>
	</root>
	`

	var numbers Numbers
	err := decodeXML("test", strings.NewReader(xmlData), func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("numbers", numbers.decode)
	})
	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case numbers.DefaultSystem != "latn":
		t.Errorf("unexpected default numbering system: %s", numbers.DefaultSystem)
	case numbers.MinGroupingDigits != 1:
		t.Errorf("unexpected minimum grouping digits: %d", numbers.MinGroupingDigits)
	case len(numbers.Symbols) != 2:
		t.Errorf("unexpected number of symbols: %d", len(numbers.Symbols))
	case numbers.Symbols["bali"].Alias != "latn":
		t.Errorf("unexpected number of symbol alias for 'bali': %s", numbers.Symbols["bali"].Alias)
	case len(numbers.DecimalFormats) != 1:
		t.Errorf("unexpected number of decimal formats: %d", len(numbers.DecimalFormats))
	case len(numbers.ScientificFormats) != 1:
		t.Errorf("unexpected number of scientific formats: %d", len(numbers.ScientificFormats))
	case len(numbers.PercentFormats) != 1:
		t.Errorf("unexpected number of percent formats: %d", len(numbers.PercentFormats))
	case len(numbers.CurrencyFormats) != 1:
		t.Errorf("unexpected number of currency formats: %d", len(numbers.CurrencyFormats))
	}
}

func TestNumberSymbolsMerge(t *testing.T) {
	fieldGetters := []func(*NumberSymbols) *string{
		func(symb *NumberSymbols) *string { return &symb.Decimal },
		func(symb *NumberSymbols) *string { return &symb.Group },
		func(symb *NumberSymbols) *string { return &symb.Percent },
		func(symb *NumberSymbols) *string { return &symb.Permille },
		func(symb *NumberSymbols) *string { return &symb.Plus },
		func(symb *NumberSymbols) *string { return &symb.Minus },
		func(symb *NumberSymbols) *string { return &symb.Exponential },
		func(symb *NumberSymbols) *string { return &symb.SuperscriptExponent },
		func(symb *NumberSymbols) *string { return &symb.Infinity },
		func(symb *NumberSymbols) *string { return &symb.NaN },
		func(symb *NumberSymbols) *string { return &symb.TimeSeparator },
		func(symb *NumberSymbols) *string { return &symb.CurrencyDecimal },
		func(symb *NumberSymbols) *string { return &symb.CurrencyGroup },
	}

	var symbols NumberSymbols
	for i, field := range fieldGetters {
		val := fmt.Sprintf("val%d", i)

		var other NumberSymbols
		*field(&other) = val

		symbols.merge(other)
		if f := *field(&symbols); f != val {
			t.Errorf("unexpected symbol value: %s", f)
		}
	}
}

func TestNumberSymbolsDecode(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<numbers>
			<symbols numberSystem="latn">
				<decimal>,</decimal>
				<group>.</group>
				<list>;</list>
				<percentSign>%</percentSign>
				<plusSign>+</plusSign>
				<minusSign>-</minusSign>
				<exponential>E</exponential>
				<superscriptingExponent>·</superscriptingExponent>
				<perMille>‰</perMille>
				<infinity>∞</infinity>
				<nan>NaN</nan>
				<timeSeparator draft="contributed">:</timeSeparator>
			</symbols>
		</numbers>
	</root>
	`

	expected := NumberSymbols{
		Decimal:             ",",
		Group:               ".",
		Percent:             "%",
		Permille:            "‰",
		Plus:                "+",
		Minus:               "-",
		Exponential:         "E",
		SuperscriptExponent: "·",
		Infinity:            "∞",
		NaN:                 "NaN",
		TimeSeparator:       ":",
		CurrencyDecimal:     ",",
		CurrencyGroup:       ".",
	}

	var symbols NumberSymbols
	err := decodeXML("test", strings.NewReader(xmlData), func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("numbers", func(d *xmlDecoder, _ xml.StartElement) {
			d.DecodeElem("symbols", symbols.decode)
		})
	})
	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case symbols != expected:
		t.Errorf("unexpected number symbols: %#q", symbols)
	}
}

func TestNumberFormatDecode(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<numbers>
			<currencyFormats>
				<currencyFormatLength>
					<currencyFormat type="standard">
						<pattern>#,##0.00 ¤</pattern>
					</currencyFormat>
					<currencyFormat type="accounting">
						<pattern>#,##0.00 ¤</pattern>
					</currencyFormat>
				</currencyFormatLength>
				<currencyFormatLength type="short">
					<currencyFormat type="standard">
						<pattern type="1000" count="one">0 Tsd'.' ¤</pattern>
					</currencyFormat>
				</currencyFormatLength>
			</currencyFormats>
		</numbers>
	</root>
	`

	var nf NumberFormat
	err := decodeXML("test", strings.NewReader(xmlData), func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("numbers", func(d *xmlDecoder, _ xml.StartElement) {
			d.DecodeElem("currencyFormats", nf.decode)
		})
	})
	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case nf.Pattern != "#,##0.00 ¤":
		t.Errorf("unexpected number format pattern: %q", nf.Pattern)
	}
}

func TestNumberFormatParser(t *testing.T) {
	formats := []NumberFormat{
		{
			Pattern:                "#,##0.###",
			PositivePrefix:         "",
			PositiveSuffix:         "",
			NegativePrefix:         "-",
			NegativeSuffix:         "",
			MinIntegerDigits:       1,
			MaxIntegerDigits:       0,
			MinFractionDigits:      0,
			MaxFractionDigits:      3,
			MinExponentDigits:      0,
			PrefixPositiveExponent: false,
			IntegerGrouping: Grouping{
				PrimarySize:   3,
				SecondarySize: 3,
			},
			FractionGrouping: Grouping{},
			Padding:          Padding{},
		},
		{
			Pattern:                "#,##0.###,#",
			PositivePrefix:         "",
			PositiveSuffix:         "",
			NegativePrefix:         "-",
			NegativeSuffix:         "",
			MinIntegerDigits:       1,
			MaxIntegerDigits:       0,
			MinFractionDigits:      0,
			MaxFractionDigits:      4,
			MinExponentDigits:      0,
			PrefixPositiveExponent: false,
			IntegerGrouping: Grouping{
				PrimarySize:   3,
				SecondarySize: 3,
			},
			FractionGrouping: Grouping{
				PrimarySize:   3,
				SecondarySize: 3,
			},
			Padding: Padding{},
		},
		{
			Pattern:                "00000.0000",
			PositivePrefix:         "",
			PositiveSuffix:         "",
			NegativePrefix:         "-",
			NegativeSuffix:         "",
			MinIntegerDigits:       5,
			MaxIntegerDigits:       0,
			MinFractionDigits:      4,
			MaxFractionDigits:      4,
			MinExponentDigits:      0,
			PrefixPositiveExponent: false,
			IntegerGrouping:        Grouping{},
			FractionGrouping:       Grouping{},
			Padding:                Padding{},
		},
		{
			Pattern:                "#E0",
			PositivePrefix:         "",
			PositiveSuffix:         "",
			NegativePrefix:         "-",
			NegativeSuffix:         "",
			MinIntegerDigits:       0,
			MaxIntegerDigits:       1,
			MinFractionDigits:      0,
			MaxFractionDigits:      0,
			MinExponentDigits:      1,
			PrefixPositiveExponent: false,
			IntegerGrouping:        Grouping{},
			FractionGrouping:       Grouping{},
			Padding:                Padding{},
		},
		{
			Pattern:                "#+E0",
			PositivePrefix:         "",
			PositiveSuffix:         "",
			NegativePrefix:         "-",
			NegativeSuffix:         "",
			MinIntegerDigits:       0,
			MaxIntegerDigits:       1,
			MinFractionDigits:      0,
			MaxFractionDigits:      0,
			MinExponentDigits:      1,
			PrefixPositiveExponent: true,
			IntegerGrouping:        Grouping{},
			FractionGrouping:       Grouping{},
			Padding:                Padding{},
		},
		{
			Pattern:                "##0.####E0",
			PositivePrefix:         "",
			PositiveSuffix:         "",
			NegativePrefix:         "-",
			NegativeSuffix:         "",
			MinIntegerDigits:       1,
			MaxIntegerDigits:       3,
			MinFractionDigits:      0,
			MaxFractionDigits:      4,
			MinExponentDigits:      1,
			PrefixPositiveExponent: false,
			IntegerGrouping:        Grouping{},
			FractionGrouping:       Grouping{},
			Padding:                Padding{},
		},
		{
			Pattern:                "¤#,##0.00;(¤#,##0.00)",
			PositivePrefix:         "¤",
			PositiveSuffix:         "",
			NegativePrefix:         "(¤",
			NegativeSuffix:         ")",
			MinIntegerDigits:       1,
			MaxIntegerDigits:       0,
			MinFractionDigits:      2,
			MaxFractionDigits:      2,
			MinExponentDigits:      0,
			PrefixPositiveExponent: false,
			IntegerGrouping: Grouping{
				PrimarySize:   3,
				SecondarySize: 3,
			},
			FractionGrouping: Grouping{},
			Padding:          Padding{},
		},
		{
			Pattern:                "$*x#,##0.00",
			PositivePrefix:         "$",
			PositiveSuffix:         "",
			NegativePrefix:         "-$",
			NegativeSuffix:         "",
			MinIntegerDigits:       1,
			MaxIntegerDigits:       0,
			MinFractionDigits:      2,
			MaxFractionDigits:      2,
			MinExponentDigits:      0,
			PrefixPositiveExponent: false,
			IntegerGrouping: Grouping{
				PrimarySize:   3,
				SecondarySize: 3,
			},
			FractionGrouping: Grouping{},
			Padding: Padding{
				Width: 8,
				Char:  'x',
				Pos:   AfterPrefix,
			},
		},
		{
			Pattern:                "* #0 o''clock",
			PositivePrefix:         "",
			PositiveSuffix:         " o'clock",
			NegativePrefix:         "-",
			NegativeSuffix:         " o'clock",
			MinIntegerDigits:       1,
			MaxIntegerDigits:       0,
			MinFractionDigits:      0,
			MaxFractionDigits:      0,
			MinExponentDigits:      0,
			PrefixPositiveExponent: false,
			IntegerGrouping:        Grouping{},
			FractionGrouping:       Grouping{},
			Padding: Padding{
				Width: 2,
				Char:  ' ',
				Pos:   BeforePrefix,
			},
		},
		{
			Pattern:                "0.##* suffix",
			PositivePrefix:         "",
			PositiveSuffix:         "suffix",
			NegativePrefix:         "-",
			NegativeSuffix:         "suffix",
			MinIntegerDigits:       1,
			MaxIntegerDigits:       0,
			MinFractionDigits:      0,
			MaxFractionDigits:      2,
			MinExponentDigits:      0,
			PrefixPositiveExponent: false,
			IntegerGrouping:        Grouping{},
			FractionGrouping:       Grouping{},
			Padding: Padding{
				Width: 4,
				Char:  ' ',
				Pos:   BeforeSuffix,
			},
		},
		{
			Pattern:                "0.##suffix* ",
			PositivePrefix:         "",
			PositiveSuffix:         "suffix",
			NegativePrefix:         "-",
			NegativeSuffix:         "suffix",
			MinIntegerDigits:       1,
			MaxIntegerDigits:       0,
			MinFractionDigits:      0,
			MaxFractionDigits:      2,
			MinExponentDigits:      0,
			PrefixPositiveExponent: false,
			IntegerGrouping:        Grouping{},
			FractionGrouping:       Grouping{},
			Padding: Padding{
				Width: 4,
				Char:  ' ',
				Pos:   AfterSuffix,
			},
		},
		{
			Pattern:                "* ##,##,#,##0.##",
			PositivePrefix:         "",
			PositiveSuffix:         "",
			NegativePrefix:         "-",
			NegativeSuffix:         "",
			MinIntegerDigits:       1,
			MaxIntegerDigits:       0,
			MinFractionDigits:      0,
			MaxFractionDigits:      2,
			MinExponentDigits:      0,
			PrefixPositiveExponent: false,
			IntegerGrouping: Grouping{
				PrimarySize:   3,
				SecondarySize: 1,
			},
			FractionGrouping: Grouping{},
			Padding: Padding{
				Width: 14,
				Char:  ' ',
				Pos:   BeforePrefix,
			},
		},
		{
			Pattern:                "'*'0.##",
			PositivePrefix:         "*",
			PositiveSuffix:         "",
			NegativePrefix:         "-*",
			NegativeSuffix:         "",
			MinIntegerDigits:       1,
			MaxIntegerDigits:       0,
			MinFractionDigits:      0,
			MaxFractionDigits:      2,
			MinExponentDigits:      0,
			PrefixPositiveExponent: false,
			IntegerGrouping:        Grouping{},
			FractionGrouping:       Grouping{},
			Padding:                Padding{},
		},
	}

	var parser numberFormatParser
	for _, expected := range formats {
		nf, err := parser.parse(expected.Pattern)
		switch {
		case err != nil:
			t.Errorf("unexpected error: %v", err)
		case nf != expected:
			t.Errorf("unexpected parse result for pattern %q", expected.Pattern)
		}
	}
}
