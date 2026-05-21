package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/components"
	"github.com/mrbryside/harness/llm"
)

// Model is the root Bubble Tea model that wires all components together.
type Model struct {
	chat      components.Chat
	input     components.Input
	sidebar   components.Sidebar
	statusbar components.StatusBar
	provider  llm.LLMProvider // always the interface — never the concrete type
	messages  []llm.Message
	streaming bool
	streamCh  <-chan llm.Chunk // active stream channel; nil when not streaming
	width     int
	height    int

	// Ctrl+C debounce state: when the user presses Ctrl+C once while not
	// streaming we show a hint and wait for a second press within this
	// window before actually quitting.
	ctrlCDebounce time.Time
}

// ctrlCDebounceWindow is how long the user has to press Ctrl+C a second
// time after the first press before the hint resets.
const ctrlCDebounceWindow = 1500 * time.Millisecond

// New creates an initialised Model with the given provider.
func New(provider llm.LLMProvider) Model {
	return Model{
		chat:      components.NewChat(80, 20),
		input:     components.NewInput(provider.Name()),
		sidebar:   components.NewSidebar(provider.Name()),
		statusbar: components.NewStatusBar(),
		provider:  provider,
		messages:  []llm.Message{},
	}
}

func (m Model) Init() tea.Cmd {
	return m.input.Init()
}

// ChatView exposes the chat component's rendered output for testing.
func (m Model) ChatView() string {
	return m.chat.View()
}

// IsStreaming reports whether a stream is in progress.
func (m Model) IsStreaming() bool {
	return m.streaming
}

// ChatAtTop reports whether the chat viewport is scrolled to the top.
// Exposed so main.go's wheel filter can drop wheel-up events when the
// viewport is already at the top (otherwise heavy trackpad scrolling
// piles up events that take seconds to drain after the user stops).
func (m Model) ChatAtTop() bool { return m.chat.AtTop() }

// ChatAtBottom reports whether the chat viewport is scrolled to the bottom.
func (m Model) ChatAtBottom() bool { return m.chat.AtBottom() }

// MouseInInput reports whether the absolute terminal cell (x, y) falls
// inside the input panel's textarea body. Exposed for main.go's wheel
// filter so wheel events over the input aren't dropped just because
// the chat is at the top/bottom — they need to scroll the input
// instead.
func (m Model) MouseInInput(x, y int) bool {
	_, _, ok := m.inputContentCoord(x, y)
	return ok
}

// StatusBarView returns the rendered status bar (for testing).
func (m Model) StatusBarView() string { return m.statusbar.View() }

// SetStreamingForTest flips the streaming flag and updates the status
// bar message. Used only by tests.
func (m Model) SetStreamingForTest(v bool) Model {
	m.streaming = v
	if v {
		m.statusbar, _ = m.statusbar.SetMessage("⟳ streaming — press Ctrl+C to interrupt", 0)
	} else {
		m.statusbar = m.statusbar.ClearMessage()
	}
	return m
}
