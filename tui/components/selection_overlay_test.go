package components

import (
	"strings"
	"testing"

	"github.com/mrbryside/harness/tui/styles"
)

func TestWrapLineRangeUsesCallerBg(t *testing.T) {
	const fakeBg = "\x1b[48;2;9;9;9m"
	got := WrapLineRange("hello world", 6, 11, fakeBg)
	if !strings.Contains(got, styles.SelectionBgSGR+"world") {
		t.Errorf("expected selection BG before 'world', got %q", got)
	}
	if !strings.Contains(got, "world"+fakeBg) {
		t.Errorf("expected caller BG after 'world', got %q", got)
	}
}

func TestWrapLineRangeSkipsANSI(t *testing.T) {
	src := "\x1b[31mhello \x1b[0mworld"
	got := WrapLineRange(src, 6, 11, "\x1b[m")
	if !strings.Contains(got, styles.SelectionBgSGR+"world") {
		t.Errorf("expected selection BG before 'world', got %q", got)
	}
}

func TestWrapLineRangeRewritesInnerBg(t *testing.T) {
	src := "abc\x1b[48;2;26;27;38m def\x1b[m"
	got := WrapLineRange(src, 0, 7, "\x1b[m")
	if strings.Contains(got, "\x1b[48;2;26;27;38m") {
		t.Errorf("inner PanelBg should be rewritten, got %q", got)
	}
	if !strings.Contains(got, styles.SelectionBgSGR) {
		t.Errorf("expected selection BG present, got %q", got)
	}
}

func TestWrapLineRangeReassertsAfterReset(t *testing.T) {
	src := "ab\x1b[mcd"
	got := WrapLineRange(src, 0, 4, "\x1b[m")
	if !strings.Contains(got, "\x1b[m"+styles.SelectionBgSGR) {
		t.Errorf("expected selectionBg re-asserted after reset, got %q", got)
	}
}
