package cldr

import (
	"encoding/xml"
	"strings"
)

const subtagSep = "-"

// Identity holds the data for an locale identity.
type Identity struct {
	Language  string
	Script    string
	Territory string
	Variant   string
}

// IsRoot returns if the i is the root identity.
func (i Identity) IsRoot() bool {
	return i.Language == "root"
}

// Truncate returns the truncated identity of i.
func (i Identity) Truncate() Identity {
	switch {
	case i.Variant != "":
		i.Variant = ""
	case i.Territory != "":
		i.Territory = ""
	case i.Script != "":
		i.Script = ""
	default:
		i.Language = "root"
	}
	return i
}

// String returns the string representation of the identity.
func (i Identity) String() string {
	str := i.Language
	if i.Script != "" {
		str += subtagSep + i.Script
	}
	if i.Territory != "" {
		str += subtagSep + i.Territory
	}
	if i.Variant != "" {
		str += subtagSep + i.Variant
	}
	return str
}

func (i Identity) empty() bool {
	return i.Language == ""
}

func (i *Identity) decode(d *xmlDecoder, _ xml.StartElement) {
	fieldDecoder := func(field *string) decodeFunc {
		return func(d *xmlDecoder, elem xml.StartElement) {
			*field = xmlAttrib(elem, "type")
			d.SkipElem()
		}
	}

	d.DecodeElems(decoders{
		"language":  fieldDecoder(&i.Language),
		"script":    fieldDecoder(&i.Script),
		"territory": fieldDecoder(&i.Territory),
		"variant":   fieldDecoder(&i.Variant),
	})
}

// ParentIdentities holds the data for the parent relationship of identities.
type ParentIdentities map[string]string // locale => parent locale

func (p *ParentIdentities) decode(d *xmlDecoder, elem xml.StartElement) {
	if xmlAttrib(elem, "component") != "" {
		d.SkipElem()
		return
	}

	parentIdents := make(ParentIdentities)
	d.DecodeElem("parentLocale", func(d *xmlDecoder, elem xml.StartElement) {
		parent := normalizeTag(xmlAttrib(elem, "parent"))
		children := strings.Split(xmlAttrib(elem, "locales"), " ")
		for _, child := range children {
			parentIdents[normalizeTag(child)] = parent
		}
		d.SkipElem()
	})

	*p = parentIdents
}

func normalizeTag(tag string) string {
	return strings.ReplaceAll(tag, "_", subtagSep)
}
