package opcode

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	} else if p.Opcode.Opcode == OP_PUSH {
		p.Operands = append(p.Operands, binary.LittleEndian.Uint32(p.Args[0:4]))
	} else if p.Opcode.Opcode > 79 {
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

func (p *Push) GetLength() int {
	return GetInstructionLength(p.GetOpcode(), p.Args[0])
}

func (p *Push) String(color string, subroutines map[int]string) string {
	var sb strings.Builder
	style := lipgloss.NewStyle()
	if color != "" {
		style = style.Foreground(lipgloss.Color(color))
	}
	name := Names[p.GetOpcode()]
	if p.GetOpcode() > 79 {
		name = Names[OP_PUSHD]
	}
	ops := p.GetOperands()
	opstrs := make([]string, len(ops))
	for j := range len(ops) {
		opstrs[j] = fmt.Sprintf("%v", ops[j])
	}
	offset := fmt.Sprintf("0x%04X", p.GetOffset())
	name = functionNameStyle.Render(name)
	offset = style.Render(offset)
	opstr := style.Render(strings.Join(opstrs, " "))
	sb.WriteString(offset + " " + name + " " + opstr)
	return sb.String()
}
