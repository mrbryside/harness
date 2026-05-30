package components

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// chatToastTickMsg is sent by the tea.Tick command scheduled by ShowToast.
type chatToastTickMsg struct{}

// SelectStart begins a new selection at (line, col) in chat content space.
func (c *Chat) SelectStart(line, col int) { c.sel.Start(line, col) }

// SelectExtend moves the selection's end anchor.
func (c *Chat) SelectExtend(line, col int) { c.sel.Extend(line, col) }

// ScrollUpAndExtend scrolls the viewport up by `lines` lines and extends the
// selection to the new top visible line (content-relative).  Used for
// edge-auto-scroll while dragging.
func (c *Chat) ScrollUpAndExtend(lines, col int) {
	if c.viewport.AtTop() {
		return
	}
	c.viewport.ScrollUp(lines)
	c.sel.Extend(c.viewport.YOffset(), col)
}

// ScrollDownAndExtend scrolls the viewport down by `lines` lines and extends
// the selection to the new bottom visible line (content-relative).  Used for
// edge-auto-scroll while dragging.
func (c *Chat) ScrollDownAndExtend(lines, col int) {
	if c.viewport.AtBottom() {
		return
	}
	c.viewport.ScrollDown(lines)
	c.sel.Extend(c.viewport.YOffset()+c.height-1, col)
}

// SelectClear drops any in-flight selection.
func (c *Chat) SelectClear() { c.sel.Clear() }

// HasSelection reports whether a selection is active.
func (c Chat) HasSelection() bool { return c.sel.Active() }

// SelectedText returns the plain text covered by the selection.
// Uses the full viewport content (not just the visible portion) so text
// scrolled out of view is still included when copying.
func (c Chat) SelectedText() string {
	text := StripANSI(c.viewport.GetContent())
	text = strings.ReplaceAll(text, "┃", "")
	return c.sel.Text(text)
}

func (c Chat) Update(msg tea.Msg) (Chat, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.viewport.SetWidth(msg.Width)
		c.viewport.SetHeight(msg.Height)
		c.width = msg.Width
		c.height = msg.Height
		// Recreate renderer with wrap width accounting for 3-space margin.
		c.renderer = newMarkdownRenderer(msg.Width)
		c.refresh()
	case tea.MouseWheelMsg:
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
	case chatToastTickMsg:
		if time.Now().After(c.toastUntil) {
			c.toast = ""
		}
		return c, nil
	default:
		c.viewport, cmd = c.viewport.Update(msg)
	}
	return c, cmd
}
