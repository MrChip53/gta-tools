package main

import (
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrchip53/gta-tools/models"
	"github.com/mrchip53/gta-tools/rage"
	"github.com/mrchip53/gta-tools/rage/img"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	sidebarStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")) // Purple
	sidebarActiveStyle = sidebarStyle.BorderForeground(lipgloss.Color("228")) // Yellow
	focusedItemStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)

	// Main content styles
	mainContentStyle       = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63"))
	mainContentActiveStyle = mainContentStyle.BorderForeground(lipgloss.Color("228"))

	// Status bar styles
	statusBarInfoStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
				Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})
	statusBarHelpStyle = lipgloss.NewStyle().Inherit(statusBarInfoStyle).Foreground(lipgloss.Color("241"))
)

type sidebarModel int

const (
	sidebarModelNone sidebarModel = iota
	sidebarModelImgFile
)

type mainContentModel int

const (
	mainContentModelNone mainContentModel = iota
)

type window int

const (
	sidebar window = iota
	mainContent
)

type model struct {
	winWidth     int
	winHeight    int
	sideWidth    int
	sideHeight   int
	mainWidth    int
	mainHeight   int
	statusWidth  int
	statusHeight int

	focusedWindow window
	ready         bool

	imgFileList      models.FileList
	mainContentModel models.ScriptView

	imgFile img.ImgFile

	statusBar models.StatusBar
}

func initialModel() model {
	img := img.LoadImgFile(imgBytes)
	sb := models.NewStatusBar()
	return model{
		imgFile:          img,
		imgFileList:      models.NewFileList(img),
		mainContentModel: models.NewScriptView(nil, 0, 0),
		statusBar:        sb,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.winWidth = msg.Width
		m.winHeight = msg.Height

		m.statusHeight = 1

		availableHeight := m.winHeight - m.statusHeight - docStyle.GetVerticalMargins() // Adjusted for docStyle vertical margins

		m.sideWidth = m.winWidth / 4
		m.sideWidth = min(max(m.sideWidth, 20), 40)

		m.mainWidth = m.winWidth - m.sideWidth

		m.sideHeight = availableHeight
		m.mainHeight = availableHeight - sidebarStyle.GetVerticalFrameSize()
		m.statusWidth = m.winWidth - docStyle.GetHorizontalMargins()/2

		m.mainContentModel = models.NewScriptView(nil, m.mainWidth, m.mainHeight)
		m.imgFileList.SetSize(m.sideWidth, m.sideHeight-sidebarStyle.GetVerticalFrameSize())
		m.ready = true
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "tab":
			if m.focusedWindow == sidebar {
				m.focusedWindow = mainContent
			} else {
				m.focusedWindow = sidebar
			}
			m.imgFileList.SetActive(m.focusedWindow == sidebar)
			m.mainContentModel.SetActive(m.focusedWindow == mainContent)
		case "s":
			b := m.imgFile.Bytes()
			imgPathFolder := filepath.Dir(imgPath)
			filePath := filepath.Join(imgPathFolder, "out.img")
			err := os.WriteFile(filePath, b, 0644)
			if err != nil {
				panic(err)
			}
			cmds = append(cmds, tea.Cmd(func() tea.Msg {
				return models.AddStatusBarMessageMsg{
					Text:     "Saved img file to disk",
					Duration: 3 * time.Second,
				}
			}))
			// case "n":
			// 	b, err := os.ReadFile("aaatest.sco")
			// 	if err != nil {
			// 		panic(err)
			// 	}
			// 	m.imgFile.AddEntry("aaatest.sco", b)
			// 	m.imgFileList = models.NewFileList(m.imgFile)
			// 	m.imgFileList.SetSize(m.sideWidth, m.sideHeight-sidebarStyle.GetVerticalFrameSize())
		}
	case models.FileSelectedMsg:
		if msg.Item().FileType() == rage.FileTypeScript {
			m.mainContentModel = models.NewScriptView(msg.Item().Entry(), m.mainWidth, m.mainHeight)
			m.focusedWindow = mainContent
			m.mainContentModel.SetActive(true)
			m.imgFileList.SetActive(false)
		}
	}

	m.statusBar, cmd = m.statusBar.Update(msg)
	cmds = append(cmds, cmd)

	if !m.statusBar.HasAction() {
		if m.focusedWindow == sidebar {
			m.imgFileList, cmd = m.imgFileList.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			m.mainContentModel, cmd = m.mainContentModel.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.ready == false || m.winWidth == 0 || m.winHeight == 0 {
		return "Initializing..."
	}

	// Build main content view
	mainContentViewStr := m.mainContentModel.View()
	mStyle := mainContentStyle.Width(m.mainWidth).Height(m.mainHeight)
	if m.focusedWindow == mainContent {
		mStyle = mainContentActiveStyle.Width(m.mainWidth).Height(m.mainHeight)
	}
	mainContentView := mStyle.Render(mainContentViewStr)

	// Build side bar view
	sidebarContentStr := m.imgFileList.View()
	sbStyle := sidebarStyle.Width(m.sideWidth - sidebarStyle.GetHorizontalFrameSize()).Height(m.sideHeight - sidebarStyle.GetVerticalFrameSize())
	if m.focusedWindow == sidebar {
		sbStyle = sidebarActiveStyle.Width(m.sideWidth - sidebarActiveStyle.GetHorizontalFrameSize()).Height(m.sideHeight - sidebarActiveStyle.GetVerticalFrameSize())
	}
	sidebarContent := sbStyle.Render(sidebarContentStr)
	// Build status bar view
	statusBarText := m.statusBar.View()
	statusBarView := statusBarInfoStyle.Width(m.statusWidth).Render(statusBarText)
	if statusBarText == "" {
		statusBarView = statusBarHelpStyle.Width(m.statusWidth).Render("q: quit | tab: switch focus | s: save")
	}

	// Combine views
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebarContent, mainContentView)
	finalView := lipgloss.JoinVertical(lipgloss.Left, mainView, statusBarView) // Combine main view and status bar

	return docStyle.Render(finalView)
}
