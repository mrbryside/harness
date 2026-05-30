package components_test

import (
	"strings"
	"testing"

	"github.com/mrbryside/harness/tui/components"
)

func TestChatAppendUserMessage(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("user", "hello")
	if !strings.Contains(c.View(), "hello") {
		t.Errorf("expected View() to contain %q, got:\n%s", "hello", c.View())
	}
}

func TestChatAppendAssistantMessage(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("assistant", "hi")
	if !strings.Contains(c.View(), "hi") {
		t.Errorf("expected View() to contain %q, got:\n%s", "hi", c.View())
	}
}

func TestChatAppendChunkAppendsToLastMessage(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("assistant", "hello ")
	c.AppendChunk("world")
	if !strings.Contains(c.View(), "world") {
		t.Errorf("expected View() to contain chunk text %q, got:\n%s", "world", c.View())
	}
}

func TestChatBothMessagesPresent(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("user", "ping")
	c.AppendMessage("assistant", "pong")
	view := c.View()
	if !strings.Contains(view, "ping") {
		t.Errorf("expected View() to contain %q, got:\n%s", "ping", view)
	}
	if !strings.Contains(view, "pong") {
		t.Errorf("expected View() to contain %q, got:\n%s", "pong", view)
	}
}

func TestChatRendersMarkdownInAssistant(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("assistant", "## Hello")
	view := c.View()
	if strings.Contains(view, "## Hello") {
		t.Errorf("expected '## Hello' to be rendered (not verbatim), got:\n%s", view)
	}
	if !strings.Contains(view, "Hello") {
		t.Errorf("expected heading text 'Hello' to appear in View(), got:\n%s", view)
	}
}

func TestChatFencedCodeBlockEmptyLinesHaveBlackBg(t *testing.T) {
	c := components.NewChat(120, 30)
	c.AppendMessage("assistant",
		"```go\nfunc f() {\n\tif true {\n\t\treturn\n\t}\n}\n```")
	view := c.View()

	bad := []string{"\x1b[m    ", "\x1b[0m    "}
	for _, b := range bad {
		if strings.Contains(view, b) {
			t.Errorf("found SGR reset followed by un-styled padding spaces (grey strip): pattern %q present in view", b)
		}
	}
}

func TestChatUserMessageNotMarkdownRendered(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("user", "## Hello")
	view := c.View()
	if !strings.Contains(view, "## Hello") {
		t.Errorf("expected user message to remain verbatim, got:\n%s", view)
	}
}

func TestChatAssistantMessageHasThreeCharLeftMargin(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("assistant", "Hello world")
	view := c.View()

	// Find the line containing "Hello world" and verify it has 3 leading spaces.
	lines := strings.Split(view, "\n")
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Hello") {
			found = true
			clean := stripAnsi(line)
			if !strings.HasPrefix(clean, "   Hello") {
				t.Errorf("expected assistant line to start with 3-space margin, got: %q", clean)
			}
		}
	}
	if !found {
		t.Errorf("expected to find 'Hello' in view, got:\n%s", view)
	}
}

func TestChatAssistantMarkdownPreservesFormatWithMargin(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("assistant", "# Title\n\nParagraph text.")
	view := c.View()

	// Markdown should be rendered (no raw #).
	if strings.Contains(view, "# Title") {
		t.Errorf("expected markdown to be rendered, got raw markdown in view:\n%s", view)
	}

	// Both heading and paragraph should have the 3-char margin.
	lines := strings.Split(view, "\n")
	foundHeading := false
	foundParagraph := false
	for _, line := range lines {
		clean := stripAnsi(line)
		if strings.Contains(clean, "Title") {
			foundHeading = true
			if !strings.HasPrefix(clean, "   ") {
				t.Errorf("expected heading line to start with 3-space margin, got: %q", clean)
			}
		}
		if strings.Contains(clean, "Paragraph") {
			foundParagraph = true
			if !strings.HasPrefix(clean, "   ") {
				t.Errorf("expected paragraph line to start with 3-space margin, got: %q", clean)
			}
		}
	}
	if !foundHeading {
		t.Errorf("expected 'Title' in view, got:\n%s", view)
	}
	if !foundParagraph {
		t.Errorf("expected 'Paragraph' in view, got:\n%s", view)
	}
}

func TestChatAssistantMarginDoesNotExceedWidth(t *testing.T) {
	width := 40
	c := components.NewChat(width, 24)
	c.AppendMessage("assistant", "This is a longer message that should wrap within the chat width including the three character left margin.")
	view := c.View()

	lines := strings.Split(view, "\n")
	for _, line := range lines {
		if strings.TrimSpace(stripAnsi(line)) == "" {
			continue
		}
		w := len(stripAnsi(line))
		if w > width {
			t.Errorf("line width %d exceeds chat width %d: %q", w, width, stripAnsi(line))
		}
	}
}
