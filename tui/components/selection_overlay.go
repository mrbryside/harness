package components

import (
	"regexp"
	"strings"

	"github.com/mrbryside/harness/tui/styles"
)

// bgSgrRE matches any SGR segment that sets a background colour.
var bgSgrRE = regexp.MustCompile(`(?:48;2;\d+;\d+;\d+|48;5;\d+|10[0-7]|4[0-79])`)

// rewriteBgsToSelection replaces every background-setting segment with
// styles.SelectionBgSGR's payload. If the escape doesn't contain a BG segment
// it's returned untouched.
func rewriteBgsToSelection(esc string) string {
	if !strings.HasPrefix(esc, "\x1b[") || !strings.HasSuffix(esc, "m") {
		return esc
	}
	body := esc[2 : len(esc)-1]
	if body == "" {
		return esc
	}
	const selPayload = "48;2;73;72;62"
	if !bgSgrRE.MatchString(body) {
		return esc
	}
	return "\x1b[" + bgSgrRE.ReplaceAllString(body, selPayload) + "m"
}

// WrapLineRange paints styles.SelectionBgSGR over visible cells between start
// and end rune-columns, restoring bgSGR afterwards.
func WrapLineRange(line string, start, end int, bgSGR string) string {
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
					sb.WriteString(styles.SelectionBgSGR)
					i = j + 1
					if bgSGR != "" {
						if remaining := string(runes[i:]); strings.HasPrefix(remaining, bgSGR) {
							i += len([]rune(bgSGR))
						}
					}
					continue
				}
				esc = rewriteBgsToSelection(esc)
			}
			sb.WriteString(esc)
			i = j + 1
			continue
		}

		if !inside && col == start {
			sb.WriteString(styles.SelectionBgSGR)
			inside = true
		}
		if inside && !bgClosed && end >= 0 && col == end {
			sb.WriteString(bgSGR)
			bgClosed = true
			inside = false
		}
		sb.WriteRune(r)
		col++
		i++
	}
	if inside && !bgClosed {
		sb.WriteString(bgSGR)
	}
	return sb.String()
}
