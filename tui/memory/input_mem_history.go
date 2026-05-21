package memory

import "strings"

// HistoryStorage persists and retrieves previously-entered messages.
//
// The app layer (or tests) may swap implementations — in-memory,
// on-disk, sqlite, redis, etc. — without changing the Input component.
//
// Cursor semantics
// ----------------
// The storage maintains an internal cursor that points to the "active"
// history item while the user is browsing with Previous / Next.
//   cursor == -1   → blank slate ( newest position ); Next returns false.
//   cursor >= 0    → pointing at items[cursor]; Previous moves toward 0.
type HistoryStorage interface {
	// Add appends a non-empty message to the history and resets the
	// cursor to the blank-slate position (-1).
	Add(text string)
	// Previous moves the cursor back one item and returns it.
	// Returns ("", false) when already at the oldest item (index 0).
	Previous() (string, bool)
	// Next moves the cursor forward one item and returns it.
	// Returns ("", false) when at the blank-slate position (-1).
	Next() (string, bool)
	// ResetCursor puts the cursor back at the blank-slate position.
	ResetCursor()
	// Cursor returns the current cursor position (-1 = blank slate).
	Cursor() int
}

// InMemoryHistory is the default HistoryStorage implementation.
// It keeps all messages in a slice and is fully lost on program exit.
type InMemoryHistory struct {
	items  []string
	cursor int // -1 = blank slate; 0..len-1 = browsing
}

// NewInMemoryHistory creates an empty in-memory history store.
func NewInMemoryHistory() *InMemoryHistory {
	return &InMemoryHistory{cursor: -1}
}

// Add stores text if it is non-empty and not pure whitespace, then
// resets the browsing cursor.
func (h *InMemoryHistory) Add(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	h.items = append(h.items, text)
	h.cursor = -1
}

// Previous walks backward through history (newest → oldest).
func (h *InMemoryHistory) Previous() (string, bool) {
	if len(h.items) == 0 {
		return "", false
	}
	if h.cursor < 0 {
		h.cursor = len(h.items) - 1
	} else if h.cursor > 0 {
		h.cursor--
	}
	// if cursor was already 0 we stay at 0 and still return the item
	return h.items[h.cursor], true
}

// Next walks forward through history (oldest → newest).
// When the cursor reaches the blank-slate position it returns false.
func (h *InMemoryHistory) Next() (string, bool) {
	if h.cursor < 0 || h.cursor >= len(h.items)-1 {
		h.cursor = -1
		return "", false
	}
	h.cursor++
	return h.items[h.cursor], true
}

// ResetCursor puts the cursor back at the blank-slate position.
func (h *InMemoryHistory) ResetCursor() {
	h.cursor = -1
}

// Cursor returns the current cursor position.
func (h *InMemoryHistory) Cursor() int {
	return h.cursor
}
