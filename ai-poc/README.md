# AI-POC: OpenAI-compatible API with Tool Use

This proof-of-concept demonstrates calling an OpenAI-compatible API with API key authentication, showing thinking, text responses, and tool usage (including fake MCP tools and internal tools).

## Setup

Set the following environment variables:

```bash
export OPENAI_API_KEY="your-api-key-here"
export OPENAI_BASE_URL="https://api.openai.com/v1"  # Optional: for compatible APIs
export OPENAI_MODEL="gpt-4o"  # Optional: defaults to gpt-4o
```

## Run

```bash
cd ai-poc
go run .
```

## Features

### Tool Types

1. **Internal Tools** (direct execution):
   - `bash` - Execute bash commands
   - `find` - Find files matching patterns
   - `read_file` - Read file contents

2. **Fake MCP Tools** (simulated responses):
   - `mcp_filesystem_list` - [MCP:filesystem] List directory contents
   - `mcp_github_search` - [MCP:github] Search GitHub repositories
   - `mcp_weather_get` - [MCP:weather] Get weather information

### Output Format

The demo shows:
- 👤 User messages
- 💭 AI thinking/reasoning
- 🔧 Tool calls with arguments
- ✅ Tool execution results
- 💬 Final AI response

## Architecture

- `client.go` - OpenAI-compatible client wrapper
- `tools.go` - Tool registry and definitions
- `internal_tools.go` - Real executing tools (bash, find, read_file)
- `mcp_tools.go` - Simulated MCP tools with fake responses
- `main.go` - Demo conversation flow

## Tool Calling Loop

The conversation supports iterative tool calling:
1. User sends message
2. AI thinks and decides to use tools
3. Tools execute and return results
4. AI processes results and responds or calls more tools
5. Final text response delivered

## Bruno API Collection

Test the API directly with [Bruno](https://www.usebruno.com/) using the provided collection in `bruno/` folder.

### Setup

1. Install Bruno: https://www.usebruno.com/downloads
2. Open Bruno > Open Collection > Select `ai-poc/bruno` folder
3. Select environment `local` (top right)
4. Edit `local` environment variables:
   - `apiKey` - Your API key
   - `baseUrl` - API endpoint (e.g., `https://api.moonshot.ai/v1`)
   - `model` - Model name (e.g., `qwen3.6-plus`)

### Available Requests

| # | Request | Description |
|---|---------|-------------|
| 01 | Chat Simple (No Stream) | Basic chat, single JSON response |
| 02 | Chat Simple (Stream) | Streaming response, see chunks |
| 03 | Chat With Tools (No Stream) | Tool calls in single response |
| 04 | Chat With Tools (Stream) | Tool calls streaming (arguments come char by char!) |
| 05 | Chat With Reasoning (Stream) | See thinking before content |
| 06 | Full Conversation With Tools | Complete flow with tool result |

### What You'll See

**Non-Stream** (`stream: false`):
- Full JSON response in one go
- `choices[0].message.content` or `choices[0].message.tool_calls`
- `finish_reason`: `"stop"` or `"tool_calls"`

**Stream** (`stream: true`):
- Multiple `data:` lines (SSE format)
- `delta.content` or `delta.reasoning_content` (tiny pieces)
- `delta.tool_calls` (tool call built character by character)
- Ends with `data: [DONE]`

**Tool Calls**:
- Non-stream: Complete `tool_calls` array with `name` and `arguments`
- Stream: `tool_calls` built incrementally in `delta`
- `finish_reason: "tool_calls"` indicates AI wants to use tools

**Reasoning**:
- Stream: `delta.reasoning_content` comes before `delta.content`
- Shows AI's thought process before final answer
