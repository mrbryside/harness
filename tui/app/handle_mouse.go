package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// scrollTickMsg is sent every tick while the mouse is held past the
// edge of the chat viewport so selection auto-scrolls smoothly.
type scrollTickMsg struct{}

const scrollTickInterval = 60 * time.Millisecond // ~16 fps, smooth but not too fast
const scrollTickLines  = 3                       // lines per tick

// scrollTickCmd returns a command that fires after scrollTickInterval.
func scrollTickCmd() tea.Cmd {
	return tea.Tick(scrollTickInterval, func(time.Time) tea.Msg {
		return scrollTickMsg{}
	})
}

// handleMouseWheel routes scroll by cursor position.
func (m Model) handleMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if _, _, ok := m.inputContentCoord(msg.X, msg.Y); ok {
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	m.chat, cmd = m.chat.Update(msg)
	return m, cmd
}

// handleMouseClick starts a selection on the component under the cursor.
func (m Model) handleMouseClick(msg tea.MouseClickMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}
	if line, col, ok := m.chatContentCoord(msg.X, msg.Y); ok {
		m.chat.SelectStart(line, col)
		return m, nil
	}
	if line, col, ok := m.inputContentCoord(msg.X, msg.Y); ok {
		m.input.SelectStart(line, col)
		return m, nil
	}
	return m, nil
}

// handleMouseMotion extends an in-flight selection (sticky to component).
// When the mouse is dragged past the top or bottom edge of the chat viewport,
// auto-scroll kicks in and scrolls smoothly until the mouse comes back inside.
func (m Model) handleMouseMotion(msg tea.MouseMotionMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}
	if m.chat.HasSelection() {
		_, y0, _, h := m.chatRect()
		// Above the chat area → start scrolling up
		if msg.Y < y0 {
			if m.chatAutoScrollDir != -1 {
				m.chatAutoScrollDir = -1
				_, m.chatAutoScrollCol = m.chatContentCoordClamped(msg.X, msg.Y)
				return m, scrollTickCmd()
			}
			return m, nil
		}
		// Below the chat area → start scrolling down
		if msg.Y >= y0+h {
			if m.chatAutoScrollDir != 1 {
				m.chatAutoScrollDir = 1
				_, m.chatAutoScrollCol = m.chatContentCoordClamped(msg.X, msg.Y)
				return m, scrollTickCmd()
			}
			return m, nil
		}
		// Inside the chat area → stop auto-scroll
		m.chatAutoScrollDir = 0
		line, col, ok := m.chatContentCoord(msg.X, msg.Y)
		if !ok {
			line, col = m.chatContentCoordClamped(msg.X, msg.Y)
		}
		m.chat.SelectExtend(line, col)
		return m, nil
	}
	if m.input.HasSelection() {
		line, col := m.inputContentCoordClamped(msg.X, msg.Y)
		m.input.SelectExtend(line, col)
		return m, nil
	}
	return m, nil
}

// handleScrollTick performs one frame of auto-scroll while the mouse is held
// past the viewport edge.
func (m Model) handleScrollTick() (tea.Model, tea.Cmd) {
	if m.chatAutoScrollDir == 0 {
		return m, nil
	}
	switch m.chatAutoScrollDir {
	case -1:
		m.chat.ScrollUpAndExtend(scrollTickLines, m.chatAutoScrollCol)
	case 1:
		m.chat.ScrollDownAndExtend(scrollTickLines, m.chatAutoScrollCol)
	}
	return m, scrollTickCmd()
}

// handleMouseRelease finalises selection and copies text.
func (m Model) handleMouseRelease(msg tea.MouseReleaseMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}
	m.chatAutoScrollDir = 0 // stop any auto-scroll
	switch {
	case m.chat.HasSelection():
		text := m.chat.SelectedText()
		m.chat.SelectClear()
		if text == "" {
			return m, nil
		}
		cmd := m.chat.ShowToast("✓ copied", 2*time.Second)
		return m, tea.Batch(tea.SetClipboard(text), cmd)
	case m.input.HasSelection():
		text := m.input.SelectedText()
		m.input.SelectClear()
		if text == "" {
			return m, nil
		}
		cmd := m.chat.ShowToast("✓ copied", 2*time.Second)
		return m, tea.Batch(tea.SetClipboard(text), cmd)
	}
	return m, nil
}