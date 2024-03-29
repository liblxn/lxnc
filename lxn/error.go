package lxn

import "fmt"

// Error represents a parsing error with an additional error position.
type Error struct {
	Err error // underlying error
	Pos Pos   // position where the error occured
}

// Error implements the error interface and returns the error message.
func (e Error) Error() string {
	return e.Pos.String() + ": " + e.Err.Error()
}

// ErrorList holds multiple parsing errors.
type ErrorList []error

// Error implements the error interface and returns the error message.
func (e ErrorList) Error() string {
	switch len(e) {
	case 0:
		return "no errors"
	case 1:
		return e[0].Error()
	default:
		return fmt.Sprintf("%v (and %d more errors)", e[0], len(e)-1)
	}
}

func (e ErrorList) err() error {
	if len(e) == 0 {
		return nil
	}
	return e
}

func (e *ErrorList) add(err error, pos Pos) {
	*e = append(*e, Error{
		Err: err,
		Pos: pos,
	})
}

func (e *ErrorList) clear() {
	*e = (*e)[:0]
}
