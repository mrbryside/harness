package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/components"
	"github.com/mrbryside/harness/tui/styles"
)

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.KeyboardEnhancements.ReportAllKeysAsEscapeCodes = true
	return v
}

// overlayAtBottom replaces the last N lines of base with overlay content.
// Used to draw autocomplete flush against the input without affecting layout.
func overlayAtBottom(base, overlay string, width int) string {
	if overlay == "" {
		return base
	}
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	n := len(overlayLines)
	start := len(baseLines) - n
	if start < 0 {
		start = 0
	}

	// Build result: keep top lines, append overlay lines.
	result := make([]string, 0, len(baseLines))
	result = append(result, baseLines[:start]...)
	result = append(result, overlayLines...)

	return strings.Join(result, "\n")
}

// render builds the full screen string.
func (m Model) render() string {
	statusView := m.statusbar.View()
	statusLines := lipgloss.Height(statusView)

	sidebarWidth := components.SidebarWidth
	chatWidth := m.width - outerMarginX - innerGap - sidebarWidth
	if chatWidth < 1 {
		chatWidth = 1
	}

	var bottomView string
	var bottomLines int
	if m.activeQuestion != nil && m.activeQuestion.question != nil && m.activeQuestion.question.Active() {
		bottomView = m.activeQuestion.question.OverlayView(chatWidth)
		bottomLines = lipgloss.Height(bottomView)
	} else {
		bottomView = m.input.View()
		bottomLines = lipgloss.Height(bottomView)
	}

	chatHeight := m.height - bottomLines - statusLines - outerMarginY - chatInputGap
	if chatHeight < 1 {
		chatHeight = 1
	}
	sidebarHeight := m.height

	chatBlock := lipgloss.Place(
		chatWidth, chatHeight,
		lipgloss.Left, lipgloss.Top,
		m.chat.View(),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(styles.ChatBackground)),
	)

	// Overlay autocomplete at the very bottom of the chat block.
	if m.autocomplete.Active() {
		autoView := m.autocomplete.View(chatWidth)
		chatBlock = overlayAtBottom(chatBlock, autoView, chatWidth)
	}

	// Overlay permission prompt replaces input area (handled above).

	sidebarBlock := lipgloss.Place(
		sidebarWidth, sidebarHeight,
		lipgloss.Left, lipgloss.Top,
		m.sidebar.View(),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(styles.Background)),
	)

	leftColContentHeight := chatHeight + bottomLines + outerMarginY + chatInputGap
	leftMargin := lipgloss.NewStyle().
		Width(outerMarginX).
		Height(leftColContentHeight).
		Background(styles.ChatBackground).
		Render("")

	topMarginLeft := lipgloss.NewStyle().
		Width(chatWidth).
		Height(outerMarginY).
		Background(styles.ChatBackground).
		Render("")

	chatInputSpacer := lipgloss.NewStyle().
		Width(chatWidth).
		Height(chatInputGap).
		Background(styles.ChatBackground).
		Render("")

	leftStack := lipgloss.JoinVertical(lipgloss.Left, topMarginLeft, chatBlock, chatInputSpacer, bottomView)

	leftCol := lipgloss.JoinHorizontal(lipgloss.Top, leftMargin, leftStack)

	gapCol := lipgloss.NewStyle().
		Width(innerGap).
		Height(leftColContentHeight).
		Background(styles.ChatBackground).
		Render("")

	leftWithGap := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, gapCol)
	leftFull := lipgloss.JoinVertical(lipgloss.Left, leftWithGap, statusView)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftFull, sidebarBlock)
}
