package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/components"
)

// Update is the main message handler. It dispatches to focused
// handler methods so each file in app/ owns a single concern.
// listenEvents() is always re-scheduled to keep the EventBus bridge alive.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var newModel tea.Model
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		newModel, cmd = m.handleResize(msg)
	case tea.KeyPressMsg:
		newModel, cmd = m.handleKeyboard(msg)
	case tea.MouseWheelMsg:
		newModel, cmd = m.handleMouseWheel(msg)
	case tea.MouseClickMsg:
		newModel, cmd = m.handleMouseClick(msg)
	case tea.MouseMotionMsg:
		newModel, cmd = m.handleMouseMotion(msg)
	case tea.MouseReleaseMsg:
		newModel, cmd = m.handleMouseRelease(msg)
	case scrollTickMsg:
		newModel, cmd = m.handleScrollTick()
	case tea.PasteMsg:
		newModel, cmd = m.handlePaste(msg)
	case components.SendMsg:
		newModel, cmd = m.handleSendMsg(msg)
	case AssistantChunkMsg:
		newModel, cmd = m.handleAssistantChunkMsg(msg)
	case ToolEditMsg:
		newModel, cmd = m.handleToolEditMsg(msg)
	case components.StatusMsg:
		newModel, cmd = m.handleStatusMsg(msg)
	case AutocompleteShowMsg:
		newModel, cmd = m.handleAutocompleteShow(msg)
	case AutocompleteHideMsg:
		newModel, cmd = m.handleAutocompleteHide()
	case AutocompleteSelectMsg:
		newModel, cmd = m.handleAutocompleteSelect(msg)
	case components.QuestionAnswerMsg:
		newModel, cmd = m.handleQuestionAnswer(msg)
	case questionShownMsg:
		newModel, cmd = m.handleQuestionShown(msg)
	default:
		newModel, cmd = m, nil
	}

	m = newModel.(Model)
	return m, tea.Batch(cmd, m.listenEvents())
}
