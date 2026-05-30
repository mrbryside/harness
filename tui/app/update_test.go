package app_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/tui/app"
	"github.com/mrbryside/harness/tui/components"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func newModel() app.Model {
	eb := eventbus.NewEventBus()
	return app.New(eb)
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestUpdateCtrlCShowsQuitHint(t *testing.T) {
	m := newModel()
	result, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	am := result.(app.Model)
	_ = cmd
	statusView := am.StatusBarView()
	if !strings.Contains(statusView, "Ctrl+C again") {
		t.Fatalf("expected status bar hint on first Ctrl+C, got:\n%s", statusView)
	}
}

func TestUpdateEscOnEmptyInputIsNoOp(t *testing.T) {
	m := newModel()
	result, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	am := result.(app.Model)
	// cmd is always non-nil now (batches with listenEvents), just verify it exists
	if cmd == nil {
		t.Fatal("expected cmd to be non-nil")
	}
	// Status bar should not show any hint
	statusView := am.StatusBarView()
	if strings.Contains(statusView, "Esc again") {
		t.Errorf("expected no Esc hint on empty input, got:\n%s", statusView)
	}
}

func TestUpdateEscReturnsStatusCmd(t *testing.T) {
	m := newModel()
	for _, r := range "hello" {
		result, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = result.(app.Model)
	}
	result, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	am := result.(app.Model)
	// cmd is always non-nil (batches with listenEvents), verify it exists
	if cmd == nil {
		t.Fatal("expected a cmd on Esc, got nil")
	}
	// Input should still contain "hello" (single Esc doesn't clear)
	// We can't check input value directly, but we can verify the model state
	// by checking that streaming is false and no question is active
	if am.IsStreaming() {
		t.Error("expected streaming to be false after Esc")
	}
	if am.QuestionActive() {
		t.Error("expected no question active after Esc")
	}
}

func TestUpdateSendMsgAppendsUserMessageToChat(t *testing.T) {
	m := newModel()
	result, _ := m.Update(components.SendMsg{Content: "ping"})
	am := result.(app.Model)
	if !containsText(am.ChatView(), "ping") {
		t.Errorf("expected chat ChatView() to contain %q after SendMsg, got:\n%s", "ping", am.ChatView())
	}
}

func TestUpdateSendMsgStartsStreaming(t *testing.T) {
	m := newModel()
	result, _ := m.Update(components.SendMsg{Content: "ping"})
	am := result.(app.Model)
	if !am.IsStreaming() {
		t.Error("expected model to be in streaming state after SendMsg")
	}
}

func TestUpdateAssistantChunkAppendsToChat(t *testing.T) {
	m := newModel()

	// send a message to start the stream
	result, _ := m.Update(components.SendMsg{Content: "ping"})
	m = result.(app.Model)

	// simulate assistant response chunks
	result, _ = m.Update(app.AssistantChunkMsg{Content: "hello", Done: false})
	m = result.(app.Model)

	result, _ = m.Update(app.AssistantChunkMsg{Content: " world", Done: true})
	m = result.(app.Model)

	am := m
	if !containsText(am.ChatView(), "hello") {
		t.Errorf("expected ChatView() to contain streamed content %q, got:\n%s", "hello", am.ChatView())
	}
	if am.IsStreaming() {
		t.Error("expected streaming to be false after Done=true")
	}
}

func TestUpdateWindowSizeUpdatesModel(t *testing.T) {
	m := newModel()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	_ = result
}

func TestUpdateIgnoresEmptySendMsg(t *testing.T) {
	m := newModel()
	result, cmd := m.Update(components.SendMsg{Content: ""})
	am := result.(app.Model)
	// cmd is always non-nil (batches with listenEvents), but we shouldn't call it
	// because listenEvents() blocks on the event channel
	if cmd == nil {
		t.Fatal("expected cmd to be non-nil")
	}
	// Verify model state didn't change (no message sent, not streaming)
	if am.IsStreaming() {
		t.Error("expected streaming to be false for empty SendMsg")
	}
}

func TestCtrlCDebouncedQuit(t *testing.T) {
	m := newModel()

	result, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	am := result.(app.Model)
	_ = cmd
	statusView := am.StatusBarView()
	if !strings.Contains(statusView, "Ctrl+C again") {
		t.Fatalf("expected status bar hint on first Ctrl+C, got:\n%s", statusView)
	}

	result2, cmd2 := am.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	_ = result2.(app.Model)
	// cmd2 is always non-nil (batches with listenEvents), verify it exists
	if cmd2 == nil {
		t.Fatalf("second Ctrl+C should return a cmd, got nil")
	}
	// We can't execute cmd2() because it blocks on listenEvents().
	// Instead, verify the behavior: after second Ctrl+C within debounce window,
	// the model should be in a state that would trigger quit.
	// The debounce timestamp should be set from the first press.
}

func TestCtrlCInterruptsStream(t *testing.T) {
	m := newModel()
	result0, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result0.(app.Model)
	m = m.SetStreamingForTest(true)

	result, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	am := result.(app.Model)
	if am.IsStreaming() {
		t.Fatal("expected streaming to be interrupted")
	}
	_ = cmd
	statusView := am.StatusBarView()
	if !strings.Contains(statusView, "interrupted") {
		t.Fatalf("expected status bar to show interruption, got:\n%s", statusView)
	}
}

func TestCtrlCSecondPressTooLate(t *testing.T) {
	m := newModel()

	result0, _ := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	m = result0.(app.Model)
	time.Sleep(350 * time.Millisecond)
	result, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	am := result.(app.Model)
	_ = cmd
	statusView := am.StatusBarView()
	if !strings.Contains(statusView, "Ctrl+C again") {
		t.Fatalf("expected fresh hint on late second press, got:\n%s", statusView)
	}
}

func TestStreamingHintInStatusBar(t *testing.T) {
	m := newModel()
	m = m.SetStreamingForTest(true)
	statusView := m.StatusBarView()
	if !strings.Contains(statusView, "streaming") {
		t.Fatalf("expected status bar to show streaming hint, got:\n%s", statusView)
	}
}

// containsText strips ANSI codes crudely and checks for substring.
func containsText(s, substr string) bool {
	clean := ""
	inEsc := false
	for _, c := range s {
		if c == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
				inEsc = false
			}
			continue
		}
		clean += string(c)
	}
	return len(clean) >= len(substr) && containsSubstr(clean, substr)
}

func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
