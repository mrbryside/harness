package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/components"
)

// Update is the main message handler. It dispatches to focused
// handler methods so each file in app/ owns a single concern.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleResize(msg)
	case tea.KeyPressMsg:
		return m.handleKeyboard(msg)
	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg)
	case tea.MouseClickMsg:
		return m.handleMouseClick(msg)
	case tea.MouseMotionMsg:
		return m.handleMouseMotion(msg)
	case tea.MouseReleaseMsg:
		return m.handleMouseRelease(msg)
	case tea.PasteMsg:
		return m.handlePaste(msg)
	case components.SendMsg:
		return m.handleSendMsg(msg)
	case chunkMsg:
		return m.handleChunkMsg(msg)
	case components.StatusMsg:
		return m.handleStatusMsg(msg)
	}
	return m, nil
}
