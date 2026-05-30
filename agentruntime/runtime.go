package agentruntime

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/llm"
)

type AgentRuntime struct {
	eventBus       *eventbus.EventBus
	provider       llm.LLMProvider
	messages       []llm.Message
	activeRequests map[string]context.CancelFunc
	mu             sync.RWMutex
}

func New(eb *eventbus.EventBus, provider llm.LLMProvider) *AgentRuntime {
	r := &AgentRuntime{
		eventBus:       eb,
		provider:       provider,
		messages:       []llm.Message{},
		activeRequests: make(map[string]context.CancelFunc),
	}
	r.init()
	return r
}

func (r *AgentRuntime) init() {
	r.subscribeQuestionAnswered()
	r.subscribeUserMessaged()
	r.subscribeCancelRequested()

	fmt.Fprintf(os.Stderr, "[AGENT] emitting question_asked: id=%q\n", "startup-test")
	r.eventBus.Emit(eventbus.EventQuestionAsked, struct {
		QuestionID string
		Question   string
	}{
		QuestionID: "startup-test",
		Question:   "Welcome! Do you want to enable the demo mode?",
	})

	// Demo 1: Small addition (comments added)
	fmt.Fprintf(os.Stderr, "[AGENT] emitting tool_edit_file_executed: path=%q\n", "demos/agent_demo.go")
	r.eventBus.Emit(eventbus.EventToolEditFileExecuted, eventbus.ToolEditFileEvent{
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
	fmt.Fprintf(os.Stderr, "[AGENT] emitting tool_edit_file_executed: path=%q\n", "demos/agent_demo.go")
	r.eventBus.Emit(eventbus.EventToolEditFileExecuted, eventbus.ToolEditFileEvent{
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
}

func (r *AgentRuntime) streamResponse(ctx context.Context, requestID string) {
	ch, err := r.provider.ChatCompletion(ctx, r.messages, llm.Options{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[AGENT] chat completion error: %v\n", err)
		r.eventBus.Emit(eventbus.EventAssistantMessaged, eventbus.AssistantMessageEvent{
			ID:      requestID,
			Content: "\n\n*Error: " + err.Error() + "*",
			Done:    true,
		})
		return
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintf(os.Stderr, "[AGENT] request %q cancelled\n", requestID)
			return
		case chunk, ok := <-ch:
			if !ok {
				return
			}
			r.eventBus.Emit(eventbus.EventAssistantMessaged, eventbus.AssistantMessageEvent{
				ID:      requestID,
				Content: chunk.Content,
				Done:    chunk.Done,
			})
			if chunk.Done {
				return
			}
		}
	}
}

func (r *AgentRuntime) EmitToolUpdate(tool eventbus.ToolEditFileEvent) {
	r.eventBus.Emit(eventbus.EventToolEditFileExecuted, tool)
}

func (r *AgentRuntime) LastMessageID() string {
	if len(r.messages) == 0 {
		return uuid.New().String()
	}
	return uuid.New().String()
}
