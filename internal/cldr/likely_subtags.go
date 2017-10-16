package cldr

import "encoding/xml"

// LikelySubtags contains a mapping for likely subtags. These data can be used
// to add or remove subtags of a locale.
type LikelySubtags map[string]string

func (l *LikelySubtags) decode(d *xmlDecoder, _ xml.StartElement) {
	*l = make(LikelySubtags)

	d.DecodeElem("likelySubtag", func(d *xmlDecoder, elem xml.StartElement) {
		from := xmlAttrib(elem, "from")
		to := xmlAttrib(elem, "to")
		if from != "" && to != "" {
			(*l)[from] = to
		}
		d.SkipElem()
	})
}
