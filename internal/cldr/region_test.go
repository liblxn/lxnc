package cldr

import (
	"encoding/xml"
	"reflect"
	"strings"
	"testing"
)

func TestRegionsTerritories(t *testing.T) {
	regions := Regions{
		"root": Region{
			"region-1",
			"region-2",
		},
		"region-1": Region{
			"region-1-1",
			"region-1-2",
		},
		"region-2": Region{
			"region-2-1",
			"region-2-2",
		},
	}

	testcases := []struct {
		code     string
		expected []string
	}{
		{
			code:     "root",
			expected: []string{"region-1-1", "region-1-2", "region-2-1", "region-2-2"},
		},
		{
			code:     "region-1",
			expected: []string{"region-1-1", "region-1-2"},
		},
		{
			code:     "region-2",
			expected: []string{"region-2-1", "region-2-2"},
		},
		{
			code:     "region-3",
			expected: []string{"region-3"},
		},
	}

	for _, c := range testcases {
		terr := regions.Territories(c.code)
		if !reflect.DeepEqual(terr, c.expected) {
			t.Errorf("unexpected territories for %q: %v", c.code, terr)
		}
	}
}

func TestRegionsDecode(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<territoryContainment>
			<group type="001" contains="019 002 150 142 009"/>
			<group type="001" contains="EU EZ UN" status="grouping"/>
			<group type="001" contains="QU" status="deprecated"/>
			<group type="011" contains="BF BJ CI CV GH GM GN GW LR ML MR NE NG SH SL SN TG"/>
		</territoryContainment>
	</root>
	`

	var regions Regions
	err := decodeXML("test", strings.NewReader(xmlData), func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("territoryContainment", regions.decode)
	})

	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case len(regions) != 2:
		t.Errorf("unexpected number of regions: %d", len(regions))
	}
}

func TestParseRegion(t *testing.T) {
	testcases := []struct {
		containment string
		expected    Region
	}{
		{
			containment: "",
			expected:    nil,
		},
		{
			containment: "region-1",
			expected:    Region{"region-1"},
		},
		{
			containment: "region-1 region-2   region-3",
			expected:    Region{"region-1", "region-2", "region-3"},
		},
	}

	for _, c := range testcases {
		region := parseRegion(c.containment)
		if !reflect.DeepEqual(region, c.expected) {
			t.Errorf("unexpected region for %q: %+v", c.containment, region)
		}
	}
}
