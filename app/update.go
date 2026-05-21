package app

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/components"
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// ── window resize ────────────────────────────────────────────────────────
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Mirror the layout math in app/view.go so children get correct sizes.
		// Sidebar is flush to the right edge — only the left margin is subtracted.
		chatWidth := msg.Width - outerMarginX - innerGap - components.SidebarWidth
		if chatWidth < 1 {
			chatWidth = 1
		}

		// give input chatWidth so it doesn't bleed into sidebar area
		m.input, _ = m.input.Update(tea.WindowSizeMsg{Width: chatWidth, Height: msg.Height})
		// status bar only spans the left+gap area (sidebar is full-height on the right)
		statusBarWidth := outerMarginX + chatWidth + innerGap
		m.statusbar, _ = m.statusbar.Update(tea.WindowSizeMsg{Width: statusBarWidth, Height: msg.Height})
		m.sidebar, _ = m.sidebar.Update(msg)

		m = m.reflowChat()
		return m, nil

	// ── keyboard ─────────────────────────────────────────────────────────────
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "up", "down", "pgup", "pgdown":
			var cmd tea.Cmd
			m.chat, cmd = m.chat.Update(msg)
			return m, cmd
		}

		// everything else goes to the input
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
		// input may have grown/shrunk — reflow chat to match
		m = m.reflowChat()

	// ── mouse wheel (scroll over chat) ───────────────────────────────────────
	case tea.MouseWheelMsg:
		var cmd tea.Cmd
		m.chat, cmd = m.chat.Update(msg)
		return m, cmd

	// ── bracketed paste (incl. drag-and-drop file paths) ─────────────────────
	// Most terminals send dropped file paths as a bracketed-paste sequence,
	// which Bubble Tea decodes to tea.PasteMsg. Forward it to the input so
	// the textarea inserts the path at the cursor.
	case tea.PasteMsg:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m = m.reflowChat()
		return m, cmd

	// ── user pressed Enter in input ──────────────────────────────────────────
	case components.SendMsg:
		if msg.Content == "" || m.streaming {
			return m, nil
		}

		m.messages = append(m.messages, llm.Message{Role: "user", Content: msg.Content})
		m.chat.AppendMessage("user", msg.Content)
		m.chat.AppendMessage("assistant", "")
		m.streaming = true

		ch, err := m.provider.ChatCompletion(context.Background(), m.messages, llm.Options{})
		if err != nil {
			m.chat.AppendChunk("\n\n*Error: " + err.Error() + "*")
			m.streaming = false
			return m, nil
		}
		m.streamCh = ch
		cmds = append(cmds, nextChunk(ch))

	// ── streaming chunk arrived ───────────────────────────────────────────────
	case chunkMsg:
		c := msg.chunk
		if c.Err != nil {
			m.chat.AppendChunk("\n\n*Error: " + c.Err.Error() + "*")
			m.streaming = false
			m.streamCh = nil
			return m, nil
		}

		m.chat.AppendChunk(c.Content)
		m.sidebar.SetTokens(c.TokensUsed)

		if c.Done {
			m.streaming = false
			m.streamCh = nil
			return m, nil
		}

		// schedule next read from the same channel
		cmds = append(cmds, nextChunk(m.streamCh))
	}

	return m, tea.Batch(cmds...)
}

// reflowChat recomputes chat width/height based on current window size and
// the current rendered height of the input. Call this whenever the input
// may have grown or shrunk so the chat viewport stays in sync with the
// layout produced by app/view.go.
func (m Model) reflowChat() Model {
	if m.width == 0 || m.height == 0 {
		return m
	}
	chatWidth := m.width - outerMarginX - innerGap - components.SidebarWidth
	if chatWidth < 1 {
		chatWidth = 1
	}
	inputLines := lipgloss.Height(m.input.View())
	statusLines := lipgloss.Height(m.statusbar.View())
	chatHeight := m.height - inputLines - statusLines - outerMarginY - chatInputGap
	if chatHeight < 1 {
		chatHeight = 1
	}
	m.chat, _ = m.chat.Update(tea.WindowSizeMsg{Width: chatWidth, Height: chatHeight})
	return m
}
