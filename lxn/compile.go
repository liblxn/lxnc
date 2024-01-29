package lxn

import (
	"os"

	"github.com/liblxn/lxnc/locale"
)

// CompileMessages parses the given files and returns all messages found in
// these files.
func CompileMessages(filenames ...string) ([]Message, error) {
	p := parser{}
	msgs := make([]Message, 0, 128)
	for _, filename := range filenames {
		bytes, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		m, err := p.Parse(filename, bytes)
		if err != nil {
			return nil, err
		}

		msgs = append(msgs, m...)
	}

	return msgs, nil
}

// CompileCatalog parses the given files and determines the locale information which is need
// for formatting data. It returns the catalog for all the messages in the files.
func CompileCatalog(loc locale.Locale, filenames ...string) (*Catalog, error) {
	messages, err := CompileMessages(filenames...)
	if err != nil {
		return nil, err
	}
	return NewCatalog(loc, messages), nil
}
