package opcode

type Base struct {
	Opcode
}

func NewInstruction(offset int, opcode uint8, args []byte) *Base {
	p := &Base{
		Opcode: Opcode{
			Offset:   offset,
			Opcode:   opcode,
			Args:     args,
			Operands: make([]any, 0),
		},
	}
	p.Disassemble()
	return p
}

func (p *Base) Disassemble() {
}

func (p *Base) GetOffset() int {
	return p.Offset
}

func (p *Base) GetOpcode() uint8 {
	return p.Opcode.Opcode
}

func (p *Base) GetOperands() []any {
	return p.Operands
}
