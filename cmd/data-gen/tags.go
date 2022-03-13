package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/liblxn/lxnc/internal/cldr"
)

const (
	langIDBits    = 8
	langBlockSize = 3

	scriptIDBits    = 8
	scriptBlockSize = 4

	regionIDBits    = 8
	regionBlockSize = 3

	tagIDBits = 16
)

// language subtag lookup
type langLookup struct {
	_stringBlockLookup
}

func (l langLookup) generate(p *printer) {
	l._stringBlockLookup.generate(p, "lang", langIDBits, langBlockSize)
}

func (l langLookup) generateTest(p *printer) {
	l._stringBlockLookup.generateTest(p, "lang", langBlockSize)
}

type langLookupVar struct {
	_stringBlockLookupVar
}

func newLangLookupVar(name string) *langLookupVar {
	return &langLookupVar{
		_stringBlockLookupVar: _stringBlockLookupVar{
			feature:   "lang",
			idBits:    langIDBits,
			blocksize: langBlockSize,
			name:      name,
		},
	}
}

func (v *langLookupVar) langID(code string) uint {
	return v._stringBlockLookupVar.stringID(code)
}

// script subtag lookup
type scriptLookup struct {
	_stringBlockLookup
}

func (l scriptLookup) generate(p *printer) {
	l._stringBlockLookup.generate(p, "script", scriptIDBits, scriptBlockSize)
}

func (l scriptLookup) generateTest(p *printer) {
	l._stringBlockLookup.generateTest(p, "script", scriptBlockSize)
}

type scriptLookupVar struct {
	_stringBlockLookupVar
}

func newScriptLookupVar(name string) *scriptLookupVar {
	return &scriptLookupVar{
		_stringBlockLookupVar: _stringBlockLookupVar{
			feature:   "script",
			idBits:    scriptIDBits,
			blocksize: scriptBlockSize,
			name:      name,
		},
	}
}

func (v *scriptLookupVar) scriptID(code string) uint {
	return v._stringBlockLookupVar.stringID(code)
}

// region subtag lookup
type regionLookup struct {
	_stringBlockLookup
}

func (l regionLookup) generate(p *printer) {
	l._stringBlockLookup.generate(p, "region", regionIDBits, regionBlockSize)
}

func (l regionLookup) generateTest(p *printer) {
	l._stringBlockLookup.generateTest(p, "region", regionBlockSize)
}

type regionLookupVar struct {
	_stringBlockLookupVar
}

func newRegionLookupVar(name string) *regionLookupVar {
	return &regionLookupVar{
		_stringBlockLookupVar: _stringBlockLookupVar{
			feature:   "region",
			idBits:    regionIDBits,
			blocksize: regionBlockSize,
			name:      name,
		},
	}
}

func (v *regionLookupVar) regionID(code string) uint {
	return v._stringBlockLookupVar.stringID(code)
}

// region containment lookup
type regionContainmentLookup struct{}

func (l regionContainmentLookup) imports() []string {
	imp := make([]string, 1, 2)
	imp[0] = "sort"
	if regionIDBits > 8 {
		imp = append(imp, "encoding/binary")
	}
	return imp
}

