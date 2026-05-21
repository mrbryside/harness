package app

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/components"
	"github.com/mrbryside/harness/llm"
)

// chunkMsg carries a single streamed chunk from the LLM provider.
type chunkMsg struct {
	chunk llm.Chunk
}

// nextChunk reads one chunk from ch and dispatches it as a chunkMsg.
func nextChunk(ch <-chan llm.Chunk) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-ch
		if !ok {
			return chunkMsg{chunk: llm.Chunk{Done: true}}
		}
		return chunkMsg{chunk: chunk}
	}
}

// handleSendMsg starts a new LLM stream when the user presses Enter.
func (m Model) handleSendMsg(msg components.SendMsg) (tea.Model, tea.Cmd) {
	if msg.Content == "" || m.streaming {
		return m, nil
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.messages = append(m.messages, llm.Message{Role: "user", Content: msg.Content})
	m.input.AddHistory(msg.Content)
	m.chat.AppendMessage("user", msg.Content)
	m.chat.AppendMessage("assistant", "")
	m.streaming = true
	m.statusbar, _ = m.statusbar.SetMessage("⟳ streaming — press Ctrl+C to interrupt", 0)

	ch, err := m.provider.ChatCompletion(context.Background(), m.messages, llm.Options{})
	if err != nil {
		m.chat.AppendChunk("\n\n*Error: " + err.Error() + "*")
		m.streaming = false
		m.statusbar = m.statusbar.ClearMessage()
		m.statusbar, cmd = m.statusbar.SetMessage("✗ error: "+err.Error(), 3*time.Second)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}
	m.streamCh = ch
	cmds = append(cmds, nextChunk(ch))
	return m, tea.Batch(cmds...)
}

// handleChunkMsg processes one chunk from the active stream.
func (m Model) handleChunkMsg(msg chunkMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	c := msg.chunk

	if c.Err != nil {
		m.chat.AppendChunk("\n\n*Error: " + c.Err.Error() + "*")
		m.streaming = false
		m.streamCh = nil
		m.statusbar = m.statusbar.ClearMessage()
		return m, nil
	}

	m.chat.AppendChunk(c.Content)
	m.sidebar.SetTokens(c.TokensUsed)

	if c.Done {
		m.streaming = false
		m.streamCh = nil
		m.statusbar = m.statusbar.ClearMessage()
		return m, nil
	}

	cmds = append(cmds, nextChunk(m.streamCh))
	return m, tea.Batch(cmds...)
}

// handleStatusMsg pushes a transient message onto the status bar.
func (m Model) handleStatusMsg(msg components.StatusMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.statusbar, cmd = m.statusbar.SetMessage(msg.Content, msg.Duration)
	return m, cmd
}
