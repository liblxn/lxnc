package generate_cldr

import (
	"fmt"
	"sort"
	"strings"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*locale)(nil)
	_ generator.TestSnippet = (*locale)(nil)
)

type locale struct {
	packageName       string
	tags              *tagLookupVar
	parentTags        *parentTagLookupVar
	regionContainment *regionContainmentLookupVar
}

func newLocale(packageName string, tags *tagLookupVar, parentTags *parentTagLookupVar, regionContainment *regionContainmentLookupVar) *locale {
	return &locale{
		packageName:       packageName,
		tags:              tags,
		parentTags:        parentTags,
		regionContainment: regionContainment,
	}
}

func (l *locale) Imports() []string {
	return []string{"github.com/liblxn/lxnc/internal/errors"}
}

func (l *locale) Generate(p *generator.Printer) {
	root := cldr.Identity{Language: "und"}
	newLocale := "NewLocale"
	if strings.ToLower(l.packageName) == "locale" {
		newLocale = "New"
	}

	p.Println(`const root Locale = `, l.tags.tagID(root))
	p.Println()
	p.Println(`// Locale represents a reference to the data of a CLDR locale.`)
	p.Println(`type Locale tagID`)
	p.Println()
	p.Println(`// `, newLocale, ` looks up a locale from the given tag (e.g. "en-US"). If the tag is`)
	p.Println(`// malformed or cannot be found in the CLDR specification, an error will`)
	p.Println(`// be returned.`)
	p.Println(`func `, newLocale, `(tag string) (Locale, error) {`)
	p.Println(`	var p localeTagParser`)
	p.Println(`	if err := p.parse(tag); err != nil {`)
	p.Println(`		return 0, err`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	lang := langTags.langID(p.lang)`)
	p.Println(`	script := scriptTags.scriptID(p.script)`)
	p.Println(`	region := regionTags.regionID(p.region)`)
	p.Println(`	switch {`)
	p.Println(`	case lang == 0:`)
	p.Println(`		return 0, errors.Newf("unsupported language: %s", p.lang)`)
	p.Println(`	case len(p.script) != 0 && script == 0:`)
	p.Println(`		return 0, errors.Newf("unsupported script: %s", p.script)`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	var tagID tagID`)
	p.Println(`	if len(p.region) == 0 || region != 0 {`)
	p.Println(`		tagID = localeTags.tagID(lang, script, region)`)
	p.Println(`	}`)
	p.Println(`	if tagID == 0 && len(p.region) != 0 {`)
	p.Println(`		var parents [2]regionID`)
	p.Println(`		nparents := regionContainment.containmentIDs(p.region, parents[:])`)
	p.Println(`		if nparents == 0 && region == 0 {`)
	p.Println(`			return 0, errors.Newf("unsupported region: %s", p.region)`)
	p.Println(`		}`)
	p.Println()
	p.Println(`		for i := 0; i < nparents; i++ {`)
	p.Println(`			if tagID = localeTags.tagID(lang, script, parents[i]); tagID != 0 {`)
	p.Println(`				break`)
	p.Println(`			}`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	if tagID == 0 {`)
	p.Println(`		return 0, errors.Newf("locale not found: %s", tag)`)
	p.Println(`	}`)
	p.Println(`	return Locale(tagID), nil`)
	p.Println(`}`)
	p.Println()
	p.Println(`// Subtags returns the language, script, and region subtags of the locale. If one`)
	p.Println(`// of the subtags are not specified, an empty string will be returned for this subtag.`)
	p.Println(`func (l Locale) Subtags() (lang string, script string, region string) {`)
	p.Println(`	langID, scriptID, regionID := l.tagIDs()`)
	p.Println(`	if langID != 0 {`)
	p.Println(`		lang = `, l.tags.langs.name, `.lang(langID)`)
	p.Println(`	} else {`)
	p.Println(`		lang = "und"`)
	p.Println(`	}`)
	p.Println(`	if scriptID != 0 {`)
	p.Println(`		script = `, l.tags.scripts.name, `.script(scriptID)`)
	p.Println(`	}`)
	p.Println(`	if regionID != 0 {`)
	p.Println(`		region = `, l.tags.regions.name, `.region(regionID)`)
	p.Println(`	}`)
	p.Println(`	return`)
	p.Println(`}`)
	p.Println()
	p.Println(`// String returns the string represenation of the locale.`)
	p.Println(`func (l Locale) String() string {`)
	p.Println(`	const sep = '-'`)
	p.Println(`	var buf [12]byte`)
	p.Println()
	p.Println(`	lang, script, region := l.Subtags()`)
	p.Println(`	n := copy(buf[:], lang)`)
	p.Println(`	if script != "" {`)
	p.Println(`		buf[n] = sep`)
	p.Println(`		n++`)
	p.Println(`		n += copy(buf[n:], script)`)
	p.Println(`	}`)
	p.Println(`	if region != "" {`)
	p.Println(`		buf[n] = sep`)
	p.Println(`		n++`)
	p.Println(`		n += copy(buf[n:], region)`)
	p.Println(`	}`)
	p.Println(`	return string(buf[:n])`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (l Locale) tagIDs() (langID, scriptID, regionID) {`)
	p.Println(`	tag := `, l.tags.name, `.tag(tagID(l))`)
	p.Println(`	return tag.langID(), tag.scriptID(), tag.regionID()`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (l Locale) parent() Locale {`)
	p.Println(`	if parentID := `, l.parentTags.name, `.parentID(tagID(l)); parentID != 0 {`)
	p.Println(`		return Locale(parentID)`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	// truncate locale`)
	p.Println(`	langID, scriptID, regionID := l.tagIDs()`)
	p.Println(`	if regionID != 0 {`)
	p.Println(`		regionID = 0`)
	p.Println(`		if tid := `, l.tags.name, `.tagID(langID, scriptID, regionID); tid != 0 {`)
	p.Println(`			return Locale(tid)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`	if scriptID != 0 {`)
	p.Println(`		scriptID = 0`)
	p.Println(`		if tid := `, l.tags.name, `.tagID(langID, scriptID, regionID); tid != 0 {`)
	p.Println(`			return Locale(tid)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`	return root`)
	p.Println(`}`)

	// locale tag parser
	p.Println()
	p.Println(`type localeTagParser struct {`)
	p.Println(`	s   string`)
	p.Println(`	tok string`)
	p.Println(`	idx int`)
	p.Println(`	buf [10]byte`)
	p.Println()
	p.Println(`	lang   []byte`)
	p.Println(`	script []byte`)
	p.Println(`	region []byte`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (p *localeTagParser) parse(tag string) error {`)
	p.Println(`	if tag == "" {`)
	p.Println(`		return errors.New("empty locale tag")`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	p.s = tag`)
	p.Println(`	p.idx = 0`)
	p.Println(`	p.next()`)
	p.Println()
	p.Println(`	p.lang = p.parseLang()`)
	p.Println(`	if len(p.lang) == 0 {`)
	p.Println(`		if len(p.tok) == 0 {`)
	p.Println(`			return errors.Newf("malformed locale tag: %s", tag)`)
	p.Println(`		}`)
	p.Println(`		return errors.Newf("invalid language subtag: %s", p.tok)`)
	p.Println(`	}`)
	p.Println(`	p.script = p.parseScript()`)
	p.Println(`	p.region = p.parseRegion()`)
	p.Println()
	p.Println(`	if len(p.tok) != 0 {`)
	p.Println(`		return errors.Newf("unsupported locale suffix: %s", p.s[p.idx:])`)
	p.Println(`	}`)
	p.Println(`	return nil`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (p *localeTagParser) parseLang() []byte {`)
	p.Println(`	var lang []byte`)
	p.Println(`	switch len(p.tok) {`)
	p.Println(`	case 2: // alpha{2}`)
	p.Println(`		lang = p.buf[:2]`)
	p.Println(`		lang[0] = p.tok[0] | 0x20 // lowercase`)
	p.Println(`		lang[1] = p.tok[1] | 0x20 // lowercase`)
	p.Println(`	case 3: // alpha{3}`)
	p.Println(`		lang = p.buf[:3]`)
	p.Println(`		lang[0] = p.tok[0] | 0x20 // lowercase`)
	p.Println(`		lang[1] = p.tok[1] | 0x20 // lowercase`)
	p.Println(`		lang[2] = p.tok[2] | 0x20 // lowercase`)
	p.Println(`	default:`)
	p.Println(`		return nil`)
	p.Println(`	}`)
	p.Println(`	p.next()`)
	p.Println(`	return lang`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (p *localeTagParser) parseScript() []byte {`)
	p.Println(`	if len(p.tok) != 4 {`)
	p.Println(`		return nil`)
	p.Println(`	}`)
	p.Println(`	script := p.buf[3:7]`)
	p.Println(`	script[0] = p.tok[0] & 0xdf // uppercase`)
	p.Println(`	script[1] = p.tok[1] | 0x20 // lowercase`)
	p.Println(`	script[2] = p.tok[2] | 0x20 // lowercase`)
	p.Println(`	script[3] = p.tok[3] | 0x20 // lowercase`)
	p.Println(`	p.next()`)
	p.Println(`	return script`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (p *localeTagParser) parseRegion() []byte {`)
	p.Println(`	var region []byte`)
	p.Println(`	switch len(p.tok) {`)
	p.Println(`	case 2: // alpha{2}`)
	p.Println(`		region = p.buf[7:9]`)
	p.Println(`		region[0] = p.tok[0] & 0xdf // uppercase`)
	p.Println(`		region[1] = p.tok[1] & 0xdf // uppercase`)
	p.Println(`	case 3: // digit{3}`)
	p.Println(`		region = p.buf[7:10]`)
	p.Println(`		region[0] = p.tok[0]`)
	p.Println(`		region[1] = p.tok[1]`)
	p.Println(`		region[2] = p.tok[2]`)
	p.Println(`	default:`)
	p.Println(`		return nil`)
	p.Println(`	}`)
	p.Println(`	p.next()`)
	p.Println(`	return region`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (p *localeTagParser) next() {`)
	p.Println(`	start := p.idx`)
	p.Println(`	for p.idx < len(p.s) {`)
	p.Println(`		if c := p.s[p.idx]; c == '-' || c == '_' {`)
	p.Println(`			p.tok = p.s[start:p.idx]`)
	p.Println(`			p.idx++`)
	p.Println(`			return`)
	p.Println(`		}`)
	p.Println(`		p.idx++`)
	p.Println(`	}`)
	p.Println(`	p.tok = p.s[start:p.idx]`)
	p.Println(`}`)
}

func (l *locale) TestImports() []string {
	return []string{"strings"}
}

func (l *locale) GenerateTest(p *generator.Printer) {
	newLocale := "NewLocale"
	if strings.ToLower(l.packageName) == "locale" {
		newLocale = "New"
	}

	newTagID := func(id cldr.Identity) string {
		return fmt.Sprintf("%d", l.tags.tagID(id))
	}

	p.Println(`func Test`, newLocale, `(t *testing.T) {`)
	p.Println(`	expected := map[string]Locale{ // tag => locale`)

	for _, id := range l.tags.ids {
		tagID := newTagID(id)
		key := id.String()
		underscoreKey := strings.Replace(key, "-", "_", -1)
		lowerKey := strings.ToLower(key)
		upperKey := strings.ToUpper(key)

		p.Print(`		"`, key, `": `, tagID)
		if underscoreKey != key {
			p.Print(`, "`, underscoreKey, `": `, tagID)
		}
		if lowerKey != key {
			p.Print(`, "`, lowerKey, `": `, tagID)
		}
		if upperKey != key {
			p.Print(`, "`, upperKey, `": `, tagID)
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for tag, expectedLoc := range expected {`)
	p.Println(`		loc, err := `, newLocale, `(tag)`)
	p.Println(`		switch {`)
	p.Println(`		case err != nil:`)
	p.Println(`			t.Errorf("unexpected error for %s: %v", tag, err)`)
	p.Println(`		case loc != expectedLoc:`)
	p.Println(`			t.Errorf("unexpected locale for %s: %s", tag, loc.String())`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)

	p.Println()
	p.Println(`func Test`, newLocale, `WithRegionContainments(t *testing.T) {`)
	p.Println(`	expected := map[string]Locale{ // tag => locale`)

	parentIDs := make(map[string][]cldr.Identity) // truncated parent region tag => parent identities
	for _, id := range l.tags.ids {
		if len(id.Territory) == 3 && '0' <= id.Territory[0] && id.Territory[0] <= '9' {
			tag := id.Truncate().String()
			parentIDs[tag] = append(parentIDs[tag], id)
		}
	}

	parentTags := make([]string, 0, len(parentIDs))
	for tag := range parentIDs {
		parentTags = append(parentTags, tag)
	}
	sort.Strings(parentTags)

	localeContainments := make([]_localeContainment, 0, 2*len(parentIDs))
	for _, tag := range parentTags {
		pids := parentIDs[tag]
		containments := make([]_localeContainment, len(pids))
		for i, pid := range pids {
			children := l.regionContainment.childrenOf(pid.Territory)
			// filter out existing tags
			c := make([]string, 0, len(children))
			for _, child := range children {
				id := pid
				id.Territory = child
				if !l.tags.containsTag(id) {
					c = append(c, child)
				}
			}
			sort.Strings(c)
			containments[i] = _localeContainment{
				id:           pid,
				childRegions: c,
			}
		}
		sort.Sort(_localeContainmentsByChildRegionCount(containments))

		// remove duplicate child regions
		for i := 1; i < len(containments); i++ {
			existingChildren := make(map[string]struct{})
			for _, containment := range containments[:i] {
				for _, child := range containment.childRegions {
					existingChildren[child] = struct{}{}
				}
			}

			children := make([]string, 0, len(containments[i].childRegions))
			id := containments[i].id
			for _, child := range containments[i].childRegions {
				id.Territory = child
				if _, exists := existingChildren[child]; !exists {
					children = append(children, child)
				}
			}
			containments[i].childRegions = children
		}

		localeContainments = append(localeContainments, containments...)
	}

	for _, containment := range localeContainments {
		id := containment.id
		tagID := newTagID(id)

		id.Territory = containment.childRegions[0]
		p.Print(`		"`, id.String(), `": `, tagID)
		for k := 1; k < len(containment.childRegions); k++ {
			id.Territory = containment.childRegions[k]
			p.Print(`, "`, id.String(), `": `, tagID)
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for tag, expectedLoc := range expected {`)
	p.Println(`		loc, err := `, newLocale, `(tag)`)
	p.Println(`		switch {`)
	p.Println(`		case err != nil:`)
	p.Println(`			t.Errorf("unexpected error for %s: %v", tag, err)`)
	p.Println(`		case loc != expectedLoc:`)
	p.Println(`			t.Errorf("unexpected locale for %s: %s", tag, loc.String())`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)

	p.Println()
	p.Println(`func Test`, newLocale, `WithInvalidTag(t *testing.T) {`)
	p.Println(`	expectedErrors := map[string]string{ // tag => error prefix`)
	p.Println(`		"":             "empty locale tag",`)
	p.Println(`		"-DE":          "malformed locale tag",`)
	p.Println(`		"overlong-DE":  "invalid language subtag",`)
	p.Println(`		"de-DE-suffix": "unsupported locale suffix",`)
	p.Println(`		"ZZ":           "unsupported language",`)
	p.Println(`		"en-4444-US":   "unsupported script",`)
	p.Println(`		"de-zzz":       "unsupported region",`)
	p.Println(`		"de-001":       "locale not found",`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	for tag, errorPrefix := range expectedErrors {`)
	p.Println(`		_, err := `, newLocale, `(tag)`)
	p.Println(`		switch {`)
	p.Println(`		case err == nil:`)
	p.Println(`			t.Errorf("expected error for %s, got none", tag)`)
	p.Println(`		case !strings.HasPrefix(err.Error(), errorPrefix):`)
	p.Println(`			t.Errorf("unexpected error message for %s: %s", tag, err.Error())`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)

	p.Println()
	p.Println(`func TestLocaleSubtags(t *testing.T) {`)
	p.Println(`	expected := map[Locale][4]string{ // tag id => (lang, script, region, locale string)`)

	for _, id := range l.tags.ids {
		tagID := fmt.Sprintf("%#0[2]*[1]x", l.tags.tagID(id), l.tags.typ.idBits/4)
		p.Println(`		`, tagID, `: {"`, id.Language, `", "`, id.Script, `", "`, id.Territory, `", "`, id.String(), `"},`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for loc, subtags := range expected {`)
	p.Println(`		lang, script, region := loc.Subtags()`)
	p.Println(`		switch {`)
	p.Println(`		case lang != subtags[0]:`)
	p.Println(`			t.Errorf("unexpected language for %s: %s", subtags[3], lang)`)
	p.Println(`		case script != subtags[1]:`)
	p.Println(`			t.Errorf("unexpected script for %s: %s", subtags[3], script)`)
	p.Println(`		case region != subtags[2]:`)
	p.Println(`			t.Errorf("unexpected region for %s: %s", subtags[3], region)`)
	p.Println(`		case loc.String() != subtags[3]:`)
	p.Println(`			t.Errorf("unexpected region for %s : %s", subtags[3], loc.String())`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)

	p.Println()
	p.Println(`func TestLocaleParents(t *testing.T) {`)
	p.Println(`	expected := map[Locale]Locale{ // child => parent`)

	const perLine = lineLength / 6
	for i := 0; i < len(l.parentTags.data); i += perLine {
		n := i + perLine
		if n > len(l.parentTags.data) {
			n = len(l.parentTags.data)
		}

		p.Print(`		`, newTagID(l.parentTags.data[i].child), `: `, newTagID(l.parentTags.data[i].parent))
		for k := i + 1; k < n; k++ {
			p.Print(`, `, newTagID(l.parentTags.data[k].child), `: `, newTagID(l.parentTags.data[k].parent))
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for child, parent := range expected {`)
	p.Println(`		if p := child.parent(); p != parent {`)
	p.Println(`			t.Errorf("unexpected parent for %s: %s (expected %s)", child.String(), p, parent)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)

	p.Println()
	p.Println(`func TestLocaleParentsWithTruncation(t *testing.T) {`)
	p.Println(`	expected := map[Locale]Locale{ // original locale => truncated locale`)

	parents := make([]parentTagData, 0, len(l.parentTags.data))
	for _, id := range l.tags.ids {
		if l.parentTags.containsParent(id) {
			continue
		}
		truncated := normalizeIdentity(id.Truncate())
		if truncated != id {
			parents = append(parents, parentTagData{
				child:  id,
				parent: truncated,
			})
		}
	}

	for i := 0; i < len(parents); i += perLine {
		n := i + perLine
		if n > len(parents) {
			n = len(parents)
		}

		p.Print(`		`, newTagID(parents[i].child), `: `, newTagID(parents[i].parent))
		for k := i + 1; k < n; k++ {
			p.Print(`, `, newTagID(parents[k].child), `: `, newTagID(parents[k].parent))
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for loc, truncated := range expected {`)
	p.Println(`		if p := loc.parent(); p != truncated {`)
	p.Println(`			t.Errorf("unexpected parent for %s: %s (expected %s)", loc.String(), p, truncated)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}

type _localeContainment struct {
	id           cldr.Identity
	childRegions []string
}

type _localeContainmentsByChildRegionCount []_localeContainment

func (s _localeContainmentsByChildRegionCount) Len() int {
	return len(s)
}

func (s _localeContainmentsByChildRegionCount) Less(i int, j int) bool {
	return len(s[i].childRegions) < len(s[j].childRegions)
}

func (s _localeContainmentsByChildRegionCount) Swap(i int, j int) {
	s[i], s[j] = s[j], s[i]
}
