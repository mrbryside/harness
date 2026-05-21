package components

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/styles"
)

func (c Chat) View() string {
	out := c.viewport.View()
	out = strings.ReplaceAll(out, "\x1b[m", "\x1b[m"+chatBgSGR)
	out = strings.ReplaceAll(out, "\x1b[0m", "\x1b[0m"+chatBgSGR)
	out = c.sel.Overlay(out, c.viewport.YOffset(), chatBgSGR)

	// Render absolute toast overlay at top-right of the chat area.
	if c.toast != "" && time.Now().Before(c.toastUntil) {
		toastStyle := lipgloss.NewStyle().
			Background(styles.PanelBg).
			Foreground(styles.AssistantText).
			Bold(true).
			Padding(0, 1)
		toastStr := toastStyle.Render(c.toast)
		toastW := lipgloss.Width(toastStr)

		autocomplete := lipgloss.NewStyle().
			Background(styles.PanelBg).
			Foreground(styles.AssistantText).
			Width(20).
			Height(20).
			Bold(true).
			Padding(0, 1)
		autocompleteRender := autocomplete.Render(out)

		chatLayer := lipgloss.NewLayer(out)
		toastLayer := lipgloss.NewLayer(toastStr).
			X(max(0, c.width-toastW)).
			Y(0).
			Z(10)

		autocompleteLayer := lipgloss.NewLayer(autocompleteRender).
			X(0).
			Y(c.height).
			Z(12)

		comp := lipgloss.NewCompositor(chatLayer, toastLayer, autocompleteLayer)
		return comp.Render()
	}

	return out
}
