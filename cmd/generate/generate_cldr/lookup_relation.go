package generate_cldr

import (
	"fmt"
	"sort"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/generator"
)

var (
	_ generator.Snippet     = (*relationLookup)(nil)
	_ generator.TestSnippet = (*relationLookup)(nil)
)

type relationLookup struct {
	operation  *pluralOperation
	connective *connective

	modexpBits      uint
	rangeCountBits  uint
	pluralChunkBits uint
	idBits          uint
}

func newRelationLookup(operation *pluralOperation, connective *connective) *relationLookup {
	return &relationLookup{
		operation:       operation,
		connective:      connective,
		modexpBits:      4,
		rangeCountBits:  5,
		pluralChunkBits: 32,
		idBits:          13,
	}
}

func (l *relationLookup) newTestRelation(operand, modexp, operator, rangeCount, connective uint) string {
	return fmt.Sprintf("%#x",
		(operand<<(l.modexpBits+l.operation.operatorBits+l.rangeCountBits+l.connective.bits))|
			(modexp<<(l.operation.operatorBits+l.rangeCountBits+l.connective.bits))|
			(operator<<(l.rangeCountBits+l.connective.bits))|
			(rangeCount<<l.connective.bits)|
			connective,
	)
}

func (l *relationLookup) Imports() []string {
	return nil
}

func (l *relationLookup) Generate(p *generator.Printer) {
	if l.operation.operandBits+l.modexpBits+l.operation.operatorBits+l.connective.bits+l.rangeCountBits > l.pluralChunkBits {
		panic("invalid plural rule lookup configuration")
	}

	var relIDBits int
	switch {
	case l.idBits <= 8:
		relIDBits = 8
	case l.idBits <= 16:
		relIDBits = 16
	case l.idBits <= 32:
		relIDBits = 32
	case l.idBits <= 64:
		relIDBits = 64
	default:
		panic("invalid relation id bits")
	}

	operandMask := fmt.Sprintf("%#x", (1<<l.operation.operandBits)-1)
	modexpMask := fmt.Sprintf("%#x", (1<<l.modexpBits)-1)
	operatorMask := fmt.Sprintf("%#x", (1<<l.operation.operatorBits)-1)
	connectiveMask := fmt.Sprintf("%#x", (1<<l.connective.bits)-1)
	rangeCountMask := fmt.Sprintf("%#x", (1<<l.rangeCountBits)-1)

	p.Println(`type relation []uint`, l.pluralChunkBits)
	p.Println()
	p.Println(`func (r relation) operand() uint    { return uint((r[0] >> `, l.modexpBits+l.operation.operatorBits+l.rangeCountBits+l.connective.bits, `) & `, operandMask, `) }`)
	p.Println(`func (r relation) modexp() int      { return int((r[0] >> `, l.operation.operatorBits+l.rangeCountBits+l.connective.bits, `) & `, modexpMask, `) }`)
	p.Println(`func (r relation) operator() uint   { return uint((r[0] >> `, l.rangeCountBits+l.connective.bits, `) & `, operatorMask, `) }`)
	p.Println(`func (r relation) rangeCount() int  { return int((r[0] >> `, l.connective.bits, `) & `, rangeCountMask, `) }`)
	p.Println(`func (r relation) connective() uint { return uint((r[0]) & `, connectiveMask, `) }`)
	p.Println(`func (r relation) ranges() relation { return r[1 : 1+2*r.rangeCount()] }`)
	p.Println(`func (r relation) next() relation   { return r[1+2*r.rangeCount():] }`)
	p.Println()
	p.Println(`type relationID uint`, relIDBits)
	p.Println(`type relationLookup []uint`, l.pluralChunkBits)
	p.Println()
	p.Println(`func (l relationLookup) relation(id relationID) relation {`)
	p.Println(`	if id == 0 || int(id) > len(l) {`)
	p.Println(`		return nil`)
	p.Println(`	}`)
	p.Println(`	return relation(l[id-1:])`)
	p.Println(`}`)
}

func (l *relationLookup) TestImports() []string {
	return []string{"reflect"}
}

