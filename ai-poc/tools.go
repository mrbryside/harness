package main

import (
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type ToolHandler func(args json.RawMessage) (string, error)

type Tool struct {
	Definition openai.Tool
	Handler    ToolHandler
}

type ToolRegistry struct {
	tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Definition.Function.Name] = tool
}

func (r *ToolRegistry) GetOpenAITools() []openai.Tool {
	tools := make([]openai.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool.Definition)
	}
	return tools
}

func (r *ToolRegistry) Execute(name string, args json.RawMessage) (string, error) {
	tool, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}
	return tool.Handler(args)
}
