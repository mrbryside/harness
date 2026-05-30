package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Conversation struct {
	client   *Client
	registry *ToolRegistry
	messages []openai.ChatCompletionMessage
}

func NewConversation(client *Client, registry *ToolRegistry) *Conversation {
	return &Conversation{
		client:   client,
		registry: registry,
		messages: make([]openai.ChatCompletionMessage, 0),
	}
}

func (c *Conversation) Send(ctx context.Context, userMessage string) error {
	// Add user message
	c.messages = append(c.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMessage,
	})

	fmt.Printf("\n👤 User: %s\n", userMessage)

	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		fmt.Printf("\n🤖 AI Thinking...\n")

		resp, err := c.client.Chat(ctx, c.messages, c.registry.GetOpenAITools())
		if err != nil {
			return err
		}

		choice := resp.Choices[0]
		message := choice.Message

		// Show reasoning/thinking if present
		if message.ReasoningContent != "" {
			fmt.Printf("\n💭 Thinking:\n%s\n", formatThinking(message.ReasoningContent))
		}

		// Add assistant message to history
		c.messages = append(c.messages, message)

		// Check if tool calls are needed
		if len(message.ToolCalls) > 0 {
			fmt.Printf("\n🔧 Tool Calls (%d):\n", len(message.ToolCalls))

			for _, tc := range message.ToolCalls {
				fmt.Printf("  📞 Calling: %s\n", tc.Function.Name)
				fmt.Printf("  📋 Arguments: %s\n", tc.Function.Arguments)

				// Execute tool
				result, err := c.registry.Execute(tc.Function.Name, json.RawMessage(tc.Function.Arguments))
				if err != nil {
					fmt.Printf("  ❌ Error: %v\n", err)
					result = fmt.Sprintf("Error: %v", err)
				}

				fmt.Printf("  ✅ Result:\n%s\n", formatResult(result))

				// Add tool result to messages
				c.messages = append(c.messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    result,
					ToolCallID: tc.ID,
				})
			}

			// Continue loop to let AI process results
			continue
		}

		// No tool calls - show final response
		if message.Content != "" {
			fmt.Printf("\n💬 Response:\n%s\n", message.Content)
		}

		break
	}

	return nil
}

func formatThinking(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, "    "+line)
	}
	return strings.Join(result, "\n")
}

func formatResult(result string) string {
	lines := strings.Split(result, "\n")
	var formatted []string
	for _, line := range lines {
		formatted = append(formatted, "      "+line)
	}
	return strings.Join(formatted, "\n")
}

// CLI mode removed - use server.go for web interface
