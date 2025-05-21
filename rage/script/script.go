package script

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/mrchip53/gta-tools/rage/script/opcode"
	"github.com/mrchip53/gta-tools/rage/util"
)

const (
	HEADER_MAGIC                      = 0x0d524353
	HEADER_MAGIC_ENCRYPTED            = 0x0e726373
	HEADER_MAGIC_ENCRYPTED_COMPRESSED = 0x0e726353
)

var (
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	opNameStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
)

type scriptHeader struct {
	Identifier       uint32
	CodeSize         int32
	LocalVarCount    int32
	GlobalVarCount   int32
	ScriptFlags      int32
	GlobalsSignature int32
	CompressedSize   int32
}

func newScriptHeader(data []byte) (int32, scriptHeader) {
	if len(data) < 28 {
		panic("Not enough data to read header")
	}
	i := binary.LittleEndian.Uint32(data[0:4])
	var c int32
	l := int32(24)
	if i == HEADER_MAGIC_ENCRYPTED_COMPRESSED {
		c = int32(binary.LittleEndian.Uint32(data[24:28]))
		l = 28
	}
	return l, scriptHeader{
		Identifier:       i,
		CodeSize:         int32(binary.LittleEndian.Uint32(data[4:8])),
		LocalVarCount:    int32(binary.LittleEndian.Uint32(data[8:12])),
		GlobalVarCount:   int32(binary.LittleEndian.Uint32(data[12:16])),
		ScriptFlags:      int32(binary.LittleEndian.Uint32(data[16:20])),
		GlobalsSignature: int32(binary.LittleEndian.Uint32(data[20:24])),
		CompressedSize:   c,
	}
}

type RageScript struct {
	Name        string
	Header      scriptHeader
	Code        []byte
	Locals      []uint
	Globals     []uint
	Unsupported bool

	Opcodes     []opcode.Instruction
	Subroutines map[int]string
}

func NewRageScript(name string, data []byte) RageScript {
	s, h := newScriptHeader(data)

	encrypted := h.Identifier == HEADER_MAGIC_ENCRYPTED
	compressed := h.Identifier == HEADER_MAGIC_ENCRYPTED_COMPRESSED

	var code, l, g []byte
	if compressed {
		ed := data[s : s+h.CompressedSize]
		cd := util.Decrypt(ed)
		_ = cd
	} else {
		code = data[s : s+h.CodeSize]
		l = data[s+h.CodeSize : s+h.CodeSize+h.LocalVarCount*4]
		g = data[s+h.CodeSize+h.LocalVarCount*4 : s+h.CodeSize+h.LocalVarCount*4+h.GlobalVarCount*4]

		if encrypted {
			util.Decrypt(code)
			util.Decrypt(l)
			util.Decrypt(g)
		}
	}

	var locals, globals []uint

	for i := range h.LocalVarCount {
		locals = append(locals, uint(binary.LittleEndian.Uint32(l[i*4:i*4+4])))
	}
	for i := range h.GlobalVarCount {
		globals = append(globals, uint(binary.LittleEndian.Uint32(g[i*4:i*4+4])))
	}

	script := RageScript{
		Name:        name,
		Header:      h,
		Code:        code,
		Locals:      locals,
		Globals:     globals,
		Unsupported: compressed,
		Subroutines: make(map[int]string),
	}

	script.Disassemble()

	return script
}

func (r *RageScript) Disassemble() {
	r.Opcodes = make([]opcode.Instruction, 0)
	var ptr int
	for ptr < len(r.Code) {
		c := r.Code[ptr]
		l := opcode.GetInstructionLength(c, r.Code[ptr+1])
		args := make([]byte, l)
		copy(args, r.Code[ptr+1:ptr+l])
		var ins opcode.Instruction = opcode.NewInstruction(ptr, c, args)
		if c > 0x4F && c <= 0xFF {
			ins = opcode.NewPush(ptr, c, args)
		} else {
			f, ok := opcode.Instructions[c]
			if ok {
				ins = f(ptr, c, args)
			}
		}
		if ins.GetOpcode() == opcode.OP_FN_BEGIN {
			r.Subroutines[ptr] = fmt.Sprintf("sub_0x%04X", ptr)
		}
		r.Opcodes = append(r.Opcodes, ins)
		ptr += l
	}
}

func (r *RageScript) InsertInstruction(offset int, opc uint8, args []byte) {
	cb := make([]byte, len(r.Code)+len(args)+1)
	copy(cb, r.Code[:offset])
	data := append([]byte{opc}, args...)
	copy(cb[offset:], data)
	copy(cb[offset+len(data):], r.Code[offset:])
	r.Code = cb
}

func (r RageScript) String(y int, offset, height int) string {
	var sb strings.Builder

	// sb.WriteString("Name: " + r.Name + "\n")
	// sb.WriteString("Header:\n")
	// sb.WriteString(fmt.Sprintf("%+v\n", r.Header))
	for i := offset; i < offset+height; i++ {
		if i >= len(r.Opcodes) {
			break
		}

		ins := r.Opcodes[i]

		color := ""
		if i == y {
			color = "#FFFF00"
		}

		str := ins.String(color, r.Subroutines)

		context := ""
		switch ins.(type) {
		case *opcode.Branch:
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#44FF44"))
			context += style.Render("-> ")
			next := "?"
			for _, v := range r.Opcodes {
				if v.GetOffset() == int(ins.GetOperands()[0].(uint32)) {
					next = v.String("", r.Subroutines)
					break
				}
			}
			context += next
		}

		sb.WriteString(str)

		if i == y {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
			sb.WriteString(style.Render(" <-"))
		}

		if context != "" {
			sb.WriteString(" " + context)
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
