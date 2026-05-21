package components_test

import (
	"strings"
	"testing"

	"github.com/mrbryside/harness/components"
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

// TestChatRendersMarkdownInAssistant verifies that assistant messages are
// rendered through glamour, so raw markdown syntax (e.g. "## ") is NOT
// present verbatim in the output, but the heading text itself still is.
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

// TestChatStripsInlineCodeBackground verifies that the grey background
// glamour applies to inline `code` spans is stripped, so inline code
// blends with the chat Background (no grey box). The foreground color
// must be preserved so the code is still visually distinguishable.
func TestChatStripsInlineCodeBackground(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("assistant", "use `go test` to run")
	view := c.View()
	// Glamour's inline code uses 256-color BG 236 ("\x1b[...;48;5;236m"
	// or "\x1b[48;5;236m"). It must not survive into the rendered output.
	if strings.Contains(view, "48;5;236") {
		t.Errorf("expected inline code BG 48;5;236 to be stripped, got:\n%q", view)
	}
	// FG (color 203) for inline code should remain so it's distinguishable.
	if !strings.Contains(view, "38;5;203") {
		t.Errorf("expected inline code FG 38;5;203 to be preserved, got:\n%q", view)
	}
	// The text itself must still be there.
	if !strings.Contains(view, "go test") {
		t.Errorf("expected 'go test' text in view, got:\n%s", view)
	}
}

// TestChatFencedCodeBlockEmptyLinesHaveBlackBg guards against the
// regression where empty lines *between* code lines inside a fenced
// code block were padded with plain spaces (no background SGR), so
// terminals would render that padding with their default background
// — appearing as a grey strip on every blank line inside the block.
//
// We verify the rendered chat output never contains an SGR reset
// followed by a run of plain spaces; every space cell after a reset
// must be preceded by the chat-background SGR.
func TestChatFencedCodeBlockEmptyLinesHaveBlackBg(t *testing.T) {
	c := components.NewChat(120, 30)
	c.AppendMessage("assistant",
		"```go\nfunc f() {\n\tif true {\n\t\treturn\n\t}\n}\n```")
	view := c.View()

	// Look for "\x1b[m" or "\x1b[0m" immediately followed by 4+ plain
	// spaces (i.e. padding without an intervening BG re-assertion).
	bad := []string{"\x1b[m    ", "\x1b[0m    "}
	for _, b := range bad {
		if strings.Contains(view, b) {
			t.Errorf("found SGR reset followed by un-styled padding spaces (grey strip): pattern %q present in view", b)
		}
	}
}
// verbatim (no markdown rendering) — users see exactly what they typed.
func TestChatUserMessageNotMarkdownRendered(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("user", "## Hello")
	view := c.View()
	if !strings.Contains(view, "## Hello") {
		t.Errorf("expected user message to remain verbatim, got:\n%s", view)
	}
}
