package models

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrchip53/gta-tools/models/statusbar"
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

	searchText string

	marker1 int
	marker2 int

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
	str := lipgloss.NewStyle().Width(w).Render(script.String(0, 0, h, -1, -1))
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
		marker1:     -1,
		marker2:     -1,
	}
}

func (m ScriptView) Init() tea.Cmd {
	return nil
}

func (m ScriptView) Update(msg tea.Msg) (ScriptView, tea.Cmd) {
	var cmds []tea.Cmd

	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			m.scroll(-1)
		case "down":
			m.scroll(1)
		case "pgup":
			m.scroll(-m.vp.Height)
		case "pgdown":
			m.scroll(m.vp.Height)
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
		case "/":
			cmds = append(cmds, func() tea.Msg {
				return statusbar.ActivateInputActionMsg{
					ID:     "search",
					Prompt: "Search",
				}
			})
		case "n":
			m.jumpToNextSearch(false)
		case "b":
			m.jumpToNextSearch(true)
		case "i":
			offset := m.script.GetOffset(m.highlightedLine)
			cmds = append(cmds, func() tea.Msg {
				return statusbar.ActivateOpcodeAndArgsInputMsg{
					ID:     "insert",
					Offset: offset,
				}
			})
		case "e":
			offset := m.script.GetOffset(m.highlightedLine)
			cmds = append(cmds, func() tea.Msg {
				return statusbar.ActivateOpcodeAndArgsInputMsg{
					ID:     "edit",
					Offset: offset,
				}
			})
		case "o":
			cmds = append(cmds, func() tea.Msg {
				return statusbar.AddStatusBarMessageMsg{
					Text:     fmt.Sprintf("%+v", m.script.Header),
					Duration: 5 * time.Second,
				}
			})
		case "?":
			m.script.ToggleByteCode()
			m.Refresh()
		case "r":
			m.script.RemoveInstruction(m.highlightedLine)
			m.Refresh()
		case "d":
			m.script.DuplicateInstruction(m.highlightedLine)
			m.Refresh()
		case "m":
			m.script.MoveInstruction(m.highlightedLine, m.highlightedLine+1)
			m.highlightedLine++
			m.Refresh()
		case "M":
			m.script.MoveInstruction(m.highlightedLine, m.highlightedLine-1)
			m.highlightedLine--
			m.Refresh()
		case " ":
			if m.marker1 == -1 {
				m.marker1 = m.highlightedLine
			} else if m.marker2 == -1 {
				m.marker2 = m.highlightedLine
				if m.marker1 > m.marker2 {
					m.marker1, m.marker2 = m.marker2, m.marker1
				}
			} else {
				m.marker1 = -1
				m.marker2 = -1
			}
			m.Refresh()
		case "t":
			cmds = append(cmds, func() tea.Msg {
				return statusbar.AddStatusBarMessageMsg{
					Text:     fmt.Sprintf("%+v %d", m.script.Entry.Toc(), len(m.script.Entry.Data())),
					Duration: 5 * time.Second,
				}
			})
		}
	case statusbar.SubmitInputActionMsg:
		if msg.ID == "search" {
			m.searchText = msg.InputText
			//m.jumpToNextSearch(false)
		}
	case statusbar.OpcodeAndArgsInputResultMsg:
		o := m.script.GetOffset(m.highlightedLine)
		op := opcode.NewInstruction(o, msg.Opcode, msg.Args)
		if msg.ID == "insert" {
			m.script.InsertInstruction(m.highlightedLine, op)
		} else if msg.ID == "edit" {
			m.script.EditInstruction(m.highlightedLine, op)
		}
		m.Refresh()
	}

	return m, tea.Batch(cmds...)
}

func (m *ScriptView) scroll(offset int) {
	d := m.highlightedLine - m.codeOffset
	m.highlightedLine += offset
	if m.highlightedLine < 0 {
		m.highlightedLine = 0
	}
	if m.highlightedLine > len(m.script.Opcodes)-1 {
		m.highlightedLine = len(m.script.Opcodes) - 1
	}
	if m.highlightedLine < m.codeOffset || m.highlightedLine > m.codeOffset+m.vp.Height-1 {
		m.codeOffset = max(m.highlightedLine-d, 0)
	}
	m.Refresh()
}

func (m *ScriptView) jumpToNextSearch(reverse bool) {
	nextIdx := m.script.FindNextOpcode(m.searchText, m.highlightedLine, reverse)
	if nextIdx != -1 {
		m.scroll(nextIdx - m.highlightedLine)
	}
	m.Refresh()
}

func (m *ScriptView) Refresh() {
	if m.data == nil {
		m.vp.SetContent("No script selected")
		return
	}

	str := lipgloss.NewStyle().Width(m.vp.Width).Render(m.script.String(m.highlightedLine, m.codeOffset, m.vp.Height, m.marker1, m.marker2))
	m.vp.SetContent(str)
}

func (m ScriptView) View() string {
	subsList := m.subsList.View()
	// localsList := m.localsList.View()
	// globalsList := ""
	// rightPane := lipgloss.JoinVertical(lipgloss.Left, m.listStyle.Render(subsList), m.listStyle.Render(localsList), m.listStyle.Render(globalsList))
	bottomPane := lipgloss.JoinHorizontal(lipgloss.Top, m.vp.View(), m.listStyle.Render(subsList))
	str := lipgloss.JoinVertical(lipgloss.Center, m.script.Name, bottomPane)
	return str
}

func (m *ScriptView) SetActive(active bool) {
	m.active = active
}
