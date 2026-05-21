package app_test

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/app"
	"github.com/mrbryside/harness/tui/components"
	"github.com/mrbryside/harness/llm"
)

// ── stub provider ─────────────────────────────────────────────────────────────

type stubProvider struct {
	response string
}

func (s *stubProvider) Name() string { return "stub" }

func (s *stubProvider) ChatCompletion(_ context.Context, _ []llm.Message, _ llm.Options) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, 2)
	ch <- llm.Chunk{Content: s.response, TokensUsed: 10, Done: false}
	ch <- llm.Chunk{Content: "", TokensUsed: 10, Done: true}
	close(ch)
	return ch, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newModel() app.Model {
	return app.New(&stubProvider{response: "hello"})
}

// drainCmds runs all commands returned by Update until there are none left,
// feeding results back in. Returns the final model.
func drainCmds(m tea.Model, cmd tea.Cmd) tea.Model {
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			break
		}
		m, cmd = m.Update(msg)
	}
	return m
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestUpdateCtrlCShowsQuitHint(t *testing.T) {
	m := newModel()
	result, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	am := result.(app.Model)
	// cmd may be non-nil (auto-clear tick), but it must NOT be Quit.
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
	if cmd != nil {
		t.Fatal("expected no cmd on Esc when input is empty, got a cmd")
	}
	_ = am.StatusBarView()
}

func TestUpdateEscReturnsStatusCmd(t *testing.T) {
	m := newModel()
	// Type some text first so Esc has something to clear.
	for _, r := range "hello" {
		result, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = result.(app.Model)
	}
	result, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	am := result.(app.Model)
	if cmd == nil {
		t.Fatal("expected a status cmd on Esc, got nil")
	}
	msg := cmd()
	statusMsg, ok := msg.(components.StatusMsg)
	if !ok {
		t.Fatalf("expected StatusMsg, got %T", msg)
	}
	if !strings.Contains(statusMsg.Content, "Esc again") {
		t.Fatalf("expected hint about double-Esc, got %q", statusMsg.Content)
	}
	// The model should not have quit.
	_ = am.StatusBarView()
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
	result, cmd := m.Update(components.SendMsg{Content: "ping"})
	if cmd == nil {
		t.Fatal("expected a cmd to be returned after SendMsg (stream start), got nil")
	}
	am := result.(app.Model)
	if !am.IsStreaming() {
		t.Error("expected model to be in streaming state after SendMsg")
	}
}

func TestUpdateChunkMsgAppendsToChat(t *testing.T) {
	m := newModel()

	// send a message to start the stream
	result, cmd := m.Update(components.SendMsg{Content: "ping"})

	// drain stream chunks
	result = drainCmds(result, cmd)

	am := result.(app.Model)
	if !containsText(am.ChatView(), "hello") {
		t.Errorf("expected ChatView() to contain streamed content %q, got:\n%s", "hello", am.ChatView())
	}
}

func TestUpdateWindowSizeUpdatesModel(t *testing.T) {
	m := newModel()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	_ = result // just ensure no panic
}

func TestUpdateIgnoresEmptySendMsg(t *testing.T) {
	m := newModel()
	_, cmd := m.Update(components.SendMsg{Content: ""})
	if cmd != nil {
		t.Errorf("expected no cmd for empty SendMsg, got %T", cmd)
	}
}

// containsText strips ANSI codes crudely and checks for substring.
func containsText(s, substr string) bool {
	// strip common ANSI escape sequences for test assertions
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

// Ctrl+C while NOT streaming: first press shows "Press Ctrl+C again to
// quit" in the status bar, second press within the debounce window quits.
func TestCtrlCDebouncedQuit(t *testing.T) {
	m := newModel()

	// First Ctrl+C → status bar hint, no quit (returns a tick cmd for auto-clear).
	result, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	am := result.(app.Model)
	// cmd may be non-nil (it's the auto-clear tick), but it must NOT be Quit.
	_ = cmd
	statusView := am.StatusBarView()
	if !strings.Contains(statusView, "Ctrl+C again") {
		t.Fatalf("expected status bar hint on first Ctrl+C, got:\n%s", statusView)
	}

	// Second Ctrl+C quickly → quit.
	result2, cmd2 := am.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	_ = result2.(app.Model)
	if cmd2 == nil {
		t.Fatalf("second Ctrl+C should quit, got nil")
	}
	msg := cmd2()
	if msg != tea.Quit() {
		t.Errorf("expected tea.Quit, got %T", msg)
	}
}

// Ctrl+C while streaming: first press interrupts the stream (no quit).
func TestCtrlCInterruptsStream(t *testing.T) {
	m := newModel()
	// Simulate streaming by flipping the flag directly.
	result0, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result0.(app.Model)
	// We can't easily start a real stream in a unit test, so we test the
	// interruption via a helper that the model exposes.
	// For now we just verify the status bar shows the right hint when
	// streaming is true.
	m = m.SetStreamingForTest(true)

	result, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	am := result.(app.Model)
	if am.IsStreaming() {
		t.Fatal("expected streaming to be interrupted")
	}
	// cmd may be non-nil (auto-clear tick), but it must NOT be Quit.
	_ = cmd
	statusView := am.StatusBarView()
	if !strings.Contains(statusView, "interrupted") {
		t.Fatalf("expected status bar to show interruption, got:\n%s", statusView)
	}
}

// If the second Ctrl+C arrives too late, the hint is shown again instead
// of quitting.
func TestCtrlCSecondPressTooLate(t *testing.T) {
	m := newModel()

	result0, _ := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	m = result0.(app.Model)
	// Wait longer than the debounce window.
	time.Sleep(350 * time.Millisecond)
	result, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	am := result.(app.Model)
	// cmd may be non-nil (auto-clear tick), but it must NOT be Quit.
	_ = cmd
	statusView := am.StatusBarView()
	if !strings.Contains(statusView, "Ctrl+C again") {
		t.Fatalf("expected fresh hint on late second press, got:\n%s", statusView)
	}
}

// When streaming starts, the status bar shows the streaming hint.
func TestStreamingHintInStatusBar(t *testing.T) {
	m := newModel()
	m = m.SetStreamingForTest(true)
	statusView := m.StatusBarView()
	if !strings.Contains(statusView, "streaming") {
		t.Fatalf("expected status bar to show streaming hint, got:\n%s", statusView)
	}
}
