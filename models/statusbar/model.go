package statusbar

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

type SubmitScriptFlagsMsg struct {
	Flags int32
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

type ActivateScriptFlagsActionMsg struct {
	ID string
}

type ImportFileActionMsg struct {
	ID          string
	HostPath    string
	ArchivePath string
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

type Model struct {
	currentAction Action

	segments     []Segment
	messageQueue []DisplayMessage
}

func New() Model {
	return Model{
		segments:     make([]Segment, 0),
		messageQueue: make([]DisplayMessage, 0),
	}
}

func (m Model) Init() tea.Cmd {
	if m.currentAction != nil {
		return m.currentAction.Init()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.currentAction != nil {
		var actionCmd tea.Cmd
		m.currentAction, actionCmd = m.currentAction.Update(msg)
		if actionCmd != nil {
			cmds = append(cmds, actionCmd)
		}

		if m.currentAction != nil && m.currentAction.IsDone() {
			resultMsg := m.currentAction.Result()
			if resultMsg != nil {
				switch res := resultMsg.(type) {
				case SubmitScriptFlagsMsg:
					cmds = append(cmds, func() tea.Msg { return res })
				case OpcodeAndArgsInputResultMsg:
					cmds = append(cmds, func() tea.Msg { return res })
				case ImportFileActionMsg:
					cmds = append(cmds, func() tea.Msg { return res })
				case SubmitInputActionMsg:
					cmds = append(cmds, func() tea.Msg { return res })
				case ActionCancelledMsg:
					// do nothing
				default:
					// Potentially log an unhandled result type
				}
			}
			m.currentAction = nil
		}
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case ActivateScriptFlagsActionMsg:
		m.currentAction = NewSetScriptFlagsAction(msg.ID, "Enter Script Flags (integer):")
		if initCmd := m.currentAction.Init(); initCmd != nil {
			cmds = append(cmds, initCmd)
		}

	case ActivateOpcodeAndArgsInputMsg:
		m.currentAction = NewOpcodeAndArgsAction(msg.ID, msg.Offset)
		if initCmd := m.currentAction.Init(); initCmd != nil {
			cmds = append(cmds, initCmd)
		}

	case ActivateImportFileActionMsg:
		m.currentAction = NewImportFileAction(msg.ID)
		if initCmd := m.currentAction.Init(); initCmd != nil {
			cmds = append(cmds, initCmd)
		}

	case ActivateInputActionMsg:
		m.currentAction = NewGeneralPurposeInputAction(msg.ID, msg.Prompt, "")
		if initCmd := m.currentAction.Init(); initCmd != nil {
			cmds = append(cmds, initCmd)
		}

	case CancelInputActionMsg:
		if m.currentAction != nil {
			m.currentAction = nil
		}

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
		if m.currentAction != nil {
			updatedAction, actionCmd := m.currentAction.Update(msg)
			m.currentAction = updatedAction
			if actionCmd != nil {
				cmds = append(cmds, actionCmd)
			}
		} else if m.currentAction != nil && msg.String() == "esc" {
			m.currentAction = nil
		}

	default:
		// This default case for active action is now handled at the top
		// if m.currentAction != nil
		// If no action is active, this default case might handle other non-action messages
	}

	var newQueue []DisplayMessage
	stillActiveMessages := false
	now := time.Now()
	for _, item := range m.messageQueue {
		if now.Before(item.DisplayUntil) {
			newQueue = append(newQueue, item)
			stillActiveMessages = true
		}
	}
	m.messageQueue = newQueue

	if stillActiveMessages && len(m.messageQueue) > 0 {
		// If there are still messages, ensure a tick is scheduled to remove the next one if not already handled by action
		// This is a simplified re-check; ideally, AddStatusBarMessageMsg itself schedules the removal tick.
		// The current AddStatusBarMessageMsg handler *does* schedule a RemoveStatusBarMessageMsg, so this might be redundant
		// or could be a fallback if an action somehow clears messages without their specific timeout cmd.
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	actionView := ""
	actionDescription := ""
	if m.currentAction != nil {
		actionView = m.currentAction.View()
		actionDescription = m.currentAction.Description()
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

func (m *Model) HasAction() bool {
	return m.currentAction != nil
}
