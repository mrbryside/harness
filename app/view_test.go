package app_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/app"
	"github.com/mrbryside/harness/components"
)

func TestViewContainsChatContent(t *testing.T) {
	m := app.New(&stubProvider{response: "hi"})
	// set window size so chat viewport has proper dimensions
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	result, _ := m2.Update(components.SendMsg{Content: "hello"})
	view := result.View().Content
	if !containsText(view, "hello") {
		t.Errorf("expected View() to contain user message %q, got:\n%s", "hello", view)
	}
}

func TestViewContainsConnected(t *testing.T) {
	m := app.New(&stubProvider{response: "hi"})
	if !strings.Contains(m.View().Content, "Connected") {
		t.Errorf("expected View() to contain %q, got:\n%s", "Connected", m.View().Content)
	}
}

func TestViewContainsCtrlP(t *testing.T) {
	m := app.New(&stubProvider{response: "hi"})
	if !strings.Contains(m.View().Content, "ctrl+p") {
		t.Errorf("expected View() to contain %q, got:\n%s", "ctrl+p", m.View().Content)
	}
}

func TestViewResizesWithWindowSize(t *testing.T) {
	m := app.New(&stubProvider{response: "hi"})
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	view := result.View().Content
	if view == "" {
		t.Error("expected non-empty View() after window resize")
	}
}
