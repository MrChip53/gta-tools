package opcode

import (
	_ "embed"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	OP_PUSHD = iota
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV

	OP_MOD

	OP_IS_ZERO

	OP_NEG

	OP_CMP_EQ
	OP_CMP_NE
	OP_CMP_GT
	OP_CMP_GE
	OP_CMP_LT
	OP_CMP_LE

	OP_ADDF
	OP_SUBF
	OP_MULF
	OP_DIVF

	OP_MODF

	OP_NEGF

	OP_CMP_EQF
	OP_CMP_NEF
	OP_CMP_GTF
	OP_CMP_GEF
	OP_CMP_LTF
	OP_CMP_LEF

	OP_ADD_VEC
	OP_SUB_VEC
	OP_MUL_VEC
	OP_DIV_VEC

	OP_NEG_VEC

	OP_AND
	OP_OR
	OP_XOR

	OP_JUMP
	OP_JUMP_FALSE
	OP_JUMP_TRUE

	OP_TO_F
	OP_FROM_F

	OP_VEC_FROM_F

	OP_PUSHS
	OP_PUSH
	OP_PUSHF

	OP_DUP
	OP_POP

	OP_CALL_NATIVE
	OP_CALL

	OP_FN_BEGIN
	OP_FN_END

	OP_REF_GET
	OP_REF_SET
	OP_REF_PEEK_SET

	OP_ARRAY_EXPLODE
	OP_ARRAY_IMPLODE

	OP_VAR0
	OP_VAR1
	OP_VAR2
	OP_VAR3
	OP_VAR4
	OP_VAR5
	OP_VAR6
	OP_VAR7
	OP_VAR

	OP_LOCAL_VAR
	OP_GLOBAL_VAR

	OP_ARRAY_REF
	OP_SWITCH
	OP_PUSH_STRING
	OP_NULL_OBJ
	OP_STR_CPY
	OP_INT_TO_STR
	OP_STR_CAT
	OP_STR_CAT_I
	OP_CATCH
	OP_THROW
	OP_STR_VAR_CPY
	OP_GET_PROTECT
	OP_SET_PROTECT
	OP_REF_PROTECT
	OP_ABORT_79
)

var Names = map[uint8]string{
	OP_ADD:           "Add",
	OP_SUB:           "Sub",
	OP_MUL:           "Mul",
	OP_DIV:           "Div",
	OP_MOD:           "Mod",
	OP_IS_ZERO:       "IsZero",
	OP_NEG:           "Neg",
	OP_CMP_EQ:        "CmpEq",
	OP_CMP_NE:        "CmpNe",
	OP_CMP_GT:        "CmpGt",
	OP_CMP_GE:        "CmpGe",
	OP_CMP_LT:        "CmpLt",
	OP_CMP_LE:        "CmpLe",
	OP_ADDF:          "AddF",
	OP_SUBF:          "SubF",
	OP_MULF:          "MulF",
	OP_DIVF:          "DivF",
	OP_MODF:          "ModF",
	OP_NEGF:          "NegF",
	OP_CMP_EQF:       "CmpEqf",
	OP_CMP_NEF:       "CmpNef",
	OP_CMP_GTF:       "CmpGtf",
	OP_CMP_GEF:       "CmpGef",
	OP_CMP_LTF:       "CmpLtf",
	OP_CMP_LEF:       "CmpLef",
	OP_ADD_VEC:       "AddVec",
	OP_SUB_VEC:       "SubVec",
	OP_MUL_VEC:       "MulVec",
	OP_DIV_VEC:       "DivVec",
	OP_NEG_VEC:       "NegVec",
	OP_AND:           "And",
	OP_OR:            "Or",
	OP_XOR:           "Xor",
	OP_JUMP:          "Jump",
	OP_JUMP_FALSE:    "JumpFalse",
	OP_JUMP_TRUE:     "JumpTrue",
	OP_TO_F:          "ToF",
	OP_FROM_F:        "FromF",
	OP_VEC_FROM_F:    "VecFromF",
	OP_PUSHS:         "PushS",
	OP_PUSH:          "Push",
	OP_PUSHF:         "PushF",
	OP_DUP:           "Dup",
	OP_POP:           "Pop",
	OP_CALL_NATIVE:   "CallNative",
	OP_CALL:          "Call",
	OP_FN_BEGIN:      "FnBegin",
	OP_FN_END:        "FnEnd",
	OP_REF_GET:       "RefGet",
	OP_REF_SET:       "RefSet",
	OP_REF_PEEK_SET:  "RefPeekSet",
	OP_ARRAY_EXPLODE: "ArrayExplode",
	OP_ARRAY_IMPLODE: "ArrayImplode",
	OP_VAR0:          "Var0",
	OP_VAR1:          "Var1",
	OP_VAR2:          "Var2",
	OP_VAR3:          "Var3",
	OP_VAR4:          "Var4",
	OP_VAR5:          "Var5",
	OP_VAR6:          "Var6",
	OP_VAR7:          "Var7",
	OP_VAR:           "Var",
	OP_LOCAL_VAR:     "LocalVar",
	OP_GLOBAL_VAR:    "GlobalVar",
	OP_ARRAY_REF:     "ArrayRef",
	OP_SWITCH:        "Switch",
	OP_PUSH_STRING:   "PushString",
	OP_NULL_OBJ:      "NullObj",
	OP_STR_CPY:       "StrCpy",
	OP_INT_TO_STR:    "IntToStr",
	OP_STR_CAT:       "StrCat",
	OP_STR_CAT_I:     "StrCatI",
	OP_CATCH:         "Catch",
	OP_THROW:         "Throw",
	OP_STR_VAR_CPY:   "StrVarCpy",
	OP_GET_PROTECT:   "GetProtect",
	OP_SET_PROTECT:   "SetProtect",
	OP_REF_PROTECT:   "RefProtect",
	OP_ABORT_79:      "Abort",
	OP_PUSHD:         "PushD",
}

