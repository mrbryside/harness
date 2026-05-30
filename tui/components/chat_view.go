package components

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/styles"
)

func (c Chat) View() string {
	out := c.viewport.View()
	out = strings.ReplaceAll(out, "\x1b[m", "\x1b[m"+styles.ChatBgSGR)
	out = strings.ReplaceAll(out, "\x1b[0m", "\x1b[0m"+styles.ChatBgSGR)
	// Selection stores content-relative coordinates (shifted by scrolls),
	// so we pass the viewport YOffset so Overlay can map back to visible
	// lines in viewport.View().
	out = c.sel.Overlay(out, c.viewport.YOffset(), styles.ChatBgSGR, c.width)

	// Render absolute toast overlay at top-right of the chat area.
	if c.toast != "" && time.Now().Before(c.toastUntil) {
		toastStyle := lipgloss.NewStyle().
			Background(styles.PanelBg).
			Foreground(styles.AssistantText).
			Bold(true).
			Padding(1, 1).
			BorderStyle(lipgloss.Border{
				Left:  "┃",
				Right: "┃",
			}).
			BorderLeft(true).
			BorderRight(true).
			BorderForeground(styles.UserBorder)
		toastStr := toastStyle.Render(c.toast)
		toastW := lipgloss.Width(toastStr)

		chatLayer := lipgloss.NewLayer(out)
		toastLayer := lipgloss.NewLayer(toastStr).
			X(max(0, c.width-toastW)).
			Y(0).
			Z(10)

		comp := lipgloss.NewCompositor(chatLayer, toastLayer)
		return comp.Render()
	}

	return out
}
