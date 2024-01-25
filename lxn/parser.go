package lxn

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/liblxn/lxnc/internal/errors"
)

type parser struct {
	t      tokenizer
	tokens <-chan token
	tok    token
	errs   ErrorList
}

func (p *parser) Parse(filename string, input []byte) ([]Message, error) {
	p.tokens = p.t.Scan(filename, input)
	p.errs.clear()
	p.next() // scan initial token

	var m []Message
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

func (p *parser) parseMessage(section string) (msg Message) {
	msg.Section = section
	msg.Key = p.tok.val
	p.next()
	return p.parseMessageFragments(msg)
}

func (p *parser) parseMessageFragments(msg Message) Message {
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

func (p *parser) parseReplacement(textPos int) (repl Replacement) {
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
		repl.Type = StringReplacement
		repl.Details = p.parseStringDetails()
	case "number":
		repl.Type = NumberReplacement
		repl.Details = p.parseNumberDetails()
	case "percent":
		repl.Type = PercentReplacement
		repl.Details = p.parsePercentDetails()
	case "money":
		repl.Type = MoneyReplacement
		repl.Details = p.parseMoneyDetails()
	case "plural":
		repl.Type = PluralReplacement
		repl.Details = p.parsePluralDetails()
	case "select":
		repl.Type = SelectReplacement
		repl.Details = p.parseSelectDetails()
	default:
		p.errorf("invalid replacement type: %s", typ)
		p.skipReplacementOptions()
	}

	p.expect(replacementEnd)
	return repl
}

func (p *parser) parseStringDetails() ReplacementDetails {
	p.skipReplacementOptions()
	return ReplacementDetails{}
}

func (p *parser) parseNumberDetails() ReplacementDetails {
	p.skipReplacementOptions()
	return ReplacementDetails{}
}

func (p *parser) parsePercentDetails() ReplacementDetails {
	p.skipReplacementOptions()
	return ReplacementDetails{}
}

func (p *parser) parseMoneyDetails() ReplacementDetails {
	details := MoneyDetails{}

	hasCurrency := false
	for p.tok.typ == replacementOptionStart {
		option := p.tok.val
		p.next()

		msg := p.parseMessageFragments(Message{})
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
	return ReplacementDetails{details}
}

func (p *parser) parsePluralDetails() ReplacementDetails {
	details := PluralDetails{
		Type:     Cardinal,
		Variants: make(map[PluralTag]Message),
		Custom:   make(map[int64]Message),
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
			details.Custom[n] = p.parseMessageFragments(Message{Key: option})
		} else {
			option = strings.ToLower(option)
			tag := PluralTag(-1)
			switch option {
			case "cardinal", "ordinal":
				if typ != "" && typ != option {
					p.errorf("multiple plural types defined (%s and %s)", typ, option)
				} else if option == "ordinal" {
					details.Type = Ordinal
				} else {
					details.Type = Cardinal
				}
				typ = option
			case "zero":
				tag = Zero
			case "one":
				tag = One
			case "two":
				tag = Two
			case "few":
				tag = Few
			case "many":
				tag = Many
			case "other":
				tag = Other
			default:
				p.errorf("invalid plural option: .%s", option)
			}

			if tag < 0 {
				p.parseMessageFragments(Message{})
			} else {
				if details.Variants[tag].Key != "" {
					p.errorf("plural option already defined: .%s", option)
				}
				details.Variants[tag] = p.parseMessageFragments(Message{Key: option})
			}
		}

		p.expect(replacementOptionEnd)
	}

	if details.Variants[Other].Key == "" {
		p.errorf("plural option .other required")
	}
	return ReplacementDetails{details}
}

func (p *parser) parseSelectDetails() ReplacementDetails {
	details := SelectDetails{
		Cases: make(map[string]Message),
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
			details.Cases[option] = p.parseMessageFragments(Message{Key: option})
		} else {
			if _, has := opts[option]; has {
				p.errorf("select option already defined: .%s", option)
			}
			opts[option] = struct{}{}

			optval := ""
			msg := p.parseMessageFragments(Message{})
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
	return ReplacementDetails{details}
}

func (p *parser) skipReplacementOptions() {
	for p.tok.typ == replacementOptionStart {
		p.next() // skip option name
		p.parseMessageFragments(Message{})
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
	p.errs.add(errors.Newf(format, args...), p.tok.pos)
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
