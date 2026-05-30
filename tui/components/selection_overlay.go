package components

import "strings"

// WrapLineRange highlights visible cells between start and end rune-columns
// using reverse video (ANSI 7), then restores bgSGR afterwards.
// Reverse video swaps foreground/background colours, so it works on any
// underlying background (including code diff add/remove colours) without
// painting over them.
// If width > 0 and end < 0 (full-line selection), the line is padded with
// spaces to the full width.
func WrapLineRange(line string, start, end int, bgSGR string, width int) string {
	var sb strings.Builder
	col := 0
	inside := false
	bgClosed := false

	runes := []rune(line)
	i := 0
	for i < len(runes) {
		r := runes[i]
		if r == 0x1b && i+1 < len(runes) && runes[i+1] == '[' {
			j := i + 2
			for j < len(runes) {
				rr := runes[j]
				if (rr >= 'a' && rr <= 'z') || (rr >= 'A' && rr <= 'Z') {
					break
				}
				j++
			}
			if j >= len(runes) {
				sb.WriteString(string(runes[i:]))
				return sb.String()
			}
			esc := string(runes[i : j+1])

			if inside {
				if esc == "\x1b[m" || esc == "\x1b[0m" {
					sb.WriteString(esc)
					// Re-assert reverse video after a reset so the highlight
					// stays active on the default background.
					sb.WriteString("\x1b[7m")
					i = j + 1
					if bgSGR != "" {
						if remaining := string(runes[i:]); strings.HasPrefix(remaining, bgSGR) {
							i += len([]rune(bgSGR))
						}
					}
					continue
				}
				// Preserve existing backgrounds (code diff add/remove colours, etc.)
				// instead of rewriting them to selection bg.
			}
			sb.WriteString(esc)
			i = j + 1
			continue
		}

		if !inside && col == start {
			// Start selection: reverse video (swap fg/bg).
			// This works on any background (including code diff add/remove
			// colours) because it inverts the existing colours instead of
			// painting over them.
			sb.WriteString("\x1b[7m")
			inside = true
		}
		if inside && !bgClosed && end >= 0 && col == end {
			// End selection: remove reverse video, then restore bg.
			sb.WriteString("\x1b[27m")
			sb.WriteString(bgSGR)
			bgClosed = true
			inside = false
		}
		sb.WriteRune(r)
		col++
		i++
	}
	// Full-line selection (end < 0) is kept open so the padding below
	// can extend the highlight to the container edge.  For every other
	// selection we close it now (remove reverse video, then restore bg).
	if inside && !bgClosed && end >= 0 {
		sb.WriteString("\x1b[27m")
		sb.WriteString(bgSGR)
	}

	// Pad full-line selections to extend the highlight to the right edge.
	// Partial selections are NOT padded — the underlying background (e.g.
	// code diff add/remove colours) is left untouched.
	if width > 0 && col < width && end < 0 && inside {
		sb.WriteString("\x1b[7m")
		sb.WriteString(strings.Repeat(" ", width-col))
		sb.WriteString("\x1b[27m")
		sb.WriteString(bgSGR)
	}

	return sb.String()
}
