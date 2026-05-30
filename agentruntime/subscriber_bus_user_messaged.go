package agentruntime

import (
	"context"
	"fmt"
	"os"

	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/llm"
)

func (r *AgentRuntime) subscribeUserMessaged() {
	r.eventBus.Subscribe(eventbus.EventUserMessaged, func(e eventbus.Event) {
		data := e.Data.(eventbus.UserMessageEvent)
		fmt.Fprintf(os.Stderr, "[AGENT] received user_messaged: id=%q content=%q\n", data.ID, data.Content)
		r.messages = append(r.messages, llm.Message{Role: "user", Content: data.Content})

		ctx, cancel := context.WithCancel(context.Background())
		r.mu.Lock()
		r.activeRequests[data.ID] = cancel
		r.mu.Unlock()

		go func() {
			r.streamResponse(ctx, data.ID)
			r.mu.Lock()
			delete(r.activeRequests, data.ID)
			r.mu.Unlock()
		}()
	})
}
