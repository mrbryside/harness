package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/styles"
)

type PermissionAnswerMsg struct {
	QuestionID string
	Answer     bool
}

type PermissionPrompt struct {
	active        bool
	question      string
	questionID    string
	selectedIndex int // 0=Yes, 1=No
}

func NewPermissionPrompt() PermissionPrompt {
	return PermissionPrompt{
		selectedIndex: 0,
	}
}

func (p *PermissionPrompt) Show(question, questionID string) {
	p.active = true
	p.question = question
	p.questionID = questionID
	p.selectedIndex = 0
}

func (p *PermissionPrompt) Hide() {
	p.active = false
	p.question = ""
	p.questionID = ""
	p.selectedIndex = 0
}

func (p PermissionPrompt) Active() bool {
	return p.active
}

func (p PermissionPrompt) View(width int) string {
	if !p.active {
		return ""
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(styles.ChatBackground).
		Background(styles.UserBorder).
		Bold(true)
	unselectedStyle := lipgloss.NewStyle().
		Bold(true)

	var yesLabel, noLabel string
	if p.selectedIndex == 0 {
		yesLabel = selectedStyle.Render("  Yes  ")
		noLabel = unselectedStyle.Foreground(styles.AccentOrange).Render("  No  ")
	} else {
		yesLabel = unselectedStyle.Foreground(styles.ModeBuildColor).Render("  Yes  ")
		noLabel = selectedStyle.Render("  No  ")
	}

	buttons := yesLabel + "\n" + noLabel

	boxWidth := width - 8
	if boxWidth < 20 {
		boxWidth = 20
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.UserBorder).
		Background(styles.PanelBg).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Center)

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.StatusBarAccent).
		Bold(true)

	content := titleStyle.Render("PERMISSION REQUIRED") + "\n\n" +
		p.question + "\n\n" +
		buttons

	return borderStyle.Render(content)
}

func (p *PermissionPrompt) HandleKey(msg tea.Msg) (tea.Cmd, bool) {
	if !p.active {
		return nil, false
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyUp, tea.KeyDown:
			p.toggleSelection()
			return nil, true
		case tea.KeyEnter, ' ':
			return p.emitAnswer(p.selectedIndex == 0), true
		case tea.KeyEscape:
			return p.emitAnswer(false), true
		}
		// Block all other keys from reaching input
		return nil, true
	}
	return nil, false
}

func (p *PermissionPrompt) Cancel() tea.Cmd {
	if !p.active {
		return nil
	}
	return p.emitAnswer(false)
}

func (p *PermissionPrompt) emitAnswer(answer bool) tea.Cmd {
	qid := p.questionID
	return func() tea.Msg {
		return PermissionAnswerMsg{
			QuestionID: qid,
			Answer:     answer,
		}
	}
}

func (p *PermissionPrompt) toggleSelection() {
	if p.selectedIndex == 0 {
		p.selectedIndex = 1
	} else {
		p.selectedIndex = 0
	}
}

func (p PermissionPrompt) Overlay(chatView string, width int) string {
	if !p.active {
		return chatView
	}

	promptView := p.View(width)
	promptLines := strings.Split(promptView, "\n")
	promptHeight := len(promptLines)

	chatLines := strings.Split(chatView, "\n")
	chatHeight := len(chatLines)

	startY := chatHeight - promptHeight
	if startY < 0 {
		startY = 0
	}

	var result []string
	for i := 0; i < chatHeight; i++ {
		if i >= startY && i < startY+promptHeight {
			promptLineIdx := i - startY
			if promptLineIdx < len(promptLines) {
				result = append(result, promptLines[promptLineIdx])
			} else {
				result = append(result, chatLines[i])
			}
		} else {
			result = append(result, chatLines[i])
		}
	}

	return strings.Join(result, "\n")
}

func (p PermissionPrompt) String() string {
	if !p.active {
		return "PermissionPrompt{inactive}"
	}
	return fmt.Sprintf("PermissionPrompt{question=%q, id=%q, selected=%d}",
		p.question, p.questionID, p.selectedIndex)
}

// Height returns the number of lines the prompt occupies.
func (p PermissionPrompt) Height(width int) int {
	if !p.active {
		return 0
	}
	view := p.OverlayView(width)
	return strings.Count(view, "\n") + 1
}

// OverlayView renders the prompt with the same layout as the input component.
// Uses lipgloss Width() for word wrapping so height is calculated correctly.
func (p PermissionPrompt) OverlayView(width int) string {
	if !p.active {
		return ""
	}

	innerWidth := width - 5 // 1 accent col + 4 padding cols (2 left + 2 right)
	if innerWidth < 1 {
		innerWidth = 1
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(styles.UserBorder).
		Bold(true)
	unselectedStyle := lipgloss.NewStyle().
		Foreground(styles.AccentOrange).
		Bold(true)

	var yesLabel, noLabel string
	if p.selectedIndex == 0 {
		yesLabel = selectedStyle.Render("Yes")
		noLabel = unselectedStyle.Render("No")
	} else {
		yesLabel = unselectedStyle.Render("Yes")
		noLabel = selectedStyle.Render("No")
	}

	// lipgloss Width() wraps text correctly — height will be accurate
	questionRendered := lipgloss.NewStyle().
		Foreground(styles.AssistantText).
		Background(styles.PanelBg).
		Width(innerWidth).
		Render(p.question)

	yesRow := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render("  " + yesLabel)

	noRow := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render("  " + noLabel)

	spacer := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render("")

	inner := lipgloss.JoinVertical(lipgloss.Left, questionRendered, spacer, yesRow, noRow)

	return lipgloss.NewStyle().
		Background(styles.PanelBg).
		Padding(1, 2).
		BorderStyle(lipgloss.Border{Left: "┃"}).
		BorderLeft(true).
		BorderForeground(styles.UserBorder).
		Render(inner)
}
