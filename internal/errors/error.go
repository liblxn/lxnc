package errors

import "fmt"

type errString string

func New(msg string) error {
	return errString(msg)
}

func Newf(format string, args ...interface{}) error {
	return New(fmt.Sprintf(format, args...))
}

func (e errString) Error() string {
	return string(e)
}
