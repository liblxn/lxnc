package locale

import (
	"testing"
)

func TestLangLookup(t *testing.T) {
	expected := [3]string{"a", "bb", "ccc"}
	lookup := langLookup("a  bb ccc")

	for i, expectedStr := range expected {
		if id := lookup.langID([]byte(expectedStr)); id != langID(i+1) {
			t.Errorf("unexpected lang id for %q: %d", expectedStr, id)
		}
		if str := lookup.lang(langID(i + 1)); str != expectedStr {
			t.Errorf("unexpected string for lang id %d: %s", i+1, str)
		}
	}

	if id := lookup.langID([]byte{'1'}); id != 0 {
		t.Errorf("unexpected lang id: %d", id)
	}

	if str := lookup.lang(0); str != "" {
		t.Errorf("unexpected string for id 0: %s", str)
	}
	if str := lookup.lang(langID(len(lookup) + 1)); str != "" {
		t.Errorf("unexpected string id %d: %s", len(lookup)+1, str)
	}
}

func TestScriptLookup(t *testing.T) {
	expected := [4]string{"a", "bb", "ccc", "dddd"}
	lookup := scriptLookup("a   bb  ccc dddd")

	for i, expectedStr := range expected {
		if id := lookup.scriptID([]byte(expectedStr)); id != scriptID(i+1) {
			t.Errorf("unexpected script id for %q: %d", expectedStr, id)
		}
		if str := lookup.script(scriptID(i + 1)); str != expectedStr {
			t.Errorf("unexpected string for script id %d: %s", i+1, str)
		}
	}

	if id := lookup.scriptID([]byte{'1'}); id != 0 {
		t.Errorf("unexpected script id: %d", id)
	}

	if str := lookup.script(0); str != "" {
		t.Errorf("unexpected string for id 0: %s", str)
	}
	if str := lookup.script(scriptID(len(lookup) + 1)); str != "" {
		t.Errorf("unexpected string id %d: %s", len(lookup)+1, str)
	}
}

func TestRegionLookup(t *testing.T) {
	expected := [3]string{"a", "bb", "ccc"}
	lookup := regionLookup("a  bb ccc")

	for i, expectedStr := range expected {
		if id := lookup.regionID([]byte(expectedStr)); id != regionID(i+1) {
			t.Errorf("unexpected region id for %q: %d", expectedStr, id)
		}
		if str := lookup.region(regionID(i + 1)); str != expectedStr {
			t.Errorf("unexpected string for region id %d: %s", i+1, str)
		}
	}

	if id := lookup.regionID([]byte{'1'}); id != 0 {
		t.Errorf("unexpected region id: %d", id)
	}

	if str := lookup.region(0); str != "" {
		t.Errorf("unexpected string for id 0: %s", str)
	}
	if str := lookup.region(regionID(len(lookup) + 1)); str != "" {
		t.Errorf("unexpected string id %d: %s", len(lookup)+1, str)
	}
}

func TestTag(t *testing.T) {
	const tag tag = 0x010203

	if id := tag.langID(); id != 1 {
		t.Errorf("unexpected language id: %d", id)
	}
	if id := tag.scriptID(); id != 2 {
		t.Errorf("unexpected script id: %d", id)
	}
	if id := tag.regionID(); id != 3 {
		t.Errorf("unexpected region id: %d", id)
	}
}

func TestTagLookup(t *testing.T) {
	lookup := tagLookup{0x010203, 0x040506, 0x070809}

	for i := 0; i < len(lookup); i++ {
		id := lookup.tagID(langID(1+3*i), scriptID(2+3*i), regionID(3+3*i))
		if id != tagID(i+1) {
			t.Errorf("unexpected tag id for %#x: %d", lookup[i], id)
			continue
		}

		tag := lookup.tag(id)
		if tag != lookup[i] {
			t.Errorf("unexpected tag for id %d: %#x", id, tag)
		}
	}

	if id := lookup.tagID(3, 2, 1); id != 0 {
		t.Errorf("unexpected tag id for 0x030201: %d", id)
	}

	if tag := lookup.tag(0); tag != 0 {
		t.Errorf("unexpected tag for id 0: %#x", tag)
	}
	if tag := lookup.tag(tagID(len(lookup) + 1)); tag != 0 {
		t.Errorf("unexpected tag for id %d: %#x", len(lookup)+1, tag)
	}
}

func TestRegionContainmentLookup(t *testing.T) {
	const lookup regionContainmentLookup = "AA\x01\x00\x00BB\x01\x02\x00CC\x01\x02\x03"

	var parents [3]regionID
	for c := byte('A'); c <= 'C'; c++ {
		region := [2]byte{c, c}
		n := lookup.containmentIDs(region[:], parents[:])
		if n != int(c-'A')+1 {
			t.Errorf("unexpected number of parents for region %s: %d", region, n)
		}
		for i := 0; i < n; i++ {
			if parents[i] != regionID(i+1) {
				t.Errorf("unexpected parent %d for region %s: %d", i, region, parents[i])
			}
		}
	}

	invalidRegions := [][]byte{
		[]byte{},
		[]byte{'A'},
		[]byte{'A', 'B'},
		[]byte{'A', 'B', 'C'},
		[]byte{'A', 'A', 'A'},
		[]byte{'B', 'B', 'B'},
		[]byte{'C', 'C', 'C'},
	}
	for _, region := range invalidRegions {
		n := lookup.containmentIDs(region, parents[:])
		if n != 0 {
			t.Errorf("unexpected number of parent for region %s: %d", region, n)
		}
	}
}

func TestParentTagLookup(t *testing.T) {
	lookup := parentTagLookup{0x00010002, 0x00030004, 0x00050006}

	if id := lookup.parentID(1); id != 2 {
		t.Errorf("unexpected parent for 1: %d", id)
	}
	if id := lookup.parentID(3); id != 4 {
		t.Errorf("unexpected parent for 3: %d", id)
	}
	if id := lookup.parentID(5); id != 6 {
		t.Errorf("unexpected parent for 5: %d", id)
	}
	if id := lookup.parentID(7); id != 0 {
		t.Errorf("unexpected parent for 7: %d", id)
	}
}
