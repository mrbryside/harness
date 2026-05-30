package components

import (
	"bufio"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/mrbryside/harness/tui/styles"
)

// diffContainerOverhead is the number of columns consumed by the diff
// container's left border (1) + left padding (2).
const diffContainerOverhead = 3

func (c *Chat) renderCodeDiffMessage(msg chatMessage) string {
	return c.RenderCodeDiffV2(msg.diffPath, msg.diffOld, msg.diffNew, msg.diffStart) + messageGap
}

// diffChromaStyle is a custom chroma style that matches the app theme.
var diffChromaStyle = func() *chroma.Style {
	sb := chroma.NewStyleBuilder("harness-diff")
	sb.Add(chroma.Text, "#E6EDF3")
	sb.Add(chroma.Comment, "#6E7681 italic")
	sb.Add(chroma.Keyword, "#FF7B72")
	sb.Add(chroma.KeywordType, "#79C0FF")
	sb.Add(chroma.NameFunction, "#FFB86C")
	sb.Add(chroma.NameClass, "#79C0FF")
	sb.Add(chroma.NameBuiltin, "#79C0FF")
	sb.Add(chroma.LiteralString, "#E6EDF3")
	sb.Add(chroma.LiteralNumber, "#A5D6FF")
	sb.Add(chroma.Operator, "#FF7B72")
	sb.Add(chroma.Punctuation, "#E6EDF3")
	style, _ := sb.Build()
	return style
}()

// renderSimpleDiff is the fallback when the file can't be read or oldContent isn't found.
// startLine is the 1-based line number where oldContent starts in the file.
func (c *Chat) renderSimpleDiff(path string, oldContent, newContent string, startLine int) string {
	oldRaw := strings.Split(oldContent, "\n")
	newRaw := strings.Split(newContent, "\n")
	diffLines := computeRawDiff(oldRaw, newRaw, 3)

	oldRendered := highlightCode(oldContent, langFromPath(path))
	newRendered := highlightCode(newContent, langFromPath(path))

	var sb strings.Builder

	header := lipgloss.NewStyle().
		Foreground(styles.SidebarLabel).
		Background(styles.Background).
		Padding(1, 2).
		Render("← " + path)
	sb.WriteString(header)
	sb.WriteString("\n")

	// Top padding line (blank line with background)
	padding := lipgloss.NewStyle().
		Background(styles.Background).
		Width(c.width - diffContainerOverhead).
		Render("")
	sb.WriteString(padding)
	sb.WriteString("\n")

	// Prefix is 7 chars: 4 (line num) + 1 (space) + 1 (marker) + 1 (space)
	// Subtract 2 more for right padding so code doesn't touch the edge.
	contentWidth := c.width - diffContainerOverhead - 7 - 2

	for _, dl := range diffLines {
		if dl.content == "" && dl.kind == "context" {
			continue
		}

		var highlighted string
		var lineNum int
		switch dl.kind {
		case "context":
			highlighted = getRenderedLine(newRendered, dl.newNum)
			if highlighted == "" {
				highlighted = dl.content
			}
			lineNum = startLine + dl.lineNum - 1
		case "remove":
			highlighted = getRenderedLine(oldRendered, dl.oldNum)
			if highlighted == "" {
				highlighted = dl.content
			}
			lineNum = startLine + dl.lineNum - 1
		case "add":
			highlighted = getRenderedLine(newRendered, dl.newNum)
			if highlighted == "" {
				highlighted = dl.content
			}
			lineNum = startLine + dl.lineNum - 1
		}

		switch dl.kind {
		case "context":
			numStr := fmt.Sprintf("%4d", lineNum)
			if dl.content == "..." {
				numStr = "   "
			}
			num := lipgloss.NewStyle().
				Foreground(styles.SidebarValue).
				Background(styles.Background).
				Width(4).
				Align(lipgloss.Right).
				Render(numStr)
			marker := lipgloss.NewStyle().
				Foreground(styles.SidebarValue).
				Background(styles.Background).
				Width(1).
				Render(" ")
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width-diffContainerOverhead, styles.Background)

		case "remove":
			num := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#c76b72")).
				Background(lipgloss.Color("#34232c")).
				Width(4).
				Align(lipgloss.Right).
				Render(fmt.Sprintf("%4d", lineNum))
			marker := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#c76b72")).
				Background(lipgloss.Color("#34232c")).
				Width(1).
				Render("-")
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width-diffContainerOverhead, lipgloss.Color("#34232c"))

		case "add":
			num := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#bfda90")).
				Background(lipgloss.Color("#23303a")).
				Width(4).
				Align(lipgloss.Right).
				Render(fmt.Sprintf("%4d", lineNum))
			marker := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#bfda90")).
				Background(lipgloss.Color("#23303a")).
				Width(1).
				Render("+")
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width-diffContainerOverhead, lipgloss.Color("#23303a"))
		}
	}

	return c.renderDiffContainer(sb.String())
}

