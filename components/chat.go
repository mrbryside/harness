package components

import (
	"regexp"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/styles"
)

// chatMessage holds a single turn in the conversation.
type chatMessage struct {
	role    string // "user" or "assistant"
	content string
}

// Chat is the scrollable message history component.
type Chat struct {
	messages     []chatMessage
	viewport     viewport.Model
	width        int
	height       int
	userScrolled bool // true once the user has scrolled away from the bottom
	renderer     *glamour.TermRenderer
}

func NewChat(width, height int) Chat {
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	// Style + FillHeight: paint every visible cell of the viewport (including
	// the empty trailing rows below the content) with the chat Background so
	// the chat area is solid styles.Background instead of falling back to the
	// terminal's default colour.
	vp.Style = lipgloss.NewStyle().Background(styles.Background)
	vp.FillHeight = true
	vp.MouseWheelEnabled = true
	// 2 lines per wheel event (default is 3). Slightly smaller delta
	// pairs with the wheel filter in main.go to keep scrolling responsive
	// without flooding the event queue on hard trackpad flicks.
	vp.MouseWheelDelta = 2
	vp.SoftWrap = true

	c := Chat{
		viewport: vp,
		width:    width,
		height:   height,
	}
	c.renderer = newMarkdownRenderer(width)
	return c
}

// newMarkdownRenderer builds a glamour renderer that word-wraps at the
// given width. Returns nil if construction fails.
//
// We don't strip glamour's backgrounds here: Chat.View() patches every
// SGR reset to re-assert the chat Background, so any BG glamour or
// chroma emits gets overridden at the next token boundary.
func newMarkdownRenderer(width int) *glamour.TermRenderer {
	if width < 1 {
		width = 1
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	return r
}

// AppendMessage adds a new message to the history.
// A new user turn resets the scroll-lock so the response auto-scrolls.
func (c *Chat) AppendMessage(role, content string) {
	if role == "user" {
		c.userScrolled = false
	}
	c.messages = append(c.messages, chatMessage{role: role, content: content})
	c.refresh()
}

// AppendChunk appends streamed text to the last message.
func (c *Chat) AppendChunk(chunk string) {
	if len(c.messages) == 0 {
		return
	}
	c.messages[len(c.messages)-1].content += chunk
	c.refresh()
}

// refresh re-renders all messages into the viewport.
// It only auto-scrolls to bottom when the user has not manually scrolled up.
func (c *Chat) refresh() {
	var sb strings.Builder
	for _, msg := range c.messages {
		sb.WriteString(c.renderMessage(msg))
	}
	c.viewport.SetContent(sb.String())
	if !c.userScrolled {
		c.viewport.GotoBottom()
	}
}

// messageGap is the blank lines inserted after every message so the spacing
// between any two consecutive messages (user↔assistant or assistant↔user) is
// always identical.
const messageGap = "\n\n"

// renderMessage renders a single message with appropriate styling.
//
// User messages are wrapped in a panel-coloured block with a coloured
// accent bar on the left (matches the input panel visually).
//
// Assistant messages are rendered through glamour so markdown
// (headings, lists, code fences, bold/italic, …) becomes formatted
// text. The grey background glamour applies to inline `code` spans is
// stripped so inline code blends with the chat Background. Every other
// SGR reset is re-asserted to chatBgSGR in Chat.View() so the chat
// area stays solid styles.Background end-to-end.
func (c *Chat) renderMessage(msg chatMessage) string {
	if msg.role == "user" {
		innerWidth := c.width - 1 // 1 col for the accent bar
		if innerWidth < 1 {
			innerWidth = 1
		}

		// Content block — padded panel where the user message sits.
		padded := lipgloss.NewStyle().
			Background(styles.PanelBg).
			Foreground(styles.AssistantText).
			Padding(1, 2).
			Width(innerWidth).
			Render(msg.content)

		// Left accent bar — same height as padded.
		barHeight := lipgloss.Height(padded)
		bar := lipgloss.NewStyle().
			Background(styles.UserBorder).
			Width(1).
			Height(barHeight).
			Render("")

		return lipgloss.JoinHorizontal(lipgloss.Top, bar, padded) + messageGap
	}

	// Assistant — glamour-render, then strip the inline-code grey BG.
	// Chat.View() handles the rest of the BG re-assertion.
	content := msg.content
	if c.renderer != nil {
		if out, err := c.renderer.Render(msg.content); err == nil {
			content = strings.TrimRight(out, "\n")
		}
	}
	content = stripInlineCodeBg(content)

	lineStyle := lipgloss.NewStyle().
		Background(styles.Background).
		Width(c.width)

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = lineStyle.Render(line)
	}
	return strings.Join(lines, "\n") + messageGap
}

// chatBgSGR re-asserts the chat background (styles.Background = #000000)
// after glamour's SGR resets so the chat panel stays solid black.
const chatBgSGR = "\x1b[48;2;0;0;0m"

