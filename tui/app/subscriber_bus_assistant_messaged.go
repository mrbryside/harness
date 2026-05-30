package app

import (
	"fmt"
	"os"

	"github.com/mrbryside/harness/eventbus"
)

func (m *Model) subscribeAssistantMessaged() {
	m.eventBus.Subscribe(eventbus.EventAssistantMessaged, func(e eventbus.Event) {
		data := e.Data.(eventbus.AssistantMessageEvent)
		fmt.Fprintf(os.Stderr, "[MODEL] received assistant_messaged: id=%q done=%v len=%d\n", data.ID, data.Done, len(data.Content))
		m.eventCh <- AssistantChunkMsg{Content: data.Content, Done: data.Done}
	})
}
