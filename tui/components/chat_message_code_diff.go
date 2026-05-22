package components

import (
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

func (c *Chat) renderCodeDiffMessage(msg chatMessage) string {
	return c.RenderCodeDiff(msg.diffPath, msg.diffOld, msg.diffNew) + messageGap
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

// RenderCodeDiff renders a unified diff view with syntax highlighting in git style.
// It reads the file at path, locates oldContent within it, grabs 5 lines of context
// above and below, and renders the replacement against newContent with correct line numbers.
func (c *Chat) RenderCodeDiff(path string, oldContent, newContent string) string {
	// Try to read the actual file so we can show real line numbers.
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return c.renderSimpleDiff(path, oldContent, newContent)
	}
	fileContent := string(fileBytes)

	// Find oldContent in the file (assume exactly 1 match).
	matchIdx := strings.Index(fileContent, oldContent)
	if matchIdx == -1 {
		return c.renderSimpleDiff(path, oldContent, newContent)
	}

	fileLines := strings.Split(fileContent, "\n")
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	// Locate the start line of oldContent in the file.
	startLine := 0
	charCount := 0
	for i, line := range fileLines {
		lineEnd := charCount + len(line)
		if matchIdx >= charCount && matchIdx < lineEnd+1 { // +1 for the newline
			startLine = i
			break
		}
		charCount = lineEnd + 1 // +1 for \n
	}

	endLine := startLine + len(oldLines) - 1

	// Grab 5 lines of context above and below.
	contextStart := startLine - 5
	if contextStart < 0 {
		contextStart = 0
	}
	contextEnd := endLine + 5
	if contextEnd >= len(fileLines) {
		contextEnd = len(fileLines) - 1
	}

	// Build the "before" slice: context above + old + context below.
	beforeLines := make([]string, 0)
	beforeLines = append(beforeLines, fileLines[contextStart:startLine]...)
	beforeLines = append(beforeLines, oldLines...)
	beforeLines = append(beforeLines, fileLines[endLine+1:contextEnd+1]...)

	// Build the "after" slice: context above + new + context below.
	afterLines := make([]string, 0)
	afterLines = append(afterLines, fileLines[contextStart:startLine]...)
	afterLines = append(afterLines, newLines...)
	afterLines = append(afterLines, fileLines[endLine+1:contextEnd+1]...)

	// Compute diff with a large context so nothing is hidden.
	diffLines := computeRawDiff(beforeLines, afterLines, 100)

	// The first real line in beforeLines corresponds to contextStart+1 in the file.
	baseLineNum := contextStart + 1

	// Render both versions with chroma for syntax highlighting.
	beforeRendered := highlightCode(strings.Join(beforeLines, "\n"), langFromPath(path))
	afterRendered := highlightCode(strings.Join(afterLines, "\n"), langFromPath(path))

	var sb strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Foreground(styles.SidebarLabel).
		Background(styles.Background).
		Padding(0, 2).
		Render("← " + path)
	sb.WriteString(header)
	sb.WriteString("\n")

	// Render each diff line with line numbers offset by baseLineNum.
	contentWidth := c.width - 6
	for _, dl := range diffLines {
		if dl.content == "" && dl.kind == "context" {
			continue
		}

		var highlighted string
		var lineNum int

		switch dl.kind {
		case "context", "remove":
			highlighted = getRenderedLine(beforeRendered, dl.oldNum)
			if highlighted == "" {
				highlighted = dl.content
			}
			lineNum = baseLineNum + dl.oldNum - 1
		case "add":
			highlighted = getRenderedLine(afterRendered, dl.newNum)
			if highlighted == "" {
				highlighted = dl.content
			}
			lineNum = baseLineNum + dl.newNum - 1
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
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width, styles.Background)

		case "remove":
			num := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ff6b6b")).
				Background(lipgloss.Color("#3c2c2c")).
				Width(4).
				Align(lipgloss.Right).
				Render(fmt.Sprintf("%4d", lineNum))
			marker := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ff6b6b")).
				Background(lipgloss.Color("#3c2c2c")).
				Width(1).
				Render("-")
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width, lipgloss.Color("#3c2c2c"))

		case "add":
			num := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6bff6b")).
				Background(lipgloss.Color("#2c3c2c")).
				Width(4).
				Align(lipgloss.Right).
				Render(fmt.Sprintf("%4d", lineNum))
			marker := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6bff6b")).
				Background(lipgloss.Color("#2c3c2c")).
				Width(1).
				Render("+")
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width, lipgloss.Color("#2c3c2c"))
		}
	}

	// Wrap the entire diff in a container with user-chat background.
	container := lipgloss.NewStyle().
		Background(styles.Background).
		Width(c.width).
		Render(sb.String())

	return container
}

