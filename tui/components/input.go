package components

import (
	"time"
	"unicode"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/memory"
	"github.com/mrbryside/harness/tui/styles"
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
	textarea  textarea.Model
	width     int
	mode      Mode
	model     string
	sel       Selection
	lastEscAt time.Time
	history   memory.HistoryStorage
	draft     string
}

const inputBgSGR = "\x1b[48;2;26;27;38m"
const escDebounce = 1 * time.Second

func NewInput(model string) Input {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.ShowLineNumbers = false
	ta.DynamicHeight = true
	ta.MinHeight = 1
	ta.MaxHeight = 5
	ta.MaxContentHeight = 1000
	ta.SetHeight(1)
	ta.Prompt = ""

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
	return Input{
		textarea: ta,
		mode:     ModeBuild,
		model:    model,
		history:  memory.NewInMemoryHistory(),
	}
}

func (i Input) Value() string  { return i.textarea.Value() }
func (i *Input) Reset()        { i.textarea.Reset() }
func (i *Input) SetMode(m Mode) { i.mode = m }
func (i Input) Mode() Mode     { return i.mode }
func (i Input) Init() tea.Cmd  { return nil }

// isCombiningMark reports whether s starts with a Unicode combining mark.
func isCombiningMark(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		return unicode.IsMark(r)
	}
	return false
}
