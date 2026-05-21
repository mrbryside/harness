package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/styles"
)

const maxVisibleSuggestions = 10

// Suggestion is a single autocomplete entry.
type Suggestion struct {
	Command     string
	Description string
}

// SuggestionProvider is a function that returns all available suggestions.
type SuggestionProvider func() []Suggestion

// Autocomplete is a slash-command suggestion popup.
type Autocomplete struct {
	active       bool
	filtered     []Suggestion
	selectedIdx  int
	scrollOffset int
	prefix       string
	provider     SuggestionProvider
}

// staticSuggestions is the hard-coded list of available commands.
var staticSuggestions = []Suggestion{
	{Command: "/help", Description: "Show available commands"},
}

// NewAutocomplete creates a hidden autocomplete component.
func NewAutocomplete() Autocomplete {
	return Autocomplete{
		provider: func() []Suggestion { return staticSuggestions },
	}
}

// SetProvider sets the suggestion provider for dynamic command lists.
func (a *Autocomplete) SetProvider(provider SuggestionProvider) {
	a.provider = provider
}

// Active reports whether the popup is visible.
func (a Autocomplete) Active() bool { return a.active }

// Show activates the popup and filters by the given prefix.
func (a *Autocomplete) Show(prefix string) {
	a.active = true
	a.prefix = prefix
	a.filter()
}

// Hide deactivates the popup.
func (a *Autocomplete) Hide() {
	a.active = false
	a.filtered = nil
	a.selectedIdx = 0
	a.scrollOffset = 0
}

// Next moves the selection down (wraps around).
func (a *Autocomplete) Next() {
	if len(a.filtered) == 0 {
		return
	}
	a.selectedIdx++
	if a.selectedIdx >= len(a.filtered) {
		a.selectedIdx = 0
		a.scrollOffset = 0
	} else if a.selectedIdx >= a.scrollOffset+maxVisibleSuggestions {
		a.scrollOffset = a.selectedIdx - maxVisibleSuggestions + 1
	}
}

// Prev moves the selection up (wraps around).
func (a *Autocomplete) Prev() {
	if len(a.filtered) == 0 {
		return
	}
	a.selectedIdx--
	if a.selectedIdx < 0 {
		a.selectedIdx = len(a.filtered) - 1
		if len(a.filtered) > maxVisibleSuggestions {
			a.scrollOffset = len(a.filtered) - maxVisibleSuggestions
		}
	} else if a.selectedIdx < a.scrollOffset {
		a.scrollOffset = a.selectedIdx
	}
}

// SelectedIndex returns the current selection index.
func (a Autocomplete) SelectedIndex() int { return a.selectedIdx }

// Selected returns the currently highlighted suggestion.
func (a Autocomplete) Selected() (command, description string) {
	if a.selectedIdx >= 0 && a.selectedIdx < len(a.filtered) {
		s := a.filtered[a.selectedIdx]
		return s.Command, s.Description
	}
	return "", ""
}

// filter rebuilds the filtered list from the prefix.
func (a *Autocomplete) filter() {
	a.filtered = nil
	a.selectedIdx = 0
	a.scrollOffset = 0
	lowerPrefix := strings.ToLower(a.prefix)
	all := a.provider()
	for _, s := range all {
		cmdLower := strings.ToLower(s.Command)
		if strings.HasPrefix(cmdLower, lowerPrefix) || strings.HasPrefix(strings.TrimPrefix(cmdLower, "/"), lowerPrefix) {
			a.filtered = append(a.filtered, s)
		}
	}
}

// renderRow builds one row with a consistent background that spans the full width.
func renderRow(cmd, desc string, cmdWidth int, totalWidth int, selected bool) string {
	gap := strings.Repeat(" ", cmdWidth-lipgloss.Width(cmd))

	if selected {
		bg := styles.Peach
		fg := styles.Background
		cmdStyled := lipgloss.NewStyle().Background(bg).Foreground(fg).Bold(true).Render(cmd)
		gapStyled := lipgloss.NewStyle().Background(bg).Render(gap)
		descStyled := lipgloss.NewStyle().Background(bg).Foreground(fg).Render(desc)
		row := lipgloss.JoinHorizontal(lipgloss.Left, cmdStyled, gapStyled, descStyled)
		// Pad to full width with the same background.
		rowW := lipgloss.Width(row)
		if rowW < totalWidth {
			pad := lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", totalWidth-rowW))
			row = lipgloss.JoinHorizontal(lipgloss.Left, row, pad)
		}
		return row
	}

	bg := styles.PanelBg
	cmdStyled := lipgloss.NewStyle().Background(bg).Foreground(styles.AssistantText).Bold(true).Render(cmd)
	gapStyled := lipgloss.NewStyle().Background(bg).Render(gap)
	descStyled := lipgloss.NewStyle().Background(bg).Foreground(styles.SidebarValue).Render(desc)
	row := lipgloss.JoinHorizontal(lipgloss.Left, cmdStyled, gapStyled, descStyled)
	// Pad to full width with the same background.
	rowW := lipgloss.Width(row)
	if rowW < totalWidth {
		pad := lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", totalWidth-rowW))
		row = lipgloss.JoinHorizontal(lipgloss.Left, row, pad)
	}
	return row
}

// View renders the popup. Each row spans the full width with a solid
// background. Command is bright, description is muted.
func (a Autocomplete) View(width int) string {
	if !a.active || len(a.filtered) == 0 {
		return ""
	}

	// Find max command width for tabular alignment.
	maxCmdWidth := 0
	end := a.scrollOffset + maxVisibleSuggestions
	if end > len(a.filtered) {
		end = len(a.filtered)
	}
	for i := a.scrollOffset; i < end; i++ {
		if w := lipgloss.Width(a.filtered[i].Command); w > maxCmdWidth {
			maxCmdWidth = w
		}
	}
	cmdColumnWidth := maxCmdWidth + 4

	var lines []string
	for i := a.scrollOffset; i < end; i++ {
		s := a.filtered[i]
		lines = append(lines, renderRow(s.Command, s.Description, cmdColumnWidth, width, i == a.selectedIdx))
	}

	return strings.Join(lines, "\n")
}