// highlightCode highlights an entire code block using chroma (no glamour padding).
func highlightCode(code string, lang string) []string {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	formatter := formatters.Get("terminal16m")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return strings.Split(code, "\n")
	}

	var buf strings.Builder
	err = formatter.Format(&buf, diffChromaStyle, iterator)
	if err != nil {
		return strings.Split(code, "\n")
	}

	rendered := strings.Trim(buf.String(), "\n")
	lines := strings.Split(rendered, "\n")

	// Strip any background colors chroma may have emitted
	for i, line := range lines {
		lines[i] = stripBackgroundColors(line)
	}

	return lines
}

// visibleWidth returns the number of visible characters (ANSI stripped).
func visibleWidth(s string) int {
	stripped := ansiSeqRE.ReplaceAllString(s, "")
	return lipgloss.Width(stripped)
}

// wrapAndRenderLine wraps highlighted content to fit within contentWidth and renders
// each row with the proper indentation. The first row gets num+marker; continuation rows
// get indent spaces so they align with the content column. All rows get a full-width
// background so the highlight never appears broken.
func wrapAndRenderLine(sb *strings.Builder, num, marker, highlighted string, contentWidth int, totalWidth int, bgColor color.Color) {
	chunks := wrapAnsiLine(highlighted, contentWidth)
	prefixWidth := 7 // 4 (line num) + 1 (space) + 1 (marker) + 1 (space)
	bgSGR := styles.ColorToAnsiBg(bgColor)

	// Ensure any resets inside num/marker re-apply the background so the
	// following spaces (and lipgloss padding) keep the correct colour.
	num = injectBackgroundAfterResets(num, bgSGR)
	marker = injectBackgroundAfterResets(marker, bgSGR)

	for i, chunk := range chunks {
		chunk = injectBackgroundAfterResets(chunk, bgSGR)
		var line string
		if i == 0 {
			// First row: num + space + marker + space + content
			line = num + " " + marker + " " + chunk
		} else {
			// Continuation rows: indent to align with content column
			indent := lipgloss.NewStyle().Background(bgColor).Render(strings.Repeat(" ", prefixWidth))
			indent = injectBackgroundAfterResets(indent, bgSGR)
			line = indent + chunk
		}
		// Pad the entire line to full width so the background spans edge-to-edge.
		style := lipgloss.NewStyle().Background(bgColor).Width(totalWidth)
		sb.WriteString(style.Render(line))
		sb.WriteString("\n")
	}
}

// injectBackgroundAfterResets inserts an ANSI background SGR after every bare
// reset sequence so that lipgloss padding spaces inherit the correct colour.
func injectBackgroundAfterResets(s, bgSGR string) string {
	s = strings.ReplaceAll(s, "\x1b[m", "\x1b[m"+bgSGR)
	s = strings.ReplaceAll(s, "\x1b[0m", "\x1b[0m"+bgSGR)
	return s
}

