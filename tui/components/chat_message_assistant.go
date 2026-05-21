package components

import (
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/tui/styles"
)

var inlineCodeBgRE = regexp.MustCompile(`(\x1b\[[0-9;]*?);?48;5;236;?([0-9;]*m)`)

func stripInlineCodeBg(s string) string {
	for {
		next := inlineCodeBgRE.ReplaceAllStringFunc(s, func(m string) string {
			sub := inlineCodeBgRE.FindStringSubmatch(m)
			prefix, suffix := sub[1], sub[2]
			if prefix == "\x1b[" && suffix == "m" {
				return "\x1b[m"
			}
			if prefix == "\x1b[" {
				return prefix + suffix
			}
			if suffix == "m" {
				return prefix + "m"
			}
			return prefix + ";" + suffix
		})
		if next == s {
			return s
		}
		s = next
	}
}

func (c *Chat) renderAssistantMessage(msg chatMessage) string {
	content := msg.content
	if c.renderer != nil {
		if out, err := c.renderer.Render(msg.content); err == nil {
			content = strings.TrimRight(out, "\n")
		}
	}
	content = stripInlineCodeBg(content)

	lineStyle := lipgloss.NewStyle().
		Background(styles.ChatBackground).
		Width(c.width)

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = lineStyle.Render(line)
	}
	rendered := strings.Join(lines, "\n")

	return rendered + messageGap
}
