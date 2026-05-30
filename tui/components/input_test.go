package components_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/components"
)

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
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if !strings.Contains(stripANSI(m.View()), "Type a message") {
		t.Errorf("expected View() to contain placeholder text, got:\n%s", m.View())
	}
}

// TestInputTypedTextUsesThemeForeground asserts that the cursor line uses
// styles.AssistantText (#c0caf5) as its foreground.
func TestInputTypedTextUsesThemeForeground(t *testing.T) {
	m := components.NewInput("test-model")
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	for _, r := range "abc" {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	view := m.View()
	const fgSGR = "38;2;248;248;242"
	abcIdx := strings.Index(view, "abc")
	if abcIdx < 0 {
		t.Fatalf("expected typed text \"abc\" in view, got:\n%q", view)
	}
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

// stripANSI removes ANSI escape sequences.
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

// Up arrow recalls the previous message from history.
func TestInputUpRecallsHistory(t *testing.T) {
	in := components.NewInput("test-model")
	in, _ = in.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Simulate sending two messages.
	in.AddHistory("first message")
	in.AddHistory("second message")

	// Press Up → should show "second message".
	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if in.Value() != "second message" {
		t.Fatalf("expected 'second message', got %q", in.Value())
	}

	// Press Up again → should show "first message".
	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if in.Value() != "first message" {
		t.Fatalf("expected 'first message', got %q", in.Value())
	}
}

// Down arrow walks forward through history; at the end it restores the
// draft the user was typing before pressing Up.
func TestInputDownRestoresDraft(t *testing.T) {
	in := components.NewInput("test-model")
	in, _ = in.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Type something, then browse history.
	for _, r := range "my draft" {
		in, _ = in.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	in.AddHistory("previous")

	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if in.Value() != "previous" {
		t.Fatalf("expected 'previous', got %q", in.Value())
	}

	// Down → back to draft.
	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if in.Value() != "my draft" {
		t.Fatalf("expected draft 'my draft', got %q", in.Value())
	}
}

// Empty history: Up and Down are no-ops.
func TestInputHistoryEmptyNoOp(t *testing.T) {
	in := components.NewInput("test-model")
	in, _ = in.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	for _, r := range "hello" {
		in, _ = in.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if in.Value() != "hello" {
		t.Fatalf("Up on empty history should not change value, got %q", in.Value())
	}
}

// Typing after browsing history resets the history cursor so the next
// Up starts from scratch.
func TestInputTypingResetsHistoryCursor(t *testing.T) {
	in := components.NewInput("test-model")
	in, _ = in.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	in.AddHistory("old")

	// Browse to history.
	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if in.Value() != "old" {
		t.Fatalf("expected 'old', got %q", in.Value())
	}

	// Type something — cursor should reset.
	in, _ = in.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})

	// Up → goes to history.
	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if in.Value() != "old" {
		t.Fatalf("expected 'old' after reset, got %q", in.Value())
	}
}

// Esc on empty input is a no-op (no hint, no clear).
func TestInputEscOnEmptyIsNoOp(t *testing.T) {
	in := components.NewInput("test-model")
	in, _ = in.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Input is empty; pressing Esc should NOT return a StatusMsg cmd.
	_, cmd := in.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd != nil {
		// Run the cmd and check it doesn't produce a StatusMsg.
		msg := cmd()
		if _, ok := msg.(components.StatusMsg); ok {
			t.Fatalf("Esc on empty input should not produce StatusMsg, got: %+v", msg)
		}
	}
}

// Single Esc is consumed and does NOT clear the input.
func TestInputSingleEscNoClear(t *testing.T) {
	in := components.NewInput("test-model")
	in, _ = in.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Type something first.
	for _, r := range "hello" {
		in, _ = in.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	if in.Value() != "hello" {
		t.Fatalf("expected value 'hello', got %q", in.Value())
	}

	// Single Esc → value unchanged.
	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if in.Value() != "hello" {
		t.Fatalf("single Esc should not clear, got %q", in.Value())
	}
}

// Double Esc (within debounce window) clears the input.
func TestInputDoubleEscClears(t *testing.T) {
	in := components.NewInput("test-model")
	in, _ = in.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	for _, r := range "hello" {
		in, _ = in.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	time.Sleep(50 * time.Millisecond) // well inside the 1 s window
	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if in.Value() != "" {
		t.Fatalf("double Esc should clear input, got %q", in.Value())
	}
}

// Double Esc spaced too far apart (outside debounce) does NOT clear.
func TestInputDoubleEscTooSlowNoClear(t *testing.T) {
	in := components.NewInput("test-model")
	in, _ = in.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	for _, r := range "hello" {
		in, _ = in.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	time.Sleep(1100 * time.Millisecond) // outside the 1 s window
	in, _ = in.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if in.Value() != "hello" {
		t.Fatalf("slow double Esc should not clear, got %q", in.Value())
	}
}
