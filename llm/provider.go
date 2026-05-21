package llm

import "context"

// LLMProvider is the interface all LLM backends must implement.
// Complete streams response chunks via a channel; callers must drain
// the channel until it is closed or a Chunk with Done=true is received.
type LLMProvider interface {
	Name() string
	ChatCompletion(ctx context.Context, messages []Message, opts Options) (<-chan Chunk, error)
}

// Message represents a single turn in a conversation.
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// Chunk is a single streamed piece of a response.
type Chunk struct {
	Content    string
	TokensUsed int
	Done       bool
	Err        error
}

// Options controls per-request LLM behaviour.
type Options struct {
	Model       string
	Temperature float64
}
