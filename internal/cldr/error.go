package cldr

import "fmt"

type errorString string

func errorf(format string, args ...interface{}) error {
	return errorString(fmt.Sprintf(format, args...))
}

func (e errorString) Error() string {
	return string(e)
}
