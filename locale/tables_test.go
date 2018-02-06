package locale

import (
	"reflect"
	"strings"
	"testing"
)

func TestLangTags(t *testing.T) {
	expected := map[langID]string{ // lang id => string
		0x01: "af ", 0x02: "agq", 0x03: "ak ", 0x04: "am ", 0x05: "ar ", 0x06: "as ",
		0x07: "asa", 0x08: "ast", 0x09: "az ", 0x0a: "bas", 0x0b: "be ", 0x0c: "bem",
		0x0d: "bez", 0x0e: "bg ", 0x0f: "bm ", 0x10: "bn ", 0x11: "bo ", 0x12: "br ",
		0x13: "brx", 0x14: "bs ", 0x15: "ca ", 0x16: "ce ", 0x17: "cgg", 0x18: "chr",
		0x19: "ckb", 0x1a: "cs ", 0x1b: "cu ", 0x1c: "cy ", 0x1d: "da ", 0x1e: "dav",
		0x1f: "de ", 0x20: "dje", 0x21: "dsb", 0x22: "dua", 0x23: "dyo", 0x24: "dz ",
		0x25: "ebu", 0x26: "ee ", 0x27: "el ", 0x28: "en ", 0x29: "eo ", 0x2a: "es ",
		0x2b: "et ", 0x2c: "eu ", 0x2d: "ewo", 0x2e: "fa ", 0x2f: "ff ", 0x30: "fi ",
		0x31: "fil", 0x32: "fo ", 0x33: "fr ", 0x34: "fur", 0x35: "fy ", 0x36: "ga ",
		0x37: "gd ", 0x38: "gl ", 0x39: "gsw", 0x3a: "gu ", 0x3b: "guz", 0x3c: "gv ",
		0x3d: "ha ", 0x3e: "haw", 0x3f: "he ", 0x40: "hi ", 0x41: "hr ", 0x42: "hsb",
		0x43: "hu ", 0x44: "hy ", 0x45: "id ", 0x46: "ig ", 0x47: "ii ", 0x48: "is ",
		0x49: "it ", 0x4a: "ja ", 0x4b: "jgo", 0x4c: "jmc", 0x4d: "ka ", 0x4e: "kab",
		0x4f: "kam", 0x50: "kde", 0x51: "kea", 0x52: "khq", 0x53: "ki ", 0x54: "kk ",
		0x55: "kkj", 0x56: "kl ", 0x57: "kln", 0x58: "km ", 0x59: "kn ", 0x5a: "ko ",
		0x5b: "kok", 0x5c: "ks ", 0x5d: "ksb", 0x5e: "ksf", 0x5f: "ksh", 0x60: "kw ",
		0x61: "ky ", 0x62: "lag", 0x63: "lb ", 0x64: "lg ", 0x65: "lkt", 0x66: "ln ",
		0x67: "lo ", 0x68: "lrc", 0x69: "lt ", 0x6a: "lu ", 0x6b: "luo", 0x6c: "luy",
		0x6d: "lv ", 0x6e: "mas", 0x6f: "mer", 0x70: "mfe", 0x71: "mg ", 0x72: "mgh",
		0x73: "mgo", 0x74: "mk ", 0x75: "ml ", 0x76: "mn ", 0x77: "mr ", 0x78: "ms ",
		0x79: "mt ", 0x7a: "mua", 0x7b: "my ", 0x7c: "mzn", 0x7d: "naq", 0x7e: "nb ",
		0x7f: "nd ", 0x80: "nds", 0x81: "ne ", 0x82: "nl ", 0x83: "nmg", 0x84: "nn ",
		0x85: "nnh", 0x86: "nus", 0x87: "nyn", 0x88: "om ", 0x89: "or ", 0x8a: "os ",
		0x8b: "pa ", 0x8c: "pl ", 0x8d: "prg", 0x8e: "ps ", 0x8f: "pt ", 0x90: "qu ",
		0x91: "rm ", 0x92: "rn ", 0x93: "ro ", 0x94: "rof", 0x95: "ru ", 0x96: "rw ",
		0x97: "rwk", 0x98: "sah", 0x99: "saq", 0x9a: "sbp", 0x9b: "se ", 0x9c: "seh",
		0x9d: "ses", 0x9e: "sg ", 0x9f: "shi", 0xa0: "si ", 0xa1: "sk ", 0xa2: "sl ",
		0xa3: "smn", 0xa4: "sn ", 0xa5: "so ", 0xa6: "sq ", 0xa7: "sr ", 0xa8: "sv ",
		0xa9: "sw ", 0xaa: "ta ", 0xab: "te ", 0xac: "teo", 0xad: "th ", 0xae: "ti ",
		0xaf: "tk ", 0xb0: "to ", 0xb1: "tr ", 0xb2: "twq", 0xb3: "tzm", 0xb4: "ug ",
		0xb5: "uk ", 0xb6: "und", 0xb7: "ur ", 0xb8: "uz ", 0xb9: "vai", 0xba: "vi ",
		0xbb: "vo ", 0xbc: "vun", 0xbd: "wae", 0xbe: "xog", 0xbf: "yav", 0xc0: "yi ",
		0xc1: "yo ", 0xc2: "yue", 0xc3: "zgh", 0xc4: "zh ", 0xc5: "zu ",
	}

	for id, str := range expected {
		if s := langTags.lang(id); s != strings.TrimSpace(str) {
			t.Fatalf("unexpected string for id %d: %q", uint(id), s)
		}
	}
}

func TestScriptTags(t *testing.T) {
	expected := map[scriptID]string{ // script id => string
		0x01: "Arab", 0x02: "Cyrl", 0x03: "Guru", 0x04: "Hans", 0x05: "Hant",
		0x06: "Latn", 0x07: "Tfng", 0x08: "Vaii",
	}

	for id, str := range expected {
		if s := scriptTags.script(id); s != strings.TrimSpace(str) {
			t.Fatalf("unexpected string for id %d: %q", uint(id), s)
		}
	}
}

