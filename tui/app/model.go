package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
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
	provider     llm.LLMProvider // always the interface — never the concrete type
	messages     []llm.Message
	streaming    bool
	streamCh     <-chan llm.Chunk // active stream channel; nil when not streaming
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
}

// ctrlCDebounceWindow is how long the user has to press Ctrl+C a second
// time after the first press before the hint resets.
const ctrlCDebounceWindow = 1500 * time.Millisecond

// New creates an initialised Model with the given provider.
func New(provider llm.LLMProvider) Model {
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
	// Demo 1: Small addition (comments added)
	chat.AppendCodeDiff(components.CodeDiff{
		Path: "demos/agent_demo.go",
		OldContent: `func (c *agentCommand) Name() string        { return "agent" }
func (c *agentCommand) Description() string { return "Switch AI agent" }`,
		NewContent: `// Name returns the command name.
func (c *agentCommand) Name() string        { return "agent" }
// Description returns the command description.
func (c *agentCommand) Description() string { return "Switch AI agent" }`,
		StartLine: 19,
	})

	// Demo 2: Delete Execute() and add a big new function (~20 lines)
	chat.AppendCodeDiff(components.CodeDiff{
		Path: "demos/agent_demo.go",
		OldContent: `func (c *agentCommand) Execute(args string) Result {
	return Result{
		Chat:  "Available agents: coder, reviewer, architect",
		Toast: "✓ Agent list",
	}
}`,
		NewContent: `// Execute runs the agent command with full orchestration.
// It parses the args, validates the agent name, loads config,
// initializes the provider, and streams the response back.
func (c *agentCommand) Execute(args string) Result {
	// Parse and validate input
	if strings.TrimSpace(args) == "" {
		return Result{
			Chat:  "Error: agent name required",
			Toast: "✗ Missing agent",
		}
	}

	// Load available agents from config
	agents := []string{"coder", "reviewer", "architect", "debugger"}
	found := false
	for _, a := range agents {
		if a == strings.ToLower(strings.TrimSpace(args)) {
			found = true
			break
		}
	}

	if !found {
		return Result{
			Chat:  fmt.Sprintf("Unknown agent: %s", args),
			Toast: "✗ Invalid agent",
		}
	}

	// Initialize provider and stream response
	provider := loadProvider(args)
	if provider == nil {
		return Result{
			Chat:  "Failed to initialize provider",
			Toast: "✗ Provider error",
		}
	}

	return Result{
		Chat:  fmt.Sprintf("Switched to %s agent", args),
		Toast: fmt.Sprintf("✓ Active: %s", args),
	}
}`,
		StartLine: 24,
	})

	return Model{
		chat:         chat,
		input:        components.NewInput(provider.Name()),
		sidebar:      components.NewSidebar(provider.Name()),
		statusbar:    components.NewStatusBar(),
		autocomplete: auto,
		cmdRegistry:  registry,
		provider:     provider,
		messages:     []llm.Message{},
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
