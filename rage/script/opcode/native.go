package opcode

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

func (p *Native) GetLength() int {
	l := uint8(0)
	if len(p.Args) > 0 {
		l = p.Args[0]
	}
	return GetInstructionLength(p.GetOpcode(), l)
}

func (p *Native) String(color string, subroutines map[int]string) string {
	var sb strings.Builder
	style := lipgloss.NewStyle()
	if color != "" {
		style = style.Foreground(lipgloss.Color(color))
	}
	name := Names[p.GetOpcode()]
	ops := p.GetOperands()
	opstr := fmt.Sprintf("%s in=%d out=%d", ops[0], ops[1], ops[2])
	offset := fmt.Sprintf("0x%04X", p.GetOffset())
	name = functionNameStyle.Render(name)
	offset = style.Render(offset)
	opstr = style.Render(opstr)
	sb.WriteString(offset + " " + name + " " + opstr)
	return sb.String()
}
