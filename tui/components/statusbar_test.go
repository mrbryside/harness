package components

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Default view shows the version and ctrl+p hint on the right.
func TestStatusBarDefaultView(t *testing.T) {
	sb := NewStatusBar()
	view := sb.View()
	if !strings.Contains(view, "ctrl+p") {
		t.Errorf("expected default view to contain 'ctrl+p', got:\n%s", view)
	}
	if !strings.Contains(view, "v0.1.0") {
		t.Errorf("expected default view to contain version, got:\n%s", view)
	}
}

// SetMessage shows toast content in the center.
func TestStatusBarSetMessage(t *testing.T) {
	sb := NewStatusBar()
	sb, _ = sb.SetMessage("✓ copied", 1*time.Second)
	view := sb.View()
	if !strings.Contains(view, "✓ copied") {
		t.Errorf("expected view to show center message, got:\n%s", view)
	}
}

// After the duration expires, the default content returns.
func TestStatusBarMessageExpires(t *testing.T) {
	sb := NewStatusBar()
	sb, _ = sb.SetMessage("temp", 50*time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	view := sb.View()
	if strings.Contains(view, "temp") {
		t.Errorf("expected message to expire, got:\n%s", view)
	}
	if !strings.Contains(view, "ctrl+p") {
		t.Errorf("expected default view after expiry, got:\n%s", view)
	}
}

// SetMessage with zero duration shows the message until manually cleared.
func TestStatusBarMessageZeroDuration(t *testing.T) {
	sb := NewStatusBar()
	sb, _ = sb.SetMessage("persistent", 0)
	time.Sleep(50 * time.Millisecond)
	view := sb.View()
	if !strings.Contains(view, "persistent") {
		t.Errorf("expected persistent message, got:\n%s", view)
	}
}

// ClearMessage immediately removes the center message.
func TestStatusBarClearMessage(t *testing.T) {
	sb := NewStatusBar()
	sb, _ = sb.SetMessage("temp", 1*time.Second)
	sb = sb.ClearMessage()
	view := sb.View()
	if strings.Contains(view, "temp") {
		t.Errorf("expected message cleared, got:\n%s", view)
	}
}

// Update with a WindowSizeMsg sets the width.
func TestStatusBarUpdateWidth(t *testing.T) {
	sb := NewStatusBar()
	sb, _ = sb.Update(tea.WindowSizeMsg{Width: 120})
	if sb.width != 120 {
		t.Errorf("expected width 120, got %d", sb.width)
	}
}
