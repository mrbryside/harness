package components_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/mrbryside/harness/tui/components"
)

func TestToolEditRendersFilePath(t *testing.T) {
	c := components.NewChat(80, 24)
	
	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`
	
	diff := c.RenderToolEdit("test.go", oldCode, newCode, 1)
	
	if !strings.Contains(diff, "test.go") {
		t.Errorf("expected diff to contain file path, got:\n%s", diff)
	}
}

func TestToolEditShowsRemovedLine(t *testing.T) {
	c := components.NewChat(80, 24)
	
	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`
	
	diff := c.RenderToolEdit("test.go", oldCode, newCode, 1)
	
	if !strings.Contains(diff, "old") {
		t.Errorf("expected diff to contain removed 'old' line, got:\n%s", diff)
	}
}

func TestToolEditShowsAddedLine(t *testing.T) {
	c := components.NewChat(80, 24)
	
	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`
	
	diff := c.RenderToolEdit("test.go", oldCode, newCode, 1)
	
	if !strings.Contains(diff, "new") {
		t.Errorf("expected diff to contain added 'new' line, got:\n%s", diff)
	}
}

func TestToolEditHighlightWidth(t *testing.T) {
	c := components.NewChat(80, 24)
	
	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`
	
	diff := c.RenderToolEdit("test.go", oldCode, newCode, 1)
	lines := strings.Split(diff, "\n")
	
	// Find an added line (should have green background and full width)
	for _, line := range lines {
		if strings.Contains(line, "println(\"new\")") {
			// The line should extend to near the full width (80 - 6 = 74 for content)
			// We check that there's padding (spaces with background) after the content
			if !strings.Contains(line, "\x1b[48;2;") {
				t.Errorf("expected added line to have background color, got:\n%s", line)
			}
			break
		}
	}
}

func TestToolEditLineWidthEqualsViewport(t *testing.T) {
	c := components.NewChat(80, 24)

	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`

	diff := c.RenderToolEdit("test.go", oldCode, newCode, 1)
	lines := strings.Split(diff, "\n")

	// Find an added line and verify it's exactly the viewport width
	for _, line := range lines {
		if strings.Contains(line, "+") && strings.Contains(line, "println") {
			// Strip ANSI to measure visible width
			clean := stripAnsi(line)
			// Account for left border character
			clean = strings.TrimPrefix(clean, "┃")
			// The lipgloss Width() call pads to c.width minus border width (1)
			if len(clean) != 79 {
				t.Errorf("expected line width 79 (80 - 1 for border), got %d: %q", len(clean), clean)
			}
			break
		}
	}
}

func stripAnsi(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func TestToolEditNoBareResets(t *testing.T) {
	c := components.NewChat(80, 24)

	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`

	diff := c.RenderToolEdit("test.go", oldCode, newCode, 1)
	lines := strings.Split(diff, "\n")

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Bare reset followed by a space means lipgloss padding lost its background.
		if strings.Contains(line, "\x1b[m ") || strings.Contains(line, "\x1b[0m ") {
			t.Errorf("line %d has bare reset followed by space (background leak):\n%s", i, line)
		}
	}
}

func TestToolEditHasTopPadding(t *testing.T) {
	c := components.NewChat(80, 24)

	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`

	diff := c.RenderToolEdit("test.go", oldCode, newCode, 1)
	lines := strings.Split(diff, "\n")

	// Find the header line (should be within first few lines)
	headerIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "test.go") {
			headerIdx = i
			break
		}
	}

	if headerIdx == -1 {
		t.Errorf("expected diff to contain header with file path")
		return
	}

	// There should be a blank/padding line between header and code
	// (at least one line with only background color and spaces after header)
	// Account for left border character.
	foundPadding := false
	for i := headerIdx + 1; i < len(lines) && i < headerIdx+5; i++ {
		clean := stripAnsi(lines[i])
		// Strip the border character "┃" as well
		clean = strings.TrimPrefix(clean, "┃")
		if strings.TrimSpace(clean) == "" {
			foundPadding = true
			break
		}
	}

	if !foundPadding {
		t.Errorf("expected padding/margin between header and code. First few lines:\n%s", strings.Join(lines[:5], "\n"))
	}
}

func TestToolEditAddColors(t *testing.T) {
	c := components.NewChat(80, 24)

	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`

	diff := c.RenderToolEdit("test.go", oldCode, newCode, 1)

	// Add background should be rgb(35,48,58) = #23303a
	if !strings.Contains(diff, "\x1b[48;2;35;48;58") {
		t.Error("expected add line background to be rgb(35,48,58) (#23303a)")
	}

	// Add symbol (+) should be rgb(191,218,144) = #bfda90
	if !strings.Contains(diff, "\x1b[38;2;191;218;144") {
		t.Error("expected add symbol (+) foreground to be rgb(191,218,144) (#bfda90)")
	}
}

