package components

import (
	"regexp"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/styles"
)

// AppendMessage adds a new message to the history.
func (c *Chat) AppendMessage(role, content string) {
	if role == "user" {
		c.userScrolled = false
	}
	c.messages = append(c.messages, chatMessage{role: role, content: content})
	c.refresh()
}

// AppendChunk appends streamed text to the last message.
func (c *Chat) AppendChunk(chunk string) {
	if len(c.messages) == 0 {
		return
	}
	c.messages[len(c.messages)-1].content += chunk
	c.refresh()
}

// refresh re-renders all messages into the viewport.
func (c *Chat) refresh() {
	var sb strings.Builder
	for _, msg := range c.messages {
		sb.WriteString(c.renderMessage(msg))
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

const messageGap = "\n\n"

func (c *Chat) renderMessage(msg chatMessage) string {
	if msg.role == "user" {
		innerWidth := c.width - 1
		if innerWidth < 1 {
			innerWidth = 1
		}

		padded := lipgloss.NewStyle().
			Background(styles.PanelBg).
			Foreground(styles.AssistantText).
			Padding(1, 2).
			Width(innerWidth).
			Render(msg.content)

		barHeight := lipgloss.Height(padded)
		bar := lipgloss.NewStyle().
			Background(styles.UserBorder).
			Width(1).
			Height(barHeight).
			Render("")

		return lipgloss.JoinHorizontal(lipgloss.Top, bar, padded) + messageGap
	}

	content := msg.content
	if c.renderer != nil {
		if out, err := c.renderer.Render(msg.content); err == nil {
			content = strings.TrimRight(out, "\n")
		}
	}
	content = stripInlineCodeBg(content)

	lineStyle := lipgloss.NewStyle().
		Background(styles.Background).
		Width(c.width)

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = lineStyle.Render(line)
	}
	return strings.Join(lines, "\n") + messageGap
}

const chatBgSGR = "\x1b[48;2;0;0;0m"

var inlineCodeBgRE = regexp.MustCompile(`(\x1b\[[0-9;]*?);?48;5;236;?([0-9;]*m)`)

func stripInlineCodeBg(s string) string {
	for {
		next := inlineCodeBgRE.ReplaceAllStringFunc(s, func(m string) string {
			sub := inlineCodeBgRE.FindStringSubmatch(m)
			prefix, suffix := sub[1], sub[2]
			if prefix == "\x1b[" && suffix == "m" {
				return "\x1b[m"
			}
			if prefix == "\x1b[" {
				return prefix + suffix
			}
			if suffix == "m" {
				return prefix + "m"
			}
			return prefix + ";" + suffix
		})
		if next == s {
			return s
		}
		s = next
	}
}
