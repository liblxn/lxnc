package generate_cldr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*regionContainmentLookup)(nil)
	_ generator.TestSnippet = (*regionContainmentLookup)(nil)
)

type regionContainmentLookup struct {
	region *regionLookup
}

func newRegionContainmentLookup(region *regionLookup) *regionContainmentLookup {
	return &regionContainmentLookup{
		region: region,
	}
}

func (l *regionContainmentLookup) Imports() []string {
	imp := make([]string, 1, 2)
	imp[0] = "sort"
	if l.region.idBits > 8 {
		imp = append(imp, "encoding/binary")
	}
	return imp
}

func (l *regionContainmentLookup) Generate(p *generator.Printer) {
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
	p.Println(`	blocksize := 2 + `, l.region.idBits/8, `*len(parents)`)
	p.Println(`	idx := sort.Search(len(l)/blocksize, func(i int) bool {`)
	p.Println(`		i *= blocksize`)
	p.Println(`		return l[i:i+2] >= regionContainmentLookup(region)`)
	p.Println(`	})`)
	p.Println()
	p.Println(`	idx *= blocksize`)
	p.Println(`	if idx < len(l) && l[idx:idx+2] == regionContainmentLookup(region) {`)
	p.Println(`		for i := 0; i < len(parents); i++ {`)
	p.Println(`			start := idx + i`)
	if l.region.idBits <= 8 {
		p.Println(`			parents[i] = regionID(l[start+2])`)
	} else {
		p.Println(`			parents[i] = regionID(binary.Uint`, l.region.idBits, `([]byte(l[start+2:start+blocksize])))`)
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

func (l *regionContainmentLookup) TestImports() []string {
	return nil
}

func (l *regionContainmentLookup) GenerateTest(p *generator.Printer) {
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

var (
	_ generator.Snippet     = (*regionContainmentLookupVar)(nil)
	_ generator.TestSnippet = (*regionContainmentLookupVar)(nil)
)

type regionContainmentLookupVar struct {
	name       string
	typ        *regionContainmentLookup
	regions    *regionLookupVar
	data       []regionContainmentData
	maxParents uint
}

func newRegionContainmentLookupVar(name string, typ *regionContainmentLookup, regions *regionLookupVar, data *cldr.Data) *regionContainmentLookupVar {
	regionData := make([]regionContainmentData, 0, 8)
	childMap := map[string]struct{}{}
	maxParents := uint(0)
	forEachRegionContainment(data, func(rcd regionContainmentData) {
		if _, has := childMap[rcd.childRegion]; has {
			panic(fmt.Sprintf("containment for child region %s already exists", rcd.childRegion))
		}
		regionData = append(regionData, rcd)
		childMap[rcd.childRegion] = struct{}{}
		if uint(len(rcd.parentRegions)) > maxParents {
			maxParents = uint(len(rcd.parentRegions))
		}
	})

	sort.Slice(regionData, func(i, j int) bool {
		return regionData[i].childRegion < regionData[j].childRegion
	})

	return &regionContainmentLookupVar{
		name:       name,
		typ:        typ,
		regions:    regions,
		data:       regionData,
		maxParents: maxParents,
	}
}

func (v *regionContainmentLookupVar) childrenOf(parent string) []string {
	var children []string
	for _, data := range v.data {
		for _, p := range data.parentRegions {
			if p == parent {
				children = append(children, data.childRegion)
				break
			}
		}
	}
	return children
}

func (v *regionContainmentLookupVar) Imports() []string {
	return nil
}

func (v *regionContainmentLookupVar) Generate(p *generator.Printer) {
	var regionID func(id uint) string
	switch {
	case v.regions.typ.idBits <= 8:
		regionID = func(id uint) string {
			return fmt.Sprintf(`\x%02x`, id)
		}
	case v.regions.typ.idBits <= 16:
		regionID = func(id uint) string {
			var buf [2]byte
			binary.BigEndian.PutUint16(buf[:], uint16(id))
			return string(buf[:])
		}
	case v.regions.typ.idBits <= 32:
		regionID = func(id uint) string {
			var buf [4]byte
			binary.BigEndian.PutUint32(buf[:], uint32(id))
			return string(buf[:])
		}
	case v.regions.typ.idBits <= 64:
		regionID = func(id uint) string {
			var buf [8]byte
			binary.BigEndian.PutUint64(buf[:], uint64(id))
			return string(buf[:])
		}
	default:
		panic("invalid region id bits")
	}

	var buf bytes.Buffer
	newContainment := func(data regionContainmentData) string {
		buf.Reset()
		buf.WriteString(data.childRegion)

		i := 0
		for ; i < len(data.parentRegions); i++ {
			parentID := v.regions.regionID(data.parentRegions[i])
			buf.WriteString(regionID(parentID))
		}
		for ; uint(i) < v.maxParents; i++ {
			buf.WriteString(regionID(0))
		}
		return buf.String()
	}

	parentBlocksize := v.maxParents * v.regions.typ.idBits / 8
	blocksize := 2 + parentBlocksize                     // 2-letter region code + v.maxParents subtag ids
	perLine := int(lineLength / (2 + 4*parentBlocksize)) // "\xff" for each subtag id

	p.Println(`const `, v.name, ` regionContainmentLookup = "" + // `, len(v.data), ` items, `, uint(len(v.data))*blocksize, ` bytes`)

	data := v.data
	for len(data) > perLine {
		p.Print(`	"`)
		for i := 0; i < perLine; i++ {
			p.Print(newContainment(data[i]))
		}
		p.Println(`" +`)

		data = data[perLine:]
	}

	p.Print(`	"`)
	for i := 0; i < len(data); i++ {
		p.Print(newContainment(data[i]))
	}
	p.Println(`"`)
}

func (v *regionContainmentLookupVar) TestImports() []string {
	return []string{"reflect"}
}

func (v *regionContainmentLookupVar) GenerateTest(p *generator.Printer) {
	p.Println(`func Test`, strings.Title(v.name), `(t *testing.T) {`)
	p.Println(`	expected := map[string][]regionID{ // child region => parent region ids`)

	fmtRegionID := func(id uint) string {
		return fmt.Sprintf("%#0[2]*[1]x", id, v.regions.typ.idBits/4)
	}

	printParents := func(parents []string) {
		i := 0
		p.Print(`{`)
		p.Print(fmtRegionID(v.regions.regionID(parents[i])))
		for i++; i < len(parents); i++ {
			p.Print(`, `, fmtRegionID(v.regions.regionID(parents[i])))
		}
		for ; uint(i) < v.maxParents; i++ {
			p.Print(`, `, fmtRegionID(0))
		}
		p.Print(`}`)
	}

	perLine := int(lineLength / (6 + v.maxParents*4))
	for i := 0; i < len(v.data); i += perLine {
		n := i + perLine
		if n > len(v.data) {
			n = len(v.data)
		}

		data := v.data[i]
		p.Print(`		"`, data.childRegion, `": `)
		printParents(data.parentRegions)
		for k := i + 1; k < n; k++ {
			data = v.data[k]
			p.Print(`, "`, data.childRegion, `": `)
			printParents(data.parentRegions)
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
