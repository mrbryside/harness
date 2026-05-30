package components

import (
	"strings"
	"testing"
)

func TestWrapLineRangeUsesReverseVideo(t *testing.T) {
	const fakeBg = "\x1b[48;2;9;9;9m"
	got := WrapLineRange("hello world", 6, 11, fakeBg, 0)
	// Selection should start with reverse video
	if !strings.Contains(got, "\x1b[7m") {
		t.Errorf("expected reverse video escape, got %q", got)
	}
	// After selection ends, reverse video should be reset
	if !strings.Contains(got, "\x1b[27m") {
		t.Errorf("expected reverse video reset escape, got %q", got)
	}
}

func TestWrapLineRangeSkipsANSI(t *testing.T) {
	src := "\x1b[31mhello \x1b[0mworld"
	got := WrapLineRange(src, 6, 11, "\x1b[m", 0)
	if !strings.Contains(got, "\x1b[7mworld") {
		t.Errorf("expected reverse video before 'world', got %q", got)
	}
}

// Inner backgrounds (code diff add/remove colours) must be preserved,
// not rewritten to selection bg.
func TestWrapLineRangePreservesInnerBg(t *testing.T) {
	const codeDiffBg = "\x1b[48;2;35;48;58m" // add line bg (#23303a)
	src := "abc" + codeDiffBg + " def\x1b[m"
	got := WrapLineRange(src, 0, 7, "\x1b[m", 0)
	if !strings.Contains(got, codeDiffBg) {
		t.Errorf("inner code diff bg should be preserved, got %q", got)
	}
	// Reverse video should be present to indicate selection
	if !strings.Contains(got, "\x1b[7m") {
		t.Errorf("expected reverse video escape for selection visibility, got %q", got)
	}
}

func TestWrapLineRangeReassertsAfterReset(t *testing.T) {
	src := "ab\x1b[mcd"
	got := WrapLineRange(src, 0, 4, "\x1b[m", 0)
	// After reset, should re-assert reverse video
	if !strings.Contains(got, "\x1b[m\x1b[7m") {
		t.Errorf("expected reverse video re-asserted after reset, got %q", got)
	}
}

// Partial selections must NOT pad to full width — the underlying
// background (e.g. code diff add/remove colours) must stay visible.
func TestWrapLineRangePartialDoesNotPad(t *testing.T) {
	const fakeBg = "\x1b[48;2;9;9;9m"
	got := WrapLineRange("hi", 0, 1, fakeBg, 80)
	// After the selected "h" the background is restored, but no padding
	// spaces should be appended.
	if strings.Contains(got, " ") {
		t.Errorf("partial selection should not pad with spaces, got %q", got)
	}
}

// Full-line selections (end < 0) ARE padded so the highlight extends
// to the right edge.
func TestWrapLineRangeFullLinePads(t *testing.T) {
	const fakeBg = "\x1b[48;2;9;9;9m"
	got := WrapLineRange("hi", 0, -1, fakeBg, 80)
	if !strings.Contains(got, strings.Repeat(" ", 78)) {
		t.Errorf("full-line selection should pad to width, got %q", got)
	}
}