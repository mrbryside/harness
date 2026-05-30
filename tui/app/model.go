package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/llm"
	"github.com/mrbryside/harness/tui/commands"
	"github.com/mrbryside/harness/tui/components"
)

// Model is the root Bubble Tea model that wires all components together.
type Model struct {
	chat         components.Chat
	input        components.Input
	sidebar      components.Sidebar
	statusbar    components.StatusBar
	autocomplete components.Autocomplete
	cmdRegistry  *commands.Registry
	messages     []llm.Message
	streaming    bool
	width        int
	height       int

	// Ctrl+C debounce state: when the user presses Ctrl+C once while not
	// streaming we show a hint and wait for a second press within this
	// window before actually quitting.
	ctrlCDebounce time.Time

	// Auto-scroll state for smooth selection scrolling past viewport edges.
	// -1 = scrolling up, 0 = off, 1 = scrolling down.
	chatAutoScrollDir  int
	chatAutoScrollCol  int

	// Active question overlay (pointer so EventBus subscriber can mutate it).
	activeQuestion *activeQuestionHolder
	eventBus       *eventbus.EventBus
	eventCh        chan tea.Msg

	// Current streaming request ID (for cancellation).
	currentRequestID string
}

type activeQuestionHolder struct {
	question components.Question
}

// ctrlCDebounceWindow is how long the user has to press Ctrl+C a second
// time after the first press before the hint resets.
const ctrlCDebounceWindow = 1500 * time.Millisecond

// New creates an initialised Model with the given event bus.
func New(eb *eventbus.EventBus) Model {
	registry := commands.NewRegistry()
	registry.Register(commands.NewHelpCommand())
	registry.Register(commands.NewClearCommand())
	registry.Register(commands.NewBuildCommand())
	registry.Register(commands.NewPlanCommand())
	registry.Register(commands.NewModelCommand())
	registry.Register(commands.NewHistoryCommand())
	registry.Register(commands.NewSaveCommand())
	registry.Register(commands.NewLoadCommand())
	registry.Register(commands.NewSettingsCommand())
	registry.Register(commands.NewQuitCommand())
	registry.Register(commands.NewThemeCommand())
	registry.Register(commands.NewExportCommand())
	registry.Register(commands.NewRestartCommand())
	registry.Register(commands.NewAgentCommand())

	auto := components.NewAutocomplete()
	auto.SetProvider(func() []components.Suggestion {
		var suggestions []components.Suggestion
		for _, cmd := range registry.All() {
			suggestions = append(suggestions, components.Suggestion{
				Command:     "/" + cmd.Name(),
				Description: cmd.Description(),
			})
		}
		return suggestions
	})

	chat := components.NewChat(80, 20)

	m := Model{
		chat:             chat,
		input:            components.NewInput("mock"),
		sidebar:          components.NewSidebar("mock"),
		statusbar:        components.NewStatusBar(),
		autocomplete:     auto,
		cmdRegistry:      registry,
		messages:         []llm.Message{},
		activeQuestion: &activeQuestionHolder{},
		eventBus:       eb,
		eventCh:        make(chan tea.Msg, 100),
	}

	m.subscribeQuestionAsked()
	m.subscribeAssistantMessaged()
	m.subscribeToolEditFileExecuted()

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.input.Init(), m.listenEvents())
}

func (m Model) listenEvents() tea.Cmd {
	return func() tea.Msg {
		return <-m.eventCh
	}
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

// AskPermission shows a permission question overlay and reflows chat.
// When the user answers, the event bus emits EventQuestionAnswered.
func (m *Model) AskPermission(question, questionID string) {
	m.activeQuestion.question = components.CreateQuestion(components.QuestionTypePermission, questionID, question)
	*m = m.reflowChat()
}

// QuestionActive returns whether a question is currently shown.
func (m Model) QuestionActive() bool {
	return m.activeQuestion != nil && m.activeQuestion.question != nil && m.activeQuestion.question.Active()
}

// EventChForTest returns the internal event channel for testing.
func (m Model) EventChForTest() <-chan tea.Msg {
	return m.eventCh
}
