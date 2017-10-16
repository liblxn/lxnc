package locale

import (
	"sort"
)

// A lang id is an identifier of a specific fixed-width string and defines
// a 1-based index into a lookup string. The lookup consists of concatenated
// blocks of size 3, where each block contains a lang string.
type langID uint8

type langLookup string

func (l langLookup) lang(id langID) string {
	if id == 0 || 3*int(id) > len(l) {
		return ""
	}

	code := l[int(id-1)*3 : int(id)*3]
	end := 3
	for end > 0 && code[end-1] == ' ' {
		end--
	}
	return string(code[:end])
}

func (l langLookup) langID(str []byte) langID {
	idx := sort.Search(len(l)/3, func(i int) bool {
		return l[i*3:(i+1)*3] >= langLookup(str)
	})

	if idx*3 < len(l) && l.lang(langID(idx+1)) == string(str) {
		return langID(idx + 1)
	}
	return 0
}

// A script id is an identifier of a specific fixed-width string and defines
// a 1-based index into a lookup string. The lookup consists of concatenated
// blocks of size 4, where each block contains a script string.
type scriptID uint8

type scriptLookup string

func (l scriptLookup) script(id scriptID) string {
	if id == 0 || 4*int(id) > len(l) {
		return ""
	}

	code := l[int(id-1)*4 : int(id)*4]
	end := 4
	for end > 0 && code[end-1] == ' ' {
		end--
	}
	return string(code[:end])
}

func (l scriptLookup) scriptID(str []byte) scriptID {
	idx := sort.Search(len(l)/4, func(i int) bool {
		return l[i*4:(i+1)*4] >= scriptLookup(str)
	})

	if idx*4 < len(l) && l.script(scriptID(idx+1)) == string(str) {
		return scriptID(idx + 1)
	}
	return 0
}

// A region id is an identifier of a specific fixed-width string and defines
// a 1-based index into a lookup string. The lookup consists of concatenated
// blocks of size 3, where each block contains a region string.
type regionID uint8

type regionLookup string

func (l regionLookup) region(id regionID) string {
	if id == 0 || 3*int(id) > len(l) {
		return ""
	}

	code := l[int(id-1)*3 : int(id)*3]
	end := 3
	for end > 0 && code[end-1] == ' ' {
		end--
	}
	return string(code[:end])
}

func (l regionLookup) regionID(str []byte) regionID {
	idx := sort.Search(len(l)/3, func(i int) bool {
		return l[i*3:(i+1)*3] >= regionLookup(str)
	})

	if idx*3 < len(l) && l.region(regionID(idx+1)) == string(str) {
		return regionID(idx + 1)
	}
	return 0
}

// A tag is a tuple consisting of language subtag, the script subtag,
// and the region subtag. The tag lookup is an ordered list of tags and the
// tag id is an 1-based index of this list.
type tag uint32

func (t tag) langID() langID {
	return langID((t >> 16) & 0xff)
}

func (t tag) scriptID() scriptID {
	return scriptID((t >> 8) & 0xff)
}

func (t tag) regionID() regionID {
	return regionID(t & 0xff)
}

type tagID uint16

type tagLookup []tag

func (l tagLookup) tag(id tagID) tag {
	if id == 0 || int(id) > len(l) {
		return 0
	}
	return l[id-1]
}

func (l tagLookup) tagID(lang langID, script scriptID, region regionID) tagID {
	t := (tag(lang) << 16) | (tag(script) << 8) | tag(region)
	idx := sort.Search(len(l), func(i int) bool {
		return l[i] >= t
	})

	if idx < len(l) && l[idx] == t {
		return tagID(idx + 1)
	}
	return 0
}

// A region containment is a tuple consisting of a 2-letter region code and
// a list of region subtag ids. The lookup maps an alphabetic region code (the child)
// to a list of containing regions (the parents). Each entry in the mapping is
// encoded as "RR\xnn\xp1...\xpn", where RR is the child region code, n the number
// of parent subtag ids and p1, ..., pn the parent subtag ids.
type regionContainmentLookup string

func (l regionContainmentLookup) containmentIDs(region []byte, parents []regionID) int {
	// The length of parents specifies the number of containment ids per block.
	if len(region) != 2 {
		return 0
	}
	blocksize := 2 + 1*len(parents)
	idx := sort.Search(len(l)/blocksize, func(i int) bool {
		i *= blocksize
		return l[i:i+2] >= regionContainmentLookup(region)
	})

	idx *= blocksize
	if idx < len(l) && l[idx:idx+2] == regionContainmentLookup(region) {
		for i := 0; i < len(parents); i++ {
			start := idx + i
			parents[i] = regionID(l[start+2])
			if parents[i] == 0 {
				return i
			}
		}
		return len(parents)
	}
	return 0
}

// The parent tag lookup is an ordered list of tag id pairs. Each pair consists
// of a child id and a parent id.
type parentTagLookup []uint32 // tag id => tag id

func (l parentTagLookup) parentID(child tagID) tagID {
	idx := sort.Search(len(l), func(i int) bool {
		return tagID(l[i]>>16) >= child
	})
	if idx < len(l) && tagID(l[idx]>>16) == child {
		return tagID(l[idx] & 0xffff)
	}
	return 0
}
