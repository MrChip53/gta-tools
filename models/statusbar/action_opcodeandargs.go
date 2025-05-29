package statusbar

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrchip53/gta-tools/rage/script/opcode"
)

// OpcodeAndArgsAction handles the two-step input for opcode and its arguments.
// It implements the Action interface.
type OpcodeAndArgsAction struct {
	id        string
	offset    int
	textInput textinput.Model
	done      bool
	resultMsg tea.Msg

	currentStep   int   // 1 for opcode, 2 for args
	enteredOpcode uint8 // Stores the opcode after step 1
}

// NewOpcodeAndArgsAction creates a new action for opcode and args input.
func NewOpcodeAndArgsAction(id string, offset int) *OpcodeAndArgsAction {
	ti := textinput.New()
	ti.Prompt = "Enter Opcode: "
	ti.CharLimit = 256
	ti.Width = 50

	return &OpcodeAndArgsAction{
		id:          id,
		offset:      offset,
		textInput:   ti,
		currentStep: 1,
	}
}

func (a *OpcodeAndArgsAction) Init() tea.Cmd {
	return a.textInput.Focus()
}

func (a *OpcodeAndArgsAction) Update(msg tea.Msg) (Action, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			inputText := strings.TrimSpace(a.textInput.Value())
			if a.currentStep == 1 { // Processing opcode
				if len(inputText) == 0 {
					cmds = append(cmds, func() tea.Msg {
						return AddStatusBarMessageMsg{Text: "Opcode cannot be empty", Duration: 3 * time.Second}
					})
					return a, tea.Batch(cmds...)
				}

				op := []byte{}
				var err error
				for k, v := range opcode.Names {
					if strings.EqualFold(v, inputText) {
						op = []byte{uint8(k)}
						break
					}
				}

				if len(op) == 0 {
					op, err = hex.DecodeString(inputText)
					if err != nil || len(op) != 1 {
						cmds = append(cmds, func() tea.Msg {
							return AddStatusBarMessageMsg{Text: "Invalid Opcode hex (must be 1 byte, e.g., 00 to FF) or name", Duration: 3 * time.Second}
						})
						return a, tea.Batch(cmds...)
					}
				}

				a.enteredOpcode = op[0]
				instructionLen := opcode.GetInstructionLength(a.enteredOpcode, 1)

				if instructionLen == 1 {
					a.resultMsg = OpcodeAndArgsInputResultMsg{
						ID:     a.id,
						Opcode: a.enteredOpcode,
						Args:   []byte{},
					}
					a.done = true
					return a, nil
				} else {
					a.currentStep = 2
					a.textInput.SetValue("")
					a.textInput.Prompt = "Enter Args (hex): "
					return a, a.textInput.Focus()
				}
			} else if a.currentStep == 2 { // Processing args
				var decodedArgs []byte
				var err error
				peekByte := uint8(0)

				if inputText != "" {
					decodedArgs, err = hex.DecodeString(inputText)
					if err != nil {
						cmds = append(cmds, func() tea.Msg {
							return AddStatusBarMessageMsg{Text: "Invalid Args hex string", Duration: 3 * time.Second}
						})
						return a, tea.Batch(cmds...)
					}
					if len(decodedArgs) > 0 {
						peekByte = decodedArgs[0]
					}
				} else {
					decodedArgs = []byte{}
				}

				expectedArgLength := opcode.GetInstructionLength(a.enteredOpcode, peekByte) - 1
				if expectedArgLength < 0 {
					expectedArgLength = 0
				}

				if len(decodedArgs) != expectedArgLength {
					msgText := fmt.Sprintf("Invalid Args length. Expected %d, got %d", expectedArgLength, len(decodedArgs))
					cmds = append(cmds, func() tea.Msg {
						return AddStatusBarMessageMsg{Text: msgText, Duration: 3 * time.Second}
					})
					return a, tea.Batch(cmds...)
				}

				a.resultMsg = OpcodeAndArgsInputResultMsg{
					ID:     a.id,
					Opcode: a.enteredOpcode,
					Args:   decodedArgs,
				}
				a.done = true
				return a, nil
			}

		case tea.KeyEsc:
			a.resultMsg = ActionCancelledMsg{ActionID: a.id}
			a.done = true
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.textInput, cmd = a.textInput.Update(msg)
	cmds = append(cmds, cmd)
	return a, tea.Batch(cmds...)
}

func (a *OpcodeAndArgsAction) View() string {
	return a.textInput.View()
}

func (a *OpcodeAndArgsAction) Description() string {
	v := strings.TrimSpace(a.textInput.Value())
	if a.currentStep == 1 { // Opcode input
		if len(v) == 2 { // Potential hex opcode
			h, err := hex.DecodeString(v)
			if err == nil && len(h) == 1 {
				if name, ok := opcode.Names[h[0]]; ok {
					return fmt.Sprintf("%s (0x%02X)", name, h[0])
				}
			}
		}
		for k, name := range opcode.Names {
			if strings.EqualFold(v, name) {
				return fmt.Sprintf("%s (0x%02X)", name, k)
			}
		}
		return "Input opcode name or hex (e.g., 0A, FF)"
	} else if a.currentStep == 2 { // Args input
		if v == "" {
			if opcode.GetInstructionLength(a.enteredOpcode, 0)-1 == 0 {
				opc := opcode.NewInstruction(a.offset, a.enteredOpcode, []byte{})
				return opc.String("#FFFF00", nil)
			}
			return "Enter arguments in hex (e.g., 01020A). Leave empty if none."
		}
		h, err := hex.DecodeString(v)
		if err != nil {
			return "Invalid hex string for arguments"
		}

		peekByte := uint8(0)
		if len(h) > 0 {
			peekByte = h[0]
		}
		expectedLen := opcode.GetInstructionLength(a.enteredOpcode, peekByte) - 1
		if expectedLen < 0 {
			expectedLen = 0
		}

		if len(h) == expectedLen {
			opc := opcode.NewInstruction(a.offset, a.enteredOpcode, h)
			return opc.String("#FFFF00", nil)
		} else {
			return fmt.Sprintf("Args hex (expected length %d based on first byte or type)", expectedLen)
		}
	}
	return ""
}

func (a *OpcodeAndArgsAction) ID() string {
	return a.id
}

func (a *OpcodeAndArgsAction) IsDone() bool {
	return a.done
}

func (a *OpcodeAndArgsAction) Result() tea.Msg {
	return a.resultMsg
}
