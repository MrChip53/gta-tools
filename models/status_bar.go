package models

import (
	"fmt"
	"strings"
	"time"

	"encoding/hex"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrchip53/gta-tools/rage/script/opcode"
)

type Segment struct {
	Text string
}

type DisplayMessage struct {
	ID           string
	Text         string
	DisplayUntil time.Time
}

type ActivateInputActionMsg struct {
	ID     string
	Prompt string
}

type SubmitInputActionMsg struct {
	ID        string
	InputText string
}

type CancelInputActionMsg struct{}

type AddStatusBarMessageMsg struct {
	Text     string
	Duration time.Duration
}

type RemoveStatusBarMessageMsg struct {
	ID string
}

type ActivateImportFileActionMsg struct {
	ID string
}

type ImportFileActionMsg struct {
	ID          string
	HostPath    string
	ArchivePath string
}

type InputAction struct {
	id              string
	prompt          string
	textInput       textinput.Model
	descriptionFunc func(*InputAction) string
}

func NewInputAction(id string, userPrompt string, descriptionFunc func(*InputAction) string) *InputAction {
	ti := textinput.New()
	ti.Prompt = "❯ "
	if userPrompt != "" {
		if !strings.HasSuffix(userPrompt, " ") {
			ti.Prompt = userPrompt + " "
		} else {
			ti.Prompt = userPrompt
		}
	}
	ti.CharLimit = 256
	ti.Width = 40
	return &InputAction{
		id:              id,
		prompt:          userPrompt,
		textInput:       ti,
		descriptionFunc: descriptionFunc,
	}
}

func (ia *InputAction) Init() tea.Cmd {
	return ia.textInput.Focus()
}

func (ia *InputAction) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			inputValue := ia.textInput.Value()
			return ia, func() tea.Msg {
				return SubmitInputActionMsg{ID: ia.id, InputText: inputValue}
			}
		case tea.KeyEsc:
			return ia, func() tea.Msg {
				return CancelInputActionMsg{}
			}
		}
	}

	updatedTi, tiCmd := ia.textInput.Update(msg)
	ia.textInput = updatedTi
	return ia, tiCmd
}

func (ia *InputAction) View() string {
	return ia.textInput.View()
}

func (ia *InputAction) Description() string {
	if ia.descriptionFunc != nil {
		return ia.descriptionFunc(ia)
	}
	return ""
}

type ActivateOpcodeAndArgsInputMsg struct {
	ID     string
	Offset int
}

type OpcodeAndArgsInputResultMsg struct {
	ID     string
	Opcode uint8
	Args   []byte
}

type OpcodeArgsState struct {
	ID          string
	Opcode      uint8
	CurrentStep int
	Offset      int
}

// New state for file import
type ImportFileState struct {
	ID          string
	CurrentStep int // 1 for host path, 2 for archive path
	HostPath    string
}

type StatusBar struct {
	action tea.Model

	active bool

	segments        []Segment
	messageQueue    []DisplayMessage
	opcodeArgsState *OpcodeArgsState
	importFileState *ImportFileState
}

func NewStatusBar() StatusBar {
	return StatusBar{
		active:       false,
		segments:     make([]Segment, 0),
		messageQueue: make([]DisplayMessage, 0),
	}
}

func (m StatusBar) Init() tea.Cmd {
	return nil
}

