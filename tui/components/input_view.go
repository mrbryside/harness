package components

import (
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/styles"
)

func (i Input) View() string {
	innerWidth := i.width - 5 // 1 accent col + 4 padding cols (2 left + 2 right)
	if innerWidth < 1 {
		innerWidth = 1
	}

	modeColor := styles.ModeBuildColor
	if i.mode == ModePlan {
		modeColor = styles.ModePlanColor
	}
	modeStyle := lipgloss.NewStyle().Foreground(modeColor).Background(styles.PanelBg).Bold(true)
	dim := lipgloss.NewStyle().Foreground(styles.SidebarLabel).Background(styles.PanelBg)

	footerInner := modeStyle.Render(string(i.mode)) + dim.Render(" · "+i.model)
	footer := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render(footerInner)

	// Selection overlay applied before lipgloss width wrapper.
	textareaBody := i.textarea.View()
	textareaBody = i.sel.Overlay(textareaBody, 0, styles.PanelBgSGR)

	body := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render(textareaBody)

	spacer := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render("")

	inner := lipgloss.JoinVertical(lipgloss.Left, body, spacer, footer)

	return lipgloss.NewStyle().
		Background(styles.PanelBg).
		Padding(1, 2).
		BorderStyle(lipgloss.Border{Left: "┃"}).
		BorderLeft(true).
		BorderForeground(styles.UserBorder).
		Render(inner)
}
