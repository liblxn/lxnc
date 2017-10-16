package cldr

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestNumberingSystemsDecode(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<numberingSystems>
			<numberingSystem id="hant" type="algorithmic" rules="zh_Hant/SpelloutRules/spellout-cardinal"/>
			<numberingSystem id="latn" type="numeric" digits="0123456789"/>
			<numberingSystem id="hmng" type="numeric" digits="&#x16B50;&#x16B51;&#x16B52;&#x16B53;&#x16B54;&#x16B55;&#x16B56;&#x16B57;&#x16B58;&#x16B59;"/>
		</numberingSystems>
	</root>
	`

	var systems NumberingSystems
	err := decodeXML("test", strings.NewReader(xmlData), func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("numberingSystems", systems.decode)
	})

	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case len(systems) != 2:
		t.Errorf("unexpected number of numbering systems: %d", len(systems))
	case string(systems["latn"].Digits) != "0123456789":
		t.Errorf("unexpected digits for latn: %q", systems["latn"].Digits)
	case string(systems["hmng"].Digits) != "\U00016B50\U00016B51\U00016B52\U00016B53\U00016B54\U00016B55\U00016B56\U00016B57\U00016B58\U00016B59":
		t.Errorf("unexpected digits for hmng: %+q", systems["hmng"].Digits)
	}
}
