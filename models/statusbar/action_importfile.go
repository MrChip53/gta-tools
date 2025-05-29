package statusbar

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ImportFileAction handles the two-step input for importing a file.
// It implements the Action interface.
type ImportFileAction struct {
	id        string
	textInput textinput.Model
	done      bool
	resultMsg tea.Msg

	currentStep     int    // 1 for host path, 2 for archive path
	hostOSPath      string // Stores the host OS path after step 1
	descriptionText string
}

// NewImportFileAction creates a new action for importing a file.
func NewImportFileAction(id string) *ImportFileAction {
	ti := textinput.New()
	ti.Prompt = "Enter host OS file path: "
	ti.CharLimit = 1024
	ti.Width = 60

	return &ImportFileAction{
		id:              id,
		textInput:       ti,
		currentStep:     1,
		descriptionText: "Enter the full path to the file on your computer.",
	}
}

func (a *ImportFileAction) Init() tea.Cmd {
	return a.textInput.Focus()
}

func (a *ImportFileAction) Update(msg tea.Msg) (Action, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			inputText := strings.TrimSpace(a.textInput.Value())
			if a.currentStep == 1 { // Processing host OS path
				if inputText == "" {
					cmds = append(cmds, func() tea.Msg {
						return AddStatusBarMessageMsg{Text: "Host OS file path cannot be empty", Duration: 3 * time.Second}
					})
					return a, tea.Batch(cmds...)
				}
				a.hostOSPath = inputText
				a.currentStep = 2
				a.textInput.SetValue("")
				a.textInput.Prompt = "Enter desired archive filename: "
				a.descriptionText = "Enter the name for the file as it will appear in the archive."
				return a, a.textInput.Focus()
			} else if a.currentStep == 2 { // Processing archive filename
				if inputText == "" {
					cmds = append(cmds, func() tea.Msg {
						return AddStatusBarMessageMsg{Text: "Archive filename cannot be empty", Duration: 3 * time.Second}
					})
					return a, tea.Batch(cmds...)
				}
				a.resultMsg = ImportFileActionMsg{
					ID:          a.id,
					HostPath:    a.hostOSPath,
					ArchivePath: inputText,
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

func (a *ImportFileAction) View() string {
	return a.textInput.View()
}

func (a *ImportFileAction) Description() string {
	return a.descriptionText
}

func (a *ImportFileAction) ID() string {
	return a.id
}

func (a *ImportFileAction) IsDone() bool {
	return a.done
}

func (a *ImportFileAction) Result() tea.Msg {
	return a.resultMsg
}
