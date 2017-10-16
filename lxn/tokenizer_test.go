package lxn

import (
	"fmt"
	"testing"
)

func TestTokenizer(t *testing.T) {
	type test struct {
		input  string
		tokens []token
	}

	tests := [...]test{
		{
			input:  "",
			tokens: []token{},
		},
		{
			input:  "\t\n",
			tokens: []token{},
		},
		{
			input:  "// comment one\n\n//comment two",
			tokens: []token{},
		},
		{
			input:  "// comment one\r\n\r\n//comment two",
			tokens: []token{},
		},
		{
			input: "[[section.header]]",
			tokens: []token{
				newToken(sectionHeader, "section.header"),
			},
		},
		{
			input: "\ufeff[[section.header]]",
			tokens: []token{
				newToken(sectionHeader, "section.header"),
			},
		},
		{
			input: "message-key:",
			tokens: []token{
				newToken(messageKey, "message-key"),
			},
		},
		{
			input: "message-key:\n\tmessage-text",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(messageText, "message-text"),
			},
		},
		{
			input: "message-key:\n\n\t\n\tmessage-text",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(messageText, "message-text"),
			},
		},
		{
			input: "message-key:\n\n\t\n\t$message-text",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(messageText, "$message-text"),
			},
		},
		{
			input: "message-key:\n\n\t\n\tmultiline\n\n\t\n\tmessage-text",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(messageText, "multiline"),
				newToken(messageNewline, "\n\n\t\n\t"),
				newToken(messageText, "message-text"),
			},
		},
		{
			input: "message-key:\n\t${replacement}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementEnd, ""),
			},
		},
		{
			input: "message-key:\n\t${replacement:type}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementEnd, ""),
			},
		},
		{
			input: "message-key:\n\t${replacement:type.opt}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementOptionStart, "opt"),
				newToken(replacementOptionEnd, ""),
				newToken(replacementEnd, ""),
			},
		},
		{
			input: "message-key:\n\t${replacement:type.opt{}}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementOptionStart, "opt"),
				newToken(replacementOptionEnd, ""),
				newToken(replacementEnd, ""),
			},
		},
		{
			input: "message-key:\n\t${replacement:type\n\t.opt{}}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementOptionStart, "opt"),
				newToken(replacementOptionEnd, ""),
				newToken(replacementEnd, ""),
			},
		},
		{
			input: "message-key:\n\t${replacement:type\n\t.[opt]{}}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementOptionStart, "[opt]"),
				newToken(replacementOptionEnd, ""),
				newToken(replacementEnd, ""),
			},
		},
		{
			input: "message-key:\n\tmessage-text ${replacement:type}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(messageText, "message-text "),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementEnd, ""),
			},
		},
		{
			input: "message-key:\n\t${replacement:type} message-text",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementEnd, ""),
				newToken(messageText, " message-text"),
			},
		},
		{
			input: "message-key:\n\tmessage-prefix ${replacement:type} message-suffix",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(messageText, "message-prefix "),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementEnd, ""),
				newToken(messageText, " message-suffix"),
			},
		},
		{
			input: "message-key:\n\t${replacement:type\n\t.opt{inner-text}}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementOptionStart, "opt"),
				newToken(messageText, "inner-text"),
				newToken(replacementOptionEnd, ""),
				newToken(replacementEnd, ""),
			},
		},
		{
			input: "message-key:\n\t${replacement:type\n\t.opt{inner-text ${inner-replacement}}}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementOptionStart, "opt"),
				newToken(messageText, "inner-text "),
				newToken(replacementStart, "inner-replacement"),
				newToken(replacementEnd, ""),
				newToken(replacementOptionEnd, ""),
				newToken(replacementEnd, ""),
			},
		},
		{
			input: "message-key:\n\t${replacement:type\n\t.opt{${inner-replacement} inner-text}}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementOptionStart, "opt"),
				newToken(replacementStart, "inner-replacement"),
				newToken(replacementEnd, ""),
				newToken(messageText, " inner-text"),
				newToken(replacementOptionEnd, ""),
				newToken(replacementEnd, ""),
			},
		},
		{
			input: "message-key:\n\t${replacement:type\n\t.opt{inner-prefix ${inner-replacement} inner-suffix}}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementOptionStart, "opt"),
				newToken(messageText, "inner-prefix "),
				newToken(replacementStart, "inner-replacement"),
				newToken(replacementEnd, ""),
				newToken(messageText, " inner-suffix"),
				newToken(replacementOptionEnd, ""),
				newToken(replacementEnd, ""),
			},
		},
		{
			input: "message-key:\n\t${replacement:type\n\t.opt1{}\n\t.opt2{}}",
			tokens: []token{
				newToken(messageKey, "message-key"),
				newToken(replacementStart, "replacement"),
				newToken(replacementType, "type"),
				newToken(replacementOptionStart, "opt1"),
				newToken(replacementOptionEnd, ""),
				newToken(replacementOptionStart, "opt2"),
				newToken(replacementOptionEnd, ""),
				newToken(replacementEnd, ""),
			},
		},
	}

	for _, test := range tests {
		var tk tokenizer
		tokens := tk.Scan("test", []byte(test.input))

		for _, expected := range test.tokens {
			tok, ok := <-tokens
			switch {
			case !ok:
				t.Errorf("unexpected eof for %q", test.input)
			case tok.typ == invalid:
				t.Errorf("unexpected error for %q: %v", test.input, tok.val)
			case tok.typ != expected.typ:
				t.Errorf("unexpected token type for %q: %s (%s expected)", test.input, tokenTypeName(tok.typ), tokenTypeName(expected.typ))
			case tok.val != expected.val:
				t.Errorf("unexpected token value for %q: %q", test.input, tok.val)
			}
		}
	}
}

