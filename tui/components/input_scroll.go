package components

import tea "charm.land/bubbletea/v2"

// handleMouseWheel translates wheel events into synthetic PageUp/PageDown
// keypresses so the textarea viewport scrolls internally.
func (i Input) handleMouseWheel(msg tea.MouseWheelMsg) (Input, tea.Cmd) {
	var synthetic tea.KeyPressMsg
	switch msg.Button {
	case tea.MouseWheelUp:
		synthetic = tea.KeyPressMsg{Code: tea.KeyPgUp}
	case tea.MouseWheelDown:
		synthetic = tea.KeyPressMsg{Code: tea.KeyPgDown}
	default:
		return i, nil
	}
	var cmd tea.Cmd
	i.textarea, cmd = i.textarea.Update(synthetic)
	return i, cmd
}

// handleWindowSize updates the textarea width to match the layout.
func (i Input) handleWindowSize(msg tea.WindowSizeMsg) (Input, tea.Cmd) {
	w := msg.Width - 3 // 1 accent col + 2 padding cols
	if w < 1 {
		w = 1
	}
	i.width = msg.Width
	i.textarea.SetWidth(w)
	return i, nil
}
