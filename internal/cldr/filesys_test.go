package cldr

import "strings"

type memoryFileTree map[string]string // path => contents

func (t memoryFileTree) Close() error {
	return nil
}

func (t memoryFileTree) Walk(root string, fn walkFunc) error {
	for path, contents := range t {
		if !strings.HasPrefix(path, root) {
			continue
		}

		err := fn(path, strings.NewReader(contents))
		if err != nil {
			return err
		}
	}
	return nil
}
