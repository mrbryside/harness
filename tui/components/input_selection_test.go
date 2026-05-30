package components_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/components"
)

func TestInputSelectionLifecycle(t *testing.T) {
	in := components.NewInput("test-model")
	if in.HasSelection() {
		t.Fatalf("fresh Input must not have a selection")
	}
	in.SelectStart(0, 0)
	in.SelectExtend(0, 3)
	if !in.HasSelection() {
		t.Fatalf("expected HasSelection() after SelectStart+Extend")
	}
	in.SelectClear()
	if in.HasSelection() {
		t.Fatalf("expected no selection after SelectClear")
	}
	if got := in.SelectedText(); got != "" {
		t.Errorf("expected empty SelectedText after clear, got %q", got)
	}
}

func TestInputSelectedTextSingleLine(t *testing.T) {
	in := components.NewInput("test-model")
	for _, r := range "hello world" {
		in, _ = in.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	in.SelectStart(0, 6)
	in.SelectExtend(0, 11)
	if got := in.SelectedText(); got != "world" {
		t.Errorf("expected %q, got %q", "world", got)
	}
}

func TestInputViewOverlaysSelection(t *testing.T) {
	in := components.NewInput("test-model")
	for _, r := range "hello" {
		in, _ = in.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	in.SelectStart(0, 0)
	in.SelectExtend(0, 5)

	view := in.View()
	if !strings.Contains(view, "\x1b[7m") {
		t.Errorf("expected reverse video escape in input View() while selecting, got:\n%q", view)
	}
}
