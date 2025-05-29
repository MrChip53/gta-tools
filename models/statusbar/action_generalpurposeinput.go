package statusbar

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// GeneralPurposeInputAction provides a generic text input action.
// It implements the Action interface.
type GeneralPurposeInputAction struct {
	id        string
	textInput textinput.Model
	done      bool
	resultMsg tea.Msg
	desc      string
}

// NewGeneralPurposeInputAction creates a new generic input action.
func NewGeneralPurposeInputAction(id, userPrompt, description string) *GeneralPurposeInputAction {
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

	return &GeneralPurposeInputAction{
		id:        id,
		textInput: ti,
		desc:      description,
	}
}

func (a *GeneralPurposeInputAction) Init() tea.Cmd {
	return a.textInput.Focus()
}

func (a *GeneralPurposeInputAction) Update(msg tea.Msg) (Action, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			inputValue := a.textInput.Value() // No trim, raw value
			a.resultMsg = SubmitInputActionMsg{ID: a.id, InputText: inputValue}
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

func (a *GeneralPurposeInputAction) View() string {
	return a.textInput.View()
}

func (a *GeneralPurposeInputAction) Description() string {
	return a.desc
}

func (a *GeneralPurposeInputAction) ID() string {
	return a.id
}

func (a *GeneralPurposeInputAction) IsDone() bool {
	return a.done
}

func (a *GeneralPurposeInputAction) Result() tea.Msg {
	return a.resultMsg
}
