package components

import tea "charm.land/bubbletea/v2"

func (i Input) Update(msg tea.Msg) (Input, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return i.handleKeyPress(msg)
	case tea.MouseWheelMsg:
		return i.handleMouseWheel(msg)
	case tea.WindowSizeMsg:
		return i.handleWindowSize(msg)
	}

	var cmd tea.Cmd
	i.textarea, cmd = i.textarea.Update(msg)
	return i, cmd
}
