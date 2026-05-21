package components

import "testing"

func TestExtract(t *testing.T) {
	got := Extract("alpha\nbeta\ngamma", 0, 2, 2, 3)
	if got != "pha\nbeta\ngam" {
		t.Errorf("got %q", got)
	}
}

func TestSliceRunesEndSentinel(t *testing.T) {
	got := SliceRunes("hello", 2, -1)
	if got != "llo" {
		t.Errorf("expected %q, got %q", "llo", got)
	}
}

func TestSliceRunesStartGreaterThanLen(t *testing.T) {
	got := SliceRunes("abc", 10, 20)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestExtractSelectionThai(t *testing.T) {
	got := Extract("สวัสดี ครับ", 0, 0, 0, 6)
	if got != "สวัสดี" {
		t.Errorf("expected %q, got %q", "สวัสดี", got)
	}
}

func TestExtractSelectionStripsANSI(t *testing.T) {
	src := "\x1b[31mhello\x1b[0m world"
	got := Extract(src, 0, 0, 0, 5)
	if got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
}
