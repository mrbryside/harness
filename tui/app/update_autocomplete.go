package app

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// handleAutocompleteShow activates the autocomplete popup with the given prefix.
func (m Model) handleAutocompleteShow(msg AutocompleteShowMsg) (tea.Model, tea.Cmd) {
	m.autocomplete.Show(msg.Prefix)
	return m, nil
}

// handleAutocompleteHide deactivates the autocomplete popup.
func (m Model) handleAutocompleteHide() (tea.Model, tea.Cmd) {
	m.autocomplete.Hide()
	return m, nil
}

// handleAutocompleteSelect clears the input and executes the selected command.
// Results may contain both chat output and a toast notification.
func (m Model) handleAutocompleteSelect(msg AutocompleteSelectMsg) (tea.Model, tea.Cmd) {
	m.autocomplete.Hide()
	m.input.Reset()

	cmdName := strings.TrimPrefix(msg.Command, "/")
	cmd, ok := m.cmdRegistry.Get(cmdName)
	if !ok {
		cmd := m.chat.ShowToast("✗ Unknown command: "+msg.Command, 3*time.Second)
		return m, cmd
	}

	result := cmd.Execute("")
	var cmds []tea.Cmd

	if result.Chat != "" {
		m.chat.AppendMessage("system", result.Chat)
	}
	if result.Toast != "" {
		cmds = append(cmds, m.chat.ShowToast(result.Toast, 2*time.Second))
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// checkAutocomplete inspects the current input value and shows/hides the
// autocomplete popup accordingly.
func (m Model) checkAutocomplete() (tea.Model, tea.Cmd) {
	val := m.input.Value()
	if strings.HasPrefix(val, "/") {
		prefix := strings.TrimPrefix(val, "/")
		m.autocomplete.Show(prefix)
	} else {
		m.autocomplete.Hide()
	}
	return m, nil
}
