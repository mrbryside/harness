package app

import (
	"fmt"
	"os"

	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/tui/components"
)

func (m *Model) subscribeQuestionAsked() {
	m.eventBus.Subscribe(eventbus.EventQuestionAsked, func(e eventbus.Event) {
		data := e.Data.(struct {
			QuestionID string
			Question   string
		})
		fmt.Fprintf(os.Stderr, "[MODEL] received question_asked: id=%q question=%q\n", data.QuestionID, data.Question)
		m.activeQuestion.question = components.CreateQuestion(components.QuestionTypePermission, data.QuestionID, data.Question)
		if m.activeQuestion.question != nil {
			m.activeQuestion.question.Show(data.Question)
		}
		m.eventCh <- questionShownMsg{}
		fmt.Fprintf(os.Stderr, "[MODEL] question shown, active=%v\n", m.activeQuestion.question != nil && m.activeQuestion.question.Active())
	})
}