// renderSimpleDiff is the fallback when the file can't be read or oldContent isn't found.
func (c *Chat) renderSimpleDiff(path string, oldContent, newContent string) string {
	oldRaw := strings.Split(oldContent, "\n")
	newRaw := strings.Split(newContent, "\n")
	diffLines := computeRawDiff(oldRaw, newRaw, 3)

	oldRendered := highlightCode(oldContent, langFromPath(path))
	newRendered := highlightCode(newContent, langFromPath(path))

	var sb strings.Builder

	header := lipgloss.NewStyle().
		Foreground(styles.SidebarLabel).
		Background(styles.Background).
		Padding(0, 2).
		Render("← " + path)
	sb.WriteString(header)
	sb.WriteString("\n")

	contentWidth := c.width - 6
	for _, dl := range diffLines {
		if dl.content == "" && dl.kind == "context" {
			continue
		}

		var highlighted string
		switch dl.kind {
		case "context", "remove":
			highlighted = getRenderedLine(oldRendered, dl.oldNum)
			if highlighted == "" {
				highlighted = dl.content
			}
		case "add":
			highlighted = getRenderedLine(newRendered, dl.newNum)
			if highlighted == "" {
				highlighted = dl.content
			}
		}

		switch dl.kind {
		case "context":
			numStr := fmt.Sprintf("%4d", dl.newNum)
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
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width, styles.Background)

		case "remove":
			num := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ff6b6b")).
				Background(lipgloss.Color("#3c2c2c")).
				Width(4).
				Align(lipgloss.Right).
				Render(fmt.Sprintf("%4d", dl.oldNum))
			marker := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ff6b6b")).
				Background(lipgloss.Color("#3c2c2c")).
				Width(1).
				Render("-")
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width, lipgloss.Color("#3c2c2c"))

		case "add":
			num := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6bff6b")).
				Background(lipgloss.Color("#2c3c2c")).
				Width(4).
				Align(lipgloss.Right).
				Render(fmt.Sprintf("%4d", dl.newNum))
			marker := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6bff6b")).
				Background(lipgloss.Color("#2c3c2c")).
				Width(1).
				Render("+")
			wrapAndRenderLine(&sb, num, marker, highlighted, contentWidth, c.width, lipgloss.Color("#2c3c2c"))
		}
	}

	// Wrap the entire diff in a container with user-chat background.
	container := lipgloss.NewStyle().
		Background(styles.Background).
		Width(c.width).
		Render(sb.String())

	return container
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
	for i, chunk := range chunks {
		var line string
		if i == 0 {
			// First row: num + space + marker + space + content
			line = num + " " + marker + " " + chunk
		} else {
			// Continuation rows: indent to align with content column
			indent := lipgloss.NewStyle().Background(bgColor).Render(strings.Repeat(" ", prefixWidth))
			line = indent + chunk
		}
		// Pad the entire line to full width so the background spans edge-to-edge.
		style := lipgloss.NewStyle().Background(bgColor).Width(totalWidth)
		sb.WriteString(style.Render(line))
		sb.WriteString("\n")
	}
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
	count := 0
	for _, line := range rendered {
		if isVisualEmpty(line) {
			continue
		}
		count++
		if count == n {
			return line
		}
	}
	return ""
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
			// Context line
			result = append(result, diffLine{
				kind:    "context",
				content: oldLines[oldIdx],
				oldNum:  oldIdx + 1,
				newNum:  newIdx + 1,
			})
			oldIdx++
			newIdx++
		} else if oldIdx < len(oldLines) && (newIdx >= len(newLines) || matches[oldIdx] == -1) {
			// Removed line
			result = append(result, diffLine{
				kind:    "remove",
				content: oldLines[oldIdx],
				oldNum:  oldIdx + 1,
			})
			oldIdx++
		} else if newIdx < len(newLines) {
			// Added line
			result = append(result, diffLine{
				kind:    "add",
				content: newLines[newIdx],
				newNum:  newIdx + 1,
			})
			newIdx++
		} else {
			break
		}
	}

	return filterContext(result, context)
}

// findMatches finds the best matching between old and new lines using LCS.
func findMatches(oldLines, newLines []string) []int {
	matches := make([]int, len(oldLines))
	for i := range matches {
		matches[i] = -1
	}

	if len(oldLines) == 0 || len(newLines) == 0 {
		return matches
	}

	lcs := longestCommonSubsequence(oldLines, newLines)

	oldIdx, newIdx := 0, 0
	for _, lcsLine := range lcs {
		for oldIdx < len(oldLines) && oldLines[oldIdx] != lcsLine {
			oldIdx++
		}
		for newIdx < len(newLines) && newLines[newIdx] != lcsLine {
			newIdx++
		}
		if oldIdx < len(oldLines) && newIdx < len(newLines) {
			matches[oldIdx] = newIdx
			oldIdx++
			newIdx++
		}
	}

	return matches
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
