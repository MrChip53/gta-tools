package models

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrchip53/gta-tools/rage/img"
	"github.com/mrchip53/gta-tools/rage/script"
	"github.com/mrchip53/gta-tools/rage/script/opcode"
)

var highlightedLine = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)

type basicItem struct {
	name string
}

func (i basicItem) FilterValue() string { return i.name }

type ScriptView struct {
	data            []byte
	script          script.RageScript
	vp              viewport.Model
	active          bool
	highlightedLine int
	codeOffset      int

	height int
	width  int

	subsList    list.Model
	localsList  list.Model
	globalsList list.Model

	listStyle lipgloss.Style
}

func NewScriptView(entry *img.ImgEntry, w, h int) ScriptView {
	vp := viewport.New(w/3*2, h-1)

	if entry == nil {
		return ScriptView{vp: vp}
	}

	data := entry.Data()

	if data == nil {
		vp.SetContent("No script selected")
		return ScriptView{data: data, vp: vp}
	}

	script := script.NewRageScript(entry)
	str := lipgloss.NewStyle().Width(w).Render(script.String(0, 0, h))
	vp.SetContent(str)

	boxHeight := (h-1)/3 - 2
	bs := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).Padding(1, 2).Width(40).Height(boxHeight)

	d := customDelegate{}

	var subs []list.Item
	var keys []int
	for k := range script.Subroutines {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, v := range keys {
		sn := script.Subroutines[v]
		subs = append(subs, listItem{name: fmt.Sprintf("%s", sn)})
	}
	sl := list.New(subs, d, 30, boxHeight-bs.GetVerticalFrameSize())
	sl.Title = "Subroutines"
	sl.SetShowStatusBar(false)
	sl.SetShowHelp(false)

	var locals []list.Item
	for i, v := range script.Locals {
		locals = append(locals, listItem{name: fmt.Sprintf("Local %d: %v", i, v)})
	}
	ll := list.New(locals, d, 30, boxHeight-bs.GetVerticalFrameSize())
	ll.Title = "Locals"
	ll.SetShowStatusBar(false)
	ll.SetShowHelp(false)

	var globals []list.Item
	for i, v := range script.Globals {
		globals = append(globals, listItem{name: fmt.Sprintf("Global %d: %v", i, v)})
	}
	gl := list.New(globals, d, 30, boxHeight-bs.GetVerticalFrameSize())
	gl.Title = "Globals"
	gl.SetShowStatusBar(false)
	gl.SetShowHelp(false)

	return ScriptView{
		data:        data,
		script:      script,
		vp:          vp,
		subsList:    sl,
		localsList:  ll,
		globalsList: gl,
		height:      h,
		width:       w,
		listStyle:   bs,
	}
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
		case "enter":
			op := m.script.Opcodes[m.highlightedLine]
			opc := op.GetOpcode()
			if opc == opcode.OP_JUMP || opc == opcode.OP_JUMP_FALSE || opc == opcode.OP_JUMP_TRUE || opc == opcode.OP_CALL {
				offset := op.GetOperands()[0].(uint32)
				for i, o := range m.script.Opcodes {
					if o.GetOffset() == int(offset) {
						m.highlightedLine = i
						break
					}
				}
				if m.highlightedLine < m.codeOffset || m.highlightedLine > m.codeOffset+m.vp.Height-1 {
					m.codeOffset = max(m.highlightedLine-5, 0)
				}
				m.Refresh()
			}
		case "a":
			nextOpIdx := m.highlightedLine + 1
			var offset int
			if nextOpIdx >= len(m.script.Opcodes) {
				offset = len(m.script.Code)
			} else {
				offset = m.script.Opcodes[nextOpIdx].GetOffset()
			}
			m.script.InsertInstruction(offset, opcode.OP_PUSHS, []byte{0x39, 0x05})
			m.script.Disassemble()
			m.Refresh()
		case "d":
			m.script.RemoveInstruction(m.highlightedLine)
			m.script.Disassemble()
			m.Refresh()
		case "t":
			cmds = append(cmds, func() tea.Msg {
				return AddStatusBarMessageMsg{
					Text:     fmt.Sprintf("%+v %d", m.script.Entry.Toc(), len(m.script.Entry.Data())),
					Duration: 5 * time.Second,
				}
			})
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

	str := lipgloss.NewStyle().Width(m.vp.Width).Render(m.script.String(m.highlightedLine, m.codeOffset, m.vp.Height))
	m.vp.SetContent(str)
}

func (m ScriptView) View() string {
	subsList := m.subsList.View()
	localsList := m.localsList.View()
	globalsList := m.globalsList.View()
	rightPane := lipgloss.JoinVertical(lipgloss.Left, m.listStyle.Render(subsList), m.listStyle.Render(localsList), m.listStyle.Render(globalsList))
	bottomPane := lipgloss.JoinHorizontal(lipgloss.Top, m.vp.View(), rightPane)
	str := lipgloss.JoinVertical(lipgloss.Center, m.script.Name, bottomPane)
	return str
}

func (m *ScriptView) SetActive(active bool) {
	m.active = active
}
