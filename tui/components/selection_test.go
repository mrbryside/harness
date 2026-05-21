package components

import (
	"strings"
	"testing"
)

// Lifecycle: Start activates, Clear deactivates, Extend updates end.
func TestSelectionLifecycle(t *testing.T) {
	var s Selection
	if s.Active() {
		t.Fatalf("zero Selection must not be Active")
	}
	s.Start(2, 5)
	if !s.Active() {
		t.Fatalf("Active() expected true after Start")
	}
	if s.HasRange() {
		t.Fatalf("HasRange() must be false immediately after Start (start==end)")
	}
	s.Extend(2, 9)
	if !s.HasRange() {
		t.Fatalf("HasRange() expected true after Extend moves end")
	}
	s.Clear()
	if s.Active() {
		t.Fatalf("Active() expected false after Clear")
	}
}

// Extend on an inactive selection must be a no-op.
func TestSelectionExtendInactiveIsNoop(t *testing.T) {
	var s Selection
	s.Extend(0, 5)
	if s.Active() {
		t.Fatalf("Extend on inactive selection must not activate it")
	}
}

// Backwards drag: start is after end. Normalised() must sort them.
func TestSelectionNormalisedBackwardsDrag(t *testing.T) {
	var s Selection
	s.Start(3, 4)
	s.Extend(1, 2)
	sl, sc, el, ec := s.Normalised()
	if sl != 1 || sc != 2 || el != 3 || ec != 4 {
		t.Errorf("expected (1,2,3,4), got (%d,%d,%d,%d)", sl, sc, el, ec)
	}
}

// Text() returns "" when there's no range, even if active.
func TestSelectionTextEmptyWhenNoRange(t *testing.T) {
	var s Selection
	s.Start(0, 0)
	if got := s.Text("hello"); got != "" {
		t.Errorf("expected empty Text when start==end, got %q", got)
	}
}

// Text() reads from supplied content (decoupled from any component).
func TestSelectionTextFromContent(t *testing.T) {
	var s Selection
	s.Start(0, 0)
	s.Extend(1, 3)
	got := s.Text("abcdef\nghijkl")
	if got != "abcdef\nghi" {
		t.Errorf("expected %q, got %q", "abcdef\nghi", got)
	}
}

// Overlay paints SelectionBgSGR over the selected range using the
// caller-supplied background reset SGR.
func TestSelectionOverlayUsesCallerBg(t *testing.T) {
	const fakeBg = "\x1b[48;2;1;2;3m"
	var s Selection
	s.Start(0, 0)
	s.Extend(0, 5)

	rendered := "hello world"
	got := s.Overlay(rendered, 0, fakeBg)
	if !strings.Contains(got, SelectionBgSGR+"hello") {
		t.Errorf("expected selection BG before 'hello', got %q", got)
	}
	if !strings.Contains(got, "hello"+fakeBg) {
		t.Errorf("expected caller BG restored after 'hello', got %q", got)
	}
}

// Overlay translates content-space line numbers via yoff.
func TestSelectionOverlayYOffsetTranslation(t *testing.T) {
	var s Selection
	s.Start(2, 0)
	s.Extend(2, 3)

	rendered := "row0\nrow1\nrow2"
	const fakeBg = "\x1b[48;2;0;0;0m"

	got := s.Overlay(rendered, 2, fakeBg)
	first := strings.Split(got, "\n")[0]
	if !strings.HasPrefix(first, SelectionBgSGR) {
		t.Errorf("expected first visible line to start with selection BG, got %q", first)
	}
}

// Inactive selection: Overlay must be a no-op.
func TestSelectionOverlayInactiveNoop(t *testing.T) {
	var s Selection
	rendered := "hello"
	if got := s.Overlay(rendered, 0, "\x1b[m"); got != rendered {
		t.Errorf("Overlay on inactive selection should return input unchanged")
	}
}
