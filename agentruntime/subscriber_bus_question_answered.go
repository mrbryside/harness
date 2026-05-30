package agentruntime

import (
	"fmt"
	"os"

	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/tui/components"
)

func (r *AgentRuntime) subscribeQuestionAnswered() {
	r.eventBus.Subscribe(eventbus.EventQuestionAnswered, func(e eventbus.Event) {
		data := e.Data.(struct {
			QuestionID string
			Answer     components.QuestionChoice
		})
		fmt.Fprintf(os.Stderr, "[AGENT] question %q answered: index=%d label=%q\n",
			data.QuestionID, data.Answer.Index, data.Answer.Label)
	})
}