func TestRegionTags(t *testing.T) {
	expected := map[regionID]string{ // region id => string
		0x01: "001", 0x02: "150", 0x03: "419", 0x04: "AD ", 0x05: "AE ", 0x06: "AF ",
		0x07: "AG ", 0x08: "AI ", 0x09: "AL ", 0x0a: "AM ", 0x0b: "AO ", 0x0c: "AR ",
		0x0d: "AS ", 0x0e: "AT ", 0x0f: "AU ", 0x10: "AW ", 0x11: "AX ", 0x12: "AZ ",
		0x13: "BA ", 0x14: "BB ", 0x15: "BD ", 0x16: "BE ", 0x17: "BF ", 0x18: "BG ",
		0x19: "BH ", 0x1a: "BI ", 0x1b: "BJ ", 0x1c: "BL ", 0x1d: "BM ", 0x1e: "BN ",
		0x1f: "BO ", 0x20: "BQ ", 0x21: "BR ", 0x22: "BS ", 0x23: "BT ", 0x24: "BW ",
		0x25: "BY ", 0x26: "BZ ", 0x27: "CA ", 0x28: "CC ", 0x29: "CD ", 0x2a: "CF ",
		0x2b: "CG ", 0x2c: "CH ", 0x2d: "CI ", 0x2e: "CK ", 0x2f: "CL ", 0x30: "CM ",
		0x31: "CN ", 0x32: "CO ", 0x33: "CR ", 0x34: "CU ", 0x35: "CV ", 0x36: "CW ",
		0x37: "CX ", 0x38: "CY ", 0x39: "CZ ", 0x3a: "DE ", 0x3b: "DG ", 0x3c: "DJ ",
		0x3d: "DK ", 0x3e: "DM ", 0x3f: "DO ", 0x40: "DZ ", 0x41: "EA ", 0x42: "EC ",
		0x43: "EE ", 0x44: "EG ", 0x45: "EH ", 0x46: "ER ", 0x47: "ES ", 0x48: "ET ",
		0x49: "FI ", 0x4a: "FJ ", 0x4b: "FK ", 0x4c: "FM ", 0x4d: "FO ", 0x4e: "FR ",
		0x4f: "GA ", 0x50: "GB ", 0x51: "GD ", 0x52: "GE ", 0x53: "GF ", 0x54: "GG ",
		0x55: "GH ", 0x56: "GI ", 0x57: "GL ", 0x58: "GM ", 0x59: "GN ", 0x5a: "GP ",
		0x5b: "GQ ", 0x5c: "GR ", 0x5d: "GT ", 0x5e: "GU ", 0x5f: "GW ", 0x60: "GY ",
		0x61: "HK ", 0x62: "HN ", 0x63: "HR ", 0x64: "HT ", 0x65: "HU ", 0x66: "IC ",
		0x67: "ID ", 0x68: "IE ", 0x69: "IL ", 0x6a: "IM ", 0x6b: "IN ", 0x6c: "IO ",
		0x6d: "IQ ", 0x6e: "IR ", 0x6f: "IS ", 0x70: "IT ", 0x71: "JE ", 0x72: "JM ",
		0x73: "JO ", 0x74: "JP ", 0x75: "KE ", 0x76: "KG ", 0x77: "KH ", 0x78: "KI ",
		0x79: "KM ", 0x7a: "KN ", 0x7b: "KP ", 0x7c: "KR ", 0x7d: "KW ", 0x7e: "KY ",
		0x7f: "KZ ", 0x80: "LA ", 0x81: "LB ", 0x82: "LC ", 0x83: "LI ", 0x84: "LK ",
		0x85: "LR ", 0x86: "LS ", 0x87: "LT ", 0x88: "LU ", 0x89: "LV ", 0x8a: "LY ",
		0x8b: "MA ", 0x8c: "MC ", 0x8d: "MD ", 0x8e: "ME ", 0x8f: "MF ", 0x90: "MG ",
		0x91: "MH ", 0x92: "MK ", 0x93: "ML ", 0x94: "MM ", 0x95: "MN ", 0x96: "MO ",
		0x97: "MP ", 0x98: "MQ ", 0x99: "MR ", 0x9a: "MS ", 0x9b: "MT ", 0x9c: "MU ",
		0x9d: "MW ", 0x9e: "MX ", 0x9f: "MY ", 0xa0: "MZ ", 0xa1: "NA ", 0xa2: "NC ",
		0xa3: "NE ", 0xa4: "NF ", 0xa5: "NG ", 0xa6: "NI ", 0xa7: "NL ", 0xa8: "NO ",
		0xa9: "NP ", 0xaa: "NR ", 0xab: "NU ", 0xac: "NZ ", 0xad: "OM ", 0xae: "PA ",
		0xaf: "PE ", 0xb0: "PF ", 0xb1: "PG ", 0xb2: "PH ", 0xb3: "PK ", 0xb4: "PL ",
		0xb5: "PM ", 0xb6: "PN ", 0xb7: "PR ", 0xb8: "PS ", 0xb9: "PT ", 0xba: "PW ",
		0xbb: "PY ", 0xbc: "QA ", 0xbd: "RE ", 0xbe: "RO ", 0xbf: "RS ", 0xc0: "RU ",
		0xc1: "RW ", 0xc2: "SA ", 0xc3: "SB ", 0xc4: "SC ", 0xc5: "SD ", 0xc6: "SE ",
		0xc7: "SG ", 0xc8: "SH ", 0xc9: "SI ", 0xca: "SJ ", 0xcb: "SK ", 0xcc: "SL ",
		0xcd: "SM ", 0xce: "SN ", 0xcf: "SO ", 0xd0: "SR ", 0xd1: "SS ", 0xd2: "ST ",
		0xd3: "SV ", 0xd4: "SX ", 0xd5: "SY ", 0xd6: "SZ ", 0xd7: "TC ", 0xd8: "TD ",
		0xd9: "TG ", 0xda: "TH ", 0xdb: "TK ", 0xdc: "TL ", 0xdd: "TM ", 0xde: "TN ",
		0xdf: "TO ", 0xe0: "TR ", 0xe1: "TT ", 0xe2: "TV ", 0xe3: "TW ", 0xe4: "TZ ",
		0xe5: "UA ", 0xe6: "UG ", 0xe7: "UM ", 0xe8: "US ", 0xe9: "UY ", 0xea: "UZ ",
		0xeb: "VA ", 0xec: "VC ", 0xed: "VE ", 0xee: "VG ", 0xef: "VI ", 0xf0: "VN ",
		0xf1: "VU ", 0xf2: "WF ", 0xf3: "WS ", 0xf4: "XK ", 0xf5: "YE ", 0xf6: "YT ",
		0xf7: "ZA ", 0xf8: "ZM ", 0xf9: "ZW ",
	}

	for id, str := range expected {
		if s := regionTags.region(id); s != strings.TrimSpace(str) {
			t.Fatalf("unexpected string for id %d: %q", uint(id), s)
		}
	}
}