func (l regionContainmentLookup) generate(p *printer) {
	p.Println(`// A region containment is a tuple consisting of a 2-letter region code and`)
	p.Println(`// a list of region subtag ids. The lookup maps an alphabetic region code (the child)`)
	p.Println(`// to a list of containing regions (the parents). Each entry in the mapping is`)
	p.Println(`// encoded as "RR\xnn\xp1...\xpn", where RR is the child region code, n the number`)
	p.Println(`// of parent subtag ids and p1, ..., pn the parent subtag ids.`)
	p.Println(`type regionContainmentLookup string`)
	p.Println()
	p.Println(`func (l regionContainmentLookup) containmentIDs(region []byte, parents []regionID) int {`)
	p.Println(`	// The length of parents specifies the number of containment ids per block.`)
	p.Println(`	if len(region) != 2 {`)
	p.Println(`		return 0`)
	p.Println(`	}`)
	p.Println(`	blocksize := 2 + `, regionIDBits/8, `*len(parents)`)
	p.Println(`	idx := sort.Search(len(l)/blocksize, func(i int) bool {`)
	p.Println(`		i *= blocksize`)
	p.Println(`		return l[i:i+2] >= regionContainmentLookup(region)`)
	p.Println(`	})`)
	p.Println()
	p.Println(`	idx *= blocksize`)
	p.Println(`	if idx < len(l) && l[idx:idx+2] == regionContainmentLookup(region) {`)
	p.Println(`		for i := 0; i < len(parents); i++ {`)
	p.Println(`			start := idx + i`)
	if regionIDBits <= 8 {
		p.Println(`			parents[i] = regionID(l[start+2])`)
	} else {
		p.Println(`			parents[i] = regionID(binary.Uint`, regionIDBits, `([]byte(l[start+2:start+blocksize])))`)
	}
	p.Println(`			if parents[i] == 0 {`)
	p.Println(`				return i`)
	p.Println(`			}`)
	p.Println(`		}`)
	p.Println(`		return len(parents)`)

	p.Println(`	}`)
	p.Println(`	return 0`)
	p.Println(`}`)
}

func (l regionContainmentLookup) testImports() []string {
	return nil
}

