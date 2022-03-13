package cldr

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestPluralsDecode(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<plurals type="cardinal">
			<pluralRules locales="bm bo dz id">
				<pluralRule count="other"> @integer 0~15, 100, 1000, 10000, 100000, 1000000, … @decimal 0.0~1.5, 10.0, 100.0, 1000.0, 10000.0, 100000.0, 1000000.0, …</pluralRule>
			</pluralRules>
		</plurals>
		<plurals type="ordinal">
			<pluralRules locales="bm bo dz id">
				<pluralRule count="other"> @integer 0~15, 100, 1000, 10000, 100000, 1000000, … @decimal 0.0~1.5, 10.0, 100.0, 1000.0, 10000.0, 100000.0, 1000000.0, …</pluralRule>
			</pluralRules>
		</plurals>
	</root>
	`

	var plurals Plurals
	err := decodeXML("test", strings.NewReader(xmlData), func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("plurals", plurals.decode)
	})

	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case len(plurals.Cardinal) != 1:
		t.Errorf("unexpected number of cardinal rules: %d", len(plurals.Cardinal))
	case len(plurals.Ordinal) != 1:
		t.Errorf("unexpected number of ordinal rules: %d", len(plurals.Ordinal))
	}
}

func TestPluralRulesDecode(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<plurals type="cardinal">
			<pluralRules locales="bm bo dz id">
				<pluralRule count="other"> @integer 0~15, 100, 1000, 10000, 100000, 1000000, … @decimal 0.0~1.5, 10.0, 100.0, 1000.0, 10000.0, 100000.0, 1000000.0, …</pluralRule>
			</pluralRules>
		</plurals>
	</root>
	`

	var rules PluralRules
	err := decodeXML("test", strings.NewReader(xmlData), func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("plurals", func(d *xmlDecoder, _ xml.StartElement) {
			d.DecodeElem("pluralRules", rules.decode)
		})
	})

	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case !reflect.DeepEqual(rules.Locales, []string{"bm", "bo", "dz", "id"}):
		t.Errorf("unexpected locales: %v", rules.Locales)
	case len(rules.Rules) != 1:
		t.Errorf("unexpected number of rules: %d", len(rules.Rules))
	}
}

func TestIntRangeString(t *testing.T) {
	testcases := []struct {
		str string
		rng IntRange
	}{
		{
			str: "1",
			rng: IntRange{LowerBound: 1, UpperBound: 1},
		},
		{
			str: "1..2",
			rng: IntRange{LowerBound: 1, UpperBound: 2},
		},
	}

	for _, c := range testcases {
		if s := c.rng.String(); s != c.str {
			t.Errorf("unexpected int range for %q: %s", c.str, s)
		}
	}
}

func TestFloatRangeString(t *testing.T) {
	testcases := []struct {
		str string
		rng FloatRange
	}{
		{
			str: "1.00",
			rng: FloatRange{LowerBound: 1, UpperBound: 1, Decimals: 2},
		},
		{
			str: "1.0~2.0",
			rng: FloatRange{LowerBound: 1, UpperBound: 2, Decimals: 1},
		},
	}

	for _, c := range testcases {
		if s := c.rng.String(); s != c.str {
			t.Errorf("unexpected float range for %q: %s", c.str, s)
		}
	}
}

func TestRelationString(t *testing.T) {
	rel := Relation{
		Operand:  AbsoluteValue,
		Modulo:   10,
		Operator: Equal,
		Ranges: []IntRange{
			{LowerBound: 1, UpperBound: 1},
			{LowerBound: 2, UpperBound: 3},
		},
	}

	if s := rel.String(); s != "n % 10 = 1, 2..3" {
		t.Errorf("unexpected relation: %s", s)
	}
}

func TestConjunctionString(t *testing.T) {
	conj := Conjunction{
		{
			Operand:  AbsoluteValue,
			Operator: Equal,
			Ranges: []IntRange{
				{LowerBound: 1, UpperBound: 1},
			},
		},
		{
			Operand:  IntegerDigits,
			Operator: NotEqual,
			Ranges: []IntRange{
				{LowerBound: 1, UpperBound: 1},
			},
		},
	}

	if s := conj.String(); s != "n = 1 and i != 1" {
		t.Errorf("unexpected conjunction: %s", s)
	}
}

func TestPluralSampleString(t *testing.T) {
	sample := PluralSample{
		Ranges: []FloatRange{
			{LowerBound: 1, UpperBound: 2, Decimals: 2},
			{LowerBound: 3, UpperBound: 4, Decimals: 0},
		},
		Infinite: true,
	}

	if s := sample.String(); s != "1.00~2.00, 3~4, …" {
		t.Errorf("unexpected plural sample: %s", s)
	}
}

func TestPluralRuleString(t *testing.T) {
	rule := PluralRule{
		Condition: []Conjunction{
			{
				{
					Operand:  IntegerDigits,
					Operator: Equal,
					Ranges: []IntRange{
						{LowerBound: 1, UpperBound: 1},
					},
				},
			},
			{
				{
					Operand:  AbsoluteValue,
					Operator: NotEqual,
					Ranges: []IntRange{
						{LowerBound: 1, UpperBound: 1},
					},
				},
			},
		},
		IntegerSample: PluralSample{
			Ranges: []FloatRange{
				{LowerBound: 1, UpperBound: 2},
			},
		},
		DecimalSample: PluralSample{
			Ranges: []FloatRange{
				{LowerBound: 1, UpperBound: 2, Decimals: 1},
			},
		},
	}

	if s := rule.String(); s != "i = 1 or n != 1 @integer 1~2 @decimal 1.0~2.0" {
		t.Errorf("unexpected plural rule: %s", s)
	}
}

