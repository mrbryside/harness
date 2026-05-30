package app

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/eventbus"
)

func newTestModelForMouse() Model {
	eb := eventbus.NewEventBus()
	return New(eb)
}

func TestScrollTickCmd(t *testing.T) {
	cmd := scrollTickCmd()
	if cmd == nil {
		t.Fatal("scrollTickCmd should not return nil")
	}
	msg := cmd()
	if _, ok := msg.(scrollTickMsg); !ok {
		t.Fatalf("expected scrollTickMsg, got %T", msg)
	}
}

func TestAutoScrollStateTransitions(t *testing.T) {
	m := newTestModelForMouse()

	// Initially no auto-scroll
	if m.chatAutoScrollDir != 0 {
		t.Fatalf("expected initial dir 0, got %d", m.chatAutoScrollDir)
	}

	// Simulate starting auto-scroll up by setting state directly
	// (the coordinate math is complex, test the state machine instead)
	m.chatAutoScrollDir = -1
	m.chatAutoScrollCol = 5

	// Handle tick while scrolling up
	model, cmd := m.handleScrollTick()
	m = model.(Model)
	if m.chatAutoScrollDir != -1 {
		t.Fatalf("expected dir -1 to continue, got %d", m.chatAutoScrollDir)
	}
	if cmd == nil {
		t.Fatal("expected next tick command")
	}

	// Verify tick command produces scrollTickMsg
	msg := cmd()
	if _, ok := msg.(scrollTickMsg); !ok {
		t.Fatalf("expected scrollTickMsg, got %T", msg)
	}

	// Simulate stopping auto-scroll
	m.chatAutoScrollDir = 0
	model, cmd = m.handleScrollTick()
	m = model.(Model)
	if m.chatAutoScrollDir != 0 {
		t.Fatalf("expected dir 0, got %d", m.chatAutoScrollDir)
	}
	if cmd != nil {
		t.Fatal("expected no command when auto-scroll is off")
	}
}

func TestAutoScrollDown(t *testing.T) {
	m := newTestModelForMouse()

	m.chatAutoScrollDir = 1
	m.chatAutoScrollCol = 10

	model, cmd := m.handleScrollTick()
	m = model.(Model)
	if m.chatAutoScrollDir != 1 {
		t.Fatalf("expected dir 1 to continue, got %d", m.chatAutoScrollDir)
	}
	if cmd == nil {
		t.Fatal("expected next tick command")
	}
}

func TestMouseReleaseStopsAutoScroll(t *testing.T) {
	m := newTestModelForMouse()

	m.chatAutoScrollDir = -1
	model, _ := m.handleMouseRelease(tea.MouseReleaseMsg{Button: tea.MouseLeft})
	m = model.(Model)
	if m.chatAutoScrollDir != 0 {
		t.Fatalf("expected dir 0 after release, got %d", m.chatAutoScrollDir)
	}
}

func TestScrollTickInterval(t *testing.T) {
	if scrollTickInterval <= 0 {
		t.Fatal("scrollTickInterval must be positive")
	}
	if scrollTickInterval > 200*time.Millisecond {
		t.Fatalf("scrollTickInterval %v is too slow for smooth scrolling", scrollTickInterval)
	}
	if scrollTickLines <= 0 {
		t.Fatal("scrollTickLines must be positive")
	}
}