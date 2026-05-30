package components

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

const messageGap = "\n\n"

// ToolEdit represents a code change with line numbers.
type ToolEdit struct {
	Path       string // file path
	OldContent string // old code
	NewContent string // new code
	StartLine  int    // 1-based line number where newContent starts in the file
}

// AppendMessage adds a new message to the history.
func (c *Chat) AppendMessage(role, content string) {
	if role == "user" {
		c.userScrolled = false
	}
	c.messages = append(c.messages, chatMessage{role: role, content: content})
	c.refresh()
}

// AppendToolEdit adds a tool-edit message with explicit line numbers.
// EndLine is computed automatically from NewContent (StartLine + newline count).
func (c *Chat) AppendToolEdit(edit ToolEdit) {
	c.messages = append(c.messages, chatMessage{
		role:          "tool_edit",
		toolEditPath:  edit.Path,
		toolEditOld:   edit.OldContent,
		toolEditNew:   edit.NewContent,
		toolEditStart: edit.StartLine,
	})
	c.refresh()
}

// AppendChunk appends streamed text to the last message.
// Optimised: only re-renders the last assistant message; earlier messages
// use their cached rendered string.
func (c *Chat) AppendChunk(chunk string) {
	if len(c.messages) == 0 {
		return
	}
	lastIdx := len(c.messages) - 1
	c.messages[lastIdx].content += chunk
	c.messages[lastIdx].rendered = "" // invalidate cache

	var sb strings.Builder
	for i := 0; i < lastIdx; i++ {
		if c.messages[i].rendered == "" {
			c.messages[i].rendered = c.renderMessage(c.messages[i])
		}
		sb.WriteString(c.messages[i].rendered)
	}
	sb.WriteString(c.renderMessage(c.messages[lastIdx]))
	c.viewport.SetContent(sb.String())
	if !c.userScrolled {
		c.viewport.GotoBottom()
	}
}

// refresh re-renders all messages into the viewport.
func (c *Chat) refresh() {
	var sb strings.Builder
	for i := range c.messages {
		c.messages[i].rendered = c.renderMessage(c.messages[i])
		sb.WriteString(c.messages[i].rendered)
	}
	c.viewport.SetContent(sb.String())
	if !c.userScrolled {
		c.viewport.GotoBottom()
	}
}

// ShowToast displays a transient notification at the top-right of the chat
// area. The message auto-clears after the given duration. Returns a Cmd
// that schedules a re-render so the toast actually disappears.
func (c *Chat) ShowToast(msg string, d time.Duration) tea.Cmd {
	c.toast = msg
	c.toastUntil = time.Now().Add(d)
	return tea.Tick(d, func(time.Time) tea.Msg {
		return chatToastTickMsg{}
	})
}

func (c *Chat) renderMessage(msg chatMessage) string {
	switch msg.role {
	case "user":
		return c.renderUserMessage(msg)
	case "tool_edit":
		return c.renderToolEditMessage(msg)
	default:
		return c.renderAssistantMessage(msg)
	}
}
