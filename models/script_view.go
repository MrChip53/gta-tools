package models

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrchip53/gta-tools/rage/script"
)

var highlightedLine = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)

type ScriptView struct {
	data            []byte
	script          script.RageScript
	vp              viewport.Model
	active          bool
	highlightedLine int
	codeOffset      int
}

func NewScriptView(name string, data []byte, w, h int) ScriptView {
	vp := viewport.New(w, h)

	if data == nil {
		vp.SetContent("No script selected")
		return ScriptView{data: data, vp: vp}
	}

	script := script.NewRageScript(name, data)
	str := lipgloss.NewStyle().Width(w).Render(script.String(0, highlightedLine, 0, h))
	vp.SetContent(str)
	return ScriptView{data: data, script: script, vp: vp}
}

func (m ScriptView) Init() tea.Cmd {
	return nil
}

func (m ScriptView) Update(msg tea.Msg) (ScriptView, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			m.highlightedLine--
			if m.highlightedLine < 0 {
				m.highlightedLine = 0
			}

			if m.highlightedLine < m.codeOffset {
				m.codeOffset--
			}
			m.Refresh()
		case "down":
			m.highlightedLine++

			if m.highlightedLine > m.codeOffset+m.vp.Height-1 {
				m.codeOffset++
			}
			m.Refresh()
		case "pgup":
			m.highlightedLine -= m.vp.Height
			if m.highlightedLine < 0 {
				m.highlightedLine = 0
				m.codeOffset = 0
			}
			if m.highlightedLine < m.codeOffset {
				m.codeOffset -= m.vp.Height
			}
			m.Refresh()
		case "pgdown":
			m.highlightedLine += m.vp.Height
			if m.highlightedLine > m.codeOffset+m.vp.Height-1 {
				m.codeOffset += m.vp.Height
			}
			m.Refresh()
		}
	}

	// m.vp, cmd = m.vp.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *ScriptView) Refresh() {
	if m.data == nil {
		m.vp.SetContent("No script selected")
		return
	}

	str := lipgloss.NewStyle().Width(m.vp.Width).Render(m.script.String(m.highlightedLine, highlightedLine, m.codeOffset, m.vp.Height))
	m.vp.SetContent(str)
}

func (m ScriptView) View() string {
	return m.vp.View()
}

func (m *ScriptView) SetActive(active bool) {
	m.active = active
}
