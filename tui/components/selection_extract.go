package components

import (
	"regexp"
	"strings"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes ANSI CSI escape sequences from s.
func StripANSI(s string) string { return ansiRE.ReplaceAllString(s, "") }

// SliceRunes returns s[start:end] in *rune* units. end == -1 means
// "to the end of the string". Out-of-range indices clamp to boundaries.
func SliceRunes(s string, start, end int) string {
	runes := []rune(s)
	if start < 0 {
		start = 0
	}
	if start > len(runes) {
		start = len(runes)
	}
	if end < 0 || end > len(runes) {
		end = len(runes)
	}
	if start > end {
		start = end
	}
	return string(runes[start:end])
}

// Extract returns the plain text between (sl, sc) and (el, ec).
func Extract(content string, sl, sc, el, ec int) string {
	lines := strings.Split(content, "\n")
	for i, ln := range lines {
		lines[i] = StripANSI(ln)
	}
	if sl < 0 {
		sl = 0
	}
	if el >= len(lines) {
		el = len(lines) - 1
	}
	if sl > el || sl >= len(lines) {
		return ""
	}
	if sl == el {
		return SliceRunes(lines[sl], sc, ec)
	}
	var sb strings.Builder
	sb.WriteString(SliceRunes(lines[sl], sc, -1))
	sb.WriteByte('\n')
	for i := sl + 1; i < el; i++ {
		sb.WriteString(lines[i])
		sb.WriteByte('\n')
	}
	sb.WriteString(SliceRunes(lines[el], 0, ec))
	return sb.String()
}
