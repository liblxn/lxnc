package generate_cldr

import (
	"fmt"
	"sort"
	"strings"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*tagLookup)(nil)
	_ generator.TestSnippet = (*tagLookup)(nil)
)

type tagLookup struct {
	idBits uint
	lang   *langLookup
	script *scriptLookup
	region *regionLookup
}

func newTagLookup(lang *langLookup, script *scriptLookup, region *regionLookup) *tagLookup {
	return &tagLookup{
		idBits: 16,
		lang:   lang,
		script: script,
		region: region,
	}
}

func (l *tagLookup) Imports() []string {
	return []string{"sort"}
}

func (l *tagLookup) Generate(p *generator.Printer) {
	bits := l.lang.idBits + l.script.idBits + l.region.idBits
	switch {
	case bits <= 8:
		bits = 8
	case bits <= 16:
		bits = 16
	case bits <= 32:
		bits = 32
	case bits <= 64:
		bits = 64
	default:
		panic(fmt.Sprintf("tag id exceeds the maximum bit size: %d", bits))
	}

	langMask := fmt.Sprintf("%#x", (1<<l.lang.idBits)-1)
	scriptMask := fmt.Sprintf("%#x", (1<<l.script.idBits)-1)
	regionMask := fmt.Sprintf("%#x", (1<<l.region.idBits)-1)

	p.Println(`// A tag is a tuple consisting of language subtag, the script subtag,`)
	p.Println(`// and the region subtag. The tag lookup is an ordered list of tags and the`)
	p.Println(`// tag id is an 1-based index of this list.`)
	p.Println(`type tag uint`, bits)
	p.Println()
	p.Println(`func (t tag) langID() langID {`)
	p.Println(`	return langID((t >> `, l.script.idBits+l.region.idBits, `) & `, langMask, `)`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (t tag) scriptID() scriptID {`)
	p.Println(`	return scriptID((t >> `, l.region.idBits, `) & `, scriptMask, `)`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (t tag) regionID() regionID {`)
	p.Println(`	return regionID(t & `, regionMask, `)`)
	p.Println(`}`)
	p.Println()
	p.Println(`type tagID uint`, l.idBits)
	p.Println()
	p.Println(`type tagLookup []tag`)
	p.Println()
	p.Println(`func (l tagLookup) tag(id tagID) tag {`)
	p.Println(`	if id == 0 || int(id) > len(l) {`)
	p.Println(`		return 0`)
	p.Println(`	}`)
	p.Println(`	return l[id-1]`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (l tagLookup) tagID(lang langID, script scriptID, region regionID) tagID {`)
	p.Println(`	t := (tag(lang) << `, l.script.idBits+l.region.idBits, `) | (tag(script) << `, l.region.idBits, `) | tag(region)`)
	p.Println(`	idx := sort.Search(len(l), func(i int) bool {`)
	p.Println(`		return l[i] >= t`)
	p.Println(`	})`)
	p.Println()
	p.Println(`	if idx < len(l) && l[idx] == t {`)
	p.Println(`		return tagID(idx + 1)`)
	p.Println(`	}`)
	p.Println(`	return 0`)
	p.Println(`}`)
}

func (l *tagLookup) TestImports() []string {
	return nil
}

func (l *tagLookup) GenerateTest(p *generator.Printer) {
	bits := l.lang.idBits + l.script.idBits + l.region.idBits

	tag := func(langID, scriptID, regionID uint) string {
		return fmt.Sprintf("%#0[2]*[1]x", (langID<<(l.script.idBits+l.region.idBits))|(scriptID<<l.region.idBits)|regionID, bits/4)
	}

	p.Println(`func TestTag(t *testing.T) {`)
	p.Println(`	const tag tag = `, tag(1, 2, 3))
	p.Println()
	p.Println(`	if id := tag.langID(); id != 1 {`)
	p.Println(`		t.Errorf("unexpected language id: %d", id)`)
	p.Println(`	}`)
	p.Println(`	if id := tag.scriptID(); id != 2 {`)
	p.Println(`		t.Errorf("unexpected script id: %d", id)`)
	p.Println(`	}`)
	p.Println(`	if id := tag.regionID(); id != 3 {`)
	p.Println(`		t.Errorf("unexpected region id: %d", id)`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`func TestTagLookup(t *testing.T) {`)
	p.Println(`	lookup := tagLookup{`, tag(1, 2, 3), `, `, tag(4, 5, 6), `, `, tag(7, 8, 9), `}`)
	p.Println()
	p.Println(`	for i := 0; i < len(lookup); i++ {`)
	p.Println(`		id := lookup.tagID(langID(1+3*i), scriptID(2+3*i), regionID(3+3*i))`)
	p.Println(`		if id != tagID(i+1) {`)
	p.Println(`			t.Errorf("unexpected tag id for %#x: %d", lookup[i], id)`)
	p.Println(`			continue`)
	p.Println(`		}`)
	p.Println()
	p.Println(`		tag := lookup.tag(id)`)
	p.Println(`		if tag != lookup[i] {`)
	p.Println(`			t.Errorf("unexpected tag for id %d: %#x", id, tag)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	if id := lookup.tagID(3, 2, 1); id != 0 {`)
	p.Println(`		t.Errorf("unexpected tag id for `, tag(3, 2, 1), `: %d", id)`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	if tag := lookup.tag(0); tag != 0 {`)
	p.Println(`		t.Errorf("unexpected tag for id 0: %#x", tag)`)
	p.Println(`	}`)
	p.Println(`	if tag := lookup.tag(tagID(len(lookup) + 1)); tag != 0 {`)
	p.Println(`		t.Errorf("unexpected tag for id %d: %#x", len(lookup)+1, tag)`)
	p.Println(`	}`)
	p.Println(`}`)
}

var (
	_ generator.Snippet     = (*tagLookupVar)(nil)
	_ generator.TestSnippet = (*tagLookupVar)(nil)
)

type tagLookupVar struct {
	name    string
	typ     *tagLookup
	langs   *langLookupVar
	scripts *scriptLookupVar
	regions *regionLookupVar
	ids     []cldr.Identity
}

func newTagLookupVar(name string, typ *tagLookup, langs *langLookupVar, scripts *scriptLookupVar, regions *regionLookupVar, data *cldr.Data) *tagLookupVar {
	idMap := map[cldr.Identity]struct{}{}
	forEachIdentity(data, func(id cldr.Identity) {
		idMap[id] = struct{}{}
	})

	if len(idMap) == (1 << typ.idBits) {
		panic(fmt.Sprintf("number of tags exceeds the maximum"))
	}

	ids := make([]cldr.Identity, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		return identityLess(ids[i], ids[j])
	})

	return &tagLookupVar{
		name:    name,
		typ:     typ,
		langs:   langs,
		scripts: scripts,
		regions: regions,
		ids:     ids,
	}
}

func (v *tagLookupVar) newTag(id cldr.Identity) uint64 {
	langID := v.langs.langID(id.Language) & ((1 << v.langs.typ.idBits) - 1)
	scriptID := uint(0)
	if id.Script != "" {
		scriptID = v.scripts.scriptID(id.Script) & ((1 << v.scripts.typ.idBits) - 1)
	}
	regionID := uint(0)
	if id.Territory != "" {
		regionID = v.regions.regionID(id.Territory) & ((1 << v.regions.typ.idBits) - 1)
	}
	return uint64(langID<<(v.scripts.typ.idBits+v.regions.typ.idBits)) | uint64(scriptID<<v.regions.typ.idBits) | uint64(regionID)
}

func (v *tagLookupVar) containsTag(id cldr.Identity) bool {
	for i := 0; i < len(v.ids); i++ {
		if v.ids[i] == id {
			return true
		}
	}
	return false
}

func (v *tagLookupVar) tagID(id cldr.Identity) uint {
	for i := 0; i < len(v.ids); i++ {
		if v.ids[i] == id {
			return uint(i + 1)
		}
	}
	panic(fmt.Sprintf("tag not found: %s", id.String()))
}

func (v *tagLookupVar) Imports() []string {
	return nil
}

func (v *tagLookupVar) Generate(p *generator.Printer) {
	digitsPerTag := (v.langs.typ.idBits + v.scripts.typ.idBits + v.regions.typ.idBits) / 4
	perLine := int(lineLength / (digitsPerTag + 4)) // additional "0x" and ", "

	hex := func(id cldr.Identity) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.newTag(id), digitsPerTag)
	}

	p.Println(`var `, v.name, ` = tagLookup{ // `, len(v.ids), ` items, `, uint(len(v.ids))*digitsPerTag/2, ` bytes`)

	for i := 0; i < len(v.ids); i += perLine {
		n := i + perLine
		if n > len(v.ids) {
			n = len(v.ids)
		}

		p.Print(`	`, hex(v.ids[i]), `, `)
		for _, id := range v.ids[i+1 : n] {
			p.Print(hex(id), `, `)
		}
		p.Print(`// `)
		for k, id := range v.ids[i:n] {
			p.Print(id.String())
			if k+i < n-1 {
				p.Print(`, `)
			}
		}
		p.Println()
	}

	p.Println(`}`)
}

func (v *tagLookupVar) TestImports() []string {
	return nil
}

func (v *tagLookupVar) GenerateTest(p *generator.Printer) {
	newID := func(i int) string {
		return fmt.Sprintf("%#0[2]*[1]x", i+1, v.typ.idBits/4)
	}

	newTag := func(id cldr.Identity) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.newTag(id), (v.langs.typ.idBits+v.scripts.typ.idBits+v.regions.typ.idBits)/4)
	}

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[tagID]tag{ // tag id => tag id`)

	perLine := int(lineLength / ((v.typ.idBits + v.langs.typ.idBits + v.scripts.typ.idBits + v.regions.typ.idBits) / 4))
	for i := 0; i < len(v.ids); i += perLine {
		n := i + perLine
		if n > len(v.ids) {
			n = len(v.ids)
		}

		p.Print(`		`, newID(i), `: `, newTag(v.ids[i]))
		for k := i + 1; k < n; k++ {
			p.Print(`, `, newID(k), `: `, newTag(v.ids[k]))
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for expectedTagID, expectedTag := range expected {`)
	p.Println(`		if tag := `, v.name, `.tag(expectedTagID); tag != expectedTag {`)
	p.Println(`			t.Errorf("unexpected tag for id %d, %#x", uint(expectedTagID), tag)`)
	p.Println(`		}`)
	p.Println(`		if tagID := `, v.name, `.tagID(expectedTag.langID(), expectedTag.scriptID(), expectedTag.regionID()); tagID != expectedTagID {`)
	p.Println(`			t.Errorf("unexpected tag id for tag %#x: %d", expectedTag, uint(tagID))`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}