func TestToolEditRemoveColors(t *testing.T) {
	c := components.NewChat(80, 24)

	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`

	diff := c.RenderToolEdit("test.go", oldCode, newCode, 1)

	// Remove background should be rgb(52,35,44) = #34232c
	if !strings.Contains(diff, "\x1b[48;2;52;35;44") {
		t.Error("expected remove line background to be rgb(52,35,44) (#34232c)")
	}

	// Remove symbol (-) should be rgb(199,107,114) = #c76b72
	if !strings.Contains(diff, "\x1b[38;2;199;107;114") {
		t.Error("expected remove symbol (-) foreground to be rgb(199,107,114) (#c76b72)")
	}
}

func TestToolEditWrappedLinesHaveBackground(t *testing.T) {
	// Use a narrow width so content wraps
	c := components.NewChat(40, 24)

	// Long line that will wrap
	oldCode := `func hello() {
	println("this is a very long line that should wrap to multiple rows")
}`
	newCode := `func hello() {
	println("this is a very long line that should wrap to multiple rows with new content")
}`

	diff := c.RenderToolEdit("test.go", oldCode, newCode, 1)
	lines := strings.Split(diff, "\n")

	// Find the added line rows - they should all have add-line background
	foundAdded := false
	for i, line := range lines {
		if strings.Contains(line, "+") && strings.Contains(line, "println") {
			foundAdded = true
			// This row and the next continuation rows should have add bg
			if !strings.Contains(line, "\x1b[48;2;35;48;58") {
				t.Errorf("added line row %d missing add-line background:\n%s", i, line)
			}
		} else if foundAdded && strings.Contains(line, "\x1b[48;2;35;48;58") {
			// Continuation row with add-line background - good
			continue
		} else if foundAdded {
			// We've passed the added lines
			break
		}
	}

	if !foundAdded {
		t.Error("did not find added line in diff")
	}
}

func TestToolEditStartLineOnlyComputesEndLine(t *testing.T) {
	c := components.NewChat(80, 24)

	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	// line 1
	// line 2
	println("new")
}`

	// Agent only passes StartLine=1, framework should compute EndLine=4
	// (StartLine 1 + 3 newlines = 4 lines total)
	diff := c.RenderToolEdit("test.go", oldCode, newCode, 1)

	// Verify the diff contains all 4 new lines
	cleanDiff := stripAnsi(diff)
	if !strings.Contains(cleanDiff, "// line 1") {
		t.Errorf("expected diff to contain '// line 1', got:\n%s", cleanDiff)
	}
	if !strings.Contains(cleanDiff, "// line 2") {
		t.Errorf("expected diff to contain '// line 2', got:\n%s", cleanDiff)
	}
	if !strings.Contains(cleanDiff, `println("new")`) {
		t.Errorf("expected diff to contain 'println(\"new\")', got:\n%s", cleanDiff)
	}
	// Verify old content is also present (as removed lines)
	if !strings.Contains(cleanDiff, `println("old")`) {
		t.Errorf("expected diff to contain removed 'println(\"old\")', got:\n%s", cleanDiff)
	}
}

func TestToolEditSingleLineStartLineOnly(t *testing.T) {
	c := components.NewChat(80, 24)

	oldCode := `return 42`
	newCode := `return 43`

	// Single line, StartLine=5, EndLine should be 5
	// (StartLine 5 + 0 newlines = 1 line total, so EndLine = 5)
	diff := c.RenderToolEdit("test.go", oldCode, newCode, 5)

	// Verify both old and new content appear
	cleanDiff := stripAnsi(diff)
	if !strings.Contains(cleanDiff, "return 42") {
		t.Errorf("expected diff to contain removed 'return 42', got:\n%s", cleanDiff)
	}
	if !strings.Contains(cleanDiff, "return 43") {
		t.Errorf("expected diff to contain added 'return 43', got:\n%s", cleanDiff)
	}
}

func TestToolEditRemovedLinesUseOldFileNumbers(t *testing.T) {
	c := components.NewChat(80, 24)

	oldCode := `func hello() {
	println("old")
}
func world() {}`

	newCode := `// New hello
func hello() {
	println("new")
}
func world() {}`

	// startLine=10 means old lines are 10,11,12,13 and new lines are 10,11,12,13
	diff := c.RenderToolEdit("test.go", oldCode, newCode, 10)
	cleanDiff := stripAnsi(diff)

	// The removed line "println("old")" was at old file line 11
	// It should show line number 11 (old file), not 12 (new file position)
	if !strings.Contains(cleanDiff, "  11") {
		t.Errorf("expected removed line to show old file line number 11, got:\n%s", cleanDiff)
	}
}

func TestToolEditAddedLinesUseNewFileNumbers(t *testing.T) {
	c := components.NewChat(80, 24)

	oldCode := `func hello() {
	println("old")
}`

	newCode := `// New comment
// Another comment
func hello() {
	println("new")
}`

	// startLine=5: old lines 5,6,7; new lines 5,6,7,8
	diff := c.RenderToolEdit("test.go", oldCode, newCode, 5)
	cleanDiff := stripAnsi(diff)

	// Added lines should show new file line numbers: 5, 6, 7, 8
	if !strings.Contains(cleanDiff, "   5") {
		t.Errorf("expected added line to show new file line number 5, got:\n%s", cleanDiff)
	}
}


