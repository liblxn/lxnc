package lxn

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

const (
	bom     = 0xfeff
	runeEOF = -1
)

type tokenType int

const (
	invalid tokenType = iota
	eof
	sectionHeader
	messageKey
	messageText
	messageNewline
	replacementStart
	replacementEnd
	replacementOptionStart
	replacementOptionEnd
	replacementType
)

type token struct {
	typ tokenType
	val string
	pos Pos
}

type tokenizer struct {
	buf    []byte
	off    int
	tokens chan token

	ch  rune
	pos Pos
}

func (t *tokenizer) Scan(filename string, input []byte) <-chan token {
	t.buf = input
	t.off = 0
	t.tokens = make(chan token)
	t.ch = '\n' // initialize pos correctly
	t.pos = Pos{File: filename, Offset: -1}

	t.next()
	if t.ch == bom {
		t.next()
	}

	go func() {
		defer close(t.tokens)
		for {
			t.skipNewlines()

			switch t.ch {
			case runeEOF, utf8.RuneError:
				return
			case '[':
				t.scanSectionHeader()
			case '/':
				t.scanComment()
			default:
				if t.skipNbSpaces() == 0 {
					t.scanMessage()
				} else if t.ch != runeEOF && t.skipNewlines() == 0 {
					t.errorf("unexpected indentation")
					t.nextLine()
					return
				}
			}
		}
	}()
	return t.tokens
}

func (t *tokenizer) scanSectionHeader() {
	t.expect('[', '[')
	t.skipNbSpaces()

	startPos := t.pos
	t.skipIdent()
	for t.ch == '.' {
		t.next()
		t.skipIdent()
	}
	t.emit(sectionHeader, startPos)

	t.skipNbSpaces()
	t.expect(']', ']')
	t.skipNbSpaces()

	if t.ch != runeEOF && t.skipNewlines() == 0 {
		t.errorf("newline expected after section header")
	}
}

func (t *tokenizer) scanComment() {
	t.expect('/', '/')
	for {
		switch t.ch {
		case '\n':
			t.next()
			return
		case bom:
			t.errorf("invalid byte order mark")
			t.nextLine()
			return
		case runeEOF, utf8.RuneError:
			return
		}
		t.next()
	}
}

func (t *tokenizer) scanMessage() {
	t.scanIdent(messageKey)
	t.expect(':')
	t.skipNbSpaces()
	if t.ch != runeEOF && t.skipNewlines() == 0 {
		t.errorf("newline expected after message key")
		t.nextLine()
	}

	if t.skipNbSpaces() != 0 {
		for t.skipNewlines() != 0 {
			if t.skipNbSpaces() == 0 {
				return
			}
		}
		t.scanMessageBlock()
	}
}

func (t *tokenizer) scanMessageBlock() {
	for {
		switch t.ch {
		case utf8.RuneError, runeEOF, '}':
			return
		case '$':
			if t.peek() == '{' {
				t.scanMessageReplacement()
				break
			}
			fallthrough
		default:
			t.scanMessageText()
		}

		startPos := t.pos
		for t.skipNewlines() != 0 {
			if t.skipNbSpaces() == 0 {
				return
			}
		}
		if startPos != t.pos {
			t.emit(messageNewline, startPos)
		}
	}
}

func (t *tokenizer) scanMessageText() {
	startPos := t.pos
	defer func() {
		if startPos != t.pos {
			t.emit(messageText, startPos)
		}
	}()

	for {
		switch t.ch {
		case bom:
			t.errorf("invalid byte order mark")
			t.next()
			return
		case utf8.RuneError, runeEOF, '\n', '}':
			return
		case '$':
			if t.peek() == '{' {
				return
			}
		}
		t.next()
	}
}

func (t *tokenizer) scanMessageReplacement() {
	t.expect('$', '{')

	t.scanIdent(replacementStart)
	if t.ch == ':' {
		t.next() // skip ':'
		t.scanIdent(replacementType)
	}

	t.skipNbSpaces()
	if t.skipNewlines() != 0 && t.skipNbSpaces() == 0 {
		t.errorf("unclosed replacement ('}' expected)")
		return
	}

	for t.ch == '.' {
		t.next() // skip '.'

		startPos := t.pos
		if t.ch == '[' {
			t.next() // skip '['
			t.skipIdent()
			t.expect(']')
		} else {
			t.skipIdent()
		}

		t.emit(replacementOptionStart, startPos)
		if t.ch == '{' {
			t.next()
			t.scanMessageBlock()
			t.expect('}')
		}
		t.emit(replacementOptionEnd, t.pos)

		t.skipNbSpaces()
		if t.skipNewlines() != 0 && t.skipNbSpaces() == 0 {
			t.errorf("unclosed replacement ('}' expected)")
			return
		}
	}

	t.expect('}')
	t.emit(replacementEnd, t.pos)
}

func (t *tokenizer) scanIdent(typ tokenType) {
	startPos := t.pos
	t.skipIdent()
	t.emit(typ, startPos)
}

func (t *tokenizer) skipIdent() {
	for t.ch == '-' || t.ch == '_' || unicode.IsLetter(t.ch) || unicode.IsDigit(t.ch) {
		t.next()
	}
}

func (t *tokenizer) skipNbSpaces() int {
	n := 0
	for t.ch == '\t' || unicode.Is(unicode.Zs, t.ch) {
		t.next()
		n++
	}
	return n
}

func (t *tokenizer) skipNewlines() int {
	n := 0
	for {
		if t.ch != '\n' {
			if t.ch != '\r' || t.peek() != '\n' {
				break
			}
			t.next() // skip '\r'
		}
		t.next()
		n++
	}
	return n
}

func (t *tokenizer) expect(chars ...rune) {
	for _, ch := range chars {
		if t.ch != ch {
			switch t.ch {
			case runeEOF:
				t.errorf("unexpected eof (%q expected)", ch)
			case '\n':
				t.errorf("unexpected newline (%q expected)", ch)
			default:
				t.errorf("unexpected token %q (%q expected)", t.ch, ch)
			}
			return
		}
		t.next()
	}
}

func (t *tokenizer) errorf(msg string, args ...interface{}) {
	if len(args) != 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	t.tokens <- token{
		typ: invalid,
		val: msg,
		pos: t.pos,
	}
}

func (t *tokenizer) emit(typ tokenType, start Pos) {
	t.tokens <- token{
		typ: typ,
		val: string(t.buf[start.Offset:t.pos.Offset]),
		pos: t.pos,
	}
}

func (t *tokenizer) nextLine() {
	for t.ch != runeEOF && t.skipNewlines() == 0 {
		t.next()
	}
}

func (t *tokenizer) next() {
	if t.ch == runeEOF {
		return
	}

	for {
		t.pos.advance(t.ch)
		t.ch = t.peek()
		switch t.ch {
		case 0:
			t.errorf("invalid character nul")
		case utf8.RuneError:
			t.errorf("invalid encoding")
		case runeEOF:
			return
		default:
			t.off += utf8.RuneLen(t.ch)
			return
		}
	}
}

func (t *tokenizer) peek() rune {
	ch, n := utf8.DecodeRune(t.buf[t.off:])
	if ch == utf8.RuneError && n == 0 {
		ch = runeEOF
	}
	return ch
}
