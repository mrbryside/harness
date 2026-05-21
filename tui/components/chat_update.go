package components

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// chatToastTickMsg is sent by the tea.Tick command scheduled by ShowToast.
type chatToastTickMsg struct{}

// SelectStart begins a new selection at (line, col) in chat content space.
func (c *Chat) SelectStart(line, col int) { c.sel.Start(line, col) }

// SelectExtend moves the selection's end anchor.
func (c *Chat) SelectExtend(line, col int) { c.sel.Extend(line, col) }

// SelectClear drops any in-flight selection.
func (c *Chat) SelectClear() { c.sel.Clear() }

// HasSelection reports whether a selection is active.
func (c Chat) HasSelection() bool { return c.sel.Active() }

// SelectedText returns the plain text covered by the selection.
func (c Chat) SelectedText() string { return c.sel.Text(StripANSI(c.viewport.View())) }

func (c Chat) Update(msg tea.Msg) (Chat, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.viewport.SetWidth(msg.Width)
		c.viewport.SetHeight(msg.Height)
		c.width = msg.Width
		c.height = msg.Height
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
