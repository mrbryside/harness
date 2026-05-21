package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/components"
	"github.com/mrbryside/harness/styles"
)

// Outer margin around the entire layout so chat / input / sidebar all share
// the same breathing room from the terminal edges.
const (
	outerMarginX = 2 // cols of Background on left & right
	outerMarginY = 1 // rows of Background on top
	innerGap     = 2 // cols between chat/input column and sidebar
	chatInputGap = 1 // rows between chat and input
)

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	// Declarative terminal-feature flags (v2): alt-screen + mouse capture
	// replace what used to be tea.WithAltScreen() / tea.WithMouseCellMotion()
	// program options.
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	// Ask the terminal to disambiguate keys via the Kitty keyboard protocol
	// so Shift+Enter arrives as a distinct event (with tea.ModShift set)
	// instead of being indistinguishable from plain Enter. Terminals that
	// don't support the protocol simply ignore the request.
	v.KeyboardEnhancements.ReportAllKeysAsEscapeCodes = true
	return v
}

// render builds the full screen string. Separated from View() so the rest
// of the app (and tests) can work with a plain string when convenient.
func (m Model) render() string {
	statusView := m.statusbar.View()
	statusLines := lipgloss.Height(statusView)

	// Sidebar is flush to the top/right/bottom edges (no outer margin on those sides).
	// Only the chat + input column gets the left/top outer margin and reserves
	// space for the status bar at the bottom.
	sidebarWidth := components.SidebarWidth
	chatWidth := m.width - outerMarginX - innerGap - sidebarWidth
	if chatWidth < 1 {
		chatWidth = 1
	}

	inputView := m.input.View()
	inputLines := lipgloss.Height(inputView)

	chatHeight := m.height - inputLines - statusLines - outerMarginY - chatInputGap
	if chatHeight < 1 {
		chatHeight = 1
	}
	// Sidebar fills the entire terminal height — no top or bottom margin.
	sidebarHeight := m.height

	chatBlock := lipgloss.Place(
		chatWidth, chatHeight,
		lipgloss.Left, lipgloss.Top,
		m.chat.View(),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(styles.Background)),
	)

	sidebarBlock := lipgloss.Place(
		sidebarWidth, sidebarHeight,
		lipgloss.Left, lipgloss.Top,
		m.sidebar.View(),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(styles.PanelBg)),
	)

	// Left column = top margin + chat + gap + input + status bar.
	// Its height equals sidebarHeight so the horizontal join lines up cleanly.
	leftColContentHeight := chatHeight + inputLines + outerMarginY + chatInputGap
	leftMargin := lipgloss.NewStyle().
		Width(outerMarginX).
		Height(leftColContentHeight).
		Background(styles.Background).
		Render("")

	topMarginLeft := lipgloss.NewStyle().
		Width(chatWidth).
		Height(outerMarginY).
		Background(styles.Background).
		Render("")

	chatInputSpacer := lipgloss.NewStyle().
		Width(chatWidth).
		Height(chatInputGap).
		Background(styles.Background).
		Render("")

	leftStack := lipgloss.JoinVertical(lipgloss.Left, topMarginLeft, chatBlock, chatInputSpacer, inputView)
	leftCol := lipgloss.JoinHorizontal(lipgloss.Top, leftMargin, leftStack)

	// Gap column between left stack and sidebar (chat-colored).
	gapCol := lipgloss.NewStyle().
		Width(innerGap).
		Height(leftColContentHeight).
		Background(styles.Background).
		Render("")

	leftWithGap := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, gapCol)

	// Stack the status bar under the left+gap column so its total height
	// matches the sidebar (which spans the full terminal height).
	leftFull := lipgloss.JoinVertical(lipgloss.Left, leftWithGap, statusView)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftFull, sidebarBlock)
}
