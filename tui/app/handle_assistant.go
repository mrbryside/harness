package app

import (
	tea "charm.land/bubbletea/v2"
)

// AssistantChunkMsg carries streamed text from the agent runtime.
type AssistantChunkMsg struct {
	Content string
	Done    bool
}

// handleAssistantChunkMsg processes one chunk from the assistant stream.
func (m Model) handleAssistantChunkMsg(msg AssistantChunkMsg) (tea.Model, tea.Cmd) {
	if msg.Content != "" {
		m.chat.AppendChunk(msg.Content)
	}

	if msg.Done {
		m.streaming = false
		m.statusbar = m.statusbar.ClearMessage()
	}

	return m, nil
}
