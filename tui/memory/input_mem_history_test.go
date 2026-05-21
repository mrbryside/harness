package memory

import "testing"

// InMemoryHistory must satisfy HistoryStorage.
var _ HistoryStorage = (*InMemoryHistory)(nil)

func TestInMemoryHistoryAddAndRetrieve(t *testing.T) {
	h := NewInMemoryHistory()
	h.Add("first")
	h.Add("second")
	h.Add("third")

	text, ok := h.Previous()
	if !ok || text != "third" {
		t.Fatalf("expected 'third', got %q (ok=%v)", text, ok)
	}
	text, ok = h.Previous()
	if !ok || text != "second" {
		t.Fatalf("expected 'second', got %q (ok=%v)", text, ok)
	}
}

// Next after Previous walks forward again.
func TestInMemoryHistoryNext(t *testing.T) {
	h := NewInMemoryHistory()
	h.Add("alpha")
	h.Add("beta")

	h.Previous() // beta
	h.Previous() // alpha
	text, ok := h.Next()
	if !ok || text != "beta" {
		t.Fatalf("expected 'beta', got %q (ok=%v)", text, ok)
	}
}

// Next at the newest position returns empty.
func TestInMemoryHistoryNextAtEnd(t *testing.T) {
	h := NewInMemoryHistory()
	h.Add("only")

	h.Previous() // only
	text, ok := h.Next()
	if ok {
		t.Fatalf("expected no item at end, got %q", text)
	}
}

// Empty history returns false.
func TestInMemoryHistoryEmpty(t *testing.T) {
	h := NewInMemoryHistory()
	_, ok := h.Previous()
	if ok {
		t.Fatal("expected false on empty history")
	}
}

// Add resets cursor to the end.
func TestInMemoryHistoryAddResetsCursor(t *testing.T) {
	h := NewInMemoryHistory()
	h.Add("a")
	h.Add("b")
	h.Previous() // b
	h.Previous() // a

	h.Add("c") // cursor resets to end
	text, ok := h.Previous()
	if !ok || text != "c" {
		t.Fatalf("expected 'c' after Add reset, got %q", text)
	}
}

// Blank messages are ignored.
func TestInMemoryHistoryIgnoresBlank(t *testing.T) {
	h := NewInMemoryHistory()
	h.Add("")
	h.Add("  ")
	_, ok := h.Previous()
	if ok {
		t.Fatal("expected blank messages to be ignored")
	}
}
