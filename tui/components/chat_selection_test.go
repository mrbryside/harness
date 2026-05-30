package components_test

import (
	"testing"

	"github.com/mrbryside/harness/tui/components"
)

func TestChatSelectionLifecycle(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("assistant", "hello")

	if c.HasSelection() {
		t.Fatalf("expected no selection on fresh Chat")
	}
	c.SelectStart(0, 0)
	c.SelectExtend(0, 3)
	if !c.HasSelection() {
		t.Fatalf("expected HasSelection() after SelectStart+Extend")
	}
	c.SelectClear()
	if c.HasSelection() {
		t.Fatalf("expected no selection after SelectClear")
	}
	if got := c.SelectedText(); got != "" {
		t.Errorf("expected empty SelectedText after clear, got %q", got)
	}
}

func TestChatSelectionEmptyWhenInactive(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("assistant", "hello")
	if got := c.SelectedText(); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}