package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func RegisterInternalTools(registry *ToolRegistry) {
	registry.Register(Tool{
		Definition: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "bash",
				Description: "Execute a bash command in the terminal",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{
							"type":        "string",
							"description": "The bash command to execute",
						},
					},
					"required": []string{"command"},
				},
			},
		},
		Handler: func(args json.RawMessage) (string, error) {
			var params struct {
				Command string `json:"command"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "", err
			}

			fmt.Printf("  → Executing bash: %s\n", params.Command)

			cmd := exec.Command("bash", "-c", params.Command)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Sprintf("Error: %v\nOutput: %s", err, string(output)), nil
			}
			return string(output), nil
		},
	})

	registry.Register(Tool{
		Definition: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "find",
				Description: "Find files matching a pattern in a directory",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Directory path to search in",
						},
						"pattern": map[string]interface{}{
							"type":        "string",
							"description": "File pattern to match (e.g., *.go)",
						},
					},
					"required": []string{"path", "pattern"},
				},
			},
		},
		Handler: func(args json.RawMessage) (string, error) {
			var params struct {
				Path    string `json:"path"`
				Pattern string `json:"pattern"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "", err
			}

			fmt.Printf("  → Finding files: %s in %s\n", params.Pattern, params.Path)

			matches, err := filepath.Glob(filepath.Join(params.Path, params.Pattern))
			if err != nil {
				return "", err
			}

			if len(matches) == 0 {
				return "No files found", nil
			}
			return strings.Join(matches, "\n"), nil
		},
	})

	registry.Register(Tool{
		Definition: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "read_file",
				Description: "Read the contents of a file",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the file",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		Handler: func(args json.RawMessage) (string, error) {
			var params struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "", err
			}

			fmt.Printf("  → Reading file: %s\n", params.Path)

			cmd := exec.Command("cat", params.Path)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Sprintf("Error reading file: %v", err), nil
			}
			return string(output), nil
		},
	})
}
