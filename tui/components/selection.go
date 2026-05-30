package components

import "strings"

// Selection tracks an in-flight click-and-drag selection.
type Selection struct {
	active              bool
	startLine, startCol int
	endLine, endCol     int
}

// Start begins a new selection anchored at (line, col).
func (s *Selection) Start(line, col int) {
	*s = Selection{
		active:    true,
		startLine: line, startCol: col,
		endLine: line, endCol: col,
	}
}

// Extend moves only the end anchor.
func (s *Selection) Extend(line, col int) {
	if !s.active {
		return
	}
	s.endLine = line
	s.endCol = col
}

// Clear drops the selection entirely.
func (s *Selection) Clear() { *s = Selection{} }

// Active reports whether a selection has been started and not cleared.
func (s Selection) Active() bool { return s.active }

// HasRange reports whether the selection covers at least one cell.
func (s Selection) HasRange() bool {
	if !s.active {
		return false
	}
	return s.startLine != s.endLine || s.startCol != s.endCol
}

// Normalised returns (startLine, startCol, endLine, endCol) in natural
// reading order regardless of drag direction.
func (s Selection) Normalised() (sl, sc, el, ec int) {
	sl, sc = s.startLine, s.startCol
	el, ec = s.endLine, s.endCol
	if sl > el || (sl == el && sc > ec) {
		sl, el = el, sl
		sc, ec = ec, sc
	}
	return
}

// Text returns the plain text covered by the selection within content.
func (s Selection) Text(content string) string {
	if !s.HasRange() {
		return ""
	}
	sl, sc, el, ec := s.Normalised()
	return Extract(content, sl, sc, el, ec)
}

// Overlay paints the selection highlight on top of rendered output.
// If width > 0, the highlight is padded to the full width so it extends
// to the right edge of the container.
func (s Selection) Overlay(rendered string, yoff int, bgSGR string, width int) string {
	if !s.HasRange() {
		return rendered
	}
	sl, sc, el, ec := s.Normalised()
	lines := strings.Split(rendered, "\n")
	for i := range lines {
		contentLine := yoff + i
		if contentLine < sl || contentLine > el {
			continue
		}
		startCol, endCol := 0, -1
		if contentLine == sl {
			startCol = sc
		}
		if contentLine == el {
			endCol = ec
		}
		if startCol < 0 {
			startCol = 0
		}
		lines[i] = WrapLineRange(lines[i], startCol, endCol, bgSGR, width)
	}
	return strings.Join(lines, "\n")
}
