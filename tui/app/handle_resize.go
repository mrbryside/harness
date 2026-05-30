package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/components"
)

// handleResize propagates window size to all children.
func (m Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	chatWidth := msg.Width - outerMarginX - innerGap - components.SidebarWidth
	if chatWidth < 1 {
		chatWidth = 1
	}

	m.input, _ = m.input.Update(tea.WindowSizeMsg{Width: chatWidth, Height: msg.Height})
	statusBarWidth := outerMarginX + chatWidth + innerGap
	m.statusbar, _ = m.statusbar.Update(tea.WindowSizeMsg{Width: statusBarWidth, Height: msg.Height})
	m.sidebar, _ = m.sidebar.Update(msg)

	m = m.reflowChat()
	return m, nil
}

// handlePaste forwards bracketed-paste data to the input.
func (m Model) handlePaste(msg tea.PasteMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m = m.reflowChat()
	return m, cmd
}