func TestTokenizerWithErrors(t *testing.T) {
	type test struct {
		input  string
		errmsg string
	}

	tests := [...]test{
		{
			input:  "  foo",
			errmsg: "unexpected indentation",
		},
		{
			input:  "// \ufeff foobar",
			errmsg: "invalid byte order mark",
		},
		{
			input:  "message-key:\n\tfoo \ufeff bar",
			errmsg: "invalid byte order mark",
		},
		{
			input:  "message-key:\n\t${foo\nmessage-key:",
			errmsg: "unclosed replacement ('}' expected)",
		},
		{
			input:  "message-key:\n\t${foo.opt{}\nmessage-key:",
			errmsg: "unclosed replacement ('}' expected)",
		},
		{
			input:  "[section]]",
			errmsg: "unexpected token 's' ('[' expected)",
		},
		{
			input:  "[[section]",
			errmsg: "unexpected eof (']' expected)",
		},
		{
			input:  "[[section]]  f:",
			errmsg: "newline expected after section header",
		},
		{
			input:  "/ comment",
			errmsg: "unexpected token ' ' ('/' expected)",
		},
		{
			input:  "message-key",
			errmsg: "unexpected eof (':' expected)",
		},
		{
			input:  "message-key: foo",
			errmsg: "newline expected after message key",
		},
		{
			input:  "message-key:\n\t${foo.[opt{}}",
			errmsg: "unexpected token '{' (']' expected)",
		},
		{
			input:  "message-key:\n\t${foo.opt{",
			errmsg: "unexpected eof ('}' expected)",
		},
		{
			input:  "message-key:\n\t${foo.opt{}",
			errmsg: "unexpected eof ('}' expected)",
		},
	}

	for _, test := range tests {
		tk := tokenizer{}
		token := token{typ: eof}
		for tok := range tk.Scan("test", []byte(test.input)) {
			if tok.typ == invalid {
				token = tok
			}
		}

		switch {
		case token.typ != invalid:
			t.Errorf("expected error for %q, got none", test.input)
		case token.val != test.errmsg:
			t.Errorf("unexpected error for %q: %s", test.input, token.val)
		}
	}
}

func newToken(typ tokenType, val string) token {
	return token{typ: typ, val: val}
}

func tokenTypeName(typ tokenType) string {
	switch typ {
	case invalid:
		return "invalid"
	case eof:
		return "eof"
	case sectionHeader:
		return "sectionHeader"
	case messageKey:
		return "messageKey"
	case messageText:
		return "messageText"
	case messageNewline:
		return "messageNewline"
	case replacementStart:
		return "replacementStart"
	case replacementEnd:
		return "replacementEnd"
	case replacementOptionStart:
		return "replacementOptionStart"
	case replacementOptionEnd:
		return "replacementOptionEnd"
	case replacementType:
		return "replacementType"
	default:
		return fmt.Sprintf("tokenType%d", typ)
	}
}