// wrapAnsiLine splits a string with ANSI codes into chunks where each chunk's
// visible width is at most maxWidth. ANSI sequences are preserved and carried
// across chunk boundaries.
func wrapAnsiLine(s string, maxWidth int) []string {
	if visibleWidth(s) <= maxWidth {
		return []string{s}
	}

	matches := ansiSeqRE.FindAllStringIndex(s, -1)
	var chunks []string
	var current strings.Builder
	visibleCount := 0
	lastEnd := 0
	activeAnsi := ""

	for _, match := range matches {
		start, end := match[0], match[1]

		// Add visible text between lastEnd and start
		text := s[lastEnd:start]
		for _, r := range text {
			if visibleCount >= maxWidth {
				chunks = append(chunks, current.String())
				current.Reset()
				if activeAnsi != "" {
					current.WriteString(activeAnsi)
				}
				visibleCount = 0
			}
			current.WriteRune(r)
			visibleCount++
		}

		// Add the ANSI sequence
		seq := s[start:end]
		current.WriteString(seq)
		lastEnd = end

		// Track active ANSI sequences (simple heuristic: if it's not a reset, carry it)
		if !strings.Contains(seq, "\x1b[0m") && !strings.Contains(seq, "\x1b[m") {
			activeAnsi = seq
		} else {
			activeAnsi = ""
		}
	}

	// Add remaining text
	text := s[lastEnd:]
	for _, r := range text {
		if visibleCount >= maxWidth {
			chunks = append(chunks, current.String())
			current.Reset()
			if activeAnsi != "" {
				current.WriteString(activeAnsi)
			}
			visibleCount = 0
		}
		current.WriteRune(r)
		visibleCount++
	}

	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	return chunks
}

// getRenderedLine returns the n-th non-empty rendered line (1-based).
func getRenderedLine(rendered []string, n int) string {
	if n <= 0 {
		return ""
	}
	idx := n - 1
	if idx >= len(rendered) {
		return ""
	}
	return rendered[idx]
}

// isVisualEmpty checks if a line is visually empty (no visible characters).
func isVisualEmpty(line string) bool {
	clean := ansiSeqRE.ReplaceAllString(line, "")
	return strings.TrimSpace(clean) == ""
}

// ansiSeqRE matches ANSI escape sequences.
var ansiSeqRE = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

// stripBackgroundColors removes background color codes from ANSI sequences.
func stripBackgroundColors(s string) string {
	return ansiSeqRE.ReplaceAllStringFunc(s, func(m string) string {
		match := ansiSeqRE.FindStringSubmatch(m)
		if len(match) < 2 {
			return m
		}
		codes := match[1]
		if codes == "" || codes == "0" {
			return m
		}

		parts := strings.Split(codes, ";")
		var newParts []string

		i := 0
		for i < len(parts) {
			switch parts[i] {
			case "48":
				if i+1 < len(parts) {
					if parts[i+1] == "5" {
						i += 3
						continue
					} else if parts[i+1] == "2" {
						i += 5
						continue
					}
				}
				i++
			default:
				newParts = append(newParts, parts[i])
				i++
			}
		}

		if len(newParts) == 0 {
			return "\x1b[m"
		}
		return "\x1b[" + strings.Join(newParts, ";") + "m"
	})
}

// langFromPath extracts the chroma language identifier from a file path.
func langFromPath(path string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	switch ext {
	case "go":
		return "go"
	case "js", "jsx":
		return "javascript"
	case "ts", "tsx":
		return "typescript"
	case "py":
		return "python"
	case "rs":
		return "rust"
	case "java":
		return "java"
	case "c", "h":
		return "c"
	case "cpp", "hpp", "cc":
		return "cpp"
	case "rb":
		return "ruby"
	case "sh", "bash", "zsh":
		return "bash"
	case "json":
		return "json"
	case "yaml", "yml":
		return "yaml"
	case "md", "markdown":
		return "markdown"
	case "html", "htm":
		return "html"
	case "css", "scss", "sass":
		return "css"
	case "sql":
		return "sql"
	case "xml":
		return "xml"
	case "dockerfile":
		return "dockerfile"
	case "makefile", "mk":
		return "makefile"
	default:
		return ext
	}
}

