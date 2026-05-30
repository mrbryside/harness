package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/mrbryside/harness/agentruntime"
	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/llm"
	"github.com/mrbryside/harness/tui/app"
	"github.com/mrbryside/harness/tui/components"
)

type testProvider struct {
	response string
}

func (t *testProvider) Name() string { return "test" }

func (t *testProvider) ChatCompletion(ctx context.Context, messages []llm.Message, opts llm.Options) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, 2)
	ch <- llm.Chunk{Content: t.response, TokensUsed: 10, Done: false}
	ch <- llm.Chunk{Content: "", TokensUsed: 5, Done: true}
	close(ch)
	return ch, nil
}

func newFullModel() (app.Model, *agentruntime.AgentRuntime) {
	eb := eventbus.NewEventBus()
	m := app.New(eb)
	ar := agentruntime.New(eb, &testProvider{response: "hello from agent"})
	return m, ar
}

func drainEventCh(m app.Model, timeout time.Duration) app.Model {
	ch := m.EventChForTest()
	if ch == nil {
		return m
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case msg := <-ch:
			result, _ := m.Update(msg)
			m = result.(app.Model)
		case <-time.After(50 * time.Millisecond):
			return m
		}
	}
	return m
}

func TestFullFlowQuestionEmittedOnInit(t *testing.T) {
	m, _ := newFullModel()

	time.Sleep(100 * time.Millisecond)
	m = drainEventCh(m, 200*time.Millisecond)

	if !m.QuestionActive() {
		t.Fatal("expected question to be active after init")
	}
}

func TestFullFlowToolEditsEmittedOnInit(t *testing.T) {
	m, _ := newFullModel()

	time.Sleep(100 * time.Millisecond)
	m = drainEventCh(m, 200*time.Millisecond)

	view := m.ChatView()
	if !containsText(view, "Switched to") {
		t.Errorf("expected chat to contain tool edit content, got:\n%s", view)
	}
	if !containsText(view, "Available agents") {
		t.Errorf("expected chat to contain removed line, got:\n%s", view)
	}
}

func TestFullFlowUserMessageToAssistantResponse(t *testing.T) {
	m, _ := newFullModel()

	time.Sleep(100 * time.Millisecond)
	m = drainEventCh(m, 200*time.Millisecond)

	result, _ := m.Update(components.SendMsg{Content: "ping"})
	m = result.(app.Model)

	if !m.IsStreaming() {
		t.Fatal("expected streaming to start after SendMsg")
	}

	time.Sleep(500 * time.Millisecond)
	m = drainEventCh(m, 1*time.Second)

	view := m.ChatView()
	if !containsText(view, "hello from agent") {
		t.Errorf("expected chat to contain assistant response, got:\n%s", view)
	}

	if m.IsStreaming() {
		t.Error("expected streaming to be false after response complete")
	}
}

func TestSendMsgEmitsUserMessagedEvent(t *testing.T) {
	eb := eventbus.NewEventBus()
	m := app.New(eb)

	var received eventbus.UserMessageEvent
	eb.Subscribe(eventbus.EventUserMessaged, func(e eventbus.Event) {
		received = e.Data.(eventbus.UserMessageEvent)
	})

	m.Update(components.SendMsg{Content: "test message"})

	if received.Content != "test message" {
		t.Errorf("expected Content='test message', got %q", received.Content)
	}
	if received.ID == "" {
		t.Error("expected non-empty ID")
	}
}
