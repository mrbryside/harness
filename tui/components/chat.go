package components

import (
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/styles"
)

// chatMessage holds a single turn in the conversation.
type chatMessage struct {
	role     string // "user", "assistant", or "code_diff"
	content  string
	rendered string // cached rendered output

	// codeDiff data — only used when role == "code_diff"
	diffPath    string
	diffOld     string
	diffNew     string
}

// Chat is the scrollable message history component.
type Chat struct {
	messages     []chatMessage
	viewport     viewport.Model
	width        int
	height       int
	userScrolled bool // true once the user has scrolled away from the bottom
	renderer     *glamour.TermRenderer
	sel          Selection

	// Toast notification shown at top-right of the chat area.
	toast      string
	toastUntil time.Time
}

func NewChat(width, height int) Chat {
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	vp.Style = lipgloss.NewStyle().Background(styles.ChatBackground)
	vp.FillHeight = true
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 4
	vp.SoftWrap = true

	c := Chat{
		viewport: vp,
		width:    width,
		height:   height,
	}
	c.renderer = newMarkdownRenderer(width)
	return c
}

func newMarkdownRenderer(width int) *glamour.TermRenderer {
	if width < 1 {
		width = 1
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes([]byte(styles.MonokaiGlamourStyle)),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	return r
}

func (c Chat) Init() tea.Cmd { return nil }

func (c Chat) AtTop() bool    { return c.viewport.AtTop() }
func (c Chat) AtBottom() bool { return c.viewport.AtBottom() }
func (c Chat) YOffset() int   { return c.viewport.YOffset() }