func TestRegionContainment(t *testing.T) {
	expected := map[string][]regionID{ // child region => parent region ids
		"AC": {0x01, 0x00}, "AD": {0x02, 0x01}, "AE": {0x01, 0x00}, "AF": {0x01, 0x00},
		"AG": {0x03, 0x01}, "AI": {0x03, 0x01}, "AL": {0x02, 0x01}, "AM": {0x01, 0x00},
		"AO": {0x01, 0x00}, "AQ": {0x01, 0x00}, "AR": {0x03, 0x01}, "AS": {0x01, 0x00},
		"AT": {0x02, 0x01}, "AU": {0x01, 0x00}, "AW": {0x03, 0x01}, "AX": {0x02, 0x01},
		"AZ": {0x01, 0x00}, "BA": {0x02, 0x01}, "BB": {0x03, 0x01}, "BD": {0x01, 0x00},
		"BE": {0x02, 0x01}, "BF": {0x01, 0x00}, "BG": {0x02, 0x01}, "BH": {0x01, 0x00},
		"BI": {0x01, 0x00}, "BJ": {0x01, 0x00}, "BL": {0x03, 0x01}, "BM": {0x01, 0x00},
		"BN": {0x01, 0x00}, "BO": {0x03, 0x01}, "BQ": {0x03, 0x01}, "BR": {0x03, 0x01},
		"BS": {0x03, 0x01}, "BT": {0x01, 0x00}, "BV": {0x01, 0x00}, "BW": {0x01, 0x00},
		"BY": {0x02, 0x01}, "BZ": {0x03, 0x01}, "CA": {0x01, 0x00}, "CC": {0x01, 0x00},
		"CD": {0x01, 0x00}, "CF": {0x01, 0x00}, "CG": {0x01, 0x00}, "CH": {0x02, 0x01},
		"CI": {0x01, 0x00}, "CK": {0x01, 0x00}, "CL": {0x03, 0x01}, "CM": {0x01, 0x00},
		"CN": {0x01, 0x00}, "CO": {0x03, 0x01}, "CP": {0x01, 0x00}, "CR": {0x03, 0x01},
		"CU": {0x03, 0x01}, "CV": {0x01, 0x00}, "CW": {0x03, 0x01}, "CX": {0x01, 0x00},
		"CY": {0x01, 0x00}, "CZ": {0x02, 0x01}, "DE": {0x02, 0x01}, "DG": {0x01, 0x00},
		"DJ": {0x01, 0x00}, "DK": {0x02, 0x01}, "DM": {0x03, 0x01}, "DO": {0x03, 0x01},
		"DZ": {0x01, 0x00}, "EA": {0x01, 0x00}, "EC": {0x03, 0x01}, "EE": {0x02, 0x01},
		"EG": {0x01, 0x00}, "EH": {0x01, 0x00}, "ER": {0x01, 0x00}, "ES": {0x02, 0x01},
		"ET": {0x01, 0x00}, "FI": {0x02, 0x01}, "FJ": {0x01, 0x00}, "FK": {0x03, 0x01},
		"FM": {0x01, 0x00}, "FO": {0x02, 0x01}, "FR": {0x02, 0x01}, "GA": {0x01, 0x00},
		"GB": {0x02, 0x01}, "GD": {0x03, 0x01}, "GE": {0x01, 0x00}, "GF": {0x03, 0x01},
		"GG": {0x02, 0x01}, "GH": {0x01, 0x00}, "GI": {0x02, 0x01}, "GL": {0x01, 0x00},
		"GM": {0x01, 0x00}, "GN": {0x01, 0x00}, "GP": {0x03, 0x01}, "GQ": {0x01, 0x00},
		"GR": {0x02, 0x01}, "GS": {0x01, 0x00}, "GT": {0x03, 0x01}, "GU": {0x01, 0x00},
		"GW": {0x01, 0x00}, "GY": {0x03, 0x01}, "HK": {0x01, 0x00}, "HM": {0x01, 0x00},
		"HN": {0x03, 0x01}, "HR": {0x02, 0x01}, "HT": {0x03, 0x01}, "HU": {0x02, 0x01},
		"IC": {0x01, 0x00}, "ID": {0x01, 0x00}, "IE": {0x02, 0x01}, "IL": {0x01, 0x00},
		"IM": {0x02, 0x01}, "IN": {0x01, 0x00}, "IO": {0x01, 0x00}, "IQ": {0x01, 0x00},
		"IR": {0x01, 0x00}, "IS": {0x02, 0x01}, "IT": {0x02, 0x01}, "JE": {0x02, 0x01},
		"JM": {0x03, 0x01}, "JO": {0x01, 0x00}, "JP": {0x01, 0x00}, "KE": {0x01, 0x00},
		"KG": {0x01, 0x00}, "KH": {0x01, 0x00}, "KI": {0x01, 0x00}, "KM": {0x01, 0x00},
		"KN": {0x03, 0x01}, "KP": {0x01, 0x00}, "KR": {0x01, 0x00}, "KW": {0x01, 0x00},
		"KY": {0x03, 0x01}, "KZ": {0x01, 0x00}, "LA": {0x01, 0x00}, "LB": {0x01, 0x00},
		"LC": {0x03, 0x01}, "LI": {0x02, 0x01}, "LK": {0x01, 0x00}, "LR": {0x01, 0x00},
		"LS": {0x01, 0x00}, "LT": {0x02, 0x01}, "LU": {0x02, 0x01}, "LV": {0x02, 0x01},
		"LY": {0x01, 0x00}, "MA": {0x01, 0x00}, "MC": {0x02, 0x01}, "MD": {0x02, 0x01},
		"ME": {0x02, 0x01}, "MF": {0x03, 0x01}, "MG": {0x01, 0x00}, "MH": {0x01, 0x00},
		"MK": {0x02, 0x01}, "ML": {0x01, 0x00}, "MM": {0x01, 0x00}, "MN": {0x01, 0x00},
		"MO": {0x01, 0x00}, "MP": {0x01, 0x00}, "MQ": {0x03, 0x01}, "MR": {0x01, 0x00},
		"MS": {0x03, 0x01}, "MT": {0x02, 0x01}, "MU": {0x01, 0x00}, "MV": {0x01, 0x00},
		"MW": {0x01, 0x00}, "MX": {0x03, 0x01}, "MY": {0x01, 0x00}, "MZ": {0x01, 0x00},
		"NA": {0x01, 0x00}, "NC": {0x01, 0x00}, "NE": {0x01, 0x00}, "NF": {0x01, 0x00},
		"NG": {0x01, 0x00}, "NI": {0x03, 0x01}, "NL": {0x02, 0x01}, "NO": {0x02, 0x01},
		"NP": {0x01, 0x00}, "NR": {0x01, 0x00}, "NU": {0x01, 0x00}, "NZ": {0x01, 0x00},
		"OM": {0x01, 0x00}, "PA": {0x03, 0x01}, "PE": {0x03, 0x01}, "PF": {0x01, 0x00},
		"PG": {0x01, 0x00}, "PH": {0x01, 0x00}, "PK": {0x01, 0x00}, "PL": {0x02, 0x01},
		"PM": {0x01, 0x00}, "PN": {0x01, 0x00}, "PR": {0x03, 0x01}, "PS": {0x01, 0x00},
		"PT": {0x02, 0x01}, "PW": {0x01, 0x00}, "PY": {0x03, 0x01}, "QA": {0x01, 0x00},
		"RE": {0x01, 0x00}, "RO": {0x02, 0x01}, "RS": {0x02, 0x01}, "RU": {0x02, 0x01},
		"RW": {0x01, 0x00}, "SA": {0x01, 0x00}, "SB": {0x01, 0x00}, "SC": {0x01, 0x00},
		"SD": {0x01, 0x00}, "SE": {0x02, 0x01}, "SG": {0x01, 0x00}, "SH": {0x01, 0x00},
		"SI": {0x02, 0x01}, "SJ": {0x02, 0x01}, "SK": {0x02, 0x01}, "SL": {0x01, 0x00},
		"SM": {0x02, 0x01}, "SN": {0x01, 0x00}, "SO": {0x01, 0x00}, "SR": {0x03, 0x01},
		"SS": {0x01, 0x00}, "ST": {0x01, 0x00}, "SV": {0x03, 0x01}, "SX": {0x03, 0x01},
		"SY": {0x01, 0x00}, "SZ": {0x01, 0x00}, "TA": {0x01, 0x00}, "TC": {0x03, 0x01},
		"TD": {0x01, 0x00}, "TF": {0x01, 0x00}, "TG": {0x01, 0x00}, "TH": {0x01, 0x00},
		"TJ": {0x01, 0x00}, "TK": {0x01, 0x00}, "TL": {0x01, 0x00}, "TM": {0x01, 0x00},
		"TN": {0x01, 0x00}, "TO": {0x01, 0x00}, "TR": {0x01, 0x00}, "TT": {0x03, 0x01},
		"TV": {0x01, 0x00}, "TW": {0x01, 0x00}, "TZ": {0x01, 0x00}, "UA": {0x02, 0x01},
		"UG": {0x01, 0x00}, "UM": {0x01, 0x00}, "US": {0x01, 0x00}, "UY": {0x03, 0x01},
		"UZ": {0x01, 0x00}, "VA": {0x02, 0x01}, "VC": {0x03, 0x01}, "VE": {0x03, 0x01},
		"VG": {0x03, 0x01}, "VI": {0x03, 0x01}, "VN": {0x01, 0x00}, "VU": {0x01, 0x00},
		"WF": {0x01, 0x00}, "WS": {0x01, 0x00}, "XK": {0x02, 0x01}, "YE": {0x01, 0x00},
		"YT": {0x01, 0x00}, "ZA": {0x01, 0x00}, "ZM": {0x01, 0x00}, "ZW": {0x01, 0x00},
	}

	for child, expectedParents := range expected {
		expectedN := 2
		for expectedParents[expectedN-1] == 0 {
			expectedN--
		}

		var parents [2]regionID
		n := regionContainment.containmentIDs([]byte(child), parents[:])
		switch {
		case n != expectedN:
			t.Errorf("unexpected number of parents for %s: %d (expected %d)", child, n, expectedN)
		case reflect.DeepEqual(parents, expectedParents):
			t.Errorf("unexpected parents: %v (expected %v)", parents, expectedParents)
		}
	}
}

