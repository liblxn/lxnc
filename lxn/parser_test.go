package lxn

import (
	"reflect"
	"testing"

	schema "github.com/liblxn/lxn/schema/golang"
)

func TestParser(t *testing.T) {
	const input = `
key-one:
	message for key one
key-two:
	message for key two with ${param}

// first section header
[[section-one]]

key-three:
	this is a
	
	multiline text

[[   section.two   ]]
key-four:
	this is a text with multiline ${param:number
	.foo{}
	.bar{}
	}

key-five:
	we have ${param:percent} test coverage

key-six:
	we have to pay ${param:money.currency{curr}}

key-seven:
	this is a cardinal plural: ${count:plural
		.cardinal
		.zero{foo}
		.One{bar}
		.OTHER{foobar}
		.[7]{seven}
	}

key-eight:
	this is an ordinal plural: ${count:plural
		.ordinal
		.TwO{foo}
		.feW{bar}
		.OTHER{foobar}
		.[7]{seven}
	}

key-nine:
	${gender:select
	.default{other}
	.[male]{he}
	.[female]{she}
	.[other]{they}
	}
	created this message

key-ten:
	line 1
	${param1}
	${param2} line 3
	line 4 ${param3}
	line 5.1 ${param4} line 5.2
	`

	expected := []schema.Message{
		{
			Section: "",
			Key:     "key-one",
			Text:    []string{"message for key one"},
		},
		{
			Section: "",
			Key:     "key-two",
			Text:    []string{"message for key two with "},
			Replacements: []schema.Replacement{
				{
					Key:     "param",
					TextPos: 1,
					Type:    schema.StringReplacement,
				},
			},
		},
		{
			Section: "section-one",
			Key:     "key-three",
			Text:    []string{"this is a multiline text"},
		},
		{
			Section: "section.two",
			Key:     "key-four",
			Text:    []string{"this is a text with multiline "},
			Replacements: []schema.Replacement{
				{
					Key:     "param",
					TextPos: 1,
					Type:    schema.NumberReplacement,
				},
			},
		},
		{
			Section: "section.two",
			Key:     "key-five",
			Text:    []string{"we have ", " test coverage"},
			Replacements: []schema.Replacement{
				{
					Key:     "param",
					TextPos: 1,
					Type:    schema.PercentReplacement,
				},
			},
		},
		{
			Section: "section.two",
			Key:     "key-six",
			Text:    []string{"we have to pay "},
			Replacements: []schema.Replacement{
				{
					Key:     "param",
					TextPos: 1,
					Type:    schema.MoneyReplacement,
					Details: schema.ReplacementDetails{schema.MoneyDetails{
						Currency: "curr",
					}},
				},
			},
		},
		{
			Section: "section.two",
			Key:     "key-seven",
			Text:    []string{"this is a cardinal plural: "},
			Replacements: []schema.Replacement{
				{
					Key:     "count",
					TextPos: 1,
					Type:    schema.PluralReplacement,
					Details: schema.ReplacementDetails{schema.PluralDetails{
						Variants: map[schema.PluralTag]schema.Message{
							schema.Zero:  schema.Message{Key: "zero", Text: []string{"foo"}},
							schema.One:   schema.Message{Key: "one", Text: []string{"bar"}},
							schema.Other: schema.Message{Key: "other", Text: []string{"foobar"}},
						},
						Custom: map[int64]schema.Message{
							7: schema.Message{Key: "7", Text: []string{"seven"}},
						},
					}},
				},
			},
		},
		{
			Section: "section.two",
			Key:     "key-eight",
			Text:    []string{"this is an ordinal plural: "},
			Replacements: []schema.Replacement{
				{
					Key:     "count",
					TextPos: 1,
					Type:    schema.PluralReplacement,
					Details: schema.ReplacementDetails{schema.PluralDetails{
						Type: schema.Ordinal,
						Variants: map[schema.PluralTag]schema.Message{
							schema.Two:   schema.Message{Key: "two", Text: []string{"foo"}},
							schema.Few:   schema.Message{Key: "few", Text: []string{"bar"}},
							schema.Other: schema.Message{Key: "other", Text: []string{"foobar"}},
						},
						Custom: map[int64]schema.Message{
							7: schema.Message{Key: "7", Text: []string{"seven"}},
						},
					}},
				},
			},
		},

		{
			Section: "section.two",
			Key:     "key-nine",
			Text:    []string{" created this message"},
			Replacements: []schema.Replacement{
				{
					Key:     "gender",
					TextPos: 0,
					Type:    schema.SelectReplacement,
					Details: schema.ReplacementDetails{schema.SelectDetails{
						Cases: map[string]schema.Message{
							"male":   schema.Message{Key: "male", Text: []string{"he"}},
							"female": schema.Message{Key: "female", Text: []string{"she"}},
							"other":  schema.Message{Key: "other", Text: []string{"they"}},
						},
						Fallback: "other",
					}},
				},
			},
		},
		{
			Section: "section.two",
			Key:     "key-ten",
			Text:    []string{"line 1 ", " ", " line 3 line 4 ", " line 5.1 ", " line 5.2"},
			Replacements: []schema.Replacement{
				{Key: "param1", TextPos: 1, Type: schema.StringReplacement},
				{Key: "param2", TextPos: 2, Type: schema.StringReplacement},
				{Key: "param3", TextPos: 3, Type: schema.StringReplacement},
				{Key: "param4", TextPos: 4, Type: schema.StringReplacement},
			},
		},
	}

	var p parser
	messages, err := p.Parse("test", []byte(input))
	if err != nil {
		t.Fatalf("unexpected parsing error: %v", err)
	}

	if len(messages) != len(expected) {
		t.Errorf("unexpected number of messages: %d", len(messages))
	} else {
		for i := range expected {
			if !reflect.DeepEqual(messages[i], expected[i]) {
				t.Errorf("unexpected message for %s: %#v", messages[i].Key, messages[i])
			}
		}
	}
}

