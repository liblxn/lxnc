package locale

const root Locale = 671

// Locale represents a reference to the data of a CLDR locale.
type Locale tagID

// New looks up a locale from the given tag (e.g. "en-US"). If the tag is
// malformed or cannot be found in the CLDR specification, an error will
// be returned.
func New(tag string) (Locale, error) {
	var p localeTagParser
	if err := p.parse(tag); err != nil {
		return 0, err
	}

	lang := langTags.langID(p.lang)
	script := scriptTags.scriptID(p.script)
	region := regionTags.regionID(p.region)
	switch {
	case lang == 0:
		return 0, errorf("unsupported language: %s", p.lang)
	case len(p.script) != 0 && script == 0:
		return 0, errorf("unsupported script: %s", p.script)
	}

	var tagID tagID
	if len(p.region) == 0 || region != 0 {
		tagID = localeTags.tagID(lang, script, region)
	}
	if tagID == 0 && len(p.region) != 0 {
		var parents [2]regionID
		nparents := regionContainment.containmentIDs(p.region, parents[:])
		if nparents == 0 && region == 0 {
			return 0, errorf("unsupported region: %s", p.region)
		}

		for i := 0; i < nparents; i++ {
			if tagID = localeTags.tagID(lang, script, parents[i]); tagID != 0 {
				break
			}
		}
	}

	if tagID == 0 {
		return 0, errorf("locale not found: %s", tag)
	}
	return Locale(tagID), nil
}

// Subtags returns the language, script, and region subtags of the locale. If one
// of the subtags are not specified, an empty string will be returned for this subtag.
func (l Locale) Subtags() (lang string, script string, region string) {
	langID, scriptID, regionID := l.tagIDs()
	if langID != 0 {
		lang = langTags.lang(langID)
	} else {
		lang = "und"
	}
	if scriptID != 0 {
		script = scriptTags.script(scriptID)
	}
	if regionID != 0 {
		region = regionTags.region(regionID)
	}
	return
}

// String returns the string represenation of the locale.
func (l Locale) String() string {
	const sep = '-'
	var buf [12]byte

	lang, script, region := l.Subtags()
	n := copy(buf[:], lang)
	if script != "" {
		buf[n] = sep
		n++
		n += copy(buf[n:], script)
	}
	if region != "" {
		buf[n] = sep
		n++
		n += copy(buf[n:], region)
	}
	return string(buf[:n])
}

func (l Locale) tagIDs() (langID, scriptID, regionID) {
	tag := localeTags.tag(tagID(l))
	return tag.langID(), tag.scriptID(), tag.regionID()
}

func (l Locale) parent() Locale {
	if parentID := parentLocaleTags.parentID(tagID(l)); parentID != 0 {
		return Locale(parentID)
	}

	// truncate locale
	langID, scriptID, regionID := l.tagIDs()
	if regionID != 0 {
		regionID = 0
		if tid := localeTags.tagID(langID, scriptID, regionID); tid != 0 {
			return Locale(tid)
		}
	}
	if scriptID != 0 {
		scriptID = 0
		if tid := localeTags.tagID(langID, scriptID, regionID); tid != 0 {
			return Locale(tid)
		}
	}
	return root
}

type localeTagParser struct {
	s   string
	tok string
	idx int
	buf [10]byte

	lang   []byte
	script []byte
	region []byte
}

func (p *localeTagParser) parse(tag string) error {
	if tag == "" {
		return errorString("empty locale tag")
	}

	p.s = tag
	p.idx = 0
	p.next()

	p.lang = p.parseLang()
	if len(p.lang) == 0 {
		if len(p.tok) == 0 {
			return errorf("malformed locale tag: %s", tag)
		}
		return errorf("invalid language subtag: %s", p.tok)
	}
	p.script = p.parseScript()
	p.region = p.parseRegion()

	if len(p.tok) != 0 {
		return errorf("unsupported locale suffix: %s", p.s[p.idx:])
	}
	return nil
}

func (p *localeTagParser) parseLang() []byte {
	var lang []byte
	switch len(p.tok) {
	case 2: // alpha{2}
		lang = p.buf[:2]
		lang[0] = p.tok[0] | 0x20 // lowercase
		lang[1] = p.tok[1] | 0x20 // lowercase
	case 3: // alpha{3}
		lang = p.buf[:3]
		lang[0] = p.tok[0] | 0x20 // lowercase
		lang[1] = p.tok[1] | 0x20 // lowercase
		lang[2] = p.tok[2] | 0x20 // lowercase
	default:
		return nil
	}
	p.next()
	return lang
}

func (p *localeTagParser) parseScript() []byte {
	if len(p.tok) != 4 {
		return nil
	}
	script := p.buf[3:7]
	script[0] = p.tok[0] & 0xdf // uppercase
	script[1] = p.tok[1] | 0x20 // lowercase
	script[2] = p.tok[2] | 0x20 // lowercase
	script[3] = p.tok[3] | 0x20 // lowercase
	p.next()
	return script
}

func (p *localeTagParser) parseRegion() []byte {
	var region []byte
	switch len(p.tok) {
	case 2: // alpha{2}
		region = p.buf[7:9]
		region[0] = p.tok[0] & 0xdf // uppercase
		region[1] = p.tok[1] & 0xdf // uppercase
	case 3: // digit{3}
		region = p.buf[7:10]
		region[0] = p.tok[0]
		region[1] = p.tok[1]
		region[2] = p.tok[2]
	default:
		return nil
	}
	p.next()
	return region
}

func (p *localeTagParser) next() {
	start := p.idx
	for p.idx < len(p.s) {
		if c := p.s[p.idx]; c == '-' || c == '_' {
			p.tok = p.s[start:p.idx]
			p.idx++
			return
		}
		p.idx++
	}
	p.tok = p.s[start:p.idx]
}
