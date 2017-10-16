package cldr

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestLikelySubtagsDecode(t *testing.T) {
	const xmlData = `<?xml version="1.0" encoding="UTF-8" ?>
	<root>
		<likelySubtags>
			<likelySubtag from="zh" to="zh_Hans_CN"/>
			<likelySubtag from="zh_HK" to="zh_Hant_HK"/>
			<likelySubtag from="zh_Hani" to="zh_Hani_CN"/>
		</likelySubtags>
	</root>
	`

	var subtags LikelySubtags
	err := decodeXML("test", strings.NewReader(xmlData), func(d *xmlDecoder, _ xml.StartElement) {
		d.DecodeElem("likelySubtags", subtags.decode)
	})

	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case len(subtags) != 3:
		t.Errorf("unexpected number of likely subtags: %d", len(subtags))
	case subtags["zh"] != "zh_Hans_CN":
		t.Errorf("unexpected likely subtag for zh: %s", subtags["zh"])
	case subtags["zh_HK"] != "zh_Hant_HK":
		t.Errorf("unexpected likely subtag for zh_HK: %s", subtags["zh_HK"])
	case subtags["zh_Hani"] != "zh_Hani_CN":
		t.Errorf("unexpected likely subtag for zh_Hani: %s", subtags["zh_Hani"])
	}
}
