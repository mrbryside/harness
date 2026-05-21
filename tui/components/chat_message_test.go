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

func TestChatStripsInlineCodeBackground(t *testing.T) {
	c := components.NewChat(80, 24)
	c.AppendMessage("assistant", "use `go test` to run")
	view := c.View()
	if strings.Contains(view, "48;5;236") {
		t.Errorf("expected inline code BG 48;5;236 to be stripped, got:\n%q", view)
	}
	if !strings.Contains(view, "38;5;203") {
		t.Errorf("expected inline code FG 38;5;203 to be preserved, got:\n%q", view)
	}
	if !strings.Contains(view, "go test") {
		t.Errorf("expected 'go test' text in view, got:\n%s", view)
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