func TestZeros(t *testing.T) {
	expected := map[zeroID]string{ // zero id => zero
		0x01: "0", 0x02: "٠", 0x04: "۰", 0x06: "०", 0x09: "০", 0x0c: "༠", 0x0f: "၀",
	}

	for id, expectedZero := range expected {
		if zero := zeros.zero(id); string(zero) != expectedZero {
			t.Fatalf("unexpected zero for id %d: %s", uint(id), string(zero))
		}
	}
}

func TestAffixes(t *testing.T) {
	expected := map[affixID]affix{ // affix id => affix
		0x01: "\x00\x00", 0x02: "\x00\x01%", 0x03: "\x00\x02¤",
		0x04: "\x00\x03 %", 0x05: "\x00\x04 ¤", 0x06: "\x00\x06 ؜¤",
		0x07: "\x01\x01%", 0x08: "\x01\x01-", 0x09: "\x01\x02-%",
		0x0a: "\x01\x03-¤", 0x0b: "\x01\x04- %", 0x0c: "\x01\x05- ¤",
		0x0d: "\x02\x02-%", 0x0e: "\x02\x02¤", 0x0f: "\x03\x03% ",
		0x10: "\x03\x03-¤", 0x11: "\x03\x03¤-", 0x12: "\x03\x07‏ ¤",
		0x13: "\x03\x09؜- ؜¤", 0x14: "\x04\x04% -", 0x15: "\x04\x04-% ",
		0x16: "\x04\x04¤ ", 0x17: "\x04\x08‏- ¤", 0x18: "\x05\x05-¤ ",
		0x19: "\x05\x05¤- ", 0x1a: "\x05\x05¤ -",
	}

	for expectedID, expectedStr := range expected {
		if s := affixes.affix(expectedID); s != expectedStr {
			t.Fatalf("unexpected affix string for id %d: %s", uint(expectedID), s)
		}
	}
}

func TestPatterns(t *testing.T) {
	expected := map[patternID]pattern{
		0x01: 0x0209100320, 0x02: 0x0108103320, 0x03: 0x030a122320, 0x04: 0x0209100330, 0x05: 0x0108103330,
		0x06: 0x050c122330, 0x07: 0x0613122330, 0x08: 0x030a122330, 0x09: 0x040b100330, 0x0a: 0x070d100330,
		0x0b: 0x0f15100330, 0x0c: 0x0f14100330, 0x0d: 0x050c122000, 0x0e: 0x0e10122320, 0x0f: 0x0e10122330,
		0x10: 0x0e11122330, 0x11: 0x0e19122330, 0x12: 0x1618122320, 0x13: 0x1618122330, 0x14: 0x1611122330,
		0x15: 0x1616122330, 0x16: 0x161a122330, 0x17: 0x1217122330,
	}

	for id, expectedPattern := range expected {
		if pattern := patterns.pattern(id); pattern != expectedPattern {
			t.Fatalf("unexpected pattern for id %d: %#x", uint(id), pattern)
		}
	}
}

func TestNumberSymbols(t *testing.T) {
	expected := map[symbolsID]symbols{ // symbols id => symbols
		0x01: "\x01\x02\x03\x04\x07\x09\x0a\x0b,.%-∞ND,.", 0x02: "\x01\x02\x03\x04\x07\x09\x0a\x0b.,%-∞TF.,", 0x03: "\x01\x02\x03\x04\x07\x0a\x0b\x0c,.%-∞NaN,.",
		0x04: "\x01\x02\x03\x04\x07\x0a\x0b\x0c.,%-∞NaN.,", 0x05: "\x01\x02\x03\x04\x07\x0a\x0b\x0c.,%-∞mnn.,", 0x06: "\x01\x02\x03\x04\x07\x10\x11\x12.,%-∞非數值.,",
		0x07: "\x01\x02\x03\x04\x07\x1d\x1e\x1f.,%-∞Терхьаш дац.,", 0x08: "\x01\x02\x03\x04\x07\x2e\x2f\x30.,%-∞ဂဏန်းမဟုတ်သော.,", 0x09: "\x01\x02\x03\x04\x07\x34\x35\x36,.%-∞ບໍ່​ແມ່ນ​ໂຕ​ເລກ,.",
		0x0a: "\x01\x02\x03\x04\x1c\x2b\x2c\x2d.,%-གྲངས་མེདཨང་མད.,", 0x0b: "\x01\x02\x03\x06\x09\x0c\x0d\x0e,.%−∞NaN,.", 0x0c: "\x01\x02\x03\x06\x09\x0f\x10\x11,.%−∞¤¤¤,.",
		0x0d: "\x01\x02\x03\x07\x0a\x0d\x0e\x0f.,%‎-∞NaN.,", 0x0e: "\x01\x02\x03\x0a\x0d\x10\x11\x12.,%‎-‎∞NaN.,", 0x0f: "\x01\x02\x09\x0d\x10\x22\x23\x24,.‎%‎‎-∞ليس رقمًا,.",
		0x10: "\x01\x02\x09\x0d\x10\x22\x23\x24.,‎%‎‎-∞ليس رقمًا.,", 0x11: "\x01\x03\x04\x05\x08\x0a\x0b\x0d, %-∞NS, ", 0x12: "\x01\x03\x04\x05\x08\x0b\x0c\x0d, %-∞NaN,.",
		0x13: "\x01\x03\x04\x05\x08\x0b\x0c\x0e, %-∞NaN, ", 0x14: "\x01\x03\x04\x05\x08\x0b\x0c\x0e, %-∞NaN. ", 0x15: "\x01\x03\x04\x05\x08\x0b\x0c\x0e. %-∞NaN. ",
		0x16: "\x01\x03\x04\x05\x08\x0c\x0d\x0f, %-∞НН, ", 0x17: "\x01\x03\x04\x05\x08\x0e\x0f\x11, %-∞ՈչԹ, ", 0x18: "\x01\x03\x04\x05\x08\x0f\x10\x12, %-∞epiloho, ",
		0x19: "\x01\x03\x04\x05\x08\x11\x12\x14, %-∞san däl, ", 0x1a: "\x01\x03\x04\x05\x08\x18\x19\x1b, %-∞не число, ", 0x1b: "\x01\x03\x04\x05\x08\x18\x19\x1b, %-∞сан емес, ",
		0x1c: "\x01\x03\x04\x05\x08\x18\x19\x1b, %-∞сан эмес, ", 0x1d: "\x01\x03\x04\x05\x08\x1a\x1b\x1d, %-∞haqiqiy son emas, ", 0x1e: "\x01\x03\x04\x05\x08\x24\x25\x27, %-∞чыыһыла буотах, ",
		0x1f: "\x01\x03\x04\x05\x08\x28\x29\x2b, %-∞ҳақиқий сон эмас, ", 0x20: "\x01\x03\x04\x05\x08\x30\x31\x33, %-∞არ არის რიცხვი, ", 0x21: "\x01\x03\x04\x07\x0a\x0d\x0e\x10, %−∞NaN, ",
		0x22: "\x01\x03\x04\x07\x0a\x10\x11\x13, %−∞¤¤¤, ", 0x23: "\x01\x03\x04\x07\x0a\x12\x13\x15, %−∞epäluku, ", 0x24: "\x01\x04\x05\x06\x09\x0c\x0d\x10,’%-∞NaN,’",
		0x25: "\x01\x04\x05\x06\x09\x0c\x0d\x10.’%-∞NaN.’", 0x26: "\x01\x04\x05\x08\x0b\x0e\x0f\x12.’%−∞NaN.’", 0x27: "\x02\x04\x06\x09\x0c\x0f\x11\x13٫٬٪؜-∞NaN٫٬",
		0x28: "\x02\x04\x06\x0d\x10\x13\x15\x17٫٬٪‎-‎∞NaN٫٬", 0x29: "\x02\x04\x08\x0b\x0e\x1c\x1e\x20٫٬٪؜؜-∞ليس رقم٫٬", 0x2a: "\x02\x04\x09\x0f\x12\x1c\x1e\x20٫٬‎٪‎−∞ناعدد٫٬",
	}

	for expectedID, expectedStr := range expected {
		if s := numberSymbols.symbols(expectedID); s != expectedStr {
			t.Fatalf("unexpected symbols string for id %d: %s", uint(expectedID), s)
		}
	}
}

