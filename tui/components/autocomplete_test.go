package components

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// Default autocomplete is hidden.
func TestAutocompleteDefaultHidden(t *testing.T) {
	a := NewAutocomplete()
	if a.Active() {
		t.Error("expected fresh Autocomplete to be hidden")
	}
}

// Show with empty prefix displays all suggestions.
func TestAutocompleteShowDisplaysSuggestions(t *testing.T) {
	a := NewAutocomplete()
	a.Show("")
	if !a.Active() {
		t.Error("expected Show() to activate autocomplete")
	}
	view := a.View(80)
	if !strings.Contains(view, "/help") {
		t.Fatalf("expected /help in view, got:\n%s", view)
	}
}

// Show with prefix filters suggestions.
func TestAutocompleteShowFiltersByPrefix(t *testing.T) {
	a := NewAutocomplete()
	a.Show("he")
	view := a.View(80)
	if !strings.Contains(view, "/help") {
		t.Fatalf("expected /help for prefix 'he', got:\n%s", view)
	}

	a.Show("xx")
	view = a.View(80)
	if strings.Contains(view, "/help") {
		t.Fatalf("expected no suggestions for prefix 'xx', got:\n%s", view)
	}
}

// Hide deactivates autocomplete.
func TestAutocompleteHide(t *testing.T) {
	a := NewAutocomplete()
	a.Show("")
	a.Hide()
	if a.Active() {
		t.Error("expected Hide() to deactivate autocomplete")
	}
}

// Next cycles through suggestions.
func TestAutocompleteNextCycles(t *testing.T) {
	a := NewAutocomplete()
	a.Show("")
	idx := a.SelectedIndex()
	a.Next()
	if a.SelectedIndex() != idx {
		t.Error("expected Next() to cycle (only 1 item, should wrap to 0)")
	}
}

// Prev cycles backwards.
func TestAutocompletePrevCycles(t *testing.T) {
	a := NewAutocomplete()
	a.Show("")
	a.Prev()
	if a.SelectedIndex() != 0 {
		t.Errorf("expected Prev() on single item to stay at 0, got %d", a.SelectedIndex())
	}
}

// Selected returns the current suggestion.
func TestAutocompleteSelected(t *testing.T) {
	a := NewAutocomplete()
	a.Show("")
	cmd, desc := a.Selected()
	if cmd != "/help" {
		t.Errorf("expected selected command '/help', got %q", cmd)
	}
	if desc == "" {
		t.Error("expected non-empty description")
	}
}

// View respects width.
func TestAutocompleteViewRespectsWidth(t *testing.T) {
	a := NewAutocomplete()
	a.Show("")
	view := a.View(40)
	lines := strings.Split(view, "\n")
	for _, line := range lines {
		if lipgloss.Width(line) > 40 {
			t.Errorf("line exceeds width 40 (visible=%d): %q", lipgloss.Width(line), line)
		}
	}
}
