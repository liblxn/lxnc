package cldr

import (
	"encoding/xml"
	"io"
)

// Decode decodes the data from the given CLDR directory.
func Decode(directory string) (*Data, error) {
	return decodeData(newOsFileTree(directory))
}

// DecodeZip decodes the data from the given CLDR zip file.
func DecodeZip(zipFile string) (*Data, error) {
	tree, err := newZipFileTree(zipFile)
	if err != nil {
		return nil, err
	}
	return decodeData(tree)
}

// Data contains all the relevant data which is decoded from the CLDR repository.
type Data struct {
	Identities       map[string]Identity // locale => identity
	Numbers          map[string]Numbers  // locale => numbers
	NumberingSystems NumberingSystems
	Plurals          Plurals
	Regions          Regions
	LikelySubtags    LikelySubtags
	ParentIdentities ParentIdentities
}

// ParentIdentity returns the parent identity of id. The parent is determined by the <parentLocale>
// data. If this data does not define any parent relationship, id will be truncated.
func (data *Data) ParentIdentity(id Identity) Identity {
	if locale, has := data.ParentIdentities[id.String()]; has {
		if parent, has := data.Identities[locale]; has {
			return parent
		}
	}
	return id.Truncate()
}

// DefaultNumberingSystem return the default numbering system for the given identity.
func (data *Data) DefaultNumberingSystem(id Identity) string {
	for {
		numbers, has := data.Numbers[id.String()]
		if id.IsRoot() || (has && numbers.DefaultSystem != "") {
			return numbers.DefaultSystem
		}
		id = data.ParentIdentity(id)
	}
}

// FullNumberSymbols returns the number symbols filled with all available data.
func (data *Data) FullNumberSymbols(id Identity, numberingSystem string) NumberSymbols {
	symbols := data.Numbers[id.String()].Symbols[numberingSystem]
	for !id.IsRoot() {
		id = data.ParentIdentity(id)
		symbols.merge(data.Numbers[id.String()].Symbols[numberingSystem])
	}
	return symbols
}

func decodeData(f fileTree) (*Data, error) {
	dirs := [...]string{
		"common/main",
		"common/supplemental",
	}

	data := &Data{
		Identities: make(map[string]Identity),
		Numbers:    make(map[string]Numbers),
	}

	for _, dir := range dirs {
		err := f.Walk(dir, func(path string, r io.Reader) error {
			return decodeXML(path, r, data.decode)
		})
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (data *Data) decode(d *xmlDecoder, root xml.StartElement) {
	switch root.Name.Local {
	case "ldml":
		data.decodeLDML(d, root)
	case "supplementalData":
		data.decodeSupplemental(d, root)
	}
}

func (data *Data) decodeLDML(d *xmlDecoder, _ xml.StartElement) {
	var identity Identity
	var numbers Numbers
	d.DecodeElems(decoders{
		"identity": identity.decode,
		"numbers":  numbers.decode,
	})

	if !identity.empty() {
		loc := identity.String()
		data.Identities[loc] = identity
		if !numbers.empty() {
			data.Numbers[loc] = numbers
		}
	}
}

func (data *Data) decodeSupplemental(d *xmlDecoder, _ xml.StartElement) {
	d.DecodeElems(decoders{
		"numberingSystems":     data.NumberingSystems.decode,
		"plurals":              data.Plurals.decode,
		"territoryContainment": data.Regions.decode,
		"likelySubtags":        data.LikelySubtags.decode,
		"parentLocales":        data.ParentIdentities.decode,
	})
}
