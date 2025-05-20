package script

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

var opNames = map[uint8]string{
	OP_ADD:           "ADD",
	OP_SUB:           "SUB",
	OP_MUL:           "MUL",
	OP_DIV:           "DIV",
	OP_MOD:           "MOD",
	OP_IS_ZERO:       "IS_ZERO",
	OP_NEG:           "NEG",
	OP_CMP_EQ:        "CMP_EQ",
	OP_CMP_NE:        "CMP_NE",
	OP_CMP_GT:        "CMP_GT",
	OP_CMP_GE:        "CMP_GE",
	OP_CMP_LT:        "CMP_LT",
	OP_CMP_LE:        "CMP_LE",
	OP_ADDF:          "ADDF",
	OP_SUBF:          "SUBF",
	OP_MULF:          "MULF",
	OP_DIVF:          "DIVF",
	OP_MODF:          "MODF",
	OP_NEGF:          "NEGF",
	OP_CMP_EQF:       "CMP_EQF",
	OP_CMP_NEF:       "CMP_NEF",
	OP_CMP_GTF:       "CMP_GTF",
	OP_CMP_GEF:       "CMP_GEF",
	OP_CMP_LTF:       "CMP_LTF",
	OP_CMP_LEF:       "CMP_LEF",
	OP_ADD_VEC:       "ADD_VEC",
	OP_SUB_VEC:       "SUB_VEC",
	OP_MUL_VEC:       "MUL_VEC",
	OP_DIV_VEC:       "DIV_VEC",
	OP_NEG_VEC:       "NEG_VEC",
	OP_AND:           "AND",
	OP_OR:            "OR",
	OP_XOR:           "XOR",
	OP_JUMP:          "JUMP",
	OP_JUMP_FALSE:    "JUMP_FALSE",
	OP_JUMP_TRUE:     "JUMP_TRUE",
	OP_TO_F:          "TO_F",
	OP_FROM_F:        "FROM_F",
	OP_VEC_FROM_F:    "VEC_FROM_F",
	OP_PUSHS:         "PUSHS",
	OP_PUSH:          "PUSH",
	OP_PUSHF:         "PUSHF",
	OP_DUP:           "DUP",
	OP_POP:           "POP",
	OP_CALL_NATIVE:   "CALL_NATIVE",
	OP_CALL:          "CALL",
	OP_FN_BEGIN:      "FN_BEGIN",
	OP_FN_END:        "FN_END",
	OP_REF_GET:       "REF_GET",
	OP_REF_SET:       "REF_SET",
	OP_REF_PEEK_SET:  "REF_PEEK_SET",
	OP_ARRAY_EXPLODE: "ARRAY_EXPLODE",
	OP_ARRAY_IMPLODE: "ARRAY_IMPLODE",
	OP_VAR0:          "VAR0",
	OP_VAR1:          "VAR1",
	OP_VAR2:          "VAR2",
	OP_VAR3:          "VAR3",
	OP_VAR4:          "VAR4",
	OP_VAR5:          "VAR5",
	OP_VAR6:          "VAR6",
	OP_VAR7:          "VAR7",
	OP_VAR:           "VAR",
	OP_LOCAL_VAR:     "LOCAL_VAR",
	OP_GLOBAL_VAR:    "GLOBAL_VAR",
	OP_ARRAY_REF:     "ARRAY_REF",
	OP_SWITCH:        "SWITCH",
	OP_PUSH_STRING:   "PUSH_STRING",
	OP_NULL_OBJ:      "NULL_OBJ",
	OP_STR_CPY:       "STR_CPY",
	OP_INT_TO_STR:    "INT_TO_STR",
	OP_STR_CAT:       "STR_CAT",
	OP_STR_CAT_I:     "STR_CAT_I",
	OP_CATCH:         "CATCH",
	OP_THROW:         "THROW",
	OP_STR_VAR_CPY:   "STR_VAR_CPY",
	OP_GET_PROTECT:   "GET_PROTECT",
	OP_SET_PROTECT:   "SET_PROTECT",
	OP_REF_PROTECT:   "REF_PROTECT",
	OP_ABORT_79:      "ABORT_79",
	OP_PUSHD:         "PUSHD",
}

func getInstructionLength(opcode, p1 uint8) int {
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