func (l *relationLookup) GenerateTest(p *generator.Printer) {
	p.Println(`func TestRelation(t *testing.T) {`)
	p.Println(`	rel := relation{`, l.newTestRelation(5, 4, 1, 2, 3), `, 0x0, 0x0, 0x0, 0x0, 0x77}`)
	p.Println()
	p.Println(`	if op := rel.operand(); op != 5 {`)
	p.Println(`		t.Errorf("unexpected operand: %d", op)`)
	p.Println(`	}`)
	p.Println(`	if modexp := rel.modexp(); modexp != 4 {`)
	p.Println(`		t.Errorf("unexpected modulo exponent: %d", modexp)`)
	p.Println(`	}`)
	p.Println(`	if op := rel.operator(); op != 1 {`)
	p.Println(`		t.Errorf("unexpected operator: %d", op)`)
	p.Println(`	}`)
	p.Println(`	if rc := rel.rangeCount(); rc != 2 {`)
	p.Println(`		t.Errorf("unexpected range count: %d", rc)`)
	p.Println(`	}`)
	p.Println(`	if c := rel.connective(); c != 3 {`)
	p.Println(`		t.Errorf("unexpected connective: %d", c)`)
	p.Println(`	}`)
	p.Println(`	if r := rel.ranges(); !reflect.DeepEqual(r, rel[1:5]) {`)
	p.Println(`		t.Errorf("unexpected ranges: %v", r)`)
	p.Println(`	}`)
	p.Println(`	if nxt := rel.next(); !reflect.DeepEqual(nxt, rel[5:]) {`)
	p.Println(`		t.Errorf("unexpected next value: %v", nxt)`)
	p.Println(`	}`)
	p.Println(`}`)
	p.Println()
	p.Println(`func TestRelationLookup(t *testing.T) {`)
	p.Println(`	lookup := relationLookup{1, 2, 3}`)
	p.Println()
	p.Println(`	if r := lookup.relation(0); len(r) != 0 {`)
	p.Println(`		t.Errorf("unexpected relation: %v", r)`)
	p.Println(`	}`)
	p.Println(`	if r := lookup.relation(1); !reflect.DeepEqual(r, relation(lookup)) {`)
	p.Println(`		t.Errorf("unexpected relation: %v", r)`)
	p.Println(`	}`)
	p.Println(`	if r := lookup.relation(2); !reflect.DeepEqual(r, relation(lookup[1:])) {`)
	p.Println(`		t.Errorf("unexpected relation: %v", r)`)
	p.Println(`	}`)
	p.Println(`}`)
}

var (
	_ generator.Snippet = (*relationLookupVar)(nil)
)

type relationRule struct {
	key       string
	numChunks int
	condition []cldr.Conjunction
}

type relationLookupVar struct {
	name      string
	typ       *relationLookup
	operation *pluralOperation
	rules     []relationRule
}

func newRelationLookupVar(name string, typ *relationLookup, operation *pluralOperation, data *cldr.Data) *relationLookupVar {
	isPowerOfTen := func(n int) bool {
		switch n {
		case 1, 10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000:
			return true
		default:
			return false
		}
	}

	v := &relationLookupVar{
		name:      name,
		typ:       typ,
		operation: operation,
	}

	totalChunks := 0
	keyMap := map[string]struct{}{}
	forEachPluralRelation(data, func(rule cldr.PluralRule) {
		if len(rule.Condition) == 0 {
			return
		}

		key := v.keyOf(rule)
		if _, has := keyMap[key]; has {
			return
		}

		numChunks := 0
		for _, conjunctions := range rule.Condition {
			for _, conj := range conjunctions {
				numChunks += 1 + 2*len(conj.Ranges)
				switch {
				case conj.Modulo != 0 && !isPowerOfTen(conj.Modulo):
					panic(fmt.Sprintf("relation modulo is not a power of 10, cannot add %q", conj.String()))
				case len(conj.Ranges) > (1<<v.typ.rangeCountBits)-1:
					panic(fmt.Sprintf("number of ranges exceeds the maximum, cannot add %q", conj.String()))
				}
				for _, rng := range conj.Ranges {
					maxBound := (1 << v.typ.pluralChunkBits) - 1
					if rng.LowerBound > maxBound || rng.UpperBound > maxBound {
						panic(fmt.Sprintf("range values exceed the maximum, cannot add %q", conj.String()))
					}
				}
			}
		}

		v.rules = append(v.rules, relationRule{key: key, condition: rule.Condition, numChunks: numChunks})
	})

	if totalChunks >= (1 << v.typ.idBits) {
		panic("number of relations exceeds the maximum")
	}

	sort.Slice(v.rules, func(i, j int) bool {
		return v.rules[i].key < v.rules[j].key
	})

	return v
}

