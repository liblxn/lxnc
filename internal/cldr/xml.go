package cldr

import (
	"encoding/xml"
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/liblxn/lxnc/internal/errors"
)

type decodeFunc func(*xmlDecoder, xml.StartElement)
type decoders map[string]decodeFunc // element name => decode function

func decodeXML(file string, r io.Reader, decode decodeFunc) error {
	d := xmlDecoder{
		file: file,
		d:    xml.NewDecoder(r),
	}

	// find root
	var root xml.StartElement
	for {
		tok := d.token()
		if elem, ok := tok.(xml.StartElement); ok {
			root = elem
			break
		}
	}

	// decode contents
	decode(&d, root)
	if d.err == io.EOF {
		return nil
	}
	return d.err
}

type xmlDecoder struct {
	file string
	d    *xml.Decoder
	err  error
}

func (d *xmlDecoder) DecodeElem(name string, decode decodeFunc) {
	d.DecodeElems(decoders{name: decode})
}

func (d *xmlDecoder) DecodeElems(dec decoders) {
	for {
		switch tok := d.token().(type) {
		case xml.StartElement:
			if fn, has := dec[tok.Name.Local]; has {
				fn(d, tok)
			} else {
				d.SkipElem()
			}

		case xml.EndElement:
			return

		default:
			if tok == nil {
				return
			}
		}
	}
}

func (d *xmlDecoder) SkipElem() {
	depth := 0
	for {
		tok := d.token()
		if tok == nil {
			return
		}

		switch tok.(type) {
		case xml.StartElement:
			depth++

		case xml.EndElement:
			if depth == 0 {
				return
			}
			depth--

		default:
			if tok == nil {
				return
			}
		}
	}
}

func (d *xmlDecoder) ReadString(parent xml.StartElement) string {
	return string(d.readBytes(parent))
}

func (d *xmlDecoder) ReadChar(parent xml.StartElement) rune {
	p := d.readBytes(parent)
	r, n := utf8.DecodeRune(p)
	if r == utf8.RuneError {
		if n != 0 {
			d.ReportErr(errors.New("invalid utf8 encoding"), parent)
		}
		return 0
	}
	if n != len(p) {
		d.ReportErr(errors.New("single character expected"), parent)
	}
	return r
}

func (d *xmlDecoder) ReadInt(parent xml.StartElement) int {
	p := d.readBytes(parent)
	n, err := strconv.ParseInt(string(p), 10, 64)
	if err != nil {
		d.ReportErr(errors.New("integer expected"), parent)
	}
	return int(n)
}

func (d *xmlDecoder) ReportErr(err error, parent xml.StartElement) {
	if d.err != nil && d.err != io.EOF {
		return
	}

	if tag := parent.Name.Local; tag != "" {
		d.err = errors.Newf("%s (<%s>): %v", d.file, tag, err)
	} else {
		d.err = errors.Newf("%s: %v", d.file, err)
	}
}

func (d *xmlDecoder) readBytes(parent xml.StartElement) []byte {
	tok := d.token()
	if cdata, ok := tok.(xml.CharData); ok {
		return cdata
	}
	d.ReportErr(errors.New("character data expected"), parent)
	return nil
}

func (d *xmlDecoder) token() xml.Token {
	if d.err != nil {
		return nil
	}

	tok, err := d.d.Token()
	if err != nil {
		d.ReportErr(err, xml.StartElement{})
		return nil
	}
	return tok
}

func xmlAttrib(elem xml.StartElement, name string) string {
	for _, a := range elem.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}
