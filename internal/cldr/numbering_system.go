package cldr

import (
	"encoding/xml"
	"unicode/utf8"
)

// NumberingSystems holds all numbering systems that could be found.
// The key specifies the system id, e.g. "latn".
type NumberingSystems map[string]NumberingSystem // system id => numbering system

func (n *NumberingSystems) decode(d *xmlDecoder, _ xml.StartElement) {
	*n = make(map[string]NumberingSystem)

	d.DecodeElem("numberingSystem", func(d *xmlDecoder, elem xml.StartElement) {
		if xmlAttrib(elem, "type") == "numeric" {
			var sys NumberingSystem
			if err := sys.decode(elem); err != nil {
				d.ReportErr(err, elem)
				return
			}
			(*n)[sys.ID] = sys
		}
		d.SkipElem()
	})
}

// NumberingSystem holds the relevant data for a single numbering system.
type NumberingSystem struct {
	ID     string
	Digits []rune
}

func (s *NumberingSystem) decode(elem xml.StartElement) error {
	s.ID = xmlAttrib(elem, "id")
	s.Digits = make([]rune, 0, 10)

	digits := xmlAttrib(elem, "digits")
	for digits != "" {
		ch, n := utf8.DecodeRuneInString(digits)
		s.Digits = append(s.Digits, ch)
		digits = digits[n:]
	}
	return nil
}
