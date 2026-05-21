package components_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/components"
)

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
