package components

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/styles"
)

const statusBarVersion = "v0.1.0"

// StatusMsg is sent by any component (or the app layer) to push a
// transient message onto the status bar. Duration == 0 means the
// message persists until another message overwrites it or ClearMessage
// is called.
type StatusMsg struct {
	Content  string
	Duration time.Duration
}

// StatusBar is the single-line bar rendered at the bottom of the TUI.
//
// Layout: center | right
//   - Center: transient toast messages set via SetMessage().
//   - Right: default "ctrl+p · commands · version" hint.
type StatusBar struct {
	width        int
	message      string    // active center toast (empty = none)
	messageUntil time.Time // when the toast expires
}

func NewStatusBar() StatusBar {
	return StatusBar{
		width: 80,
	}
}

func (s StatusBar) Init() tea.Cmd { return nil }

// SetMessage pushes a transient toast message into the center of the
// status bar. The message replaces any previous toast and auto-clears
// after `duration`. Pass duration == 0 for a persistent toast.
func (s StatusBar) SetMessage(msg string, duration time.Duration) (StatusBar, tea.Cmd) {
	s.message = msg
	if duration > 0 {
		s.messageUntil = time.Now().Add(duration)
		return s, tea.Tick(duration, func(time.Time) tea.Msg {
			return statusBarTickMsg{}
		})
	}
	s.messageUntil = time.Time{} // zero = never expires
	return s, nil
}

// ClearMessage immediately drops the center toast.
func (s StatusBar) ClearMessage() StatusBar {
	s.message = ""
	s.messageUntil = time.Time{}
	return s
}

// statusBarTickMsg is sent by the tea.Tick command scheduled by
// SetMessage. It forces a re-render so the status bar can evaluate
// whether the transient message has expired.
type statusBarTickMsg struct{}

// isMessageActive reports whether the current center toast has not
// yet expired.
func (s StatusBar) isMessageActive() bool {
	if s.message == "" {
		return false
	}
	if s.messageUntil.IsZero() {
		return true // zero = persistent
	}
	return time.Now().Before(s.messageUntil)
}

func (s StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		s.width = msg.Width
	}
	// Swallow our own tick — the only purpose is to trigger a View()
	// re-render so isMessageActive() is re-evaluated.
	if _, ok := msg.(statusBarTickMsg); ok {
		return s, nil
	}
	return s, nil
}

func (s StatusBar) View() string {
	bg := lipgloss.NewStyle().Background(styles.ChatBackground)
	muted := lipgloss.NewStyle().Foreground(styles.SidebarValue).Background(styles.ChatBackground)
	white := lipgloss.NewStyle().Foreground(styles.AssistantText).Background(styles.ChatBackground).Bold(true)

	// Right-side default hint.
	right := lipgloss.JoinHorizontal(lipgloss.Left,
		white.Render("ctrl+p"),
		muted.Render("  commands  "),
		muted.Render(statusBarVersion),
	)

	innerWidth := s.width - 4
	if innerWidth < 1 {
		innerWidth = 1
	}

	if s.isMessageActive() {
		// Left-aligned message (no background).
		leftMsg := white.Render(s.message)
		leftWidth := lipgloss.Width(leftMsg)
		rightWidth := lipgloss.Width(right)

		gapWidth := innerWidth - leftWidth - rightWidth
		if gapWidth < 0 {
			gapWidth = 0
		}

		return bg.Width(s.width).Padding(1, 2).Render(
			leftMsg +
				lipgloss.NewStyle().Background(styles.ChatBackground).Width(gapWidth).Render("") +
				right,
		)
	}

	// No center toast — right-aligned default.
	rightWidth := lipgloss.Width(right)
	return bg.Width(s.width).Padding(1, 2).Render(
		lipgloss.NewStyle().Background(styles.ChatBackground).Width(innerWidth-rightWidth).Render("") + right,
	)
}
