package opcode

import (
	"encoding/binary"
	"math"
)

type Push struct {
	Opcode
}

func NewPush(offset int, opcode uint8, args []byte) *Push {
	p := &Push{
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

func (p *Push) Disassemble() {
	if p.Opcode.Opcode == OP_PUSHS {
		p.Operands = append(p.Operands, binary.LittleEndian.Uint16(p.Args[0:2]))
	} else if p.Opcode.Opcode == OP_PUSHF {
		bits := binary.LittleEndian.Uint32(p.Args[0:4])
		f := math.Float32frombits(bits)
		p.Operands = append(p.Operands, f)
	} else if p.Opcode.Opcode == OP_PUSH_STRING {
		str := make([]byte, len(p.Args)-1)
		copy(str, p.Args[1:])
		p.Operands = append(p.Operands, string(str))
	} else if p.Opcode.Opcode >= 0x4F && p.Opcode.Opcode <= 0xFF {
		p.Operands = append(p.Operands, p.GetOpcode()-96)
	}
}

func (p *Push) GetOffset() int {
	return p.Offset
}

func (p *Push) GetOpcode() uint8 {
	return p.Opcode.Opcode
}

func (p *Push) GetOperands() []any {
	return p.Operands
}
