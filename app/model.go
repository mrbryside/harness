package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/components"
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
}

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
