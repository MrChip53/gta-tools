package opcode

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Branch struct {
	Opcode
	TargetInstruction Instruction
}

func NewBranch(offset int, opcode uint8, args []byte) *Branch {
	p := &Branch{
		Opcode: Opcode{
			Offset:   offset,
			Opcode:   opcode,
			Args:     args,
			Operands: make([]any, 0),
		},
		TargetInstruction: nil,
	}
	p.Disassemble()
	return p
}

func (p *Branch) Disassemble() {
	bo := binary.LittleEndian.Uint32(p.Args[0:4])
	p.Operands = append(p.Operands, bo)
}

func (p *Branch) GetOffset() int {
	return p.Offset
}

func (p *Branch) GetOpcode() uint8 {
	return p.Opcode.Opcode
}

func (p *Branch) GetOperands() []any {
	return p.Operands
}

func (p *Branch) GetLength() int {
	l := uint8(0)
	if len(p.Args) > 0 {
		l = p.Args[0]
	}
	return GetInstructionLength(p.GetOpcode(), l)
}

func (p *Branch) UpdateTargetOffset(newOffset int) {
	newTarget := uint32(newOffset)
	p.Operands[0] = newTarget
	binary.LittleEndian.PutUint32(p.Args[0:4], newTarget)
}

func (p *Branch) String(color string, subroutines map[int]string) string {
	var sb strings.Builder
	style := lipgloss.NewStyle()
	if color != "" {
		style = style.Foreground(lipgloss.Color(color))
	}
	text := p.Text(style, subroutines)
	offset := fmt.Sprintf("0x%04X", p.GetOffset())
	offset = style.Render(offset)
	sb.WriteString(offset + " " + text)
	return sb.String()
}

func (p *Branch) Text(style lipgloss.Style, subroutines map[int]string) string {
	name := functionNameStyle.Render(Names[p.GetOpcode()])
	co := p.GetOperands()[0].(uint32)
	opstr := fmt.Sprintf("0x%04X", co)
	if subroutines != nil {
		if subroutine, ok := subroutines[int(co)]; ok {
			opstr = fmt.Sprintf("%s", subroutine)
		}
	}
	opstr = style.Render(opstr)

	return name + " " + opstr
}
