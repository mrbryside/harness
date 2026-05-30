package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/tui/styles"
)

const (
	permOptionYes = 0
	permOptionNo  = 1
	permOptionCount = 2
)

// PermissionQuestion is a yes/no confirmation question.
type PermissionQuestion struct {
	id            string
	question      string
	selectedIndex int // 0=Yes, 1=No
	active        bool
}

// NewPermissionQuestion creates a new permission question (active by default).
func NewPermissionQuestion(id, question string) *PermissionQuestion {
	return &PermissionQuestion{
		id:       id,
		question: question,
		active:   true,
	}
}

var permOptionLabels = []string{
	"Yes, allow this permission.",
	"No, not allow.",
}

func (q *PermissionQuestion) Type() string        { return QuestionTypePermission }
func (q *PermissionQuestion) ID() string          { return q.id }
func (q *PermissionQuestion) Active() bool        { return q.active }
func (q *PermissionQuestion) OptionCount() int    { return permOptionCount }
func (q *PermissionQuestion) SelectedIndex() int  { return q.selectedIndex }
func (q *PermissionQuestion) OptionLabel(i int) string {
	if i < 0 || i >= len(permOptionLabels) {
		return ""
	}
	return permOptionLabels[i]
}

func (q *PermissionQuestion) Show(question string) {
	q.active = true
	q.question = question
	q.selectedIndex = 0
}

func (q *PermissionQuestion) Hide() {
	q.active = false
	q.question = ""
	q.selectedIndex = 0
}

func (q *PermissionQuestion) HandleKey(msg tea.Msg) (tea.Cmd, bool) {
	if !q.active {
		return nil, false
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyUp:
			q.MoveSelection(-1)
			return nil, true
		case tea.KeyDown:
			q.MoveSelection(1)
			return nil, true
		case tea.KeyEnter, ' ':
			return q.emitAnswer(), true
		case tea.KeyEscape:
			return q.emitCancel(), true
		}
		return nil, true
	}
	return nil, false
}

func (q *PermissionQuestion) Cancel() tea.Cmd {
	if !q.active {
		return nil
	}
	return q.emitCancel()
}

func (q *PermissionQuestion) emitAnswer() tea.Cmd {
	idx := q.selectedIndex
	label := q.OptionLabel(idx)
	qid := q.id
	return func() tea.Msg {
		return QuestionAnswerMsg{
			QuestionID: qid,
			Data: QuestionChoice{
				Index: idx,
				Label: label,
			},
		}
	}
}

func (q *PermissionQuestion) emitCancel() tea.Cmd {
	qid := q.id
	return func() tea.Msg {
		return QuestionAnswerMsg{
			QuestionID: qid,
			Data: QuestionChoice{
				Index: -1,
				Label: "cancelled",
			},
		}
	}
}

func (q *PermissionQuestion) MoveSelection(dir int) {
	q.selectedIndex += dir
	if q.selectedIndex < 0 {
		q.selectedIndex = 0
	}
	if q.selectedIndex >= permOptionCount {
		q.selectedIndex = permOptionCount - 1
	}
}

func (q *PermissionQuestion) OverlayView(width int) string {
	if !q.active {
		return ""
	}

	innerWidth := width - 5
	if innerWidth < 1 {
		innerWidth = 1
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(styles.AccentOrange).
		Bold(true)
	unselectedStyle := lipgloss.NewStyle().
		Foreground(styles.SidebarValue)

	var yesLabel, noLabel string
	if q.selectedIndex == 0 {
		yesLabel = selectedStyle.Render("Yes, allow this permission.")
		noLabel = unselectedStyle.Render("No, not allow.")
	} else {
		yesLabel = unselectedStyle.Render("Yes, allow this permission.")
		noLabel = selectedStyle.Render("No, not allow.")
	}

	questionRendered := lipgloss.NewStyle().
		Foreground(styles.AssistantText).
		Background(styles.PanelBg).
		Width(innerWidth).
		Render(q.question)

	yesRow := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render(yesLabel)

	noRow := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render(noLabel)

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

func (q *PermissionQuestion) Height(width int) int {
	if !q.active {
		return 0
	}
	return countLines(q.OverlayView(width))
}

func (q *PermissionQuestion) String() string {
	if !q.active {
		return "PermissionQuestion{inactive}"
	}
	return fmt.Sprintf("PermissionQuestion{question=%q, id=%q, selected=%d}",
		q.question, q.id, q.selectedIndex)
}
