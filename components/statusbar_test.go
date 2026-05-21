package components_test

import (
	"strings"
	"testing"

	"github.com/mrbryside/harness/components"
)

func TestStatusBarViewContainsCtrlP(t *testing.T) {
	m := components.NewStatusBar()
	if !strings.Contains(m.View(), "ctrl+p") {
		t.Errorf("expected View() to contain %q, got:\n%s", "ctrl+p", m.View())
	}
}

func TestStatusBarViewContainsVersion(t *testing.T) {
	m := components.NewStatusBar()
	if !strings.Contains(m.View(), "v0.1.0") {
		t.Errorf("expected View() to contain version %q, got:\n%s", "v0.1.0", m.View())
	}
}