func TestLocaleTags(t *testing.T) {
	expected := map[tagID]tag{ // tag id => tag id
		0x0001: 0x010000, 0x0002: 0x0100a1, 0x0003: 0x0100f7, 0x0004: 0x020000, 0x0005: 0x020030, 0x0006: 0x030000,
		0x0007: 0x030055, 0x0008: 0x040000, 0x0009: 0x040048, 0x000a: 0x050000, 0x000b: 0x050001, 0x000c: 0x050005,
		0x000d: 0x050019, 0x000e: 0x05003c, 0x000f: 0x050040, 0x0010: 0x050044, 0x0011: 0x050045, 0x0012: 0x050046,
		0x0013: 0x050069, 0x0014: 0x05006d, 0x0015: 0x050073, 0x0016: 0x050079, 0x0017: 0x05007d, 0x0018: 0x050081,
		0x0019: 0x05008a, 0x001a: 0x05008b, 0x001b: 0x050099, 0x001c: 0x0500ad, 0x001d: 0x0500b8, 0x001e: 0x0500bc,
		0x001f: 0x0500c2, 0x0020: 0x0500c5, 0x0021: 0x0500cf, 0x0022: 0x0500d1, 0x0023: 0x0500d5, 0x0024: 0x0500d8,
		0x0025: 0x0500de, 0x0026: 0x0500f5, 0x0027: 0x060000, 0x0028: 0x06006b, 0x0029: 0x070000, 0x002a: 0x0700e4,
		0x002b: 0x080000, 0x002c: 0x080047, 0x002d: 0x090000, 0x002e: 0x090200, 0x002f: 0x090212, 0x0030: 0x090600,
		0x0031: 0x090612, 0x0032: 0x0a0000, 0x0033: 0x0a0030, 0x0034: 0x0b0000, 0x0035: 0x0b0025, 0x0036: 0x0c0000,
		0x0037: 0x0c00f8, 0x0038: 0x0d0000, 0x0039: 0x0d00e4, 0x003a: 0x0e0000, 0x003b: 0x0e0018, 0x003c: 0x0f0000,
		0x003d: 0x0f0093, 0x003e: 0x100000, 0x003f: 0x100015, 0x0040: 0x10006b, 0x0041: 0x110000, 0x0042: 0x110031,
		0x0043: 0x11006b, 0x0044: 0x120000, 0x0045: 0x12004e, 0x0046: 0x130000, 0x0047: 0x13006b, 0x0048: 0x140000,
		0x0049: 0x140200, 0x004a: 0x140213, 0x004b: 0x140600, 0x004c: 0x140613, 0x004d: 0x150000, 0x004e: 0x150004,
		0x004f: 0x150047, 0x0050: 0x15004e, 0x0051: 0x150070, 0x0052: 0x160000, 0x0053: 0x1600c0, 0x0054: 0x170000,
		0x0055: 0x1700e6, 0x0056: 0x180000, 0x0057: 0x1800e8, 0x0058: 0x190000, 0x0059: 0x19006d, 0x005a: 0x19006e,
		0x005b: 0x1a0000, 0x005c: 0x1a0039, 0x005d: 0x1b0000, 0x005e: 0x1b00c0, 0x005f: 0x1c0000, 0x0060: 0x1c0050,
		0x0061: 0x1d0000, 0x0062: 0x1d003d, 0x0063: 0x1d0057, 0x0064: 0x1e0000, 0x0065: 0x1e0075, 0x0066: 0x1f0000,
		0x0067: 0x1f000e, 0x0068: 0x1f0016, 0x0069: 0x1f002c, 0x006a: 0x1f003a, 0x006b: 0x1f0070, 0x006c: 0x1f0083,
		0x006d: 0x1f0088, 0x006e: 0x200000, 0x006f: 0x2000a3, 0x0070: 0x210000, 0x0071: 0x21003a, 0x0072: 0x220000,
		0x0073: 0x220030, 0x0074: 0x230000, 0x0075: 0x2300ce, 0x0076: 0x240000, 0x0077: 0x240023, 0x0078: 0x250000,
		0x0079: 0x250075, 0x007a: 0x260000, 0x007b: 0x260055, 0x007c: 0x2600d9, 0x007d: 0x270000, 0x007e: 0x270038,
		0x007f: 0x27005c, 0x0080: 0x280000, 0x0081: 0x280001, 0x0082: 0x280002, 0x0083: 0x280007, 0x0084: 0x280008,
		0x0085: 0x28000d, 0x0086: 0x28000e, 0x0087: 0x28000f, 0x0088: 0x280014, 0x0089: 0x280016, 0x008a: 0x28001a,
		0x008b: 0x28001d, 0x008c: 0x280022, 0x008d: 0x280024, 0x008e: 0x280026, 0x008f: 0x280027, 0x0090: 0x280028,
		0x0091: 0x28002c, 0x0092: 0x28002e, 0x0093: 0x280030, 0x0094: 0x280037, 0x0095: 0x280038, 0x0096: 0x28003a,
		0x0097: 0x28003b, 0x0098: 0x28003d, 0x0099: 0x28003e, 0x009a: 0x280046, 0x009b: 0x280049, 0x009c: 0x28004a,
		0x009d: 0x28004b, 0x009e: 0x28004c, 0x009f: 0x280050, 0x00a0: 0x280051, 0x00a1: 0x280054, 0x00a2: 0x280055,
		0x00a3: 0x280056, 0x00a4: 0x280058, 0x00a5: 0x28005e, 0x00a6: 0x280060, 0x00a7: 0x280061, 0x00a8: 0x280068,
		0x00a9: 0x280069, 0x00aa: 0x28006a, 0x00ab: 0x28006b, 0x00ac: 0x28006c, 0x00ad: 0x280071, 0x00ae: 0x280072,
		0x00af: 0x280075, 0x00b0: 0x280078, 0x00b1: 0x28007a, 0x00b2: 0x28007e, 0x00b3: 0x280082, 0x00b4: 0x280085,
		0x00b5: 0x280086, 0x00b6: 0x280090, 0x00b7: 0x280091, 0x00b8: 0x280096, 0x00b9: 0x280097, 0x00ba: 0x28009a,
		0x00bb: 0x28009b, 0x00bc: 0x28009c, 0x00bd: 0x28009d, 0x00be: 0x28009f, 0x00bf: 0x2800a1, 0x00c0: 0x2800a4,
		0x00c1: 0x2800a5, 0x00c2: 0x2800a7, 0x00c3: 0x2800aa, 0x00c4: 0x2800ab, 0x00c5: 0x2800ac, 0x00c6: 0x2800b1,
		0x00c7: 0x2800b2, 0x00c8: 0x2800b3, 0x00c9: 0x2800b6, 0x00ca: 0x2800b7, 0x00cb: 0x2800ba, 0x00cc: 0x2800c1,
		0x00cd: 0x2800c3, 0x00ce: 0x2800c4, 0x00cf: 0x2800c5, 0x00d0: 0x2800c6, 0x00d1: 0x2800c7, 0x00d2: 0x2800c8,
		0x00d3: 0x2800c9, 0x00d4: 0x2800cc, 0x00d5: 0x2800d1, 0x00d6: 0x2800d4, 0x00d7: 0x2800d6, 0x00d8: 0x2800d7,
		0x00d9: 0x2800db, 0x00da: 0x2800df, 0x00db: 0x2800e1, 0x00dc: 0x2800e2, 0x00dd: 0x2800e4, 0x00de: 0x2800e6,
		0x00df: 0x2800e7, 0x00e0: 0x2800e8, 0x00e1: 0x2800ec, 0x00e2: 0x2800ee, 0x00e3: 0x2800ef, 0x00e4: 0x2800f1,
		0x00e5: 0x2800f3, 0x00e6: 0x2800f7, 0x00e7: 0x2800f8, 0x00e8: 0x2800f9, 0x00e9: 0x290000, 0x00ea: 0x290001,
		0x00eb: 0x2a0000, 0x00ec: 0x2a0003, 0x00ed: 0x2a000c, 0x00ee: 0x2a001f, 0x00ef: 0x2a0021, 0x00f0: 0x2a0026,
		0x00f1: 0x2a002f, 0x00f2: 0x2a0032, 0x00f3: 0x2a0033, 0x00f4: 0x2a0034, 0x00f5: 0x2a003f, 0x00f6: 0x2a0041,
		0x00f7: 0x2a0042, 0x00f8: 0x2a0047, 0x00f9: 0x2a005b, 0x00fa: 0x2a005d, 0x00fb: 0x2a0062, 0x00fc: 0x2a0066,
		0x00fd: 0x2a009e, 0x00fe: 0x2a00a6, 0x00ff: 0x2a00ae, 0x0100: 0x2a00af, 0x0101: 0x2a00b2, 0x0102: 0x2a00b7,
		0x0103: 0x2a00bb, 0x0104: 0x2a00d3, 0x0105: 0x2a00e8, 0x0106: 0x2a00e9, 0x0107: 0x2a00ed, 0x0108: 0x2b0000,
		0x0109: 0x2b0043, 0x010a: 0x2c0000, 0x010b: 0x2c0047, 0x010c: 0x2d0000, 0x010d: 0x2d0030, 0x010e: 0x2e0000,
		0x010f: 0x2e0006, 0x0110: 0x2e006e, 0x0111: 0x2f0000, 0x0112: 0x2f0030, 0x0113: 0x2f0059, 0x0114: 0x2f0099,
		0x0115: 0x2f00ce, 0x0116: 0x300000, 0x0117: 0x300049, 0x0118: 0x310000, 0x0119: 0x3100b2, 0x011a: 0x320000,
		0x011b: 0x32003d, 0x011c: 0x32004d, 0x011d: 0x330000, 0x011e: 0x330016, 0x011f: 0x330017, 0x0120: 0x33001a,
		0x0121: 0x33001b, 0x0122: 0x33001c, 0x0123: 0x330027, 0x0124: 0x330029, 0x0125: 0x33002a, 0x0126: 0x33002b,
		0x0127: 0x33002c, 0x0128: 0x33002d, 0x0129: 0x330030, 0x012a: 0x33003c, 0x012b: 0x330040, 0x012c: 0x33004e,
		0x012d: 0x33004f, 0x012e: 0x330053, 0x012f: 0x330059, 0x0130: 0x33005a, 0x0131: 0x33005b, 0x0132: 0x330064,
		0x0133: 0x330079, 0x0134: 0x330088, 0x0135: 0x33008b, 0x0136: 0x33008c, 0x0137: 0x33008f, 0x0138: 0x330090,
		0x0139: 0x330093, 0x013a: 0x330098, 0x013b: 0x330099, 0x013c: 0x33009c, 0x013d: 0x3300a2, 0x013e: 0x3300a3,
		0x013f: 0x3300b0, 0x0140: 0x3300b5, 0x0141: 0x3300bd, 0x0142: 0x3300c1, 0x0143: 0x3300c4, 0x0144: 0x3300ce,
		0x0145: 0x3300d5, 0x0146: 0x3300d8, 0x0147: 0x3300d9, 0x0148: 0x3300de, 0x0149: 0x3300f1, 0x014a: 0x3300f2,
		0x014b: 0x3300f6, 0x014c: 0x340000, 0x014d: 0x340070, 0x014e: 0x350000, 0x014f: 0x3500a7, 0x0150: 0x360000,
		0x0151: 0x360068, 0x0152: 0x370000, 0x0153: 0x370050, 0x0154: 0x380000, 0x0155: 0x380047, 0x0156: 0x390000,
		0x0157: 0x39002c, 0x0158: 0x39004e, 0x0159: 0x390083, 0x015a: 0x3a0000, 0x015b: 0x3a006b, 0x015c: 0x3b0000,
		0x015d: 0x3b0075, 0x015e: 0x3c0000, 0x015f: 0x3c006a, 0x0160: 0x3d0000, 0x0161: 0x3d0055, 0x0162: 0x3d00a3,
		0x0163: 0x3d00a5, 0x0164: 0x3e0000, 0x0165: 0x3e00e8, 0x0166: 0x3f0000, 0x0167: 0x3f0069, 0x0168: 0x400000,
		0x0169: 0x40006b, 0x016a: 0x410000, 0x016b: 0x410013, 0x016c: 0x410063, 0x016d: 0x420000, 0x016e: 0x42003a,
		0x016f: 0x430000, 0x0170: 0x430065, 0x0171: 0x440000, 0x0172: 0x44000a, 0x0173: 0x450000, 0x0174: 0x450067,
		0x0175: 0x460000, 0x0176: 0x4600a5, 0x0177: 0x470000, 0x0178: 0x470031, 0x0179: 0x480000, 0x017a: 0x48006f,
		0x017b: 0x490000, 0x017c: 0x49002c, 0x017d: 0x490070, 0x017e: 0x4900cd, 0x017f: 0x4900eb, 0x0180: 0x4a0000,
		0x0181: 0x4a0074, 0x0182: 0x4b0000, 0x0183: 0x4b0030, 0x0184: 0x4c0000, 0x0185: 0x4c00e4, 0x0186: 0x4d0000,
		0x0187: 0x4d0052, 0x0188: 0x4e0000, 0x0189: 0x4e0040, 0x018a: 0x4f0000, 0x018b: 0x4f0075, 0x018c: 0x500000,
		0x018d: 0x5000e4, 0x018e: 0x510000, 0x018f: 0x510035, 0x0190: 0x520000, 0x0191: 0x520093, 0x0192: 0x530000,
		0x0193: 0x530075, 0x0194: 0x540000, 0x0195: 0x54007f, 0x0196: 0x550000, 0x0197: 0x550030, 0x0198: 0x560000,
		0x0199: 0x560057, 0x019a: 0x570000, 0x019b: 0x570075, 0x019c: 0x580000, 0x019d: 0x580077, 0x019e: 0x590000,
		0x019f: 0x59006b, 0x01a0: 0x5a0000, 0x01a1: 0x5a007b, 0x01a2: 0x5a007c, 0x01a3: 0x5b0000, 0x01a4: 0x5b006b,
		0x01a5: 0x5c0000, 0x01a6: 0x5c006b, 0x01a7: 0x5d0000, 0x01a8: 0x5d00e4, 0x01a9: 0x5e0000, 0x01aa: 0x5e0030,
		0x01ab: 0x5f0000, 0x01ac: 0x5f003a, 0x01ad: 0x600000, 0x01ae: 0x600050, 0x01af: 0x610000, 0x01b0: 0x610076,
		0x01b1: 0x620000, 0x01b2: 0x6200e4, 0x01b3: 0x630000, 0x01b4: 0x630088, 0x01b5: 0x640000, 0x01b6: 0x6400e6,
		0x01b7: 0x650000, 0x01b8: 0x6500e8, 0x01b9: 0x660000, 0x01ba: 0x66000b, 0x01bb: 0x660029, 0x01bc: 0x66002a,
		0x01bd: 0x66002b, 0x01be: 0x670000, 0x01bf: 0x670080, 0x01c0: 0x680000, 0x01c1: 0x68006d, 0x01c2: 0x68006e,
		0x01c3: 0x690000, 0x01c4: 0x690087, 0x01c5: 0x6a0000, 0x01c6: 0x6a0029, 0x01c7: 0x6b0000, 0x01c8: 0x6b0075,
		0x01c9: 0x6c0000, 0x01ca: 0x6c0075, 0x01cb: 0x6d0000, 0x01cc: 0x6d0089, 0x01cd: 0x6e0000, 0x01ce: 0x6e0075,
		0x01cf: 0x6e00e4, 0x01d0: 0x6f0000, 0x01d1: 0x6f0075, 0x01d2: 0x700000, 0x01d3: 0x70009c, 0x01d4: 0x710000,
		0x01d5: 0x710090, 0x01d6: 0x720000, 0x01d7: 0x7200a0, 0x01d8: 0x730000, 0x01d9: 0x730030, 0x01da: 0x740000,
		0x01db: 0x740092, 0x01dc: 0x750000, 0x01dd: 0x75006b, 0x01de: 0x760000, 0x01df: 0x760095, 0x01e0: 0x770000,
		0x01e1: 0x77006b, 0x01e2: 0x780000, 0x01e3: 0x78001e, 0x01e4: 0x78009f, 0x01e5: 0x7800c7, 0x01e6: 0x790000,
		0x01e7: 0x79009b, 0x01e8: 0x7a0000, 0x01e9: 0x7a0030, 0x01ea: 0x7b0000, 0x01eb: 0x7b0094, 0x01ec: 0x7c0000,
		0x01ed: 0x7c006e, 0x01ee: 0x7d0000, 0x01ef: 0x7d00a1, 0x01f0: 0x7e0000, 0x01f1: 0x7e00a8, 0x01f2: 0x7e00ca,
		0x01f3: 0x7f0000, 0x01f4: 0x7f00f9, 0x01f5: 0x800000, 0x01f6: 0x80003a, 0x01f7: 0x8000a7, 0x01f8: 0x810000,
		0x01f9: 0x81006b, 0x01fa: 0x8100a9, 0x01fb: 0x820000, 0x01fc: 0x820010, 0x01fd: 0x820016, 0x01fe: 0x820020,
		0x01ff: 0x820036, 0x0200: 0x8200a7, 0x0201: 0x8200d0, 0x0202: 0x8200d4, 0x0203: 0x830000, 0x0204: 0x830030,
		0x0205: 0x840000, 0x0206: 0x8400a8, 0x0207: 0x850000, 0x0208: 0x850030, 0x0209: 0x860000, 0x020a: 0x8600d1,
		0x020b: 0x870000, 0x020c: 0x8700e6, 0x020d: 0x880000, 0x020e: 0x880048, 0x020f: 0x880075, 0x0210: 0x890000,
		0x0211: 0x89006b, 0x0212: 0x8a0000, 0x0213: 0x8a0052, 0x0214: 0x8a00c0, 0x0215: 0x8b0000, 0x0216: 0x8b0100,
		0x0217: 0x8b01b3, 0x0218: 0x8b0300, 0x0219: 0x8b036b, 0x021a: 0x8c0000, 0x021b: 0x8c00b4, 0x021c: 0x8d0000,
		0x021d: 0x8d0001, 0x021e: 0x8e0000, 0x021f: 0x8e0006, 0x0220: 0x8f0000, 0x0221: 0x8f000b, 0x0222: 0x8f0021,
		0x0223: 0x8f002c, 0x0224: 0x8f0035, 0x0225: 0x8f005b, 0x0226: 0x8f005f, 0x0227: 0x8f0088, 0x0228: 0x8f0096,
		0x0229: 0x8f00a0, 0x022a: 0x8f00b9, 0x022b: 0x8f00d2, 0x022c: 0x8f00dc, 0x022d: 0x900000, 0x022e: 0x90001f,
		0x022f: 0x900042, 0x0230: 0x9000af, 0x0231: 0x910000, 0x0232: 0x91002c, 0x0233: 0x920000, 0x0234: 0x92001a,
		0x0235: 0x930000, 0x0236: 0x93008d, 0x0237: 0x9300be, 0x0238: 0x940000, 0x0239: 0x9400e4, 0x023a: 0x950000,
		0x023b: 0x950025, 0x023c: 0x950076, 0x023d: 0x95007f, 0x023e: 0x95008d, 0x023f: 0x9500c0, 0x0240: 0x9500e5,
		0x0241: 0x960000, 0x0242: 0x9600c1, 0x0243: 0x970000, 0x0244: 0x9700e4, 0x0245: 0x980000, 0x0246: 0x9800c0,
		0x0247: 0x990000, 0x0248: 0x990075, 0x0249: 0x9a0000, 0x024a: 0x9a00e4, 0x024b: 0x9b0000, 0x024c: 0x9b0049,
		0x024d: 0x9b00a8, 0x024e: 0x9b00c6, 0x024f: 0x9c0000, 0x0250: 0x9c00a0, 0x0251: 0x9d0000, 0x0252: 0x9d0093,
		0x0253: 0x9e0000, 0x0254: 0x9e002a, 0x0255: 0x9f0000, 0x0256: 0x9f0600, 0x0257: 0x9f068b, 0x0258: 0x9f0700,
		0x0259: 0x9f078b, 0x025a: 0xa00000, 0x025b: 0xa00084, 0x025c: 0xa10000, 0x025d: 0xa100cb, 0x025e: 0xa20000,
		0x025f: 0xa200c9, 0x0260: 0xa30000, 0x0261: 0xa30049, 0x0262: 0xa40000, 0x0263: 0xa400f9, 0x0264: 0xa50000,
		0x0265: 0xa5003c, 0x0266: 0xa50048, 0x0267: 0xa50075, 0x0268: 0xa500cf, 0x0269: 0xa60000, 0x026a: 0xa60009,
		0x026b: 0xa60092, 0x026c: 0xa600f4, 0x026d: 0xa70000, 0x026e: 0xa70200, 0x026f: 0xa70213, 0x0270: 0xa7028e,
		0x0271: 0xa702bf, 0x0272: 0xa702f4, 0x0273: 0xa70600, 0x0274: 0xa70613, 0x0275: 0xa7068e, 0x0276: 0xa706bf,
		0x0277: 0xa706f4, 0x0278: 0xa80000, 0x0279: 0xa80011, 0x027a: 0xa80049, 0x027b: 0xa800c6, 0x027c: 0xa90000,
		0x027d: 0xa90029, 0x027e: 0xa90075, 0x027f: 0xa900e4, 0x0280: 0xa900e6, 0x0281: 0xaa0000, 0x0282: 0xaa006b,
		0x0283: 0xaa0084, 0x0284: 0xaa009f, 0x0285: 0xaa00c7, 0x0286: 0xab0000, 0x0287: 0xab006b, 0x0288: 0xac0000,
		0x0289: 0xac0075, 0x028a: 0xac00e6, 0x028b: 0xad0000, 0x028c: 0xad00da, 0x028d: 0xae0000, 0x028e: 0xae0046,
		0x028f: 0xae0048, 0x0290: 0xaf0000, 0x0291: 0xaf00dd, 0x0292: 0xb00000, 0x0293: 0xb000df, 0x0294: 0xb10000,
		0x0295: 0xb10038, 0x0296: 0xb100e0, 0x0297: 0xb20000, 0x0298: 0xb200a3, 0x0299: 0xb30000, 0x029a: 0xb3008b,
		0x029b: 0xb40000, 0x029c: 0xb40031, 0x029d: 0xb50000, 0x029e: 0xb500e5, 0x029f: 0xb60000, 0x02a0: 0xb70000,
		0x02a1: 0xb7006b, 0x02a2: 0xb700b3, 0x02a3: 0xb80000, 0x02a4: 0xb80100, 0x02a5: 0xb80106, 0x02a6: 0xb80200,
		0x02a7: 0xb802ea, 0x02a8: 0xb80600, 0x02a9: 0xb806ea, 0x02aa: 0xb90000, 0x02ab: 0xb90600, 0x02ac: 0xb90685,
		0x02ad: 0xb90800, 0x02ae: 0xb90885, 0x02af: 0xba0000, 0x02b0: 0xba00f0, 0x02b1: 0xbb0000, 0x02b2: 0xbb0001,
		0x02b3: 0xbc0000, 0x02b4: 0xbc00e4, 0x02b5: 0xbd0000, 0x02b6: 0xbd002c, 0x02b7: 0xbe0000, 0x02b8: 0xbe00e6,
		0x02b9: 0xbf0000, 0x02ba: 0xbf0030, 0x02bb: 0xc00000, 0x02bc: 0xc00001, 0x02bd: 0xc10000, 0x02be: 0xc1001b,
		0x02bf: 0xc100a5, 0x02c0: 0xc20000, 0x02c1: 0xc20061, 0x02c2: 0xc30000, 0x02c3: 0xc3008b, 0x02c4: 0xc40000,
		0x02c5: 0xc40400, 0x02c6: 0xc40431, 0x02c7: 0xc40461, 0x02c8: 0xc40496, 0x02c9: 0xc404c7, 0x02ca: 0xc40500,
		0x02cb: 0xc40561, 0x02cc: 0xc40596, 0x02cd: 0xc405e3, 0x02ce: 0xc50000, 0x02cf: 0xc500f7,
	}

	for expectedTagID, expectedTag := range expected {
		if tag := localeTags.tag(expectedTagID); tag != expectedTag {
			t.Errorf("unexpected tag for id %d, %#x", uint(expectedTagID), tag)
		}
		if tagID := localeTags.tagID(expectedTag.langID(), expectedTag.scriptID(), expectedTag.regionID()); tagID != expectedTagID {
			t.Errorf("unexpected tag id for tag %#x: %d", expectedTag, uint(tagID))
		}
	}
}