// inlineCodeBgRE matches the "48;5;236" segment that glamour's dark style
// uses for inline `code` backgrounds. The regex captures the SGR prefix
// (up to but not including the BG segment) and the suffix (whatever comes
// after it, including the terminating 'm'). stripInlineCodeBg rewrites
// every match into an equivalent SGR with that one segment removed while
// preserving any foreground/attribute codes around it.
//
// We strip only this exact 256-color BG so fenced code blocks — which use
// 24-bit BG ("\x1b[48;2;...m") for their chroma background — stay intact.
var inlineCodeBgRE = regexp.MustCompile(`(\x1b\[[0-9;]*?);?48;5;236;?([0-9;]*m)`)

func stripInlineCodeBg(s string) string {
	for {
		next := inlineCodeBgRE.ReplaceAllStringFunc(s, func(m string) string {
			sub := inlineCodeBgRE.FindStringSubmatch(m)
			prefix, suffix := sub[1], sub[2]
			// "\x1b[" + "m" → bare reset
			if prefix == "\x1b[" && suffix == "m" {
				return "\x1b[m"
			}
			// "\x1b[" + "Xm" → no leading ';' needed
			if prefix == "\x1b[" {
				return prefix + suffix
			}
			// prefix has codes + suffix is just "m" → drop trailing ';'
			if suffix == "m" {
				return prefix + "m"
			}
			// both sides have codes → rejoin with ';'
			return prefix + ";" + suffix
		})
		if next == s {
			return s
		}
		s = next
	}
}

func (c Chat) Init() tea.Cmd { return nil }

// AtTop reports whether the chat viewport is scrolled to the top
// (no more content above). Used by the program-level wheel filter
// to drop wheel-up events that would otherwise pile up in the queue.
func (c Chat) AtTop() bool { return c.viewport.AtTop() }

// AtBottom reports whether the chat viewport is scrolled to the bottom.
func (c Chat) AtBottom() bool { return c.viewport.AtBottom() }

func (c Chat) Update(msg tea.Msg) (Chat, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.viewport.SetWidth(msg.Width)
		c.viewport.SetHeight(msg.Height)
		c.width = msg.Width
		c.height = msg.Height

		// rebuild markdown renderer so word-wrap matches the new chat width
		c.renderer = newMarkdownRenderer(msg.Width)

		c.refresh()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "pgup":
			c.userScrolled = true
		case "down", "pgdown":
			// if after scrolling down we hit bottom, resume auto-scroll
			c.viewport, cmd = c.viewport.Update(msg)
			if c.viewport.AtBottom() {
				c.userScrolled = false
			}
			return c, cmd
		}
		c.viewport, cmd = c.viewport.Update(msg)
	case tea.MouseWheelMsg:
		// Mouse wheel: scroll the chat viewport. Mark scrolled-up on wheel-up,
		// release scroll-lock when wheel-down reaches the bottom.
		switch msg.Button {
		case tea.MouseWheelUp:
			c.userScrolled = true
		case tea.MouseWheelDown:
			c.viewport, cmd = c.viewport.Update(msg)
			if c.viewport.AtBottom() {
				c.userScrolled = false
			}
			return c, cmd
		}
		c.viewport, cmd = c.viewport.Update(msg)
	default:
		c.viewport, cmd = c.viewport.Update(msg)
	}
	return c, cmd
}

func (c Chat) View() string {
	// Why this exists: solid black background inside the chat area.
	//
	// Three layers contribute SGR escapes to the final string:
	//   1. glamour       — emits "\x1b[m" (full reset) after every
	//                      styled token in the rendered markdown.
	//   2. lipgloss      — when it pads a line to a fixed Width, it can
	//                      emit a reset right before the trailing pad
	//                      spaces instead of carrying the BG through.
	//   3. viewport      — with FillHeight=true it pads empty rows below
	//                      the content with plain spaces preceded by a
	//                      reset.
	//
	// A bare "\x1b[m" / "\x1b[0m" clears both FG and BG, so any plain
	// space that follows is rendered with the terminal's default
	// background (which is grey in many themes, e.g. Zed). That shows
	// up as grey strips on empty lines inside fenced code blocks and on
	// the empty rows beneath the conversation.
	//
	// Fix: re-assert the chat background SGR immediately after every
	// reset, so the next cell (whatever it is — a styled token, a pad
	// space, a viewport fill space) starts with BG = styles.Background.
	// We do this once here at the outermost layer because View() runs
	// after all three sources have contributed their escapes; patching
	// earlier (in renderMessage) would miss the resets viewport adds.
	out := c.viewport.View()
	out = strings.ReplaceAll(out, "\x1b[m", "\x1b[m"+chatBgSGR)
	out = strings.ReplaceAll(out, "\x1b[0m", "\x1b[0m"+chatBgSGR)
	return out
}
