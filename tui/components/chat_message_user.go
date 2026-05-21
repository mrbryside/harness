package components

import (
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/styles"
)

func (c *Chat) renderUserMessage(msg chatMessage) string {
	padded := lipgloss.NewStyle().
		Background(styles.Background).
		Foreground(styles.AssistantText).
		Padding(1, 2).
		BorderStyle(lipgloss.Border{Left: "┃"}).
		BorderLeft(true).
		BorderForeground(styles.UserBorder).
		Width(c.width).
		Render(msg.content)

	return padded + messageGap
}
