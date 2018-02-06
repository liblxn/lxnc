package lxn

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	schema "github.com/liblxn/lxn/schema/golang"
)

type parser struct {
	t      tokenizer
	tokens <-chan token
	tok    token
	errs   ErrorList
}

func (p *parser) Parse(filename string, input []byte) ([]schema.Message, error) {
	p.tokens = p.t.Scan(filename, input)
	p.errs.clear()
	p.next() // scan initial token

	var m []schema.Message
	section := ""
	for {
		switch p.tok.typ {
		case eof:
			return m, p.errs.err()
		case invalid:
			p.next()
		case sectionHeader:
			section = p.tok.val
			p.next()
		case messageKey:
			m = append(m, p.parseMessage(section))
		default:
			p.errorf("unexpected token %q", p.tok.val)
			p.next()
		}
	}
}

func (p *parser) parseMessage(section string) (msg schema.Message) {
	msg.Section = section
	msg.Key = p.tok.val
	p.next()
	return p.parseMessageFragments(msg)
}

func (p *parser) parseMessageFragments(msg schema.Message) schema.Message {
	defer func() {
		// trim trailing whitespaces
		ntext := len(msg.Text)
		nrepl := len(msg.Replacements)
		if ntext != 0 && (nrepl == 0 || msg.Replacements[nrepl-1].TextPos < ntext) {
			trimmed := strings.TrimRightFunc(msg.Text[ntext-1], unicode.IsSpace)
			if trimmed == "" {
				msg.Text = msg.Text[:ntext-1]
			} else {
				msg.Text[ntext-1] = trimmed
			}
		}
	}()

	appendText := func(text string) {
		ntext := len(msg.Text)
		nrepl := len(msg.Replacements)
		if ntext != 0 && (nrepl == 0 || msg.Replacements[nrepl-1].TextPos < ntext) {
			msg.Text[len(msg.Text)-1] += text
		} else {
			msg.Text = append(msg.Text, text)
		}
	}

	for {
		switch p.tok.typ {
		case messageNewline:
			if len(msg.Text) != 0 || len(msg.Replacements) != 0 {
				appendText(" ")
			}
			p.next()

		case messageText:
			appendText(p.parseText())

		case replacementStart:
			repl := p.parseReplacement(len(msg.Text))
			msg.Replacements = append(msg.Replacements, repl)

		default:
			return msg
		}
	}

}

func (p *parser) parseText() string {
	txt := p.tok.val
	p.next()
	return txt
}

func (p *parser) parseReplacement(textPos int) (repl schema.Replacement) {
	repl.TextPos = textPos
	repl.Key = p.tok.val
	p.next()

	typ := "string"
	if p.tok.typ == replacementType {
		typ = p.tok.val
		p.next()
	}

	switch strings.ToLower(typ) {
	case "string":
		repl.Type = schema.StringReplacement
		repl.Details = p.parseStringDetails()
	case "number":
		repl.Type = schema.NumberReplacement
		repl.Details = p.parseNumberDetails()
	case "percent":
		repl.Type = schema.PercentReplacement
		repl.Details = p.parsePercentDetails()
	case "money":
		repl.Type = schema.MoneyReplacement
		repl.Details = p.parseMoneyDetails()
	case "plural":
		repl.Type = schema.PluralReplacement
		repl.Details = p.parsePluralDetails()
	case "select":
		repl.Type = schema.SelectReplacement
		repl.Details = p.parseSelectDetails()
	default:
		p.errorf("invalid replacement type: %s", typ)
		p.skipReplacementOptions()
	}

	p.expect(replacementEnd)
	return repl
}

func (p *parser) parseStringDetails() schema.ReplacementDetails {
	p.skipReplacementOptions()
	return schema.ReplacementDetails{}
}

func (p *parser) parseNumberDetails() schema.ReplacementDetails {
	p.skipReplacementOptions()
	return schema.ReplacementDetails{}
}

func (p *parser) parsePercentDetails() schema.ReplacementDetails {
	p.skipReplacementOptions()
	return schema.ReplacementDetails{}
}

func (p *parser) parseMoneyDetails() schema.ReplacementDetails {
	details := schema.MoneyDetails{}

	hasCurrency := false
	for p.tok.typ == replacementOptionStart {
		option := p.tok.val
		p.next()

		msg := p.parseMessageFragments(schema.Message{})
		switch strings.ToLower(option) {
		case "currency":
			switch {
			case hasCurrency:
				p.errorf("money option already defined: .currency")
			case len(msg.Replacements) != 0:
				p.errorf("replacements not allowed for money option .currency")
			case len(msg.Text) == 0:
				p.errorf("empty money option .currency")
			default:
				details.Currency = msg.Text[0]
			}
			hasCurrency = true

		default:
			p.errorf("invalid money option: .%s", option)
		}

		p.expect(replacementOptionEnd)
	}

	if !hasCurrency {
		p.errorf("money option .currency required")
	}
	return schema.ReplacementDetails{details}
}

