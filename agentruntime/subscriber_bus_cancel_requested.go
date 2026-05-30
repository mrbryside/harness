package agentruntime

import (
	"fmt"
	"os"

	"github.com/mrbryside/harness/eventbus"
)

func (r *AgentRuntime) subscribeCancelRequested() {
	r.eventBus.Subscribe(eventbus.EventCancelRequested, func(e eventbus.Event) {
		data := e.Data.(eventbus.CancelRequestEvent)
		fmt.Fprintf(os.Stderr, "[AGENT] received cancel_requested: id=%q\n", data.RequestID)
		r.mu.Lock()
		if cancel, ok := r.activeRequests[data.RequestID]; ok {
			cancel()
			delete(r.activeRequests, data.RequestID)
		}
		r.mu.Unlock()
	})
}
