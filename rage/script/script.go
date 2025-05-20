package script

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/mrchip53/gta-tools/rage/util"
)

const (
	HEADER_MAGIC                      = 0x0d524353
	HEADER_MAGIC_ENCRYPTED            = 0x0e726373
	HEADER_MAGIC_ENCRYPTED_COMPRESSED = 0x0e726353
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

	return RageScript{
		Name:        name,
		Header:      h,
		Code:        code,
		Locals:      locals,
		Globals:     globals,
		Unsupported: compressed,
	}
}

func (r RageScript) String(y int, style lipgloss.Style, offset, height int) string {
	var sb strings.Builder

	// sb.WriteString("Name: " + r.Name + "\n")
	// sb.WriteString("Header:\n")
	// sb.WriteString(fmt.Sprintf("%+v\n", r.Header))
	var ptr int
	iter := 0
	for ptr < len(r.Code) {
		opcode := r.Code[ptr]
		if iter >= offset && iter < offset+height {
			name := opNames[opcode]
			if opcode > 79 && opcode <= 255 {
				name = fmt.Sprintf("%s %d", opNames[OP_PUSHD], opcode-96)
			}
			fstr := fmt.Sprintf("0x%08X: 0x%02x - %s", ptr, opcode, name)
			if iter == y {
				fstr = style.Render(fstr + " <-")
			}
			sb.WriteString(fstr + "\n")
		}
		ptr += getInstructionLength(opcode, r.Code[ptr+1])
		iter++
	}

	return sb.String()
}
