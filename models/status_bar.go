package models

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

type AddStatusBarMessageMsg struct {
	Text     string
	Duration time.Duration
}

type RemoveStatusBarMessageMsg struct {
	ID string
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
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
		return m, timeoutCmd

	case RemoveStatusBarMessageMsg:
		var newQueue []DisplayMessage
		for _, item := range m.messageQueue {
			if item.ID != msg.ID {
				newQueue = append(newQueue, item)
			}
		}
		m.messageQueue = newQueue
		return m, nil

	case tea.KeyMsg:
		if !m.active {
			return m, nil
		}
		switch msg.String() {
		case "esc":
			m.action = nil
			m.active = false
			return m, nil
		default:
			if m.action != nil {
				m.action, cmd = m.action.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
			// Else: Potentially handle other keys for status bar itself (e.g., activate action mode)
		}

	default:
		if m.active && m.action != nil {
			m.action, cmd = m.action.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
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
