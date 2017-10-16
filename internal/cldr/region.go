package cldr

import (
	"encoding/xml"
	"strings"
	"unicode"
)

// Regions represents a collection of defined regions. Each region can either
// be a sub-region or a territory.
type Regions map[string]Region // region code => region

// Territories returns all territories of the region with the given code. It
// resolves the regions recursively. If the given region code does not exist,
// it is assumed to be a territory.
func (r Regions) Territories(regionCode string) []string {
	return r.collectTerritories(regionCode, nil)
}

func (r Regions) collectTerritories(code string, res []string) []string {
	region := r[code]
	if len(region) == 0 {
		return append(res, code)
	}

	for _, subcode := range region {
		res = r.collectTerritories(subcode, res)
	}
	return res
}

func (r *Regions) decode(d *xmlDecoder, _ xml.StartElement) {
	*r = make(Regions)

	d.DecodeElem("group", func(d *xmlDecoder, elem xml.StartElement) {
		if xmlAttrib(elem, "status") == "" {
			code := xmlAttrib(elem, "type")
			region := parseRegion(xmlAttrib(elem, "contains"))
			if code != "" && len(region) != 0 {
				(*r)[code] = region
			}
		}
		d.SkipElem()
	})
}

// Region represents a list of the sub-regions or territories it contains.
type Region []string // list of sub-region codes

func parseRegion(containment string) Region {
	var region Region
	for containment != "" {
		idx := strings.IndexFunc(containment, unicode.IsSpace)
		if idx < 0 {
			idx = len(containment)
		}
		region = append(region, containment[:idx])

		containment = containment[idx:]
		idx = strings.IndexFunc(containment, func(ch rune) bool { return !unicode.IsSpace(ch) })
		if idx < 0 {
			break
		}
		containment = containment[idx:]
	}
	return region
}
