package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// handleKeyboard routes key events. Ctrl+C has debounce logic;
// PgUp/PgDown scroll chat; everything else goes to input.
func (m Model) handleKeyboard(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c":
		return m.handleCtrlC()
	case "esc":
		if m.streaming {
			return m.handleCtrlC()
		}
		// When not streaming, fall through to input (handles double-press clear).
	case "pgup", "pgdown":
		m.chat, cmd = m.chat.Update(msg)
		return m, cmd
	}

	m.input, cmd = m.input.Update(msg)
	m = m.reflowChat()
	return m, cmd
}

// handleCtrlC: streaming → interrupt; debounce → quit; otherwise → hint.
func (m Model) handleCtrlC() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.streaming {
		m.streaming = false
		m.streamCh = nil
		m.statusbar = m.statusbar.ClearMessage()
		m.statusbar, cmd = m.statusbar.SetMessage("✓ stream interrupted", 2*time.Second)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	if time.Since(m.ctrlCDebounce) < ctrlCDebounceWindow {
		return m, tea.Quit
	}

	m.ctrlCDebounce = time.Now()
	m.statusbar, cmd = m.statusbar.SetMessage("press Ctrl+C again to quit", ctrlCDebounceWindow)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}
