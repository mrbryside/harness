package components_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mrbryside/harness/tui/components"
)

func TestChatScrollDebug(t *testing.T) {
	c := components.NewChat(80, 10)

	// Add content
	for i := 0; i < 30; i++ {
		c.AppendMessage("assistant", fmt.Sprintf("line %d", i))
	}

	t.Logf("Initial YOffset: %d", c.YOffset())

	// Scroll to top first
	for !c.AtTop() {
		c.ScrollUpAndExtend(1, 0)
	}
	t.Logf("After scrolling to top, YOffset: %d", c.YOffset())

	// Now start selection at content line 5 (which is visible)
	c.SelectStart(5, 0)
	c.SelectExtend(7, 10)
	t.Logf("After select 5-7, text: %q", c.SelectedText())

	// Scroll down
	c.ScrollDownAndExtend(5, 10)
	t.Logf("After scroll down 5, YOffset: %d", c.YOffset())
	t.Logf("After scroll down, text: %q", c.SelectedText())
}

func TestChatSelectionContentRelative(t *testing.T) {
	c := components.NewChat(80, 10)

	// Add enough content so we can scroll
	for i := 0; i < 30; i++ {
		c.AppendMessage("assistant", fmt.Sprintf("content-%d", i))
	}

	// Scroll to top first (adding messages auto-scrolls to bottom)
	for !c.AtTop() {
		c.ScrollUpAndExtend(1, 0)
	}

	// Start selection at content line 5 (visible at startup since YOffset=0)
	c.SelectStart(5, 0)
	c.SelectExtend(7, 8)

	before := c.SelectedText()
	if !strings.Contains(before, "content-") {
		t.Fatalf("expected selected text to contain content, got:\n%q", before)
	}

	// Scroll down by 5 lines — viewport now shows content lines 5-14
	c.ScrollDownAndExtend(5, 8)

	// The selection should have extended to the bottom of the new viewport
	text := c.SelectedText()
	if !strings.Contains(text, "content-") {
		t.Fatalf("expected selected text to contain content after scroll, got:\n%q", text)
	}
}

func TestChatCopyIncludesScrolledContent(t *testing.T) {
	c := components.NewChat(80, 10)

	// Add content
	for i := 0; i < 30; i++ {
		c.AppendMessage("assistant", fmt.Sprintf("content-%d", i))
	}

	// Scroll to top
	for !c.AtTop() {
		c.ScrollUpAndExtend(1, 0)
	}

	// Select content lines 5-7
	c.SelectStart(5, 0)
	c.SelectExtend(7, 10)

	// Scroll down by 10 lines
	c.ScrollDownAndExtend(10, 10)

	// SelectedText should include content from line 5 (now scrolled out of view)
	// through the extended area
	text := c.SelectedText()
	if !strings.Contains(text, "content-") {
		t.Fatalf("expected selected text to include scrolled content, got:\n%q", text)
	}

	// Should contain at least a few lines
	lines := strings.Count(text, "\n")
	if lines < 2 {
		t.Fatalf("expected multiple lines, got %d lines:\n%q", lines, text)
	}
}