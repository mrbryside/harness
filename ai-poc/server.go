package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
)

var globalHistory = NewHistoryStore()

type ChatRequest struct {
	Message string `json:"message"`
}

type SSEEvent struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Name    string `json:"name,omitempty"`
	Args    string `json:"args,omitempty"`
	Result  string `json:"result,omitempty"`
}

func sendSSE(w io.Writer, event SSEEvent) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendSSE(w, SSEEvent{Type: "error", Content: err.Error()})
		return
	}

	client := NewClient()
	registry := NewToolRegistry()
	RegisterInternalTools(registry)
	RegisterMCPTools(registry)

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: req.Message},
	}

	ctx := r.Context()
	maxIterations := 5

	for i := 0; i < maxIterations; i++ {
		// Build API request for history
		rawTools := make([]json.RawMessage, 0, len(registry.GetOpenAITools()))
		for _, tool := range registry.GetOpenAITools() {
			toolJSON, _ := json.Marshal(tool)
			rawTools = append(rawTools, toolJSON)
		}

		rawMessages := make([]json.RawMessage, 0, len(messages))
		for _, msg := range messages {
			msgJSON, _ := json.Marshal(msg)
			rawMessages = append(rawMessages, msgJSON)
		}

		round := APIRound{
			ID:        fmt.Sprintf("round-%d-%d", time.Now().Unix(), i),
			Timestamp: time.Now(),
			Request: &APIRequest{
				Model:    client.model,
				Messages: rawMessages,
				Tools:    rawTools,
				Stream:   true,
			},
			Response: &APIResponse{
				RawChunks: make([]map[string]interface{}, 0),
			},
		}

		stream, err := client.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model:    client.model,
			Messages: messages,
			Tools:    registry.GetOpenAITools(),
			Stream:   true,
		})
		if err != nil {
			globalHistory.AddRound(round)
			sendSSE(w, SSEEvent{Type: "error", Content: fmt.Sprintf("Stream error: %v", err)})
			return
		}

		var assistantMessage openai.ChatCompletionMessage
		assistantMessage.Role = openai.ChatMessageRoleAssistant

		toolCallBuffer := make(map[int]openai.ToolCall)

		for {
			response, err := stream.Recv()
			if err != nil {
				break
			}

			// Store raw chunk for history
			var rawChunk map[string]interface{}
			chunkJSON, _ := json.Marshal(response)
			json.Unmarshal(chunkJSON, &rawChunk)
			round.Response.RawChunks = append(round.Response.RawChunks, rawChunk)

			if len(response.Choices) == 0 {
				continue
			}

			choice := response.Choices[0]
			delta := choice.Delta

			// Stream reasoning content
			if delta.ReasoningContent != "" {
				round.Response.Reasoning += delta.ReasoningContent
				sendSSE(w, SSEEvent{Type: "reasoning", Content: delta.ReasoningContent})
			}

			// Stream text content
			if delta.Content != "" {
				assistantMessage.Content += delta.Content
				sendSSE(w, SSEEvent{Type: "text", Content: delta.Content})
			}

			// Accumulate tool calls
			for _, tc := range delta.ToolCalls {
				if tc.Index != nil && *tc.Index >= 0 {
					idx := *tc.Index
					existing, ok := toolCallBuffer[idx]
					if !ok {
						toolCallBuffer[idx] = openai.ToolCall{
							ID:   tc.ID,
							Type: tc.Type,
							Function: openai.FunctionCall{
								Name: tc.Function.Name,
							},
						}
					} else {
						existing.Function.Arguments += tc.Function.Arguments
						toolCallBuffer[idx] = existing
					}
				}
			}

			if choice.FinishReason != "" {
				round.Response.FinishReason = string(choice.FinishReason)
				break
			}
		}
		stream.Close()

		// Convert buffer to slice
		var toolCalls []openai.ToolCall
		for idx := 0; idx < len(toolCallBuffer); idx++ {
			if tc, ok := toolCallBuffer[idx]; ok {
				toolCalls = append(toolCalls, tc)
			}
		}
		assistantMessage.ToolCalls = toolCalls

		// Execute tool calls if any
		if len(assistantMessage.ToolCalls) > 0 {
			messages = append(messages, assistantMessage)

			for _, tc := range assistantMessage.ToolCalls {
			toolInfo := ToolCallInfo{
				ID:        tc.ID,
				Type:      string(tc.Type),
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}

				sendSSE(w, SSEEvent{
					Type: "tool_call",
					Name: tc.Function.Name,
					Args: tc.Function.Arguments,
				})

				result, err := registry.Execute(tc.Function.Name, json.RawMessage(tc.Function.Arguments))
				if err != nil {
					result = fmt.Sprintf("Error: %v", err)
				}

				toolInfo.Result = result
				round.Response.ToolCalls = append(round.Response.ToolCalls, toolInfo)

				sendSSE(w, SSEEvent{
					Type:   "tool_result",
					Name:   tc.Function.Name,
					Result: result,
				})

				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    result,
					ToolCallID: tc.ID,
				})
			}

			round.Response.Text = assistantMessage.Content
			globalHistory.AddRound(round)
			continue
		}

		// No tool calls, add assistant message and finish
		messages = append(messages, assistantMessage)
		round.Response.Text = assistantMessage.Content
		globalHistory.AddRound(round)
		break
	}

	sendSSE(w, SSEEvent{Type: "done"})
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "GET" {
		json.NewEncoder(w).Encode(globalHistory.GetAll())
		return
	}

	if r.Method == "DELETE" {
		globalHistory.Clear()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleHistoryLast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	last := globalHistory.GetLast()
	if last == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "no history"})
		return
	}

	json.NewEncoder(w).Encode(last)
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/chat", handleChat)
	http.HandleFunc("/api/history", handleHistory)
	http.HandleFunc("/api/history/last", handleHistoryLast)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 AI-POC Server starting on http://localhost:%s", port)
	log.Printf("📁 Serving static files from ./static")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
