package cldr

import "testing"

func TestDecodeData(t *testing.T) {
	ftree := memoryFileTree{
		"ignored": `no xml contents`,

		"common/main/de.xml": `<?xml version="1.0" encoding="UTF-8" ?>
			<ldml>
				<identity>
					<language type="de"/>
				</identity>
				<numbers>
					<defaultNumberingSystem>latn</defaultNumberingSystem>
				</numbers>
			</ldml>
		`,

		"common/supplemental/numberingSystems.xml": `<?xml version="1.0" encoding="UTF-8" ?>
			<supplementalData>
				<numberingSystems>
					<numberingSystem id="latn" type="numeric" digits="0123456789"/>
				</numberingSystems>
			</supplementalData>
		`,

		"common/supplemental/plurals.xml": `<?xml version="1.0" encoding="UTF-8" ?>
			<supplementalData>
				<plurals type="cardinal">
					<pluralRules locales="bm bo dz id ig ii in ja jbo jv jw kde kea km ko lkt lo ms my nqo root sah ses sg th to vi wo yo yue zh">
						<pluralRule count="other"> @integer 0~15, 100, 1000, 10000, 100000, 1000000, … @decimal 0.0~1.5, 10.0, 100.0, 1000.0, 10000.0, 100000.0, 1000000.0, …</pluralRule>
					</pluralRules>
				</plurals>
			</supplementalData>
		`,

		"common/supplemental/ordinals.xml": `<?xml version="1.0" encoding="UTF-8" ?>
			<supplementalData>
				<plurals type="ordinal">
					<pluralRules locales="bm bo dz id ig ii in ja jbo jv jw kde kea km ko lkt lo ms my nqo root sah ses sg th to vi wo yo yue zh">
						<pluralRule count="other"> @integer 0~15, 100, 1000, 10000, 100000, 1000000, … @decimal 0.0~1.5, 10.0, 100.0, 1000.0, 10000.0, 100000.0, 1000000.0, …</pluralRule>
					</pluralRules>
				</plurals>
			</supplementalData>
		`,

		"common/supplemental/supplementalData.xml": `<?xml version="1.0" encoding="UTF-8" ?>
			<supplementalData>
				<territoryContainment>
					<group type="001" contains="019 002 150 142 009"/>
				</territoryContainment>
			</supplementalData>
		`,
	}

	data, err := decodeData(ftree)
	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case len(data.Identities) != 1:
		t.Errorf("unexpected number of identities: %d", len(data.Identities))
	case len(data.Numbers) != 1:
		t.Errorf("unexpected number of numbers: %d", len(data.Numbers))
	case len(data.NumberingSystems) != 1:
		t.Errorf("unexpected number of numbering systems: %d", len(data.NumberingSystems))
	case len(data.Plurals.Cardinal) != 1:
		t.Errorf("unexpected number of cardinal plural rules: %d", len(data.Plurals.Cardinal))
	case len(data.Plurals.Ordinal) != 1:
		t.Errorf("unexpected number of ordinal plural rules: %d", len(data.Plurals.Ordinal))
	case len(data.Regions) != 1:
		t.Errorf("unexpected number of regions: %d", len(data.Regions))
	}
}

func TestDataParentIdentity(t *testing.T) {
	child := Identity{Language: "child"}
	parent := Identity{Language: "parent"}
	data := &Data{
		Identities: map[string]Identity{
			"child":  child,
			"parent": parent,
		},
		ParentIdentities: ParentIdentities{
			"child": "parent",
		},
	}

	p := data.ParentIdentity(child)
	if p != parent {
		t.Errorf("unexpected parent identity for %s: %s", child.String(), p.String())
	}

	p = data.ParentIdentity(parent)
	if p != parent.Truncate() {
		t.Errorf("unexpected parent identity for %s: %s", parent.String(), p.String())
	}
}

func TestDataDefaultNumberingSystem(t *testing.T) {
	data := &Data{
		Identities: map[string]Identity{
			"root":         Identity{Language: "root"},
			"parent":       Identity{Language: "parent"},
			"parent-child": Identity{Language: "parent", Territory: "child"},
		},
		Numbers: map[string]Numbers{
			"root":         Numbers{DefaultSystem: "root-numsys"},
			"parent":       Numbers{DefaultSystem: "parent-numsys"},
			"parent-child": Numbers{DefaultSystem: ""},
		},
	}

	numsys := data.DefaultNumberingSystem(data.Identities["parent-child"])
	if numsys != "parent-numsys" {
		t.Errorf("unexpected numbering system for the child locale: %s", numsys)
	}

	numsys = data.DefaultNumberingSystem(data.Identities["parent"])
	if numsys != "parent-numsys" {
		t.Errorf("unexpected numbering system for the parent locale: %s", numsys)
	}

	numsys = data.DefaultNumberingSystem(data.Identities["root"])
	if numsys != "root-numsys" {
		t.Errorf("unexpected numbering system for the root locale: %s", numsys)
	}
}

func TestDataFullNumberSymbols(t *testing.T) {
	const numsys = "numsys"

	data := &Data{
		Identities: map[string]Identity{
			"root":         Identity{Language: "root"},
			"parent":       Identity{Language: "parent"},
			"parent-child": Identity{Language: "parent", Territory: "child"},
		},
		Numbers: map[string]Numbers{
			"root": Numbers{
				Symbols: map[string]NumberSymbols{
					numsys: NumberSymbols{Decimal: "dec"},
				},
			},
			"parent": Numbers{
				Symbols: map[string]NumberSymbols{
					numsys: NumberSymbols{Group: "group"},
				},
			},
			"parent-child": Numbers{
				Symbols: map[string]NumberSymbols{
					numsys: NumberSymbols{Percent: "percent"},
				},
			},
		},
	}

	expected := NumberSymbols{Decimal: "dec"}
	symbols := data.FullNumberSymbols(data.Identities["root"], numsys)
	if symbols != expected {
		t.Errorf("unexpected number symbols for root: %#v", symbols)
	}

	expected.Group = "group"
	symbols = data.FullNumberSymbols(data.Identities["parent"], numsys)
	if symbols != expected {
		t.Errorf("unexpected number symbols for parent: %#v", symbols)
	}

	expected.Percent = "percent"
	symbols = data.FullNumberSymbols(data.Identities["parent-child"], numsys)
	if symbols != expected {
		t.Errorf("unexpected number symbols for child: %#v", symbols)
	}
}
