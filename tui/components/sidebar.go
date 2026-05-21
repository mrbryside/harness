package components

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/styles"
)

const SidebarWidth = 40

// Sidebar displays model info, token count, and connection status.
type Sidebar struct {
	model  string
	tokens int
	width  int
	height int
}

func NewSidebar(modelName string) Sidebar {
	return Sidebar{
		model:  modelName,
		width:  SidebarWidth,
		height: 20,
	}
}

func (s *Sidebar) SetTokens(n int) {
	s.tokens = n
}

func (s Sidebar) Init() tea.Cmd { return nil }

func (s Sidebar) Update(msg tea.Msg) (Sidebar, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.height = msg.Height
		s.width = SidebarWidth
	}
	return s, nil
}

func (s Sidebar) View() string {
	// Each line is rendered full-width with PanelBg so there's no transparent
	// strip on the right of short labels like "Model" / "Tokens".
	innerWidth := SidebarWidth - 4 // account for Padding(1,2)
	lineStyle := lipgloss.NewStyle().Background(styles.Background).Width(innerWidth)

	label := func(s string) string {
		return lineStyle.Foreground(styles.SidebarLabel).Render(s)
	}
	value := func(s string) string {
		return lineStyle.Foreground(styles.SidebarValue).Render(s)
	}
	blank := lineStyle.Render("")

	formattedTokens := formatTokens(s.tokens)

	statusLine := lineStyle.Render(
		lipgloss.NewStyle().Foreground(styles.ConnectedDot).Background(styles.Background).Render("● ") +
			lipgloss.NewStyle().Foreground(styles.SidebarValue).Background(styles.Background).Render("Connected"),
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		label("Model"),
		value(s.model),
		blank,
		label("Tokens"),
		value(formattedTokens),
		blank,
		label("Cost"),
		value("$0.00"),
		blank,
		label("Status"),
		statusLine,
	)

	return lipgloss.NewStyle().
		Width(SidebarWidth).
		Background(styles.Background).
		Padding(1, 2).
		Render(content)
}

// formatTokens formats an integer with comma separators.
func formatTokens(n int) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}
	// insert commas
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}
