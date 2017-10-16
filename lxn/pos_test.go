package lxn

import "testing"

func TestPosString(t *testing.T) {
	type test struct {
		pos Pos
		str string
	}

	tests := [...]test{
		{
			pos: Pos{Line: 1, Column: 2},
			str: "1:2",
		},
		{
			pos: Pos{File: "file", Line: 1, Column: 2},
			str: "file:1:2",
		},
	}

	for _, test := range tests {
		if s := test.pos.String(); s != test.str {
			t.Errorf("unexpected string for %+v: %s", test.pos, s)
		}
	}
}

func TestPosAdvance(t *testing.T) {
	startPos := Pos{Line: 10, Column: 20, Offset: 30}

	type test struct {
		char     rune
		expected Pos
	}

	tests := [...]test{
		{
			char: 'a',
			expected: Pos{
				Line:   startPos.Line,
				Column: startPos.Column + 1,
				Offset: startPos.Offset + 1,
			},
		},
		{
			char: 'ä',
			expected: Pos{
				Line:   startPos.Line,
				Column: startPos.Column + 1,
				Offset: startPos.Offset + 2,
			},
		},
		{
			char: '\n',
			expected: Pos{
				Line:   startPos.Line + 1,
				Column: 0,
				Offset: startPos.Offset + 1,
			},
		},
	}

	for _, test := range tests {
		pos := startPos
		pos.advance(test.char)
		if pos != test.expected {
			t.Errorf("unexpected position for %q: %+v", test.char, pos)
		}
	}
}
