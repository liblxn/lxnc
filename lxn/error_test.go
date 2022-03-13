package lxn

import (
	"testing"

	"github.com/liblxn/lxnc/internal/errors"
)

func TestErrorString(t *testing.T) {
	err := errors.New("foobar")
	if msg := err.Error(); msg != "foobar" {
		t.Errorf("unexpected error message: %q", msg)
	}
}

func TestError(t *testing.T) {
	err := &Error{
		Err: errors.New("foobar"),
		Pos: Pos{
			File:   "file",
			Offset: 12,
			Line:   2,
			Column: 4,
		},
	}

	if msg := err.Error(); msg != "file:2:4: foobar" {
		t.Errorf("unexpected error message: %q", msg)
	}
}

func TestErrorList(t *testing.T) {
	var errs ErrorList

	if err := errs.err(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if msg := errs.Error(); msg != "no errors" {
		t.Errorf("unexpected error message: %q", msg)
	}

	errs.add(errors.New("foo"), Pos{Offset: 12, Line: 2, Column: 4})
	if err := errs.err(); err == nil {
		t.Error("expected error, got none")
	}
	if msg := errs.Error(); msg != "2:4: foo" {
		t.Errorf("unexpected error message: %q", msg)
	}

	errs.add(errors.New("bar"), Pos{Offset: 16, Line: 3, Column: 1})
	if err := errs.err(); err == nil {
		t.Error("expected error, got none")
	}
	if msg := errs.Error(); msg != "2:4: foo (and 1 more errors)" {
		t.Errorf("unexpected error message: %q", msg)
	}

	errs.clear()
	if len(errs) != 0 {
		t.Errorf("unexpected number of errors: %d", len(errs))
	}
}
