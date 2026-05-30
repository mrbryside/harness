package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/components"
)

// handlePermissionAnswer hides the prompt, emits to event bus, and reflows chat.
func (m Model) handlePermissionAnswer(msg components.PermissionAnswerMsg) (tea.Model, tea.Cmd) {
	m.permissionPrompt.Hide()
	m = m.reflowChat()

	m.eventBus.Emit(EventQuestionAnswered, struct {
		QuestionID string
		Answer     bool
	}{
		QuestionID: msg.QuestionID,
		Answer:     msg.Answer,
	})

	return m, nil
}

// handleShowPermissionPrompt shows the startup permission question and reflows chat.
func (m Model) handleShowPermissionPrompt() (tea.Model, tea.Cmd) {
	m.permissionPrompt.Show("Welcome! Do you want to enable the demo mode?", "startup-demo")
	m = m.reflowChat()
	return m, nil
}
