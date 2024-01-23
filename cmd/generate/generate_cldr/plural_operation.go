package generate_cldr

import "github.com/liblxn/lxnc/internal/generator"

var (
	_ generator.Snippet = (*pluralOperation)(nil)
)

type pluralOperation struct {
	operatorBits uint
	eq           uint8
	neq          uint8

	operandBits        uint
	absValue           uint8 // n
	intDigits          uint8 // i
	numFracDigit       uint8 // v
	numFracDigitNoZero uint8 // w
	fracDigits         uint8 // f
	fracDigitsNoZero   uint8 // t
	compactDecExp      uint8 // c, e
}

func newPluralOperation() *pluralOperation {
	return &pluralOperation{
		operatorBits: 1,
		eq:           0x0,
		neq:          0x1,

		operandBits:        4,
		absValue:           0x0,
		intDigits:          0x1,
		numFracDigit:       0x2,
		numFracDigitNoZero: 0x3,
		fracDigits:         0x4,
		fracDigitsNoZero:   0x5,
		compactDecExp:      0x6,
	}
}

func (po *pluralOperation) Imports() []string {
	return nil
}

func (po *pluralOperation) Generate(p *generator.Printer) {
	p.Println(`// Operator represents an operator in a plural rule.`)
	p.Println(`type Operator uint8`)
	p.Println()
	p.Println(`// Available operators for the plural rules.`)
	p.Println(`const (`)
	p.Println(`	Equal    Operator = `, po.eq)
	p.Println(`	NotEqual Operator = `, po.neq)
	p.Println(`)`)
	p.Println()
	p.Println(`// Operand represents an operand in a plural rule.`)
	p.Println(`type Operand uint8`)
	p.Println()
	p.Println(`// Available operands for the plural rules.`)
	p.Println(`const (`)
	p.Println(`	AbsoluteValue       Operand = `, po.absValue, ` // n`)
	p.Println(`	IntegerDigits       Operand = `, po.intDigits, ` // i`)
	p.Println(`	NumFracDigit        Operand = `, po.numFracDigit, ` // v`)
	p.Println(`	NumFracDigitNoZeros Operand = `, po.numFracDigitNoZero, ` // w`)
	p.Println(`	FracDigits          Operand = `, po.fracDigits, ` // f`)
	p.Println(`	FracDigitsNoZeros   Operand = `, po.fracDigitsNoZero, ` // t`)
	p.Println(`	CompactDecimalExp   Operand = `, po.compactDecExp, ` // c, e`)
	p.Println(`)`)
}
