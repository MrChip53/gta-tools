package opcode

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Base struct {
	Opcode
}

func NewInstruction(offset int, opcode uint8, args []byte) Instruction {
	f, ok := Instructions[opcode]
	if ok {
		ins := f(offset, opcode, args)
		ins.Disassemble()
		return ins
	}

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

func (p *Base) GetLength() int {
	l := uint8(0)
	if len(p.Args) > 0 {
		l = p.Args[0]
	}
	return GetInstructionLength(p.GetOpcode(), l)
}

func (p *Base) String(color string, subroutines map[int]string) string {
	var sb strings.Builder
	style := lipgloss.NewStyle()
	if color != "" {
		style = style.Foreground(lipgloss.Color(color))
	}
	name := Names[p.GetOpcode()]
	ops := p.GetOperands()
	opstrs := make([]string, len(ops))
	for j := range len(ops) {
		opstrs[j] = fmt.Sprintf("%v", ops[j])
	}
	name = functionNameStyle.Render(name)
	opstr := style.Render(strings.Join(opstrs, " "))
	sb.WriteString(name + " " + opstr)
	return sb.String()
}
