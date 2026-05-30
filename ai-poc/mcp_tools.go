package main

import (
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type FakeMCPClient struct {
	name string
}

func NewFakeMCPClient(name string) *FakeMCPClient {
	return &FakeMCPClient{name: name}
}

func RegisterMCPTools(registry *ToolRegistry) {
	// Fake filesystem MCP server
	registry.Register(Tool{
		Definition: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "mcp_filesystem_list",
				Description: "[MCP:filesystem] List files in a directory",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"directory": map[string]interface{}{
							"type":        "string",
							"description": "Directory to list",
						},
					},
					"required": []string{"directory"},
				},
			},
		},
		Handler: func(args json.RawMessage) (string, error) {
			var params struct {
				Directory string `json:"directory"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "", err
			}

			fmt.Printf("  → [MCP:filesystem] Listing directory: %s\n", params.Directory)

			// Fake response
			return fmt.Sprintf(`Directory: %s
Files:
  - README.md
  - main.go
  - go.mod
  - go.sum
  - .gitignore`, params.Directory), nil
		},
	})

	// Fake GitHub MCP server
	registry.Register(Tool{
		Definition: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "mcp_github_search",
				Description: "[MCP:github] Search for repositories on GitHub",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "Search query",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		Handler: func(args json.RawMessage) (string, error) {
			var params struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "", err
			}

			fmt.Printf("  → [MCP:github] Searching: %s\n", params.Query)

			// Fake response
			return fmt.Sprintf(`GitHub Search Results for "%s":
1. user/awesome-project (⭐1.2k)
   Description: An awesome project
2. org/another-repo (⭐856)
   Description: Another great repo
3. dev/cool-tool (⭐432)
   Description: A cool tool`, params.Query), nil
		},
	})

	// Fake weather MCP server
	registry.Register(Tool{
		Definition: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "mcp_weather_get",
				Description: "[MCP:weather] Get weather information for a location",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "City or location name",
						},
					},
					"required": []string{"location"},
				},
			},
		},
		Handler: func(args json.RawMessage) (string, error) {
			var params struct {
				Location string `json:"location"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "", err
			}

			fmt.Printf("  → [MCP:weather] Getting weather for: %s\n", params.Location)

			// Fake response
			return fmt.Sprintf(`Weather for %s:
Temperature: 24°C
Condition: Partly cloudy
Humidity: 65%%
Wind: 12 km/h SE`, params.Location), nil
		},
	})
}
