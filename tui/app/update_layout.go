package app

import (
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/components"
)

// chatRect returns the screen-cell rectangle currently occupied by the
// chat viewport. It mirrors the layout math in app/view.go.
func (m Model) chatRect() (x0, y0, w, h int) {
	chatWidth := m.width - outerMarginX - innerGap - components.SidebarWidth
	if chatWidth < 1 {
		chatWidth = 1
	}
	inputLines := lipgloss.Height(m.input.View())
	statusLines := lipgloss.Height(m.statusbar.View())
	chatHeight := m.height - inputLines - statusLines - outerMarginY - chatInputGap
	if chatHeight < 1 {
		chatHeight = 1
	}
	return outerMarginX, outerMarginY, chatWidth, chatHeight
}

// chatContentCoord maps an absolute terminal cell (x, y) to a
// content-relative (line, col) coordinate.  The line is offset by the
// viewport's YOffset so selection anchors stay locked to content even when
// the viewport scrolls.
func (m Model) chatContentCoord(x, y int) (line, col int, ok bool) {
	x0, y0, w, h := m.chatRect()
	if x < x0 || x >= x0+w || y < y0 || y >= y0+h {
		return 0, 0, false
	}
	return m.chat.YOffset() + (y - y0), x - x0, true
}

// chatContentCoordClamped clamps to the chat rectangle and returns a
// content-relative coordinate.
func (m Model) chatContentCoordClamped(x, y int) (line, col int) {
	x0, y0, w, h := m.chatRect()
	if x < x0 {
		x = x0
	} else if x >= x0+w {
		x = x0 + w - 1
	}
	if y < y0 {
		y = y0
	} else if y >= y0+h {
		y = y0 + h - 1
	}
	return m.chat.YOffset() + (y - y0), x - x0
}

// inputBodyRect returns the screen-cell rectangle of the textarea body
// inside the input panel (excludes accent bar, padding, footer).
func (m Model) inputBodyRect() (x0, y0, w, h int) {
	chatWidth := m.width - outerMarginX - innerGap - components.SidebarWidth
	if chatWidth < 1 {
		chatWidth = 1
	}
	statusLines := lipgloss.Height(m.statusbar.View())
	inputLines := lipgloss.Height(m.input.View())

	inputY0 := m.height - statusLines - inputLines
	inputX0 := outerMarginX

	bodyX0 := inputX0 + 2
	bodyW := chatWidth - 3
	bodyY0 := inputY0 + 1
	bodyH := inputLines - 4
	if bodyW < 1 {
		bodyW = 1
	}
	if bodyH < 1 {
		bodyH = 1
	}
	return bodyX0, bodyY0, bodyW, bodyH
}

// inputContentCoord maps an absolute terminal cell to input-body coords.
func (m Model) inputContentCoord(x, y int) (line, col int, ok bool) {
	x0, y0, w, h := m.inputBodyRect()
	if x < x0 || x >= x0+w || y < y0 || y >= y0+h {
		return 0, 0, false
	}
	return y - y0, x - x0, true
}

// inputContentCoordClamped clamps to the input body rectangle.
func (m Model) inputContentCoordClamped(x, y int) (line, col int) {
	x0, y0, w, h := m.inputBodyRect()
	if x < x0 {
		x = x0
	} else if x >= x0+w {
		x = x0 + w - 1
	}
	if y < y0 {
		y = y0
	} else if y >= y0+h {
		y = y0 + h - 1
	}
	return y - y0, x - x0
}
