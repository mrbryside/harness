package components

import (
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/styles"
)

func (i Input) View() string {
	innerWidth := i.width - 3 // 1 accent col + 2 padding cols
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
	textareaBody = i.sel.Overlay(textareaBody, 0, inputBgSGR)

	body := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render(textareaBody)

	spacer := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render("")

	inner := lipgloss.JoinVertical(lipgloss.Left, body, spacer, footer)

	padded := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Padding(1, 1).
		Render(inner)

	barHeight := lipgloss.Height(padded)
	if barHeight < 1 {
		barHeight = 1
	}

	bar := lipgloss.NewStyle().
		Background(styles.UserBorder).
		Width(1).
		Height(barHeight).
		Render("")

	return lipgloss.JoinHorizontal(lipgloss.Top, bar, padded)
}