func TestParentLocaleTags(t *testing.T) {
	expected := map[tagID]tagID{
		0x002e: 0x029f, 0x0049: 0x029f, 0x0082: 0x0081, 0x0083: 0x0081, 0x0084: 0x0081, 0x0086: 0x0082,
		0x0087: 0x0081, 0x0088: 0x0081, 0x0089: 0x0081, 0x008b: 0x0081, 0x008c: 0x0081, 0x008d: 0x0081,
		0x008e: 0x0081, 0x008f: 0x0081, 0x0090: 0x0081, 0x0091: 0x0082, 0x0092: 0x0081, 0x0093: 0x0081,
		0x0094: 0x0081, 0x0095: 0x0081, 0x0096: 0x0082, 0x0097: 0x0081, 0x0098: 0x0082, 0x0099: 0x0081,
		0x009a: 0x0081, 0x009b: 0x0082, 0x009c: 0x0081, 0x009d: 0x0081, 0x009e: 0x0081, 0x009f: 0x0081,
		0x00a0: 0x0081, 0x00a1: 0x0081, 0x00a2: 0x0081, 0x00a3: 0x0081, 0x00a4: 0x0081, 0x00a6: 0x0081,
		0x00a7: 0x0081, 0x00a8: 0x0081, 0x00a9: 0x0081, 0x00aa: 0x0081, 0x00ab: 0x0081, 0x00ac: 0x0081,
		0x00ad: 0x0081, 0x00ae: 0x0081, 0x00af: 0x0081, 0x00b0: 0x0081, 0x00b1: 0x0081, 0x00b2: 0x0081,
		0x00b3: 0x0081, 0x00b4: 0x0081, 0x00b5: 0x0081, 0x00b6: 0x0081, 0x00b8: 0x0081, 0x00ba: 0x0081,
		0x00bb: 0x0081, 0x00bc: 0x0081, 0x00bd: 0x0081, 0x00be: 0x0081, 0x00bf: 0x0081, 0x00c0: 0x0081,
		0x00c1: 0x0081, 0x00c2: 0x0082, 0x00c3: 0x0081, 0x00c4: 0x0081, 0x00c5: 0x0081, 0x00c6: 0x0081,
		0x00c7: 0x0081, 0x00c8: 0x0081, 0x00c9: 0x0081, 0x00cb: 0x0081, 0x00cc: 0x0081, 0x00cd: 0x0081,
		0x00ce: 0x0081, 0x00cf: 0x0081, 0x00d0: 0x0082, 0x00d1: 0x0081, 0x00d2: 0x0081, 0x00d3: 0x0082,
		0x00d4: 0x0081, 0x00d5: 0x0081, 0x00d6: 0x0081, 0x00d7: 0x0081, 0x00d8: 0x0081, 0x00d9: 0x0081,
		0x00da: 0x0081, 0x00db: 0x0081, 0x00dc: 0x0081, 0x00dd: 0x0081, 0x00de: 0x0081, 0x00e1: 0x0081,
		0x00e2: 0x0081, 0x00e4: 0x0081, 0x00e5: 0x0081, 0x00e6: 0x0081, 0x00e7: 0x0081, 0x00e8: 0x0081,
		0x00ed: 0x00ec, 0x00ee: 0x00ec, 0x00ef: 0x00ec, 0x00f0: 0x00ec, 0x00f1: 0x00ec, 0x00f2: 0x00ec,
		0x00f3: 0x00ec, 0x00f4: 0x00ec, 0x00f5: 0x00ec, 0x00f7: 0x00ec, 0x00fa: 0x00ec, 0x00fb: 0x00ec,
		0x00fd: 0x00ec, 0x00fe: 0x00ec, 0x00ff: 0x00ec, 0x0100: 0x00ec, 0x0102: 0x00ec, 0x0103: 0x00ec,
		0x0104: 0x00ec, 0x0105: 0x00ec, 0x0106: 0x00ec, 0x0107: 0x00ec, 0x0216: 0x029f, 0x0221: 0x022a,
		0x0223: 0x022a, 0x0224: 0x022a, 0x0225: 0x022a, 0x0226: 0x022a, 0x0227: 0x022a, 0x0228: 0x022a,
		0x0229: 0x022a, 0x022b: 0x022a, 0x022c: 0x022a, 0x0256: 0x029f, 0x0273: 0x029f, 0x02a4: 0x029f,
		0x02a6: 0x029f, 0x02ab: 0x029f, 0x02ca: 0x029f, 0x02cc: 0x02cb,
	}

	for child, expectedParent := range expected {
		if parent := parentLocaleTags.parentID(child); parent != expectedParent {
			t.Errorf("unexpected parent id for child id %#x: %#x", child, parent)
		}
	}
}