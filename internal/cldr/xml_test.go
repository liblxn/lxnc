package cldr

import (
	"encoding/xml"
	"reflect"
	"strings"
	"testing"
)

func TestXMLDecoderDecodeElems(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<foo>
			<bar1 />
			<bar2 />
			<bar3 />
		</foo>
	</root>
	`

	var elems []string
	pushBack := func(d *xmlDecoder, elem xml.StartElement) {
		elems = append(elems, elem.Name.Local)
		d.SkipElem()
	}

	decode := func(d *xmlDecoder, elem xml.StartElement) {
		if elem.Name.Local != "root" {
			t.Fatalf("unexpected root element: <%s>", elem.Name.Local)
		}

		d.DecodeElem("foo", func(d *xmlDecoder, elem xml.StartElement) {
			if elem.Name.Local != "foo" {
				t.Fatalf("unexpected element: <%s>", elem.Name.Local)
			}

			d.DecodeElems(decoders{
				"bar1": pushBack,
				"bar3": pushBack,
			})
		})
	}

	err := decodeXML("test", strings.NewReader(xmlData), decode)
	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case !reflect.DeepEqual(elems, []string{"bar1", "bar3"}):
		t.Errorf("unexpected elements: %v", elems)
	}
}

func TestXMLDecoderSkipElem(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<foo>
			<bar />
			<bar />
		</foo>
		<foo>
		</foo>
	</root>
	`

	skips := 0
	decode := func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("foo", func(d *xmlDecoder, _ xml.StartElement) {
			d.SkipElem()
			skips++
		})
	}

	err := decodeXML("test", strings.NewReader(xmlData), decode)
	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case skips != 2:
		t.Errorf("unexpected number of skips: %d", skips)
	}
}

func TestXMLDecoderReadString(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<foo>bar</foo>
		<foo> </foo>
	</root>
	`

	str := make([]string, 0, 2)
	decode := func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("foo", func(d *xmlDecoder, elem xml.StartElement) {
			str = append(str, d.ReadString(elem))
			d.SkipElem()
		})
	}

	err := decodeXML("test", strings.NewReader(xmlData), decode)
	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case !reflect.DeepEqual(str, []string{"bar", " "}):
		t.Errorf("unexpected strings: %v", str)
	}

	xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<foo></foo>
	</root>
	`

	str = str[:0]
	err = decodeXML("test", strings.NewReader(xmlData), decode)
	switch {
	case err == nil:
		t.Error("expected error, got none")
	case !reflect.DeepEqual(str, []string{""}):
		t.Errorf("unexpected strings: %v", str)
	}
}

func TestXMLDecoderReadChar(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<foo>b</foo>
		<foo> </foo>
	</root>
	`

	chars := make([]rune, 0, 2)
	decode := func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("foo", func(d *xmlDecoder, elem xml.StartElement) {
			chars = append(chars, d.ReadChar(elem))
			d.SkipElem()
		})
	}

	err := decodeXML("test", strings.NewReader(xmlData), decode)
	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case !reflect.DeepEqual(chars, []rune{'b', ' '}):
		t.Errorf("unexpected chars: %q", chars)
	}

	xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<foo>bar</foo>
	</root>
	`

	chars = chars[:0]
	err = decodeXML("test", strings.NewReader(xmlData), decode)
	switch {
	case err == nil:
		t.Error("expected error, got none")
	case !reflect.DeepEqual(chars, []rune{'b'}):
		t.Errorf("unexpected strings: %q", chars)
	}
}

func TestXMLDecoderReadInt(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<foo>-7</foo>
		<foo>7</foo>
	</root>
	`

	ints := make([]int, 0, 2)
	decode := func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("foo", func(d *xmlDecoder, elem xml.StartElement) {
			ints = append(ints, d.ReadInt(elem))
			d.SkipElem()
		})
	}

	err := decodeXML("test", strings.NewReader(xmlData), decode)
	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case !reflect.DeepEqual(ints, []int{-7, 7}):
		t.Errorf("unexpected ints: %v", ints)
	}

	xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<foo>bar</foo>
	</root>
	`

	ints = ints[:0]
	err = decodeXML("test", strings.NewReader(xmlData), decode)
	switch {
	case err == nil:
		t.Error("expected error, got none")
	case !reflect.DeepEqual(ints, []int{0}):
		t.Errorf("unexpected ints: %v", ints)
	}
}

func TestXMLAttrib(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<foo bar="bar" baz="baz" />
	</root>
	`

	decode := func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("foo", func(d *xmlDecoder, elem xml.StartElement) {
			if bar := xmlAttrib(elem, "bar"); bar != "bar" {
				t.Errorf("unexpected attribute value for 'bar': %s", bar)
			}
			if baz := xmlAttrib(elem, "baz"); baz != "baz" {
				t.Errorf("unexpected attribute value for 'baz': %s", baz)
			}
			if foobar := xmlAttrib(elem, "foobar"); foobar != "" {
				t.Errorf("unexpected attribute value for 'foobar': %s", foobar)
			}
			d.SkipElem()
		})
	}

	err := decodeXML("test", strings.NewReader(xmlData), decode)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
