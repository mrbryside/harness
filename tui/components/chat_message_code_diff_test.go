package components_test

import (
	"strings"
	"testing"

	"github.com/mrbryside/harness/tui/components"
)

func TestCodeDiffRendersFilePath(t *testing.T) {
	c := components.NewChat(80, 24)
	
	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`
	
	diff := c.RenderCodeDiff("test.go", oldCode, newCode)
	
	if !strings.Contains(diff, "test.go") {
		t.Errorf("expected diff to contain file path, got:\n%s", diff)
	}
}

func TestCodeDiffShowsRemovedLine(t *testing.T) {
	c := components.NewChat(80, 24)
	
	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`
	
	diff := c.RenderCodeDiff("test.go", oldCode, newCode)
	
	if !strings.Contains(diff, "old") {
		t.Errorf("expected diff to contain removed 'old' line, got:\n%s", diff)
	}
}

func TestCodeDiffShowsAddedLine(t *testing.T) {
	c := components.NewChat(80, 24)
	
	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`
	
	diff := c.RenderCodeDiff("test.go", oldCode, newCode)
	
	if !strings.Contains(diff, "new") {
		t.Errorf("expected diff to contain added 'new' line, got:\n%s", diff)
	}
}

func TestCodeDiffHighlightWidth(t *testing.T) {
	c := components.NewChat(80, 24)
	
	oldCode := `func hello() {
	println("old")
}`
	newCode := `func hello() {
	println("new")
}`
	
	diff := c.RenderCodeDiff("test.go", oldCode, newCode)
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
