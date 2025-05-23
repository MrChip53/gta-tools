package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
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

type CancelInputActionMsg struct{}

type AddStatusBarMessageMsg struct {
	Text     string
	Duration time.Duration
}

type RemoveStatusBarMessageMsg struct {
	ID string
}

type InputAction struct {
	id        string
	prompt    string
	textInput textinput.Model
}

func NewInputAction(id string, userPrompt string) *InputAction {
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
		id:        id,
		prompt:    userPrompt,
		textInput: ti,
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

type StatusBar struct {
	action tea.Model

	active bool

	segments     []Segment
	messageQueue []DisplayMessage
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
		action := NewInputAction(msg.ID, msg.Prompt)
		m.action = action
		m.active = true
		if initCmd := m.action.Init(); initCmd != nil {
			cmds = append(cmds, initCmd)
		}

	case SubmitInputActionMsg:
		m.action = nil
		m.active = false

	case CancelInputActionMsg:
		m.action = nil
		m.active = false

	case AddStatusBarMessageMsg:
		newMessage := DisplayMessage{
			ID:           fmt.Sprintf("%d-%s", time.Now().UnixNano(), msg.Text), // Unique ID
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
	if m.active && m.action != nil {
		return m.action.View()
	}

	if len(m.messageQueue) > 0 {
		// Display the first message in the queue
		// More sophisticated rendering could show multiple or allow scrolling
		return m.messageQueue[0].Text
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
