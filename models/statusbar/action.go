package statusbar

import tea "github.com/charmbracelet/bubbletea"

// Action defines the interface for interactive operations
// managed by the status bar. Each Action is responsible for its
// own state, input handling, view, and producing a result.
type Action interface {
	// Init is called when the action becomes active, allowing it to
	// perform initial setup, like focusing a text input.
	Init() tea.Cmd

	// Update handles incoming Bubble Tea messages (like key presses).
	// It returns the potentially updated Action (e.g., if it's a multi-step
	// action transitioning to a new state) and any command to execute.
	Update(msg tea.Msg) (Action, tea.Cmd)

	// View renders the current UI for the action (e.g., a text input field).
	View() string

	// Description provides a textual hint or context for the action,
	// displayed alongside its view.
	Description() string

	// ID returns a unique identifier for the action instance or type.
	// Useful for debugging or if external components need to identify the action.
	ID() string

	// IsDone indicates whether the action has completed its lifecycle
	// (e.g., input submitted or cancelled).
	IsDone() bool

	// Result returns a tea.Msg representing the outcome of the action
	// (e.g., SubmitScriptFlagsMsg, OpcodeAndArgsInputResultMsg, or a cancellation message).
	// This should only be called if IsDone() returns true.
	Result() tea.Msg
}

// ActionCancelledMsg is a generic message to indicate an action was cancelled.
type ActionCancelledMsg struct {
	ActionID string
}
