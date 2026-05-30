package agentruntime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/llm"
	"github.com/mrbryside/harness/tui/components"
)

type testProvider struct {
	response string
}

func (t *testProvider) Name() string { return "test" }

func (t *testProvider) ChatCompletion(_ context.Context, _ []llm.Message, _ llm.Options) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, 2)
	ch <- llm.Chunk{Content: t.response, TokensUsed: 10, Done: false}
	ch <- llm.Chunk{Content: "", TokensUsed: 5, Done: true}
	close(ch)
	return ch, nil
}

func TestAgentRuntimeEmitsQuestionAskedOnInit(t *testing.T) {
	eb := eventbus.NewEventBus()

	var received struct {
		QuestionID string
		Question   string
		mu         sync.Mutex
	}

	eb.Subscribe(eventbus.EventQuestionAsked, func(e eventbus.Event) {
		data := e.Data.(struct {
			QuestionID string
			Question   string
		})
		received.mu.Lock()
		received.QuestionID = data.QuestionID
		received.Question = data.Question
		received.mu.Unlock()
	})

	New(eb, &testProvider{response: "hi"})

	received.mu.Lock()
	defer received.mu.Unlock()

	if received.QuestionID != "startup-test" {
		t.Errorf("expected QuestionID='startup-test', got %q", received.QuestionID)
	}
	if received.Question != "Welcome! Do you want to enable the demo mode?" {
		t.Errorf("expected Question='Welcome! Do you want to enable the demo mode?', got %q", received.Question)
	}
}

func TestAgentRuntimeEmitsToolUpdatesOnInit(t *testing.T) {
	eb := eventbus.NewEventBus()

	var toolUpdates []eventbus.ToolEditFileEvent
	var mu sync.Mutex
	eb.Subscribe(eventbus.EventToolEditFileExecuted, func(e eventbus.Event) {
		data := e.Data.(eventbus.ToolEditFileEvent)
		mu.Lock()
		toolUpdates = append(toolUpdates, data)
		mu.Unlock()
	})

	New(eb, &testProvider{response: "hi"})

	mu.Lock()
	defer mu.Unlock()

	if len(toolUpdates) != 2 {
		t.Fatalf("expected 2 tool updates, got %d", len(toolUpdates))
	}

	if toolUpdates[0].Path != "demos/agent_demo.go" {
		t.Errorf("expected path='demos/agent_demo.go', got %q", toolUpdates[0].Path)
	}
	if toolUpdates[0].StartLine != 19 {
		t.Errorf("expected StartLine=19, got %d", toolUpdates[0].StartLine)
	}
	if toolUpdates[1].StartLine != 24 {
		t.Errorf("expected StartLine=24, got %d", toolUpdates[1].StartLine)
	}
}

func TestAgentRuntimeSubscribesToQuestionAnswered(t *testing.T) {
	eb := eventbus.NewEventBus()

	New(eb, &testProvider{response: "hi"})

	var answered components.QuestionChoice
	var mu sync.Mutex
	eb.Subscribe(eventbus.EventQuestionAnswered, func(e eventbus.Event) {
		data := e.Data.(struct {
			QuestionID string
			Answer     components.QuestionChoice
		})
		mu.Lock()
		answered = data.Answer
		mu.Unlock()
	})

	eb.Emit(eventbus.EventQuestionAnswered, struct {
		QuestionID string
		Answer     components.QuestionChoice
	}{
		QuestionID: "startup-test",
		Answer: components.QuestionChoice{
			Index: 0,
			Label: "Yes, allow this permission.",
		},
	})

	mu.Lock()
	defer mu.Unlock()
	if answered.Index != 0 {
		t.Errorf("expected answered.Index=0, got %d", answered.Index)
	}
	if answered.Label != "Yes, allow this permission." {
		t.Errorf("expected answered.Label='Yes, allow this permission.', got %q", answered.Label)
	}
}

func TestAgentRuntimeSubscribesToUserMessagedAndStreamsResponse(t *testing.T) {
	eb := eventbus.NewEventBus()
	New(eb, &testProvider{response: "hello world"})

	var assistantMessages []eventbus.AssistantMessageEvent
	var mu sync.Mutex
	var done chan struct{}
	eb.Subscribe(eventbus.EventAssistantMessaged, func(e eventbus.Event) {
		data := e.Data.(eventbus.AssistantMessageEvent)
		mu.Lock()
		assistantMessages = append(assistantMessages, data)
		isDone := data.Done
		mu.Unlock()
		if isDone && done != nil {
			close(done)
		}
	})

	done = make(chan struct{})
	eb.Emit(eventbus.EventUserMessaged, eventbus.UserMessageEvent{
		ID:      "test-1",
		Content: "say hello",
	})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for assistant messages")
	}

	mu.Lock()
	defer mu.Unlock()

	if len(assistantMessages) < 2 {
		t.Fatalf("expected at least 2 assistant messages (content + done), got %d", len(assistantMessages))
	}

	if assistantMessages[0].Content != "hello world" {
		t.Errorf("expected first chunk content='hello world', got %q", assistantMessages[0].Content)
	}
	if assistantMessages[0].Done {
		t.Error("expected first chunk Done=false")
	}

	last := assistantMessages[len(assistantMessages)-1]
	if !last.Done {
		t.Error("expected last chunk Done=true")
	}
}