type diffLine struct {
	kind    string // "context", "remove", "add"
	content string
	oldNum  int // 1-based line number in old file (0 = not present)
	newNum  int // 1-based line number in new file (0 = not present)
	lineNum int // actual line number in the current file view (1-based from startLine)
}

// computeRawDiff computes a unified diff on raw strings with correct line numbers.
func computeRawDiff(oldLines, newLines []string, context int) []diffLine {
	if len(oldLines) == 0 && len(newLines) == 0 {
		return nil
	}

	// Find LCS matches
	matches := findMatches(oldLines, newLines)

	// Build diff by walking through both files
	var result []diffLine
	oldIdx, newIdx := 0, 0

	for oldIdx < len(oldLines) || newIdx < len(newLines) {
		if oldIdx < len(oldLines) && newIdx < len(newLines) && matches[oldIdx] == newIdx {
			// Context line — lineNum follows the new file position.
			result = append(result, diffLine{
				kind:    "context",
				content: oldLines[oldIdx],
				oldNum:  oldIdx + 1,
				newNum:  newIdx + 1,
				lineNum: newIdx + 1,
			})
			oldIdx++
			newIdx++
		} else if oldIdx < len(oldLines) && (newIdx >= len(newLines) || matches[oldIdx] == -1) {
			// Removed line — lineNum follows the old file position (git diff style).
			result = append(result, diffLine{
				kind:    "remove",
				content: oldLines[oldIdx],
				oldNum:  oldIdx + 1,
				lineNum: oldIdx + 1,
			})
			oldIdx++
		} else if newIdx < len(newLines) {
			// Added line — lineNum follows the new file position.
			result = append(result, diffLine{
				kind:    "add",
				content: newLines[newIdx],
				newNum:  newIdx + 1,
				lineNum: newIdx + 1,
			})
			newIdx++
		} else {
			break
		}
	}

	return filterContext(result, context)
}

// findMatches finds the best matching between old and new lines using a
// position-aware greedy algorithm that prefers lines close to the current
// position, avoiding incorrect matches when there are duplicate lines.
func findMatches(oldLines, newLines []string) []int {
	matches := make([]int, len(oldLines))
	for i := range matches {
		matches[i] = -1
	}

	if len(oldLines) == 0 || len(newLines) == 0 {
		return matches
	}

	used := make([]bool, len(newLines))
	for oldIdx := 0; oldIdx < len(oldLines); oldIdx++ {
		bestNewIdx := -1
		bestDist := len(newLines) + len(oldLines)
		for newIdx := 0; newIdx < len(newLines); newIdx++ {
			if used[newIdx] {
				continue
			}
			if oldLines[oldIdx] != newLines[newIdx] {
				continue
			}
			dist := abs(newIdx - oldIdx)
			if dist < bestDist {
				bestDist = dist
				bestNewIdx = newIdx
			}
		}
		if bestNewIdx != -1 {
			matches[oldIdx] = bestNewIdx
			used[bestNewIdx] = true
		}
	}

	return matches
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// longestCommonSubsequence finds the LCS of two string slices.
func longestCommonSubsequence(a, b []string) []string {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return []string{}
	}

	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] > dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	var lcs []string
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs = append([]string{a[i-1]}, lcs...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return lcs
}

// filterContext filters the diff to only show context lines around changes.
func filterContext(diff []diffLine, context int) []diffLine {
	if len(diff) == 0 {
		return diff
	}

	// Find indices of changed lines
	var changedIndices []int
	for i, dl := range diff {
		if dl.kind != "context" {
			changedIndices = append(changedIndices, i)
		}
	}

	if len(changedIndices) == 0 {
		return diff
	}

	// Determine which lines to show
	showLine := make([]bool, len(diff))
	for _, idx := range changedIndices {
		start := idx - context
		if start < 0 {
			start = 0
		}
		end := idx + context + 1
		if end > len(diff) {
			end = len(diff)
		}
		for i := start; i < end; i++ {
			showLine[i] = true
		}
	}

	// Build result, adding gap markers
	var result []diffLine
	inGap := false
	for i, dl := range diff {
		if showLine[i] {
			if inGap {
				inGap = false
			}
			result = append(result, dl)
		} else {
			if !inGap && len(result) > 0 {
				result = append(result, diffLine{
					kind:    "context",
					content: "...",
				})
				inGap = true
			}
		}
	}

	return result
}

