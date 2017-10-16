package locale

import (
	"fmt"
)

type errorString string

func errorf(msg string, args ...interface{}) error {
	return errorString(fmt.Sprintf(msg, args...))
}

func (e errorString) Error() string {
	return string(e)
}
