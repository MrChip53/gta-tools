package opcode

import (
	"encoding/binary"
	"fmt"
)

type Native struct {
	Opcode
}

func newNative(offset int, opcode uint8, args []byte) *Native {
	p := &Native{
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

func (p *Native) Disassemble() {
	in := p.Opcode.Args[0]
	out := p.Opcode.Args[1]
	native := binary.LittleEndian.Uint32(p.Args[2:6])
	nativeStr, ok := nativeFunctions[native]
	if !ok {
		nativeStr = fmt.Sprintf("Unknown (%d)", native)
	}
	p.Operands = append(p.Operands, nativeStr, in, out)
}

func (p *Native) GetOffset() int {
	return p.Offset
}

func (p *Native) GetOpcode() uint8 {
	return p.Opcode.Opcode
}

func (p *Native) GetOperands() []any {
	return p.Operands
}