// normalizeLines trims trailing empty lines and removes leading/trailing
// whitespace from each line for robust matching.
func normalizeLines(lines []string) []string {
	// Remove trailing empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// findLineSequence finds the first occurrence of needle lines in haystack lines
// and returns the starting line index, or -1 if not found.
func findLineSequence(haystack, needle []string) int {
	if len(needle) == 0 {
		return 0
	}

	for i := 0; i <= len(haystack)-len(needle); i++ {
		if lineSeqEqual(haystack[i:i+len(needle)], needle) {
			return i
		}
	}
	return -1
}

// lineSeqEqual compares two line slices for equality, ignoring leading/trailing
// whitespace differences on each line.
func lineSeqEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if strings.TrimSpace(a[i]) != strings.TrimSpace(b[i]) {
			return false
		}
	}
	return true
}

// RenderCodeDiffV2 renders a diff when the caller already knows the line numbers
// of newContent in the file. It reads only the necessary lines from disk using
// bufio.Scanner, avoiding a full file read into memory.
func (c *Chat) RenderCodeDiffV2(path, oldContent, newContent string, startLine int) string {
	oldLines := normalizeLines(strings.Split(oldContent, "\n"))
	newLines := normalizeLines(strings.Split(newContent, "\n"))

	// EndLine is computed from NewContent: StartLine + newline count
	endLine := startLine + strings.Count(newContent, "\n")

	contextStart := startLine - 5
	if contextStart < 1 {
		contextStart = 1
	}
	contextEnd := endLine + 5

	// Read only the lines we need from the file.
	file, err := os.Open(path)
	if err != nil {
		return c.renderSimpleDiff(path, oldContent, newContent, startLine)
	}
	defer file.Close()

	var contextLines []string
	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		if lineNo >= contextStart && lineNo <= contextEnd {
			contextLines = append(contextLines, scanner.Text())
		}
		if lineNo > contextEnd {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return c.renderSimpleDiff(path, oldContent, newContent, startLine)
	}

	// Determine boundaries within the sliced context.
	beforeStart := 0
	beforeEnd := startLine - contextStart         // exclusive
	afterStart := endLine - contextStart + 1      // first line after newContent
	afterEnd := len(contextLines)                 // exclusive

	beforeLines := make([]string, 0)
	beforeLines = append(beforeLines, contextLines[beforeStart:beforeEnd]...)
	beforeLines = append(beforeLines, oldLines...)
	if afterStart < afterEnd {
		beforeLines = append(beforeLines, contextLines[afterStart:afterEnd]...)
	}

	afterLines := make([]string, 0)
	afterLines = append(afterLines, contextLines[beforeStart:beforeEnd]...)
	afterLines = append(afterLines, newLines...)
	if afterStart < afterEnd {
		afterLines = append(afterLines, contextLines[afterStart:afterEnd]...)
	}

	diffLines := computeRawDiff(beforeLines, afterLines, 100)

	// Build line number mappings from diff position to actual file line.
	// beforeLines = contextBefore (contextStart..startLine-1) + oldLines + contextAfter (endLine+1..contextEnd)
	// afterLines  = contextBefore (contextStart..startLine-1) + newLines + contextAfter (endLine+1..contextEnd)
	beforeLineMap := buildLineMap(len(beforeLines), contextStart, startLine, endLine, len(oldLines), len(newLines), true)
	afterLineMap := buildLineMap(len(afterLines), contextStart, startLine, endLine, len(oldLines), len(newLines), false)

	beforeRendered := highlightCode(strings.Join(beforeLines, "\n"), langFromPath(path))
	afterRendered := highlightCode(strings.Join(afterLines, "\n"), langFromPath(path))

	var sb strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Foreground(styles.SidebarLabel).
		Background(styles.Background).
		Padding(1, 2).
		Render("← " + path)
	sb.WriteString(header)
	sb.WriteString("\n")

	// Top padding line
	padding := lipgloss.NewStyle().
		Background(styles.Background).
		Width(c.width - diffContainerOverhead).
		Render("")
	sb.WriteString(padding)
	sb.WriteString("\n")

	contentWidth := c.width - diffContainerOverhead - 7 - 2
	for _, dl := range diffLines {
		if dl.content == "" && dl.kind == "context" {
			continue
		}

		var highlighted string
		var lineNum int

		switch dl.kind {
		case "context":
			highlighted = getRenderedLine(beforeRendered, dl.oldNum)
			if highlighted == "" {
				highlighted = dl.content
			}
			// Context lines exist in both before and after; show the "after" line number
			// because that reflects the current file state.
			lineNum = afterLineMap[dl.newNum]
		case "remove":
			highlighted = getRenderedLine(beforeRendered, dl.oldNum)
			if highlighted == "" {
				highlighted = dl.content
			}
			lineNum = beforeLineMap[dl.oldNum]
		case "add":
			highlighted = getRenderedLine(afterRendered, dl.newNum)
			if highlighted == "" {
				highlighted = dl.content
			}
			lineNum = afterLineMap[dl.newNum]
		}

		switch dl.kind {
		case "context":
			numStr := fmt.Sprintf("%4d", lineNum)
			if dl.content == "..." {
				numStr = "   "
			}
			num := lipgloss.NewStyle().
				Foreground(styles.SidebarValue).
				Background(styles.Background).
				Width(4).
				Align(lipgloss.Right).
				Render(numStr)
			marker := lipgloss.NewStyle().
				Foreground(styles.SidebarValue).
				Background(styles.Background).
				Width(1).
				Render(" ")
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width-diffContainerOverhead, styles.Background)

		case "remove":
			num := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#c76b72")).
				Background(lipgloss.Color("#34232c")).
				Width(4).
				Align(lipgloss.Right).
				Render(fmt.Sprintf("%4d", lineNum))
			marker := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#c76b72")).
				Background(lipgloss.Color("#34232c")).
				Width(1).
				Render("-")
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width-diffContainerOverhead, lipgloss.Color("#34232c"))

		case "add":
			num := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#bfda90")).
				Background(lipgloss.Color("#23303a")).
				Width(4).
				Align(lipgloss.Right).
				Render(fmt.Sprintf("%4d", lineNum))
			marker := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#bfda90")).
				Background(lipgloss.Color("#23303a")).
				Width(1).
				Render("+")
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width-diffContainerOverhead, lipgloss.Color("#23303a"))
		}
	}

	return c.renderDiffContainer(sb.String())
}

