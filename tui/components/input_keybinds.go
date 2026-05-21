package components

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// handleKeyPress routes individual keys. Esc, Enter, Ctrl+J, Up/Down
// (history), and Thai combining marks are handled here; everything
// else falls through to the textarea.
func (i Input) handleKeyPress(msg tea.KeyPressMsg) (Input, tea.Cmd) {
	// Esc double-press clears input.
	if msg.Code == tea.KeyEscape {
		return i.handleEsc()
	}

	// Drop CapsLock key (produces spurious "A" text).
	if msg.Code == tea.KeyCapsLock {
		return i, nil
	}

	// Drop Thai combining (eg. ( ่),( ้)) marks at start of buffer.
	if i.textarea.Value() == "" && isCombiningMark(msg.Text) {
		return i, nil
	}

	// Ctrl+J → newline (Shift+Enter proxy).
	if msg.Code == 'j' && msg.Mod == tea.ModCtrl {
		msg = tea.KeyPressMsg{Code: tea.KeyEnter, Text: "\n"}
		var cmd tea.Cmd
		i.textarea, cmd = i.textarea.Update(msg)
		return i, cmd
	}

	// Up/Down: navigate within text first; only browse history at edges.
	if msg.Code == tea.KeyUp {
		if i.textarea.Line() > 0 {
			var cmd tea.Cmd
			i.textarea, cmd = i.textarea.Update(msg)
			return i, cmd
		}
		return i.handleHistoryUp()
	}
	if msg.Code == tea.KeyDown {
		if i.textarea.Line() < i.textarea.LineCount()-1 {
			var cmd tea.Cmd
			i.textarea, cmd = i.textarea.Update(msg)
			return i, cmd
		}
		return i.handleHistoryDown()
	}

	// Enter sends message (unless Shift/Alt modifier).
	if msg.Code == tea.KeyEnter {
		return i.handleEnter(msg)
	}

	// Any other key while browsing history → reset cursor.
	if i.history != nil && i.history.Cursor() >= 0 {
		i.history.ResetCursor()
		i.draft = i.textarea.Value()
	}

	var cmd tea.Cmd
	i.textarea, cmd = i.textarea.Update(msg)
	return i, cmd
}

func (i Input) handleEsc() (Input, tea.Cmd) {
	if i.textarea.Value() == "" {
		return i, nil
	}
	if time.Since(i.lastEscAt) < escDebounce {
		i.textarea.Reset()
		i.lastEscAt = time.Time{}
		return i, func() tea.Msg {
			return StatusMsg{Content: "✓ input cleared", Duration: 2 * time.Second}
		}
	}
	i.lastEscAt = time.Now()
	return i, func() tea.Msg {
		return StatusMsg{Content: "press Esc again to clear input", Duration: escDebounce}
	}
}

func (i Input) handleEnter(msg tea.KeyPressMsg) (Input, tea.Cmd) {
	if msg.Mod.Contains(tea.ModShift) || msg.Mod.Contains(tea.ModAlt) {
		var cmd tea.Cmd
		i.textarea, cmd = i.textarea.Update(msg)
		return i, cmd
	}
	content := i.textarea.Value()
	i.textarea.Reset()
	return i, func() tea.Msg {
		return SendMsg{Content: content}
	}
}

func (i Input) handleHistoryUp() (Input, tea.Cmd) {
	if i.history == nil {
		return i, nil
	}
	if i.history.Cursor() < 0 {
		i.draft = i.textarea.Value()
	}
	text, ok := i.history.Previous()
	if ok {
		i.textarea.SetValue(text)
		i.textarea.SetCursorColumn(len(text))
	}
	return i, nil
}

func (i Input) handleHistoryDown() (Input, tea.Cmd) {
	if i.history == nil {
		return i, nil
	}
	text, ok := i.history.Next()
	if ok {
		i.textarea.SetValue(text)
		i.textarea.SetCursorColumn(len(text))
	} else {
		i.textarea.SetValue(i.draft)
		i.textarea.SetCursorColumn(len(i.draft))
		i.history.ResetCursor()
	}
	return i, nil
}
