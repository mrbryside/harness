package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/styles"
)

const statusBarVersion = "v0.1.0"

// StatusBar is the single-line bar rendered at the bottom of the TUI.
//
// The bar shows only the global "ctrl+p · commands · version" hint on the
// right. Per-session info (provider, model) lives in the sidebar, not here.
type StatusBar struct {
	width int
}

func NewStatusBar() StatusBar {
	return StatusBar{
		width: 80,
	}
}

func (s StatusBar) Init() tea.Cmd { return nil }

func (s StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		s.width = msg.Width
	}
	return s, nil
}

func (s StatusBar) View() string {
	bg := lipgloss.NewStyle().Background(styles.Background)
	accent := lipgloss.NewStyle().Foreground(styles.StatusBarAccent).Background(styles.Background).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.SidebarValue).Background(styles.Background)

	right := lipgloss.JoinHorizontal(lipgloss.Left,
		accent.Render("ctrl+p"),
		muted.Render("  commands  "),
		muted.Render(statusBarVersion),
	)

	// Match the input's content alignment: input has a 1-col blue accent bar
	// + 1 col of left padding, so its text starts at column 2. The status bar
	// has no accent bar, so we use 2 cols of left/right padding to align.
	innerWidth := s.width - 4
	if innerWidth < 1 {
		innerWidth = 1
	}
	rightWidth := lipgloss.Width(right)
	gap := innerWidth - rightWidth
	if gap < 1 {
		gap = 1
	}

	return bg.Width(s.width).Padding(1, 2).Render(
		lipgloss.NewStyle().Background(styles.Background).Width(gap).Render("") + right,
	)
}
