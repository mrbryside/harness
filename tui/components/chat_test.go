package components_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/components"
)

func TestChatNewChatCreatesComponent(t *testing.T) {
	c := components.NewChat(80, 24)
	if c.Init() != nil {
		t.Error("expected Init() to return nil")
	}
}

func TestChatAtTopInitially(t *testing.T) {
	c := components.NewChat(80, 24)
	if !c.AtTop() {
		t.Error("expected fresh Chat to be at top")
	}
}

func TestChatUpdateWindowSize(t *testing.T) {
	c := components.NewChat(80, 24)
	updated, _ := c.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	// Should not panic; basic smoke test.
	_ = updated.View()
}
