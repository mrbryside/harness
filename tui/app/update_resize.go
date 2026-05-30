package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

// reflowChat recomputes chat viewport size after input changes.
func (m Model) reflowChat() Model {
	if m.width == 0 || m.height == 0 {
		return m
	}
	chatWidth := m.width - outerMarginX - innerGap - components.SidebarWidth
	if chatWidth < 1 {
		chatWidth = 1
	}
	statusLines := lipgloss.Height(m.statusbar.View())

	var inputLines int
	if m.permissionPrompt.Active() {
		inputLines = lipgloss.Height(m.permissionPrompt.OverlayView(chatWidth))
	} else {
		inputLines = lipgloss.Height(m.input.View())
	}

	chatHeight := m.height - inputLines - statusLines - outerMarginY - chatInputGap
	if chatHeight < 1 {
		chatHeight = 1
	}
	m.chat, _ = m.chat.Update(tea.WindowSizeMsg{Width: chatWidth, Height: chatHeight})
	return m
}

// handlePaste forwards bracketed-paste data to the input.
func (m Model) handlePaste(msg tea.PasteMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m = m.reflowChat()
	return m, cmd
}
