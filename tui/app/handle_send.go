package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/llm"
	"github.com/mrbryside/harness/tui/components"
	"github.com/google/uuid"
)

// handleSendMsg emits a user_messaged event when the user presses Enter.
func (m Model) handleSendMsg(msg components.SendMsg) (tea.Model, tea.Cmd) {
	if msg.Content == "" || m.streaming {
		return m, nil
	}

	var cmds []tea.Cmd

	m.messages = append(m.messages, llm.Message{Role: "user", Content: msg.Content})
	m.input.AddHistory(msg.Content)
	m.chat.AppendMessage("user", msg.Content)
	m.chat.AppendMessage("assistant", "")
	m.streaming = true
	m.statusbar, _ = m.statusbar.SetMessage("⟳ streaming — press Ctrl+C to interrupt", 0)

	msgID := uuid.New().String()
	m.currentRequestID = msgID
	m.eventBus.Emit(eventbus.EventUserMessaged, eventbus.UserMessageEvent{
		ID:      msgID,
		Content: msg.Content,
	})

	return m, tea.Batch(cmds...)
}
