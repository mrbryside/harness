package app

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/tui/components"
)

// questionShownMsg is emitted when a question is shown via EventBus,
// triggering a reflow of the chat layout.
type questionShownMsg struct{}

// handleQuestionShown reflows the chat after a question is displayed.
func (m Model) handleQuestionShown(msg questionShownMsg) (tea.Model, tea.Cmd) {
	m = m.reflowChat()
	return m, nil
}

// handleQuestionAnswer hides the question, emits to event bus, and reflows chat.
func (m Model) handleQuestionAnswer(msg components.QuestionAnswerMsg) (tea.Model, tea.Cmd) {
	fmt.Fprintf(os.Stderr, "[PERM] answer received: id=%q answer=%v\n", msg.QuestionID, msg.Data)
	if m.activeQuestion != nil && m.activeQuestion.question != nil {
		m.activeQuestion.question.Hide()
		m.activeQuestion.question = nil
	}
	m = m.reflowChat()

	m.eventBus.Emit(eventbus.EventQuestionAnswered, struct {
		QuestionID string
		Answer     components.QuestionChoice
	}{
		QuestionID: msg.QuestionID,
		Answer:     msg.Data,
	})

	return m, nil
}