func (l regionContainmentLookup) generateTest(p *printer) {
	p.Println(`func TestRegionContainmentLookup(t *testing.T) {`)
	p.Println(`	const lookup regionContainmentLookup = "AA\x01\x00\x00BB\x01\x02\x00CC\x01\x02\x03"`)
	p.Println()
	p.Println(`	var parents [3]regionID`)
	p.Println(`	for c := byte('A'); c <= 'C'; c++ {`)
	p.Println(`		region := [2]byte{c, c}`)
	p.Println(`		n := lookup.containmentIDs(region[:], parents[:])`)
	p.Println(`		if n != int(c-'A')+1 {`)
	p.Println(`			t.Errorf("unexpected number of parents for region %s: %d", region, n)`)
	p.Println(`		}`)
	p.Println(`		for i := 0; i < n; i++ {`)
	p.Println(`			if parents[i] != regionID(i+1) {`)
	p.Println(`				t.Errorf("unexpected parent %d for region %s: %d", i, region, parents[i])`)
	p.Println(`			}`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println()
	p.Println(`	invalidRegions := [][]byte{`)
	p.Println(`		{},`)
	p.Println(`		{'A'},`)
	p.Println(`		{'A', 'B'},`)
	p.Println(`		{'A', 'B', 'C'},`)
	p.Println(`		{'A', 'A', 'A'},`)
	p.Println(`		{'B', 'B', 'B'},`)
	p.Println(`		{'C', 'C', 'C'},`)
	p.Println(`	}`)
	p.Println(`	for _, region := range invalidRegions {`)
	p.Println(`		n := lookup.containmentIDs(region, parents[:])`)
	p.Println(`		if n != 0 {`)
	p.Println(`			t.Errorf("unexpected number of parent for region %s: %d", region, n)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}

type regionContainmentLookupVar struct {
	name       string
	regions    *regionLookupVar
	children   []string
	parents    [][]string
	maxParents int
}

func newRegionContainmentLookupVar(name string, regions *regionLookupVar) *regionContainmentLookupVar {
	return &regionContainmentLookupVar{
		name:    name,
		regions: regions,
	}
}

func (v *regionContainmentLookupVar) childrenOf(parent string) []string {
	var children []string
	for i := 0; i < len(v.children); i++ {
		for _, p := range v.parents[i] {
			if p == parent {
				children = append(children, v.children[i])
				break
			}
		}
	}
	return children
}

func (v *regionContainmentLookupVar) add(child string, parents []string) {
	idx := 0
	for idx < len(v.children) && v.children[idx] < child {
		idx++
	}
	if idx < len(v.children) && v.children[idx] == child {
		panic(fmt.Sprintf("containment for child region %s already exists", child))
	}

	v.children = append(v.children, "")
	copy(v.children[idx+1:], v.children[idx:])
	v.children[idx] = child

	v.parents = append(v.parents, nil)
	copy(v.parents[idx+1:], v.parents[idx:])
	v.parents[idx] = parents

	if len(parents) > v.maxParents {
		v.maxParents = len(parents)
	}
}

func (v *regionContainmentLookupVar) imports() []string {
	return nil
}

func (v *regionContainmentLookupVar) generate(p *printer) {
	var regionID func(id uint) string
	switch {
	case regionIDBits <= 8:
		regionID = func(id uint) string {
			return fmt.Sprintf(`\x%02x`, id)
		}
	case regionIDBits <= 16:
		regionID = func(id uint) string {
			var buf [2]byte
			binary.BigEndian.PutUint16(buf[:], uint16(id))
			return string(buf[:])
		}
	case regionIDBits <= 32:
		regionID = func(id uint) string {
			var buf [4]byte
			binary.BigEndian.PutUint32(buf[:], uint32(id))
			return string(buf[:])
		}
	default:
		regionID = func(id uint) string {
			var buf [8]byte
			binary.BigEndian.PutUint64(buf[:], uint64(id))
			return string(buf[:])
		}
	}

	var buf bytes.Buffer
	newContainment := func(child string, parents []string) string {
		buf.Reset()
		buf.WriteString(child)

		i := 0
		for ; i < len(parents); i++ {
			parentID := v.regions.regionID(parents[i])
			buf.WriteString(regionID(parentID))
		}
		for ; i < v.maxParents; i++ {
			buf.WriteString(regionID(0))
		}
		return buf.String()
	}

	parentBlocksize := v.maxParents * regionIDBits / 8
	blocksize := 2 + parentBlocksize                // 2-letter region code + v.maxParents subtag ids
	perLine := lineLength / (2 + 4*parentBlocksize) // "\xff" for each subtag id

	p.Println(`const `, v.name, ` regionContainmentLookup = "" + // `, len(v.children), ` items, `, len(v.children)*blocksize, ` bytes`)

	children := v.children
	parents := v.parents
	for len(children) > perLine {
		p.Print(`	"`)
		for i := 0; i < perLine; i++ {
			p.Print(newContainment(children[i], parents[i]))
		}
		p.Println(`" +`)

		children = children[perLine:]
		parents = parents[perLine:]
	}

	p.Print(`	"`)
	for i := 0; i < len(children); i++ {
		p.Print(newContainment(children[i], parents[i]))
	}
	p.Println(`"`)
}

func (v *regionContainmentLookupVar) testImports() []string {
	return []string{"reflect"}
}

func (v *regionContainmentLookupVar) generateTest(p *printer) {
	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[string][]regionID{ // child region => parent region ids`)

	fmtRegionID := func(id uint) string {
		return fmt.Sprintf("%#0[2]*[1]x", id, regionIDBits/4)
	}

	printParents := func(parents []string) {
		i := 0
		p.Print(`{`)
		p.Print(fmtRegionID(v.regions.regionID(parents[i])))
		for i++; i < len(parents); i++ {
			p.Print(`, `, fmtRegionID(v.regions.regionID(parents[i])))
		}
		for ; i < v.maxParents; i++ {
			p.Print(`, `, fmtRegionID(0))
		}
		p.Print(`}`)
	}

	perLine := lineLength / (6 + v.maxParents*4)
	for i := 0; i < len(v.children); i += perLine {
		n := i + perLine
		if n > len(v.children) {
			n = len(v.children)
		}

		child := v.children[i]
		p.Print(`		"`, child, `": `)
		printParents(v.parents[i])
		for k := i + 1; k < n; k++ {
			child = v.children[k]
			p.Print(`, "`, child, `": `)
			printParents(v.parents[k])
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for child, expectedParents := range expected {`)
	p.Println(`		expectedN := `, v.maxParents)
	p.Println(`		for expectedParents[expectedN-1] == 0 {`)
	p.Println(`			expectedN--`)
	p.Println(`		}`)
	p.Println()
	p.Println(`		var parents [`, v.maxParents, `]regionID`)
	p.Println(`		n := `, v.name, `.containmentIDs([]byte(child), parents[:])`)
	p.Println(`		switch {`)
	p.Println(`		case n != expectedN:`)
	p.Println(`			t.Errorf("unexpected number of parents for %s: %d (expected %d)", child, n, expectedN)`)
	p.Println(`		case reflect.DeepEqual(parents, expectedParents):`)
	p.Println(`			t.Errorf("unexpected parents: %v (expected %v)", parents, expectedParents)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}

// locale tag lookup
type tagLookup struct{}

func (l tagLookup) imports() []string {
	return []string{"sort"}
}

func (l tagLookup) generate(p *printer) {
	bits := langIDBits + scriptIDBits + regionIDBits
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

	langMask := fmt.Sprintf("%#x", (1<<langIDBits)-1)
	scriptMask := fmt.Sprintf("%#x", (1<<scriptIDBits)-1)
	regionMask := fmt.Sprintf("%#x", (1<<regionIDBits)-1)

	p.Println(`// A tag is a tuple consisting of language subtag, the script subtag,`)
	p.Println(`// and the region subtag. The tag lookup is an ordered list of tags and the`)
	p.Println(`// tag id is an 1-based index of this list.`)
	p.Println(`type tag uint`, bits)
	p.Println()
	p.Println(`func (t tag) langID() langID {`)
	p.Println(`	return langID((t >> `, scriptIDBits+regionIDBits, `) & `, langMask, `)`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (t tag) scriptID() scriptID {`)
	p.Println(`	return scriptID((t >> `, regionIDBits, `) & `, scriptMask, `)`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (t tag) regionID() regionID {`)
	p.Println(`	return regionID(t & `, regionMask, `)`)
	p.Println(`}`)
	p.Println()
	p.Println(`type tagID uint`, tagIDBits)
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
	p.Println(`	t := (tag(lang) << `, scriptIDBits+regionIDBits, `) | (tag(script) << `, regionIDBits, `) | tag(region)`)
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

func (l tagLookup) testImports() []string {
	return nil
}

func (l tagLookup) generateTest(p *printer) {
	const bits = langIDBits + scriptIDBits + regionIDBits

	tag := func(langID, scriptID, regionID uint) string {
		return fmt.Sprintf("%#0[2]*[1]x", (langID<<(scriptIDBits+regionIDBits))|(scriptID<<regionIDBits)|regionID, bits/4)
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

type tagLookupVar struct {
	name    string
	langs   *langLookupVar
	scripts *scriptLookupVar
	regions *regionLookupVar
	ids     []cldr.Identity
}

func newTagLookupVar(name string, langs *langLookupVar, scripts *scriptLookupVar, regions *regionLookupVar) *tagLookupVar {
	return &tagLookupVar{
		name:    name,
		langs:   langs,
		scripts: scripts,
		regions: regions,
	}
}

func (v *tagLookupVar) newTag(id cldr.Identity) uint64 {
	langID := v.langs.langID(id.Language) & ((1 << langIDBits) - 1)
	scriptID := uint(0)
	if id.Script != "" {
		scriptID = v.scripts.scriptID(id.Script) & ((1 << scriptIDBits) - 1)
	}
	regionID := uint(0)
	if id.Territory != "" {
		regionID = v.regions.regionID(id.Territory) & ((1 << regionIDBits) - 1)
	}
	return uint64(langID<<(scriptIDBits+regionIDBits)) | uint64(scriptID<<regionIDBits) | uint64(regionID)
}

func (v *tagLookupVar) containsTag(id cldr.Identity) bool {
	for i := 0; i < len(v.ids); i++ {
		if v.ids[i] == id {
			return true
		}
	}
	return false
}

func (v *tagLookupVar) add(id cldr.Identity) {
	if len(v.ids) == (1<<tagIDBits)-1 {
		panic(fmt.Sprintf("number of tags exceeds the maximum, cannot add %s", id.String()))
	}

	idx := 0
	for idx < len(v.ids) && identityLess(v.ids[idx], id) {
		idx++
	}
	if idx == len(v.ids) || v.ids[idx] != id {
		v.ids = append(v.ids, id)
		copy(v.ids[idx+1:], v.ids[idx:])
		v.ids[idx] = id
	}
}

func (v *tagLookupVar) tagID(id cldr.Identity) uint {
	for i := 0; i < len(v.ids); i++ {
		if v.ids[i] == id {
			return uint(i + 1)
		}
	}
	panic(fmt.Sprintf("tag not found: %s", id.String()))
}

func (v *tagLookupVar) iterate(iter func(id cldr.Identity)) {
	for _, id := range v.ids {
		iter(id)
	}
}

func (v *tagLookupVar) imports() []string {
	return nil
}

func (v *tagLookupVar) generate(p *printer) {
	const digitsPerTag = (langIDBits + scriptIDBits + regionIDBits) / 4
	const perLine = lineLength / (digitsPerTag + 4) // additional "0x" and ", "

	hex := func(id cldr.Identity) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.newTag(id), digitsPerTag)
	}

	p.Println(`var `, v.name, ` = tagLookup{ // `, len(v.ids), ` items, `, len(v.ids)*digitsPerTag/2, ` bytes`)

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

func (v *tagLookupVar) testImports() []string {
	return nil
}

func (v *tagLookupVar) generateTest(p *printer) {
	newID := func(i int) string {
		return fmt.Sprintf("%#0[2]*[1]x", i+1, tagIDBits/4)
	}

	newTag := func(id cldr.Identity) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.newTag(id), (langIDBits+scriptIDBits+regionIDBits)/4)
	}

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[tagID]tag{ // tag id => tag id`)

	perLine := lineLength / ((tagIDBits + langIDBits + scriptIDBits + regionIDBits) / 4)
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

// parent locale tag lookup
type parentTagLookup struct{}

func (l parentTagLookup) imports() []string {
	return []string{"sort"}
}

func (l parentTagLookup) generate(p *printer) {
	const tagIDMask = (1 << tagIDBits) - 1

	bits := 2 * tagIDBits
	switch {
	case bits <= 16:
		bits = 16
	case bits <= 32:
		bits = 32
	case bits <= 64:
		bits = 64
	default:
		panic(fmt.Sprintf("parent tag exceeds the maximum bit size: %d", bits))
	}

	p.Println(`// The parent tag lookup is an ordered list of tag id pairs. Each pair consists`)
	p.Println(`// of a child id and a parent id.`)
	p.Println(`type parentTagLookup []uint`, bits, ` // tag id => tag id`)
	p.Println()
	p.Println(`func (l parentTagLookup) parentID(child tagID) tagID {`)
	p.Println(`	idx := sort.Search(len(l), func(i int) bool {`)
	p.Println(`		return tagID(l[i]>>`, tagIDBits, `) >= child`)
	p.Println(`	})`)
	p.Println(`	if idx < len(l) && tagID(l[idx]>>`, tagIDBits, `) == child {`)
	p.Println(`		return tagID(l[idx] & `, fmt.Sprintf("%#x", tagIDMask), `)`)
	p.Println(`	}`)
	p.Println(`	return 0`)
	p.Println(`}`)
}

func (l parentTagLookup) testImports() []string {
	return nil
}

func (l parentTagLookup) generateTest(p *printer) {
	parentTag := func(child, parent uint) string {
		val := (child << tagIDBits) | parent
		return fmt.Sprintf("%#0[2]*[1]x", val, tagIDBits/2)
	}

	p.Println(`func TestParentTagLookup(t *testing.T) {`)
	p.Println(`	lookup := parentTagLookup{`, parentTag(1, 2), `, `, parentTag(3, 4), `, `, parentTag(5, 6), `}`)
	p.Println()
	p.Println(`	if id := lookup.parentID(1); id != 2 {`)
	p.Println(`		t.Errorf("unexpected parent for 1: %d", id)`)
	p.Println(`	}`)
	p.Println(`	if id := lookup.parentID(3); id != 4 {`)
	p.Println(`		t.Errorf("unexpected parent for 3: %d", id)`)
	p.Println(`	}`)
	p.Println(`	if id := lookup.parentID(5); id != 6 {`)
	p.Println(`		t.Errorf("unexpected parent for 5: %d", id)`)
	p.Println(`	}`)
	p.Println(`	if id := lookup.parentID(7); id != 0 {`)
	p.Println(`		t.Errorf("unexpected parent for 7: %d", id)`)
	p.Println(`	}`)
	p.Println(`}`)
}

type parentTagLookupVar struct {
	name     string
	tags     *tagLookupVar
	children []cldr.Identity
	parents  []cldr.Identity
}

func newParentTagLookupVar(name string, tags *tagLookupVar) *parentTagLookupVar {
	return &parentTagLookupVar{
		name: name,
		tags: tags,
	}
}

func (v *parentTagLookupVar) iterate(iter func(child, parent cldr.Identity)) {
	for i := 0; i < len(v.children); i++ {
		iter(v.children[i], v.parents[i])
	}
}

func (v *parentTagLookupVar) containsParent(child cldr.Identity) bool {
	for _, c := range v.children {
		if c == child {
			return true
		}
	}
	return false
}

func (v *parentTagLookupVar) add(child, parent cldr.Identity) {
	idx := 0
	for idx < len(v.children) && identityLess(v.children[idx], child) {
		idx++
	}
	if idx == len(v.children) || v.children[idx] != child {
		v.children = append(v.children, child)
		copy(v.children[idx+1:], v.children[idx:])
		v.children[idx] = child

		v.parents = append(v.parents, parent)
		copy(v.parents[idx+1:], v.parents[idx:])
		v.parents[idx] = parent
	}
}

func (v *parentTagLookupVar) imports() []string {
	return nil
}

func (v *parentTagLookupVar) generate(p *printer) {
	bits := 2 * tagIDBits
	switch {
	case bits <= 16:
		bits = 16
	case bits <= 32:
		bits = 32
	default:
		bits = 64
	}

	digitsPerElem := bits / 4
	perLine := lineLength / (digitsPerElem + 4) // additional "0x" and ", "

	hex := func(child, parent cldr.Identity) string {
		childID := v.tags.tagID(child)
		parentID := v.tags.tagID(parent)
		return fmt.Sprintf("%#0[2]*[1]x", (childID<<tagIDBits)|parentID, digitsPerElem)
	}

	p.Println(`var `, v.name, ` = parentTagLookup{ // `, len(v.children), ` items, `, len(v.children)*bits/8, ` bytes`)

	for i := 0; i < len(v.children); i += perLine {
		n := i + perLine
		if n > len(v.children) {
			n = len(v.children)
		}

		p.Print(`	`, hex(v.children[i], v.parents[i]), `, `)
		for k := i + 1; k < n; k++ {
			p.Print(hex(v.children[k], v.parents[k]), `, `)
		}
		p.Print(`// `)
		for k := i; k < n; k++ {
			p.Print(v.children[k].String(), ` -> `, v.parents[k].String())
			if k < n-1 {
				p.Print(`, `)
			}
		}
		p.Println()

	}

	p.Println(`}`)
}

func (v *parentTagLookupVar) testImports() []string {
	return nil
}

func (v *parentTagLookupVar) generateTest(p *printer) {
	tagID := func(id cldr.Identity) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.tags.tagID(id), tagIDBits/4)
	}

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[tagID]tagID{`)

	perLine := lineLength / (2 + 2*tagIDBits/4)
	for i := 0; i < len(v.children); i += perLine {
		n := i + perLine
		if n > len(v.children) {
			n = len(v.children)
		}

		p.Print(`		`, tagID(v.children[i]), `: `, tagID(v.parents[i]))
		for k := i + 1; k < n; k++ {
			p.Print(`, `, tagID(v.children[k]), `: `, tagID(v.parents[k]))
		}
		p.Println(`,`)
	}

	p.Println(`	}`)
	p.Println()
	p.Println(`	for child, expectedParent := range expected {`)
	p.Println(`		if parent := `, v.name, `.parentID(child); parent != expectedParent {`)
	p.Println(`			t.Errorf("unexpected parent id for child id %#x: %#x", child, parent)`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`}`)
}

// utility functions
func identityLess(x, y cldr.Identity) bool {
	switch {
	case x.Language != y.Language:
		return x.Language < y.Language
	case x.Script != y.Script:
		return x.Script < y.Script
	default:
		return x.Territory < y.Territory
	}
}
