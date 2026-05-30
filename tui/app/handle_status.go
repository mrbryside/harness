package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/components"
)

// handleStatusMsg pushes a transient message onto the status bar.
func (m Model) handleStatusMsg(msg components.StatusMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.statusbar, cmd = m.statusbar.SetMessage(msg.Content, msg.Duration)
	return m, cmd
}
