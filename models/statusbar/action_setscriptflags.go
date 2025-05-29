package statusbar

import (
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SetScriptFlagsAction handles the input for script flags.
// It implements the Action interface.
type SetScriptFlagsAction struct {
	id        string
	textInput textinput.Model
	done      bool
	resultMsg tea.Msg
}

// NewSetScriptFlagsAction creates a new action for setting script flags.
func NewSetScriptFlagsAction(id string, prompt string) *SetScriptFlagsAction {
	ti := textinput.New()
	ti.Prompt = "❯ "
	if prompt != "" {
		if !strings.HasSuffix(prompt, " ") {
			ti.Prompt = prompt + " "
		} else {
			ti.Prompt = prompt
		}
	}
	ti.CharLimit = 256
	ti.Width = 40

	return &SetScriptFlagsAction{
		id:        id,
		textInput: ti,
	}
}

func (a *SetScriptFlagsAction) Init() tea.Cmd {
	return a.textInput.Focus()
}

func (a *SetScriptFlagsAction) Update(msg tea.Msg) (Action, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			flagsStr := strings.TrimSpace(a.textInput.Value())
			var flags int64
			var err error
			if strings.HasPrefix(strings.ToLower(flagsStr), "0x") {
				flags, err = strconv.ParseInt(flagsStr[2:], 16, 32)
			} else {
				flags, err = strconv.ParseInt(flagsStr, 10, 32)
			}

			if err != nil {
				a.done = false
				return a, func() tea.Msg {
					return AddStatusBarMessageMsg{Text: "Invalid integer format for flags.", Duration: 3 * time.Second}
				}
			}
			a.resultMsg = SubmitScriptFlagsMsg{Flags: int32(flags)}
			a.done = true
			return a, nil

		case tea.KeyEsc:
			a.resultMsg = ActionCancelledMsg{ActionID: a.id}
			a.done = true
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.textInput, cmd = a.textInput.Update(msg)
	return a, cmd
}

func (a *SetScriptFlagsAction) View() string {
	return a.textInput.View()
}

func (a *SetScriptFlagsAction) Description() string {
	return "Enter an integer value for script flags. Press Enter to submit, Esc to cancel."
}

func (a *SetScriptFlagsAction) ID() string {
	return a.id
}

func (a *SetScriptFlagsAction) IsDone() bool {
	return a.done
}

func (a *SetScriptFlagsAction) Result() tea.Msg {
	return a.resultMsg
}
