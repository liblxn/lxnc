package cldr

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestIdentityTruncate(t *testing.T) {
	testcases := []struct {
		chain []Identity
	}{
		{
			chain: []Identity{
				{Language: "zh", Script: "Hant", Territory: "TW"},
				{Language: "zh", Script: "Hant"},
				{Language: "zh"},
				{Language: "root"},
			},
		},
		{
			chain: []Identity{
				{Language: "en", Territory: "US", Variant: "POSIX"},
				{Language: "en", Territory: "US"},
				{Language: "en"},
				{Language: "root"},
			},
		},
		{
			chain: []Identity{
				{Language: "de", Territory: "DE"},
				{Language: "de"},
				{Language: "root"},
			},
		},
	}

	for _, c := range testcases {
		for i, n := 1, len(c.chain); i < n; i++ {
			id := c.chain[i-1]
			truncated := c.chain[i]
			if truncated != id.Truncate() {
				t.Errorf("unexpected truncated identity for %s: %s", id.String(), truncated.String())
			}
		}
	}

	// check if the truncation of "root" is still "root"
	root := Identity{Language: "root"}
	if truncated := root.Truncate(); root != truncated {
		t.Errorf("unexpected truncated identity for root: %s", truncated.String())
	}
}

func TestIdentityDecode(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<identity>
			<language type="zh"/>
			<script type="Hant"/>
			<territory type="HK"/>
		</identity>
	</root>
	`

	var id Identity
	err := decodeXML("test", strings.NewReader(xmlData), func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("identity", id.decode)
	})

	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case id.String() != "zh-Hant-HK":
		t.Errorf("unexpected identity: %s", id)
	}
}

func TestParentIdentitiesDecode(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<parentLocales>
			<parentLocale parent="root" locales="az_Arab az_Cyrl bm_Nkoo"/>
			<parentLocale parent="en_001" locales="en_150"/>
			<parentLocale parent="en_150" locales="en_AT en_CH"/>
		</parentLocales>
	</root>
	`

	var parents ParentIdentities
	err := decodeXML("test", strings.NewReader(xmlData), func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("parentLocales", parents.decode)
	})

	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case len(parents) != 6:
		t.Errorf("unexpected number of parent identities: %d", len(parents))
	case parents["az-Arab"] != "root":
		t.Errorf("unexpected parent for az-Arab: %s", parents["az-Arab"])
	case parents["az-Cyrl"] != "root":
		t.Errorf("unexpected parent for az-Cyrl: %s", parents["az-Cyrl"])
	case parents["bm-Nkoo"] != "root":
		t.Errorf("unexpected parent for bm-Nkoo: %s", parents["bm-Nkoo"])
	case parents["en-150"] != "en-001":
		t.Errorf("unexpected parent for en-150: %s", parents["en-150"])
	case parents["en-AT"] != "en-150":
		t.Errorf("unexpected parent for en-AT: %s", parents["en-AT"])
	case parents["en-CH"] != "en-150":
		t.Errorf("unexpected parent for en-CH: %s", parents["en-CH"])
	}
}
