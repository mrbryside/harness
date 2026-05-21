package components_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/components"
)

// stripANSI removes ANSI escape sequences so substring checks aren't fooled
// by styling injected between characters (e.g. the textarea's virtual cursor
// styling the first character of the placeholder separately from the rest).
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, c := range s {
		if c == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
				inEsc = false
			}
			continue
		}
		b.WriteRune(c)
	}
	return b.String()
}

func TestInputInitialValueIsEmpty(t *testing.T) {
	m := components.NewInput("test-model")
	if m.Value() != "" {
		t.Errorf("expected initial Value() to be empty, got %q", m.Value())
	}
}

func TestInputResetClearsValue(t *testing.T) {
	m := components.NewInput("test-model")
	m.Reset()
	if m.Value() != "" {
		t.Errorf("expected Value() after Reset() to be empty, got %q", m.Value())
	}
}

func TestInputViewContainsPlaceholder(t *testing.T) {
	m := components.NewInput("test-model")
	// give it a realistic width so the placeholder isn't wrapped per-char
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if !strings.Contains(stripANSI(m.View()), "Type a message") {
		t.Errorf("expected View() to contain placeholder text, got:\n%s", m.View())
	}
}

func TestInputEnterTriggersSendAndResets(t *testing.T) {
	m := components.NewInput("test-model")
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if updated.Value() != "" {
		t.Errorf("expected Value() to be empty after Enter, got %q", updated.Value())
	}
	if cmd == nil {
		t.Errorf("expected a cmd to be returned on Enter (send signal), got nil")
	}
}

// TestInputTypedTextUsesThemeForeground asserts that the cursor line — the
// line the user is actively typing on — uses styles.AssistantText (#c0caf5,
// RGB 192;202;245) as its foreground, not the terminal's default colour.
//
// Why this matters: bubbles textarea renders the cursor line through
// computedCursorLine(), not computedText(). Setting only Styles.Focused.Text
// is not enough; the cursor line inherits from CursorLine, so CursorLine
// also needs a Foreground.
func TestInputTypedTextUsesThemeForeground(t *testing.T) {
	m := components.NewInput("test-model")
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Simulate typing "abc".
	for _, r := range "abc" {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	view := m.View()
	// AssistantText = #c0caf5 → SGR 38;2;192;202;245
	const fgSGR = "38;2;192;202;245"
	// The "abc" run must be immediately preceded by an SGR carrying the
	// theme foreground. We scan for the AssistantText FG segment somewhere
	// in the SGR string that opens the chunk containing "abc".
	abcIdx := strings.Index(view, "abc")
	if abcIdx < 0 {
		t.Fatalf("expected typed text \"abc\" in view, got:\n%q", view)
	}
	// Find the SGR opener directly before "abc" (the last ESC[…m before it).
	prefix := view[:abcIdx]
	lastEsc := strings.LastIndex(prefix, "\x1b[")
	if lastEsc < 0 {
		t.Fatalf("expected an SGR escape before \"abc\", got prefix:\n%q", prefix)
	}
	opener := prefix[lastEsc:]
	if !strings.Contains(opener, fgSGR) {
		t.Errorf("expected SGR opener before typed text to include theme foreground %q,\nopener was: %q\nfull view:\n%q", fgSGR, opener, view)
	}
}
