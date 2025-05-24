package script

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/mrchip53/gta-tools/rage/img"
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

func (h scriptHeader) Bytes() []byte {
	buf := make([]byte, 28)
	binary.LittleEndian.PutUint32(buf[0:4], h.Identifier)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(h.CodeSize))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(h.LocalVarCount))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(h.GlobalVarCount))
	binary.LittleEndian.PutUint32(buf[16:20], uint32(h.ScriptFlags))
	binary.LittleEndian.PutUint32(buf[20:24], uint32(h.GlobalsSignature))
	if h.Identifier == HEADER_MAGIC_ENCRYPTED_COMPRESSED {
		binary.LittleEndian.PutUint32(buf[24:28], uint32(h.CompressedSize))
	} else {
		buf = buf[:24]
	}
	return buf
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

	Entry *img.ImgEntry

	// Render options
	showBytecode bool
}

func NewRageScript(entry *img.ImgEntry) RageScript {
	data := entry.Data()
	s, h := newScriptHeader(data)

	encrypted := h.Identifier == HEADER_MAGIC_ENCRYPTED
	compressed := h.Identifier == HEADER_MAGIC_ENCRYPTED_COMPRESSED

	var code, l, g []byte
	if compressed {
		ed := data[s : s+h.CompressedSize]
		cd := util.Decrypt(ed)
		_ = cd
		panic("compressed")
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
		Name:        entry.Name(),
		Header:      h,
		Code:        code,
		Locals:      locals,
		Globals:     globals,
		Unsupported: compressed,
		Subroutines: make(map[int]string),
		Entry:       entry,
	}

	script.disassemble()

	return script
}

func (r *RageScript) disassemble() {
	r.Opcodes = make([]opcode.Instruction, 0)
	offsetToInstructionMap := make(map[int]opcode.Instruction)
	var ptr int
	for ptr < len(r.Code) {
		c := r.Code[ptr]
		l := opcode.GetInstructionLength(c, r.Code[ptr+1])
		args := make([]byte, l-1)
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
		offsetToInstructionMap[ptr] = ins
		ptr += l
	}

	for _, ins := range r.Opcodes {
		if branchIns, ok := ins.(*opcode.Branch); ok {
			targetOffset := branchIns.GetOperands()[0].(uint32)
			if targetIns, found := offsetToInstructionMap[int(targetOffset)]; found {
				branchIns.TargetInstruction = targetIns
			}
		}
	}
}

func (r *RageScript) Rebuild() {
	newCode := make([]byte, 0)
	currentOffset := 0

	for _, ins := range r.Opcodes {
		ins.SetOffset(currentOffset)
		currentOffset += ins.GetLength()
	}

	newCode = make([]byte, 0)
	r.Subroutines = make(map[int]string)
	for _, ins := range r.Opcodes {
		if ins.GetOpcode() == opcode.OP_FN_BEGIN {
			r.Subroutines[ins.GetOffset()] = fmt.Sprintf("sub_0x%04X", ins.GetOffset())
		}
		if branchIns, ok := ins.(*opcode.Branch); ok {
			if branchIns.TargetInstruction != nil {
				branchIns.UpdateTargetOffset(branchIns.TargetInstruction.GetOffset())
			}
		}

		newCode = append(newCode, ins.GetOpcode())
		newCode = append(newCode, ins.GetArgs()...)
	}

	r.Code = newCode
	r.Header.CodeSize = int32(len(r.Code))
	r.Entry.SetData(r.Bytes())
}

func (r *RageScript) MoveInstruction(index int, offset int) {
	if index < 0 || index >= len(r.Opcodes) {
		fmt.Println("Error: MoveInstruction index out of bounds")
		return
	}

	ins := r.Opcodes[index]
	r.Opcodes = append(r.Opcodes[:index], r.Opcodes[index+1:]...)
	r.Opcodes = append(r.Opcodes[:offset], append([]opcode.Instruction{ins}, r.Opcodes[offset:]...)...)

	r.Rebuild()
}

func (r *RageScript) DuplicateInstruction(index int) {
	if index < 0 || index >= len(r.Opcodes) {
		fmt.Println("Error: DuplicateInstruction index out of bounds")
		return
	}

	originalIns := r.Opcodes[index]
	originalOpcode := originalIns.GetOpcode()
	originalArgs := originalIns.GetArgs()

	newArgs := make([]byte, len(originalArgs))
	copy(newArgs, originalArgs)

	var newIns opcode.Instruction
	tempOffset := originalIns.GetOffset()

	if originalOpcode > 0x4F && originalOpcode <= 0xFF {
		newIns = opcode.NewPush(tempOffset, originalOpcode, newArgs)
	} else {
		f, ok := opcode.Instructions[originalOpcode]
		if ok {
			newIns = f(tempOffset, originalOpcode, newArgs)
		} else {
			newIns = opcode.NewInstruction(tempOffset, originalOpcode, newArgs)
		}
	}

	r.Opcodes = append(r.Opcodes[:index+1], append([]opcode.Instruction{newIns}, r.Opcodes[index+1:]...)...)

	r.Rebuild()
}

