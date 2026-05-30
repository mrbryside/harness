package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/app"
	"github.com/mrbryside/harness/llm"
)

// wheelFilter drops mouse-wheel events that can't do anything anyway:
// wheel-up when the chat is already at the top, wheel-down when at the
// bottom. Trackpads emit wheel events at very high rates; without this,
// a hard flick fills the event queue with hundreds of events that the
// main loop has to process one-by-one (each one re-renders the View),
// so scrolling feels frozen for a beat after the user stops.
//
// Wheel events over the *input* panel are never filtered — the input
// has its own scrollable viewport and the user expects wheel-up/down
// there to scroll the input regardless of the chat's scroll position.
func wheelFilter(m tea.Model, msg tea.Msg) tea.Msg {
	wheel, ok := msg.(tea.MouseWheelMsg)
	if !ok {
		return msg
	}
	model, ok := m.(app.Model)
	if !ok {
		return msg
	}
	if model.MouseInInput(wheel.X, wheel.Y) {
		return msg
	}
	switch wheel.Button {
	case tea.MouseWheelUp:
		if model.ChatAtTop() {
			return nil
		}
	case tea.MouseWheelDown:
		if model.ChatAtBottom() {
			return nil
		}
	}
	return msg
}

func main() {
	// Bubble Tea v2 no longer auto-probes the terminal's background color in
	// init() (that was the v1 bug whose stray OSC 11 reply leaked into the
	// textarea on Zed and similar terminals). All terminal feature flags —
	// alt-screen and mouse capture — are now declared on the root tea.View
	// in app/view.go.
	provider := &llm.MockProvider{}
	model := app.New(provider)

	p := tea.NewProgram(model, tea.WithFilter(wheelFilter))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
