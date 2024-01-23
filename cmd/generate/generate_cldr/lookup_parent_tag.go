package generate_cldr

import (
	"fmt"
	"sort"
	"strings"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*parentTagLookup)(nil)
	_ generator.TestSnippet = (*parentTagLookup)(nil)
)

type parentTagLookup struct {
	tag *tagLookup
}

func newParentTagLookup(tag *tagLookup) *parentTagLookup {
	return &parentTagLookup{tag: tag}
}

func (l *parentTagLookup) Imports() []string {
	return []string{"sort"}
}

func (l *parentTagLookup) Generate(p *generator.Printer) {
	tagIDMask := (1 << l.tag.idBits) - 1

	bits := 2 * l.tag.idBits
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
	p.Println(`		return tagID(l[i]>>`, l.tag.idBits, `) >= child`)
	p.Println(`	})`)
	p.Println(`	if idx < len(l) && tagID(l[idx]>>`, l.tag.idBits, `) == child {`)
	p.Println(`		return tagID(l[idx] & `, fmt.Sprintf("%#x", tagIDMask), `)`)
	p.Println(`	}`)
	p.Println(`	return 0`)
	p.Println(`}`)
}

func (l *parentTagLookup) TestImports() []string {
	return nil
}

func (l *parentTagLookup) GenerateTest(p *generator.Printer) {
	parentTag := func(child, parent uint) string {
		val := (child << l.tag.idBits) | parent
		return fmt.Sprintf("%#0[2]*[1]x", val, l.tag.idBits/2)
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

var (
	_ generator.Snippet     = (*parentTagLookupVar)(nil)
	_ generator.TestSnippet = (*parentTagLookupVar)(nil)
)

type parentTagLookupVar struct {
	name string
	typ  *parentTagLookup
	tags *tagLookupVar
	data []parentTagData
}

func newParentTagLookupVar(name string, typ *parentTagLookup, tags *tagLookupVar, data *cldr.Data) *parentTagLookupVar {
	parentData := make([]parentTagData, 0, 8)
	childMap := map[cldr.Identity]struct{}{}
	forEachParentIdentity(data, func(data parentTagData) {
		if _, has := childMap[data.child]; !has {
			parentData = append(parentData, data)
			childMap[data.child] = struct{}{}
		}
	})

	sort.Slice(parentData, func(i, j int) bool {
		return identityLess(parentData[i].child, parentData[j].child)
	})

	return &parentTagLookupVar{
		name: name,
		typ:  typ,
		tags: tags,
		data: parentData,
	}
}

func (v *parentTagLookupVar) containsParent(child cldr.Identity) bool {
	for _, data := range v.data {
		if data.child == child {
			return true
		}
	}
	return false
}

func (v *parentTagLookupVar) Imports() []string {
	return nil
}

func (v *parentTagLookupVar) Generate(p *generator.Printer) {
	bits := 2 * v.tags.typ.idBits
	switch {
	case bits <= 16:
		bits = 16
	case bits <= 32:
		bits = 32
	default:
		bits = 64
	}

	digitsPerElem := bits / 4
	perLine := int(lineLength / (digitsPerElem + 4)) // additional "0x" and ", "

	hex := func(data parentTagData) string {
		childID := v.tags.tagID(data.child)
		parentID := v.tags.tagID(data.parent)
		return fmt.Sprintf("%#0[2]*[1]x", (childID<<v.tags.typ.idBits)|parentID, digitsPerElem)
	}

	p.Println(`var `, v.name, ` = parentTagLookup{ // `, len(v.data), ` items, `, uint(len(v.data))*bits/8, ` bytes`)

	for i := 0; i < len(v.data); i += perLine {
		n := i + perLine
		if n > len(v.data) {
			n = len(v.data)
		}

		p.Print(`	`, hex(v.data[i]), `, `)
		for k := i + 1; k < n; k++ {
			p.Print(hex(v.data[k]), `, `)
		}
		p.Print(`// `)
		for k := i; k < n; k++ {
			p.Print(v.data[k].child.String(), ` -> `, v.data[k].parent.String())
			if k < n-1 {
				p.Print(`, `)
			}
		}
		p.Println()

	}

	p.Println(`}`)
}

func (v *parentTagLookupVar) TestImports() []string {
	return nil
}

func (v *parentTagLookupVar) GenerateTest(p *generator.Printer) {
	tagID := func(id cldr.Identity) string {
		return fmt.Sprintf("%#0[2]*[1]x", v.tags.tagID(id), v.tags.typ.idBits/4)
	}

	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[tagID]tagID{`)

	perLine := int(lineLength / (2 + 2*v.tags.typ.idBits/4))
	for i := 0; i < len(v.data); i += perLine {
		n := i + perLine
		if n > len(v.data) {
			n = len(v.data)
		}

		p.Print(`		`, tagID(v.data[i].child), `: `, tagID(v.data[i].parent))
		for k := i + 1; k < n; k++ {
			p.Print(`, `, tagID(v.data[k].child), `: `, tagID(v.data[k].parent))
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