func (v *relationLookupVar) newRelation(rel cldr.Relation, connective uint) uint64 {
	operand := uint8(0)
	switch rel.Operand {
	case cldr.AbsoluteValue:
		operand = v.operation.absValue
	case cldr.IntegerDigits:
		operand = v.operation.intDigits
	case cldr.FracDigitCountTrailingZeros:
		operand = v.operation.numFracDigit
	case cldr.FracDigitCount:
		operand = v.operation.numFracDigitNoZero
	case cldr.FracDigitsTrailingZeros:
		operand = v.operation.fracDigits
	case cldr.FracDigits:
		operand = v.operation.fracDigitsNoZero
	case cldr.CompactDecimalExponent, cldr.CompactDecimalExponent2:
		operand = v.operation.compactDecExp
	default:
		panic(fmt.Sprintf("unknown relation operand %q detected", rel.Operand))
	}

	modexp := 0
	if mod := rel.Modulo; mod != 0 {
		for mod > 1 {
			modexp++
			mod /= 10
		}
	}

	operator := v.operation.eq
	if rel.Operator == cldr.NotEqual {
		operator = v.operation.neq
	}

	return uint64(operand)<<(v.typ.modexpBits+v.typ.operation.operatorBits+v.typ.rangeCountBits+v.typ.connective.bits) |
		uint64(modexp)<<(v.typ.operation.operatorBits+v.typ.rangeCountBits+v.typ.connective.bits) |
		uint64(operator)<<(v.typ.rangeCountBits+v.typ.connective.bits) |
		uint64(len(rel.Ranges))<<v.typ.connective.bits |
		uint64(connective)
}

func (v *relationLookupVar) keyOf(rule cldr.PluralRule) string {
	r := cldr.PluralRule{Condition: rule.Condition} // strip samples
	return r.String()
}

func (v *relationLookupVar) relationID(rule cldr.PluralRule) uint {
	key := v.keyOf(rule)
	id := 1
	for _, r := range v.rules {
		if r.key == key {
			return uint(id)
		}
		id += r.numChunks
	}
	panic(fmt.Sprintf("relation not found: %s", key))
}

func (v *relationLookupVar) Imports() []string {
	return nil
}

func (v *relationLookupVar) Generate(p *generator.Printer) {
	numChunks := 0
	for _, rule := range v.rules {
		numChunks += rule.numChunks
	}

	hex := func(chunk uint64) string {
		return fmt.Sprintf("%#0[2]*[1]x", chunk, v.typ.pluralChunkBits/4)
	}

	p.Println(`var `, v.name, ` = relationLookup{ // `, numChunks, ` items, `, uint(numChunks)*v.typ.pluralChunkBits/8, ` bytes`)

	for _, rule := range v.rules {
		p.Print(`	`)
		for d := 0; d < len(rule.condition); d++ {
			conj := rule.condition[d]

			or := v.typ.connective.disjunction
			if d == len(conj)-1 {
				or = v.typ.connective.none
			}

			for r := 0; r < len(conj); r++ {
				rel := conj[r]
				and := v.typ.connective.conjunction
				if r == len(conj)-1 {
					and = or
				}

				p.Print(hex(v.newRelation(rel, and)), `, `)
				for _, rng := range rel.Ranges {
					p.Print(hex(uint64(rng.LowerBound)), `, `, hex(uint64(rng.UpperBound)), `, `)
				}
			}
		}

		pr := cldr.PluralRule{Condition: rule.condition}
		p.Println(`// `, pr.String())
	}
	p.Println(`}`)
}
