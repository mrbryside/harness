package components

import (
	"unicode"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/styles"
)

// SendMsg is emitted when the user presses Enter to send a message.
type SendMsg struct {
	Content string
}

// Mode is the agent mode shown in the input footer (Build / Plan).
type Mode string

const (
	ModeBuild Mode = "Build"
	ModePlan  Mode = "Plan"
)

// Input is the multi-line text input component at the bottom of the TUI.
type Input struct {
	textarea textarea.Model
	width    int
	mode     Mode
	model    string
}

func NewInput(model string) Input {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.ShowLineNumbers = false
	// DynamicHeight: let the textarea size itself between MinHeight and
	// MaxHeight as the user types (counting soft-wrapped visual rows, not
	// just logical lines). This replaces the manual SetHeight() / LineCount()
	// dance below.
	ta.DynamicHeight = true
	ta.MinHeight = 1
	ta.MaxHeight = 10
	ta.SetHeight(1)
	// Remove the textarea's built-in per-line prompt (the thick "┃ " column
	// that otherwise appears immediately to the right of our blue accent bar).
	ta.Prompt = ""

	// Paint every textarea surface with PanelBg so the input panel reads as
	// one continuous block. Bubbles' default CursorLine uses a darker grey
	// that clashes with our PanelBg; overriding it (plus EndOfBuffer and
	// Placeholder which inherit a different default) keeps the panel solid.
	//
	// Foreground note: bubbles textarea renders the line the cursor is on
	// through computedCursorLine() (not computedText()), so the line the
	// user is actively typing inherits its colour from CursorLine. We have
	// to set Foreground on BOTH Text and CursorLine — otherwise typed text
	// falls back to the terminal's default foreground (visible as a yellow
	// or off-white tone in many themes).
	panelBg := lipgloss.NewStyle().Background(styles.PanelBg)
	themedText := lipgloss.NewStyle().
		Foreground(styles.AssistantText).
		Background(styles.PanelBg)
	st := ta.Styles()
	st.Focused.Base = panelBg
	st.Blurred.Base = panelBg
	st.Focused.Text = themedText
	st.Blurred.Text = themedText
	st.Focused.CursorLine = themedText
	st.Blurred.CursorLine = themedText
	st.Focused.EndOfBuffer = panelBg
	st.Blurred.EndOfBuffer = panelBg
	st.Focused.Placeholder = lipgloss.NewStyle().Foreground(styles.SidebarValue).Background(styles.PanelBg)
	st.Blurred.Placeholder = st.Focused.Placeholder
	st.Cursor.Color = styles.AssistantText
	ta.SetStyles(st)

	ta.Focus()
	return Input{textarea: ta, mode: ModeBuild, model: model}
}

// Value returns the current text in the input.
func (i Input) Value() string {
	return i.textarea.Value()
}

// Reset clears the input.
func (i *Input) Reset() {
	i.textarea.Reset()
}

// SetMode updates the agent mode (Build / Plan) shown in the footer.
func (i *Input) SetMode(m Mode) {
	i.mode = m
}

// Mode returns the current agent mode.
func (i Input) Mode() Mode {
	return i.mode
}

func (i Input) Init() tea.Cmd {
	return nil
}

// isCombiningMark reports whether s starts with a Unicode combining mark
// (category M: Mn = non-spacing, Mc = spacing-combining, Me = enclosing).
// Used to detect Thai tone marks (ไม้เอก ◌่, ไม้โท ◌้, etc.) and vowels
// (◌ิ, ◌ี, ◌ุ, ◌ู, …) typed without a preceding base character.
func isCombiningMark(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		return unicode.IsMark(r)
	}
	return false
}

func (i Input) Update(msg tea.Msg) (Input, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Drop Thai (and other) combining marks typed at the start of the
		// buffer. bubbles textarea counts cursor position in runes, but
		// combining marks are zero-width graphemes that attach to the
		// preceding base. With no base to attach to, the textarea reserves
		// a cell for the mark anyway → visible overlap / drift on the
		// next keystroke. Just ignore the keypress in that case.
		if i.textarea.Value() == "" && isCombiningMark(msg.Text) {
			return i, nil
		}

		// Many terminals (iTerm, Zed, Terminal.app, etc.) don't disambiguate
		// Shift+Enter from Enter via modifier bits — but they DO send the
		// physical key as Ctrl+J (LF, \n) instead of Enter's Ctrl+M (CR, \r)
		// when Shift is held. Rewrite Ctrl+J as a plain Enter so the
		// textarea inserts a newline (its keymap binds Enter to InsertNewline).
		if msg.Code == 'j' && msg.Mod == tea.ModCtrl {
			msg = tea.KeyPressMsg{Code: tea.KeyEnter, Text: "\n"}
			var cmd tea.Cmd
			i.textarea, cmd = i.textarea.Update(msg)
			return i, cmd
		}
		if msg.Code == tea.KeyEnter {
			// Shift+Enter (or Alt+Enter as a fallback for terminals that
			// don't distinguish Shift+Enter from Enter) inserts a newline;
			// plain Enter sends the message.
			if msg.Mod.Contains(tea.ModShift) || msg.Mod.Contains(tea.ModAlt) {
				break // fall through to textarea.Update → newline
			}
			content := i.textarea.Value()
			i.textarea.Reset()
			return i, func() tea.Msg {
				return SendMsg{Content: content}
			}
		}
	case tea.WindowSizeMsg:
		// leave 1 col for the left accent bar + 2 for horizontal padding
		w := msg.Width - 3
		if w < 1 {
			w = 1
		}
		i.width = msg.Width
		i.textarea.SetWidth(w)
	}

	var cmd tea.Cmd
	i.textarea, cmd = i.textarea.Update(msg)

	return i, cmd
}

func (i Input) View() string {
	innerWidth := i.width - 3 // 1 accent col + 2 padding cols
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Mode-coloured label + dim model name, padded to full inner width
	// so the footer row is one continuous PanelBg strip.
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

	// Textarea body — now that all textarea Styles.* surfaces use PanelBg,
	// the textarea renders as a solid panel by itself. We just need to width-
	// pad the block so any trailing whitespace right of the cursor line is
	// also PanelBg.
	body := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render(i.textarea.View())

	// Blank spacer line so footer doesn't crowd the textarea.
	spacer := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Width(innerWidth).
		Render("")

	inner := lipgloss.JoinVertical(lipgloss.Left, body, spacer, footer)

	// Pad the inner content (1 row top/bottom, 1 col left/right gap from bar).
	padded := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Padding(1, 1).
		Render(inner)

	// Build a left accent bar exactly as tall as `padded` so it reads as one
	// continuous vertical strip instead of a stack of half-blocks.
	barHeight := lipgloss.Height(padded)
	if barHeight < 1 {
		barHeight = 1
	}

	bar := lipgloss.NewStyle().
		Background(styles.UserBorder). // bar color as bg
		Width(1).
		Height(barHeight).
		Render("")

	return lipgloss.JoinHorizontal(lipgloss.Top, bar, padded)
}
