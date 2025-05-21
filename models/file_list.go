package models

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrchip53/gta-tools/rage"
	"github.com/mrchip53/gta-tools/rage/img"
)

type FileSelectedMsg struct{ item listItem }

func (m FileSelectedMsg) Item() listItem { return m.item }

// TODO move to rage package
type listItem struct {
	name     string
	data     []byte
	tocEntry img.TocEntry
	fileType rage.FileType
}

func (i listItem) FilterValue() string { return i.name }
func (i listItem) Name() string        { return i.name }
func (i listItem) Data() []byte {
	d := make([]byte, len(i.data))
	copy(d, i.data)
	return d
}
func (i listItem) FileType() rage.FileType { return i.fileType }
func (i listItem) TocEntry() img.TocEntry  { return i.tocEntry }

func newListItem(f img.ImgEntry) list.Item {
	t := rage.GetFileType(f.Name())
	return listItem{name: f.Name(), fileType: t, data: f.Data(), tocEntry: f.Toc()}
}

type customDelegate struct{}

func (d customDelegate) Height() int                               { return 1 }
func (d customDelegate) Spacing() int                              { return 0 }
func (d customDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d customDelegate) Render(w io.Writer, m list.Model, index int, li list.Item) {
	i, ok := li.(listItem)
	if !ok {
		pf := "  "
		if m.Index() == index {
			pf = "> "
		}
		fmt.Fprint(w, itemStyle.Render(pf+li.FilterValue()))
		return
	}

	if index == m.Index() {
		color := COLOR_ACCENT
		if i.fileType == rage.FileTypeScript {
			color = COLOR_SCRIPT
		}
		fmt.Fprint(w, focusedItemStyle.Foreground(color).Render("> "+i.name))
	} else {
		fmt.Fprint(w, itemStyle.Render("  "+i.name))
	}
}

var (
	focusedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)
	itemStyle        = lipgloss.NewStyle().Foreground(Grey)
)

type FileList struct {
	list   list.Model
	active bool
}

func NewFileList(img img.ImgFile) FileList {
	var items []list.Item

	for _, v := range img.Entries() {
		items = append(items, newListItem(v))
	}

	d := customDelegate{}

	l := list.New(items, d, 0, 0)
	l.Title = "Files"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	return FileList{
		list: l,
		// TODO shouldnt do this here
		active: true,
	}
}

func (m FileList) Init() tea.Cmd {
	return nil
}

func (m FileList) Update(msg tea.Msg) (FileList, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if !m.active {
		return m, nil
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			item, ok := m.list.SelectedItem().(listItem)
			if ok {
				cmds = append(cmds, func() tea.Msg {
					return FileSelectedMsg{item}
				})
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func (m FileList) View() string {
	return m.list.View()
}

func (m *FileList) SetSize(w, h int) {
	m.list.SetSize(w, h)
}

func (m *FileList) SetActive(active bool) {
	m.active = active
}