func (r *RageScript) EditInstruction(index int, newIns opcode.Instruction) {
	if index < 0 || index >= len(r.Opcodes) {
		fmt.Println("Error: EditInstruction index out of bounds")
		return
	}

	r.Opcodes[index] = newIns

	r.Rebuild()
}

func (r *RageScript) InsertInstruction(index int, newIns opcode.Instruction) {
	if index < 0 || index > len(r.Opcodes) {
		fmt.Println("Error: InsertInstruction index out of bounds")
		return
	}

	if newIns == nil {
		fmt.Println("Error: InsertInstruction cannot insert a nil instruction")
		return
	}

	r.Opcodes = append(r.Opcodes[:index], append([]opcode.Instruction{newIns}, r.Opcodes[index:]...)...)

	r.Rebuild()
}

func (r *RageScript) RemoveInstruction(index int) {
	if index < 0 || index >= len(r.Opcodes) {
		fmt.Println("Error: RemoveInstruction index out of bounds")
		return
	}

	r.Opcodes = append(r.Opcodes[:index], r.Opcodes[index+1:]...)

	r.Rebuild()
}

func (r *RageScript) Bytes() []byte {
	headerBytes := r.Header.Bytes()

	currentCode := make([]byte, len(r.Code))
	copy(currentCode, r.Code)

	needsEncryption := r.Header.Identifier == HEADER_MAGIC_ENCRYPTED || r.Header.Identifier == HEADER_MAGIC_ENCRYPTED_COMPRESSED

	if needsEncryption {
		util.Encrypt(currentCode)
	}

	// If the script was originally compressed (and is thus 'Unsupported' by current loading logic)
	if r.Unsupported { // This implies r.Header.Identifier == HEADER_MAGIC_ENCRYPTED_COMPRESSED
		// For compressed scripts, NewRageScript (if it didn't panic) would only provide decompressed code.
		// Locals and globals are not parsed from the stream in that path.
		// The output will be header + encrypted(decompressed_code).
		// This will not match original CompressedSize, but it's the best representation possible.
		panic("compressed")
	}

	localsData := make([]byte, int(r.Header.LocalVarCount)*4)
	for i := 0; i < int(r.Header.LocalVarCount); i++ {
		val := uint32(0)
		if i < len(r.Locals) {
			val = uint32(r.Locals[i])
		}
		binary.LittleEndian.PutUint32(localsData[i*4:(i+1)*4], val)
	}

	globalsData := make([]byte, int(r.Header.GlobalVarCount)*4)
	for i := 0; i < int(r.Header.GlobalVarCount); i++ {
		val := uint32(0)
		if i < len(r.Globals) {
			val = uint32(r.Globals[i])
		}
		binary.LittleEndian.PutUint32(globalsData[i*4:(i+1)*4], val)
	}

	if r.Header.Identifier == HEADER_MAGIC_ENCRYPTED {
		util.Encrypt(localsData)
		util.Encrypt(globalsData)
	}

	var result []byte
	result = append(result, headerBytes...)
	result = append(result, currentCode...)
	result = append(result, localsData...)
	result = append(result, globalsData...)
	return result
}

func (r RageScript) GetOffset(line int) int {
	if line < 0 || line >= len(r.Opcodes) {
		return -1
	}
	return r.Opcodes[line].GetOffset()
}

func (r *RageScript) FindNextOpcode(searchTerm string, startIndex int, reverseSearch bool) int {
	if len(r.Opcodes) == 0 {
		return -1
	}

	normalizedSearchTerm := strings.ToLower(searchTerm)

	if !reverseSearch {
		for i := startIndex + 1; i < len(r.Opcodes); i++ {
			ins := r.Opcodes[i]
			opString := ins.String("", r.Subroutines)
			if strings.Contains(strings.ToLower(opString), normalizedSearchTerm) {
				return i
			}
		}

		for i := 0; i <= startIndex; i++ {
			ins := r.Opcodes[i]
			opString := ins.String("", r.Subroutines)
			if strings.Contains(strings.ToLower(opString), normalizedSearchTerm) {
				return i
			}
		}
	} else {
		for i := startIndex - 1; i >= 0; i-- {
			ins := r.Opcodes[i]
			opString := ins.String("", r.Subroutines)
			if strings.Contains(strings.ToLower(opString), normalizedSearchTerm) {
				return i
			}
		}

		for i := len(r.Opcodes) - 1; i >= startIndex; i-- {
			ins := r.Opcodes[i]
			opString := ins.String("", r.Subroutines)
			if strings.Contains(strings.ToLower(opString), normalizedSearchTerm) {
				return i
			}
		}
	}

	return -1
}

func (r *RageScript) ToggleByteCode() {
	r.showBytecode = !r.showBytecode
}

func (r RageScript) String(y int, offset, height int) string {
	var sb strings.Builder

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
			if r.showBytecode {
				strs := []string{fmt.Sprintf("%02X", ins.GetOpcode())}
				for _, b := range ins.GetArgs() {
					strs = append(strs, fmt.Sprintf("%02X", b))
				}
				sb.WriteString(style.Render(" [" + strings.Join(strs, " ") + "]"))
			}
		}

		if context != "" {
			sb.WriteString(" " + context)
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
