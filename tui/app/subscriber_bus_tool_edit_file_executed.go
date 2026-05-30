package app

import (
	"fmt"
	"os"

	"github.com/mrbryside/harness/eventbus"
)

func (m *Model) subscribeToolEditFileExecuted() {
	m.eventBus.Subscribe(eventbus.EventToolEditFileExecuted, func(e eventbus.Event) {
		data := e.Data.(eventbus.ToolEditFileEvent)
		fmt.Fprintf(os.Stderr, "[MODEL] received tool_edit_file_executed: path=%q\n", data.Path)
		m.eventCh <- ToolEditMsg{Event: data}
	})
}