func (p *parser) parsePluralDetails() schema.ReplacementDetails {
	details := schema.PluralDetails{
		Type:     schema.Cardinal,
		Variants: make(map[schema.PluralTag]schema.Message),
		Custom:   make(map[int64]schema.Message),
	}

	typ := ""
	for p.tok.typ == replacementOptionStart {
		option := p.tok.val
		p.next()

		if option[0] == '[' {
			option = option[1 : len(option)-1] // trim '[' and ']'
			n, err := strconv.ParseInt(option, 10, 64)
			if err != nil {
				p.errorf("invalid plural option: .[%s]", option)
			} else if _, has := details.Custom[n]; has {
				p.errorf("plural option already defined: .[%s]", option)
			}
			details.Custom[n] = p.parseMessageFragments(schema.Message{Key: option})
		} else {
			option = strings.ToLower(option)
			tag := schema.PluralTag(-1)
			switch option {
			case "cardinal", "ordinal":
				if typ != "" && typ != option {
					p.errorf("multiple plural types defined (%s and %s)", typ, option)
				} else if option == "ordinal" {
					details.Type = schema.Ordinal
				} else {
					details.Type = schema.Cardinal
				}
				typ = option
			case "zero":
				tag = schema.Zero
			case "one":
				tag = schema.One
			case "two":
				tag = schema.Two
			case "few":
				tag = schema.Few
			case "many":
				tag = schema.Many
			case "other":
				tag = schema.Other
			default:
				p.errorf("invalid plural option: .%s", option)
			}

			if tag < 0 {
				p.parseMessageFragments(schema.Message{})
			} else {
				if details.Variants[tag].Key != "" {
					p.errorf("plural option already defined: .%s", option)
				}
				details.Variants[tag] = p.parseMessageFragments(schema.Message{Key: option})
			}
		}

		p.expect(replacementOptionEnd)
	}

	if details.Variants[schema.Other].Key == "" {
		p.errorf("plural option .other required")
	}
	return schema.ReplacementDetails{details}
}

func (p *parser) parseSelectDetails() schema.ReplacementDetails {
	details := schema.SelectDetails{
		Cases: make(map[string]schema.Message),
	}

	opts := make(map[string]struct{})
	for p.tok.typ == replacementOptionStart {
		option := p.tok.val
		p.next()

		if option[0] == '[' {
			option = option[1 : len(option)-1] // trim '[' and ']'
			if _, has := details.Cases[option]; has {
				p.errorf("select option already defined: .[%s]", option)
			}
			details.Cases[option] = p.parseMessageFragments(schema.Message{Key: option})
		} else {
			if _, has := opts[option]; has {
				p.errorf("select option already defined: .%s", option)
			}
			opts[option] = struct{}{}

			optval := ""
			msg := p.parseMessageFragments(schema.Message{})
			switch {
			case len(msg.Replacements) != 0:
				p.errorf("replacements not allowed in select option .%s", option)
			case len(msg.Text) != 0:
				optval = msg.Text[0]
			}

			if strings.ToLower(option) == "default" {
				details.Fallback = optval
			}
		}

		p.expect(replacementOptionEnd)
	}

	if _, hasFallback := details.Cases[details.Fallback]; details.Fallback != "" && !hasFallback {
		p.errorf("default value %q not found in select options", details.Fallback)
	}
	return schema.ReplacementDetails{details}
}

func (p *parser) skipReplacementOptions() {
	for p.tok.typ == replacementOptionStart {
		p.next() // skip option name
		p.parseMessageFragments(schema.Message{})
		p.expect(replacementOptionEnd)
	}
}

func (p *parser) expect(typ tokenType) {
	if p.tok.typ != typ {
		p.errorf("unexpected token %q", p.tok.val)
	}
	p.next()
}

func (p *parser) errorf(format string, args ...interface{}) {
	err := errorString(fmt.Sprintf(format, args...))
	p.errs.add(err, p.tok.pos)
}

func (p *parser) next() {
	var ok bool
	p.tok, ok = <-p.t.tokens

	switch {
	case !ok:
		p.tok.typ = eof
	case p.tok.typ == invalid:
		p.errorf(p.tok.val)
	}
}