func TestPluralRuleParser(t *testing.T) {
	testcases := []struct {
		rule     string
		expected PluralRule
	}{
		{
			rule: " @integer 0~15, 100, 1000, 10000, 100000, 1000000, … @decimal 0.0~1.5, 10.0, 100.0, 1000.0, 10000.0, 100000.0, 1000000.0, …",
			expected: PluralRule{
				IntegerSample: PluralSample{
					Ranges: []FloatRange{
						{LowerBound: 0, UpperBound: 15},
						{LowerBound: 100, UpperBound: 100},
						{LowerBound: 1000, UpperBound: 1000},
						{LowerBound: 10000, UpperBound: 10000},
						{LowerBound: 100000, UpperBound: 100000},
						{LowerBound: 1000000, UpperBound: 1000000},
					},
					Infinite: true,
				},
				DecimalSample: PluralSample{
					Ranges: []FloatRange{
						{LowerBound: 0, UpperBound: 1.5, Decimals: 1},
						{LowerBound: 10, UpperBound: 10, Decimals: 1},
						{LowerBound: 100, UpperBound: 100, Decimals: 1},
						{LowerBound: 1000, UpperBound: 1000, Decimals: 1},
						{LowerBound: 10000, UpperBound: 10000, Decimals: 1},
						{LowerBound: 100000, UpperBound: 100000, Decimals: 1},
						{LowerBound: 1000000, UpperBound: 1000000, Decimals: 1},
					},
					Infinite: true,
				},
			},
		},
		{
			rule: "i = 0 or n = 1 @integer 0, 1 @decimal 0.0~1.0, 0.00~0.04",
			expected: PluralRule{
				Condition: []Conjunction{
					{
						{
							Operand:  IntegerDigits,
							Modulo:   0,
							Operator: Equal,
							Ranges: []IntRange{
								{LowerBound: 0, UpperBound: 0},
							},
						},
					},
					{
						{
							Operand:  AbsoluteValue,
							Modulo:   0,
							Operator: Equal,
							Ranges: []IntRange{
								{LowerBound: 1, UpperBound: 1},
							},
						},
					},
				},
				IntegerSample: PluralSample{
					Ranges: []FloatRange{
						{LowerBound: 0, UpperBound: 0},
						{LowerBound: 1, UpperBound: 1},
					},
					Infinite: false,
				},
				DecimalSample: PluralSample{
					Ranges: []FloatRange{
						{LowerBound: 0, UpperBound: 1, Decimals: 1},
						{LowerBound: 0, UpperBound: 0.04, Decimals: 2},
					},
					Infinite: false,
				},
			},
		},
		{
			rule: "i = 0..1 or n != 1 and i % 10 = 1 @integer 0, 1, ... @decimal 0.0~1.0, 0.00~0.04",
			expected: PluralRule{
				Condition: []Conjunction{
					{
						{
							Operand:  IntegerDigits,
							Modulo:   0,
							Operator: Equal,
							Ranges: []IntRange{
								{LowerBound: 0, UpperBound: 1},
							},
						},
					},
					{
						{
							Operand:  AbsoluteValue,
							Modulo:   0,
							Operator: NotEqual,
							Ranges: []IntRange{
								{LowerBound: 1, UpperBound: 1},
							},
						},
						{
							Operand:  IntegerDigits,
							Modulo:   10,
							Operator: Equal,
							Ranges: []IntRange{
								{LowerBound: 1, UpperBound: 1},
							},
						},
					},
				},
				IntegerSample: PluralSample{
					Ranges: []FloatRange{
						{LowerBound: 0, UpperBound: 0},
						{LowerBound: 1, UpperBound: 1},
					},
					Infinite: true,
				},
				DecimalSample: PluralSample{
					Ranges: []FloatRange{
						{LowerBound: 0, UpperBound: 1, Decimals: 1},
						{LowerBound: 0, UpperBound: 0.04, Decimals: 2},
					},
					Infinite: false,
				},
			},
		},
		{
			rule: "i = 0,1 or n != 1 and i % 10 = 1 @integer 0, 1, ... @decimal 0.0~1.0, 0.00~0.04",
			expected: PluralRule{
				Condition: []Conjunction{
					{
						{
							Operand:  IntegerDigits,
							Modulo:   0,
							Operator: Equal,
							Ranges: []IntRange{
								{LowerBound: 0, UpperBound: 1},
							},
						},
					},
					{
						{
							Operand:  AbsoluteValue,
							Modulo:   0,
							Operator: NotEqual,
							Ranges: []IntRange{
								{LowerBound: 1, UpperBound: 1},
							},
						},
						{
							Operand:  IntegerDigits,
							Modulo:   10,
							Operator: Equal,
							Ranges: []IntRange{
								{LowerBound: 1, UpperBound: 1},
							},
						},
					},
				},
				IntegerSample: PluralSample{
					Ranges: []FloatRange{
						{LowerBound: 0, UpperBound: 0},
						{LowerBound: 1, UpperBound: 1},
					},
					Infinite: true,
				},
				DecimalSample: PluralSample{
					Ranges: []FloatRange{
						{LowerBound: 0, UpperBound: 1, Decimals: 1},
						{LowerBound: 0, UpperBound: 0.04, Decimals: 2},
					},
					Infinite: false,
				},
			},
		},
	}

	var p pluralRuleParser
	for _, c := range testcases {
		rule, err := p.parse(c.rule)
		switch {
		case err != nil:
			t.Errorf("unexpected error: %v", err)
		case !reflect.DeepEqual(rule, c.expected):
			fmt.Println(">>>", c.expected.String())
			t.Errorf("unexpected plural rule: %s", rule.String())
		}
	}
}