// renderDiffContainer wraps diff content in a styled container with a subtle left border.
func (c *Chat) renderDiffContainer(content string) string {
	return lipgloss.NewStyle().
		Background(styles.Background).
		Width(c.width).
		PaddingLeft(2).
		BorderStyle(lipgloss.Border{Left: "┃"}).
		BorderLeft(true).
		BorderForeground(lipgloss.Color("#1a1a1a")).
		Render(content)
}

// buildLineMap creates a 1-based mapping from diff line position to actual file line number.
func buildLineMap(totalLines, contextStart, startLine, endLine, oldCount, newCount int, isBefore bool) map[int]int {
	m := make(map[int]int, totalLines)
	ctxBeforeCount := startLine - contextStart
	var changeCount int
	var afterStart int
	if isBefore {
		changeCount = oldCount
		afterStart = startLine + oldCount
	} else {
		changeCount = newCount
		afterStart = endLine + 1
	}
	for i := 1; i <= totalLines; i++ {
		switch {
		case i <= ctxBeforeCount:
			m[i] = contextStart + i - 1
		case i <= ctxBeforeCount+changeCount:
			m[i] = startLine + i - ctxBeforeCount - 1
		default:
			m[i] = afterStart + i - ctxBeforeCount - changeCount - 1
		}
	}
	return m
}