func TestParserWithErrors(t *testing.T) {
	const input = `
key-one:
	${foo:unknowntype}
	${foo:money}
	${foo:money.currency{bar}.currency{baz}}
	${foo:money.currency{bar}.unknown}
	${foo:money.currency{}}
	${foo:money.currency{${bar:string}}}
	${foo:money.currency{bar ${baz:string}}}
	${foo:plural.all{}.[a]{}.other{}}
	${foo:plural.one{}.one{}.other{}}
	${foo:plural.[7]{}.[7]{}.other{}}
	${foo:plural.one{}}
	${foo:plural.cardinal.ordinal.other{}}
	${foo:select.[male]{}.[male]{}}
	${foo:select.default.default}
	${foo:select.default{${bar:string}}}
	${foo:select.default{baz ${bar:string}}}
	${foo:select.default{a}.[b]{some text}}
	`

	expectedErrors := [...]string{
		"invalid replacement type: unknowntype",
		"money option .currency required",
		"money option already defined: .currency",
		"invalid money option: .unknown",
		"empty money option .currency",
		"replacements not allowed for money option .currency",
		"replacements not allowed for money option .currency",
		"invalid plural option: .all",
		"invalid plural option: .[a]",
		"plural option already defined: .one",
		"plural option already defined: .[7]",
		"plural option .other required",
		"multiple plural types defined (cardinal and ordinal)",
		"select option already defined: .[male]",
		"select option already defined: .default",
		"replacements not allowed in select option .default",
		"replacements not allowed in select option .default",
		"default value \"a\" not found in select options",
	}

	var p parser
	_, err := p.Parse("test", []byte(input))
	if err == nil {
		t.Fatal("unexpected parse error, got none")
	}

	errs, ok := err.(ErrorList)
	if !ok {
		t.Fatalf("unexpected error type: %T", err)
	}

	if len(errs) < len(expectedErrors) {
		t.Errorf("unexpected number of errors: %d", len(errs))
	} else {
		for i, e := range errs[:len(expectedErrors)] {
			if err, ok := e.(Error); !ok {
				t.Errorf("unexpected error type: %T", e)
			} else if msg := err.Err.Error(); msg != expectedErrors[i] {
				t.Errorf("unexpected error message: %s", msg)
			}
		}
	}
}