func (m StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ActivateInputActionMsg:
		action := NewInputAction(msg.ID, msg.Prompt, nil)
		m.action = action
		m.active = true
		m.opcodeArgsState = nil
		m.importFileState = nil
		if initCmd := m.action.Init(); initCmd != nil {
			cmds = append(cmds, initCmd)
		}

	case ActivateOpcodeAndArgsInputMsg:
		m.opcodeArgsState = &OpcodeArgsState{ID: msg.ID, CurrentStep: 1, Offset: msg.Offset}
		action := NewInputAction(msg.ID+"_opcode", "Enter Opcode:", func(ia *InputAction) string {
			v := strings.TrimSpace(ia.textInput.Value())
			if len(v) == 2 {
				h, err := hex.DecodeString(v)
				if err == nil {
					if name, ok := opcode.Names[h[0]]; ok {
						return fmt.Sprintf("%s (0x%02X)", name, h[0])
					}
				}
			}

			for k, vv := range opcode.Names {
				if strings.EqualFold(v, vv) {
					return fmt.Sprintf("%s (0x%02X)", vv, k)
				}
			}

			return "Input opcode in hexadecimal format (e.g., 0A, FF) or name"
		})
		m.action = action
		m.active = true
		m.importFileState = nil
		if initCmd := m.action.Init(); initCmd != nil {
			cmds = append(cmds, initCmd)
		}

	case ActivateImportFileActionMsg:
		m.importFileState = &ImportFileState{ID: msg.ID, CurrentStep: 1}
		action := NewInputAction(msg.ID+"_hostpath", "Enter host OS file path:", nil)
		m.action = action
		m.active = true
		m.opcodeArgsState = nil // Clear other multi-step states
		if initCmd := m.action.Init(); initCmd != nil {
			cmds = append(cmds, initCmd)
		}

	case SubmitInputActionMsg:
		if m.opcodeArgsState != nil {
			if m.opcodeArgsState.CurrentStep == 1 {
				inputText := strings.TrimSpace(msg.InputText)
				if len(inputText) == 0 {
					addMsgCmd := func() tea.Msg {
						return AddStatusBarMessageMsg{Text: "Opcode cannot be empty", Duration: 3 * time.Second}
					}
					cmds = append(cmds, addMsgCmd)
					return m, tea.Batch(cmds...)
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
						addMsgCmd := func() tea.Msg {
							return AddStatusBarMessageMsg{Text: "Invalid Opcode hex (must be 1 byte, e.g., 00 to FF) or name", Duration: 3 * time.Second}
						}
						cmds = append(cmds, addMsgCmd)
						// Keep the input active, similar to empty input case
						// Potentially clear the input field: m.action.(*InputAction).textInput.SetValue("")
						return m, tea.Batch(cmds...)
					}
				}

				l := opcode.GetInstructionLength(op[0], 1)
				if l == 1 {
					id := m.opcodeArgsState.ID
					resultCmd := func() tea.Msg {
						return OpcodeAndArgsInputResultMsg{
							ID:     id,
							Opcode: op[0],
							Args:   []byte{},
						}
					}
					cmds = append(cmds, resultCmd)
					m.action = nil
					m.active = false
					m.opcodeArgsState = nil
					return m, tea.Batch(cmds...)
				} else {
					m.opcodeArgsState.Opcode = op[0]
					m.opcodeArgsState.CurrentStep = 2
					action := NewInputAction(m.opcodeArgsState.ID+"_args", "Enter Args (hex):", func(ia *InputAction) string {
						op := strings.TrimSpace(ia.textInput.Value())
						if op == "" {
							return ""
						}
						h, err := hex.DecodeString(op)
						if err != nil {
							return "Invalid hex string"
						}
						l := opcode.GetInstructionLength(m.opcodeArgsState.Opcode, h[0])
						if l != len(h)+1 {
							return "Invalid args length"
						}
						opc := opcode.NewInstruction(m.opcodeArgsState.Offset, m.opcodeArgsState.Opcode, h)
						return opc.String("#FFFF00", nil)
					})
					m.action = action
					if initCmd := m.action.Init(); initCmd != nil {
						cmds = append(cmds, initCmd)
					}
				}
			} else if m.opcodeArgsState.CurrentStep == 2 {
				inputText := strings.TrimSpace(msg.InputText)
				var decodedArgs []byte
				var err error
				var h1 uint8
				if inputText != "" {
					decodedArgs, err = hex.DecodeString(inputText)
					if err != nil {
						addMsgCmd := func() tea.Msg {
							return AddStatusBarMessageMsg{Text: "Invalid Args hex string", Duration: 3 * time.Second}
						}
						cmds = append(cmds, addMsgCmd)

						return m, tea.Batch(cmds...)
					}
					h1 = decodedArgs[0]
				} else {
					decodedArgs = []byte{}
				}

				opc := m.opcodeArgsState.Opcode

				argLength := opcode.GetInstructionLength(opc, h1) - 1
				if argLength != len(decodedArgs) {
					addMsgCmd := func() tea.Msg {
						return AddStatusBarMessageMsg{Text: "Invalid Args length", Duration: 3 * time.Second}
					}
					cmds = append(cmds, addMsgCmd)
					return m, tea.Batch(cmds...)
				}

				m.action = nil
				m.active = false

				id := m.opcodeArgsState.ID
				resultCmd := func() tea.Msg {
					return OpcodeAndArgsInputResultMsg{
						ID:     id,
						Opcode: opc,
						Args:   decodedArgs,
					}
				}
				cmds = append(cmds, resultCmd)
				m.opcodeArgsState = nil
			}
		} else if m.importFileState != nil {
			inputText := strings.TrimSpace(msg.InputText)
			if m.importFileState.CurrentStep == 1 { // Submitted host OS path
				if inputText == "" {
					addMsgCmd := func() tea.Msg {
						return AddStatusBarMessageMsg{Text: "Host OS file path cannot be empty", Duration: 3 * time.Second}
					}
					cmds = append(cmds, addMsgCmd)
					return m, tea.Batch(cmds...)
				}
				m.importFileState.HostPath = inputText
				m.importFileState.CurrentStep = 2
				action := NewInputAction(m.importFileState.ID+"_archivepath", "Enter desired archive filename:", nil)
				m.action = action
				// No need to set m.active = true, it already is
				if initCmd := m.action.Init(); initCmd != nil {
					cmds = append(cmds, initCmd)
				}
			} else if m.importFileState.CurrentStep == 2 { // Submitted archive filename
				if inputText == "" {
					addMsgCmd := func() tea.Msg {
						return AddStatusBarMessageMsg{Text: "Archive filename cannot be empty", Duration: 3 * time.Second}
					}
					cmds = append(cmds, addMsgCmd)
					return m, tea.Batch(cmds...)
				}
				id := m.importFileState.ID
				hostPath := m.importFileState.HostPath
				resultCmd := func() tea.Msg {
					return ImportFileActionMsg{
						ID:          id,
						HostPath:    hostPath,
						ArchivePath: inputText,
					}
				}
				cmds = append(cmds, resultCmd)
				m.action = nil
				m.active = false
				m.importFileState = nil
			}
		} else {
			originalSubmitCmd := func() tea.Msg {
				return msg
			}
			cmds = append(cmds, originalSubmitCmd)

			m.action = nil
			m.active = false
		}

	case CancelInputActionMsg:
		m.action = nil
		m.active = false
		m.opcodeArgsState = nil
		m.importFileState = nil

	case AddStatusBarMessageMsg:
		newMessage := DisplayMessage{
			ID:           fmt.Sprintf("%d-%s", time.Now().UnixNano(), msg.Text),
			Text:         msg.Text,
			DisplayUntil: time.Now().Add(msg.Duration),
		}
		m.messageQueue = append(m.messageQueue, newMessage)

		messageID := newMessage.ID
		timeoutCmd := tea.Tick(msg.Duration, func(t time.Time) tea.Msg {
			return RemoveStatusBarMessageMsg{ID: messageID}
		})
		cmds = append(cmds, timeoutCmd)

	case RemoveStatusBarMessageMsg:
		var newQueue []DisplayMessage
		for _, item := range m.messageQueue {
			if item.ID != msg.ID {
				newQueue = append(newQueue, item)
			}
		}
		m.messageQueue = newQueue

	case tea.KeyMsg:
		if m.active && m.action != nil {
			updatedAction, actionCmd := m.action.Update(msg)
			m.action = updatedAction
			if actionCmd != nil {
				cmds = append(cmds, actionCmd)
			}
		} else if m.active && msg.String() == "esc" {
			m.action = nil
			m.active = false
		}

	default:
		if m.active && m.action != nil {
			updatedAction, actionCmd := m.action.Update(msg)
			m.action = updatedAction
			if actionCmd != nil {
				cmds = append(cmds, actionCmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m StatusBar) View() string {
	actionView := ""
	actionDescription := ""
	if m.active && m.action != nil {
		actionView = m.action.View()
		if inputAction, ok := m.action.(*InputAction); ok {
			actionDescription = inputAction.Description()
		}
	}

	finalView := actionView
	if actionDescription != "" {
		if finalView != "" {
			finalView += " "
		}
		finalView += "[" + actionDescription + "]"
	}

	if len(m.messageQueue) > 0 {
		if finalView != "" {
			finalView += " | "
		}
		finalView += m.messageQueue[0].Text
		return finalView
	}

	if finalView != "" {
		return finalView
	}

	if len(m.segments) > 0 {
		var parts []string
		for _, segment := range m.segments {
			parts = append(parts, segment.Text)
		}
		return strings.Join(parts, " | ")
	}

	return ""
}

func (m *StatusBar) SetActive(active bool) {
	m.active = active
}

func (m *StatusBar) HasAction() bool {
	return m.active && m.action != nil
}
