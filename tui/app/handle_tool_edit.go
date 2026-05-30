package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/tui/components"
)

// ToolEditMsg carries a tool edit event from the agent runtime.
type ToolEditMsg struct {
	Event eventbus.ToolEditFileEvent
}

// handleToolEditMsg processes a tool edit event.
func (m Model) handleToolEditMsg(msg ToolEditMsg) (tea.Model, tea.Cmd) {
	m.chat.AppendToolEdit(components.ToolEdit{
		Path:       msg.Event.Path,
		OldContent: msg.Event.OldContent,
		NewContent: msg.Event.NewContent,
		StartLine:  msg.Event.StartLine,
	})

	return m, nil
}