var Instructions = map[uint8]func(int, uint8, []byte) Instruction{
	OP_PUSHS: func(offset int, opcode uint8, args []byte) Instruction {
		return NewPush(offset, opcode, args)
	},
	OP_PUSH: func(offset int, opcode uint8, args []byte) Instruction {
		return NewPush(offset, opcode, args)
	},
	OP_PUSHF: func(offset int, opcode uint8, args []byte) Instruction {
		return NewPush(offset, opcode, args)
	},
	OP_PUSH_STRING: func(offset int, opcode uint8, args []byte) Instruction {
		return NewPush(offset, opcode, args)
	},
	OP_CALL_NATIVE: func(offset int, opcode uint8, args []byte) Instruction {
		return newNative(offset, opcode, args)
	},
	OP_JUMP: func(offset int, opcode uint8, args []byte) Instruction {
		return NewBranch(offset, opcode, args)
	},
	OP_JUMP_FALSE: func(offset int, opcode uint8, args []byte) Instruction {
		return NewBranch(offset, opcode, args)
	},
	OP_JUMP_TRUE: func(offset int, opcode uint8, args []byte) Instruction {
		return NewBranch(offset, opcode, args)
	},
	OP_CALL: func(offset int, opcode uint8, args []byte) Instruction {
		return NewBranch(offset, opcode, args)
	},
}

var (
	functionNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	branchTargetStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF"))
	highlightStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
)

//go:embed native_new.dat
var nativeNew string

var nativeFunctions map[uint32]string

func parseNativeFile(file string) map[uint32]string {
	m := make(map[uint32]string)
	lines := strings.Split(file, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" && value != "" {
				keyInt, err := strconv.ParseUint(key, 10, 32)
				if err == nil {
					m[uint32(keyInt)] = value
				}
			}
		}
	}
	return m
}

func init() {
	nativeFunctions = parseNativeFile(nativeNew)
}

type Opcode struct {
	Offset   int
	Opcode   uint8
	Args     []byte
	Operands []any
	New      bool
}

type Instruction interface {
	Disassemble()
	GetOffset() int
	GetOpcode() uint8
	GetOperands() []any
	String(color string, subroutines map[int]string) string
	GetLength() int
}

func GetInstructionLength(opcode, p1 uint8) int {
	switch opcode {
	case OP_STR_CPY, OP_INT_TO_STR, OP_STR_CAT, OP_STR_CAT_I:
		return 2
	case OP_PUSHS, OP_FN_END:
		return 3
	case OP_FN_BEGIN:
		return 4
	case OP_JUMP, OP_JUMP_FALSE, OP_JUMP_TRUE, OP_PUSH, OP_PUSHF, OP_CALL:
		return 5
	case OP_CALL_NATIVE:
		return 7
	case OP_SWITCH:
		return int(p1)*8 + 2
	case OP_PUSH_STRING:
		return int(p1) + 2
	default:
		return 1
	}
}
