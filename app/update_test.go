package app_test

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/app"
	"github.com/mrbryside/harness/components"
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

func TestUpdateQuitOnCtrlC(t *testing.T) {
	m := newModel()
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("expected a quit cmd on Ctrl+C, got nil")
	}
	msg := cmd()
	if msg != tea.Quit() {
		t.Errorf("expected tea.Quit msg, got %T", msg)
	}
}

func TestUpdateQuitOnEsc(t *testing.T) {
	m := newModel()
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected a quit cmd on Esc, got nil")
	}
	msg := cmd()
	if msg != tea.Quit() {
		t.Errorf("expected tea.Quit msg, got %T", msg)
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
