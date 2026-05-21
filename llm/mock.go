package llm

import (
	"context"
	"math"
	"math/rand"
	"regexp"
	"time"
)

const mockStreamDelay = 40 * time.Millisecond

// tokenRe matches either a run of whitespace OR a run of non-whitespace.
// This preserves newlines and spacing when chunking the mock response,
// which is essential for markdown to render correctly during streaming.
var tokenRe = regexp.MustCompile(`\s+|\S+`)

// MockProvider is a fake LLMProvider that streams canned responses
// token-by-token while preserving original whitespace (including newlines).
type MockProvider struct{}

func (m *MockProvider) Name() string { return "gpt-4o" }

func (m *MockProvider) ChatCompletion(ctx context.Context, messages []Message, opts Options) (<-chan Chunk, error) {
	response := mockResponses[rand.Intn(len(mockResponses))]
	tokens := tokenRe.FindAllString(response, -1)

	// Estimate token count from non-whitespace tokens (rough heuristic).
	wordCount := 0
	for _, t := range tokens {
		if !isWhitespace(t) {
			wordCount++
		}
	}
	totalTokens := int(math.Round(float64(wordCount) * 1.3))

	ch := make(chan Chunk)

	go func() {
		defer close(ch)

		for i, tok := range tokens {
			isLast := i == len(tokens)-1

			select {
			case <-ctx.Done():
				ch <- Chunk{Err: ctx.Err(), Done: true}
				return
			case <-time.After(mockStreamDelay):
				ch <- Chunk{
					Content:    tok,
					TokensUsed: totalTokens,
					Done:       isLast,
				}
			}
		}
	}()

	return ch, nil
}

func isWhitespace(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}

var mockResponses = []string{
	// 1. Plain paragraph
	// "## Hello\n\nThis is a simple mock response. Nothing fancy here, just plain text to verify basic rendering works correctly.",

	// // 2. Go code block
	"## Code Example\n\nHere is a simple Go function:\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc fibonacci(n int) int {\n\tif n <= 1 {\n\t\treturn n\n\t}\n\treturn fibonacci(n-1) + fibonacci(n-2)\n}\n\nfunc main() {\n\tfmt.Println(fibonacci(10))\n}\n```\n\nThis recursively computes the nth Fibonacci number.",

	// // 3. Table
	// "## Comparison Table\n\n| Language | Typing   | GC  | Compiled |\n|----------|----------|-----|----------|\n| Go       | Static   | Yes | Yes      |\n| Python   | Dynamic  | Yes | No       |\n| Rust     | Static   | No  | Yes      |\n| JS       | Dynamic  | Yes | No       |\n\nGo strikes a balance between performance and developer ergonomics.",

	// // 4. Architecture overview
	// "## System Architecture\n\nThe application is split into three layers:\n\n### Presentation\nHandles all TUI rendering via Bubble Tea components.\n\n### Domain\nContains business logic — message history, provider selection, token tracking.\n\n### Infrastructure\nLLM provider implementations behind the `LLMProvider` interface.\n\nDependencies only flow inward: infrastructure → domain ← presentation.",

	// // 5. Bullet list
	// "## Key Features\n\n- **Streaming** responses via Go channels\n- **Markdown** rendering with syntax highlighting\n- **Multiple providers** behind a single interface\n- **Keyboard-driven** — no mouse required\n- **Themeable** via a central style package\n\nAll features are designed to work together without tight coupling.",

	// 6. Numbered steps
	"## Setup Steps\n\n1. Install Go 1.21 or later\n2. Run `go mod tidy` to fetch dependencies\n3. Set your API key: `export OPENAI_API_KEY=sk-...`\n4. Run the TUI: `go run .`\n5. Press `Ctrl+C` to quit\n\nThat's it — no Docker, no database, no config file required.",

	// // 7. Mixed code + table
	// "## Token Estimation\n\nTokens are estimated using a simple heuristic:\n\n```go\ntokens := int(math.Round(float64(len(words)) * 1.3))\n```\n\nApproximate token costs by model:\n\n| Model         | Input (per 1K) | Output (per 1K) |\n|---------------|----------------|-----------------|\n| gpt-4o        | $0.005         | $0.015          |\n| gpt-4o-mini   | $0.00015       | $0.0006         |\n| claude-3-5    | $0.003         | $0.015          |\n\nActual token counts may vary.",

	// // 8. Blockquote + list
	// "## Design Philosophy\n\n> Simple things should be simple. Complex things should be possible.\n\nThis drives every decision in the codebase:\n\n- Interfaces over concrete types\n- Channels over callbacks\n- Composition over inheritance\n- Explicit over implicit\n\nWhen in doubt, favour the approach that is easiest to delete.",

	// // 9. Multi-language code blocks
	// "## Cross-Language Hello World\n\n**Go:**\n```go\nfmt.Println(\"Hello, world!\")\n```\n\n**Python:**\n```python\nprint(\"Hello, world!\")\n```\n\n**Rust:**\n```rust\nfn main() {\n    println!(\"Hello, world!\");\n}\n```\n\n**TypeScript:**\n```typescript\nconsole.log(\"Hello, world!\");\n```",

	// // 10. Long prose
	// "## How Streaming Works\n\nWhen you send a message, the app calls `ChatCompletion` on the active `LLMProvider`. This returns a read-only channel of `Chunk` values.\n\nThe app model listens on this channel inside a `tea.Cmd`. Each chunk is dispatched as a `chunkMsg` back into the Bubble Tea event loop, where it is appended to the current assistant message and the chat viewport is re-rendered.\n\nThis means the UI never blocks — it simply processes one chunk at a time as they arrive. Cancellation is handled by closing the context passed to `ChatCompletion`, which causes the provider goroutine to exit cleanly and send a final chunk with `Done: true`.",
}
