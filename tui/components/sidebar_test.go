package components_test

import (
	"strings"
	"testing"

	"github.com/mrbryside/harness/tui/components"
)

func TestSidebarViewContainsModelName(t *testing.T) {
	m := components.NewSidebar("mock")
	if !strings.Contains(m.View(), "mock") {
		t.Errorf("expected View() to contain model name %q, got:\n%s", "mock", m.View())
	}
}

func TestSidebarViewContainsTokenCount(t *testing.T) {
	m := components.NewSidebar("mock")
	m.SetTokens(1234)
	if !strings.Contains(m.View(), "1,234") {
		t.Errorf("expected View() to contain formatted token count %q, got:\n%s", "1,234", m.View())
	}
}

func TestSidebarViewContainsConnected(t *testing.T) {
	m := components.NewSidebar("mock")
	if !strings.Contains(m.View(), "Connected") {
		t.Errorf("expected View() to contain %q, got:\n%s", "Connected", m.View())
	}
}

func TestSidebarUpdateTokensReflectedInView(t *testing.T) {
	m := components.NewSidebar("mock")
	m.SetTokens(42)
	if !strings.Contains(m.View(), "42") {
		t.Errorf("expected View() to contain token count %q after update, got:\n%s", "42", m.View())
	}
	m.SetTokens(999)
	if !strings.Contains(m.View(), "999") {
		t.Errorf("expected View() to contain updated token count %q, got:\n%s", "999", m.View())
	}
}
