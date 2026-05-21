package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

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
func (m Model) handleMouseMotion(msg tea.MouseMotionMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}
	if m.chat.HasSelection() {
		line, col := m.chatContentCoordClamped(msg.X, msg.Y)
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

// handleMouseRelease finalises selection and copies text.
func (m Model) handleMouseRelease(msg tea.MouseReleaseMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}
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
