# harness — AGENTS.md

## Commands
- `make test` — run all tests (`go test ./...`)
- `make run` — build and run TUI
- `make run-log` — run TUI with stderr → `/tmp/harness.log`
- `make log` — tail `/tmp/harness.log` in another terminal
- `make build` — build to `build/harness`
- `make clean` — remove build dir and log

Go 1.26.3 (managed by asdf via `.tool-versions`).

## Architecture

EventBus-driven TUI. Three layers communicate only through `eventbus.EventBus`:

```
tui/app (Model) ◄──EventBus──► agentruntime (AgentRuntime) ◄──► llm (LLMProvider)
```

**Entry point:** `cli.go` — wiring order matters:
```go
eb := eventbus.NewEventBus()
model := app.New(eb)          // subscribes to events first
agentruntime.New(eb, provider) // emits events after
```

**Never** call `provider.ChatCompletion()` from the TUI. The TUI emits `EventUserMessaged`; AgentRuntime subscribes, calls the provider, and emits `EventAssistantMessaged` chunks back.

## EventBus Events

| Event | Emitted by | Consumed by | Payload |
|---|---|---|---|
| `question_asked` | AgentRuntime | TUI Model | `{QuestionID, Question string}` |
| `question_answered` | TUI Model | AgentRuntime | `{QuestionID, Answer QuestionChoice}` |
| `user_messaged` | TUI Model | AgentRuntime | `{ID, Content string}` |
| `assistant_messaged` | AgentRuntime | TUI Model | `{ID, Content string, Done bool}` |
| `tool_edit_file_executed` | AgentRuntime | TUI Model | `{Path, OldContent, NewContent string, StartLine int}` |

## EventBus → TUI bridge (`eventCh`)

EventBus subscribers run synchronously on the `Emit()` goroutine. They cannot mutate the Bubble Tea Model directly (it's passed by value). Instead, subscribers send `tea.Msg` to `m.eventCh` (buffered, size 100). `listenEvents()` in `Init()` reads from this channel and feeds messages back into the Tea loop.

**Critical:** After handling any EventBus-derived message, the handler MUST return `m.listenEvents()` to keep the loop alive. Forgetting this breaks the event flow silently.

## Package structure

```
cli.go                          — main(), wheelFilter, wiring
eventbus/                       — EventBus pub/sub + event constants/payloads
agentruntime/                   — AgentRuntime: owns LLMProvider, streams responses
  runtime.go                    — struct, New(), init(), streamResponse()
  subscriber_bus_*.go           — one file per EventBus subscription
llm/                            — LLMProvider interface + MockProvider
tui/app/                        — Bubble Tea Model
  model.go                      — Model struct, New(), Init()
  update.go                     — Update() dispatcher switch ONLY
  message_types.go              — message types shared across handlers
  layout.go                     — layout constants, coord math, reflowChat()
  view.go                       — View(), render()
  handle_*.go                   — one file per handler concern
  subscriber_bus_*.go           — one file per EventBus subscription
tui/components/                 — reusable Bubble Tea components
  question.go                   — Question interface + registry (OCP for question types)
  question_permission.go        — PermissionQuestion (yes/no with N options)
  chat.go / chat_message*.go    — Chat component, message rendering
  chat_message_tool_edit.go     — ToolEdit rendering (was code_diff)
tui/styles/                     — Monokai theme, lipgloss styles
tui/commands/                   — slash command registry
```

## Handler file naming

- `handle_*.go` — contains handler methods called from `Update()` dispatcher
- `subscriber_bus_*.go` — contains EventBus subscription setup methods
- Message types live in the file that uses them, or `message_types.go` if shared

## Key gotchas

1. **EventBus subscriber ordering** — `app.New(eb)` must be called before `agentruntime.New(eb, provider)` so subscriptions are registered before events are emitted.
2. **`activeQuestion` is a pointer wrapper** — `*activeQuestionHolder` so EventBus subscribers can mutate shared state (value copies are lost).
3. **`questionShownMsg`** — EventBus subscriber sends this to trigger `reflowChat()` via the Tea loop, not directly (direct mutation is lost on value copy).
4. **MockProvider** — `llm.MockProvider` streams canned markdown responses token-by-token with 40ms delay. Used by default in `cli.go`.
5. **`messages` field** — TUI Model keeps its own `[]llm.Message` for conversation history. AgentRuntime has a separate copy. Both are updated independently.
6. **`handleSendMsg` returns empty cmd** — it emits `user_messaged` but returns `tea.Batch()` (nil). The `listenEvents()` cmd from `Init()` handles the async response.

## Testing

- `go test ./...` — all packages
- Tests in `tui/app/` use `app.New(eb)` without provider (provider lives in AgentRuntime)
- Full integration tests in `full_integration_test.go` use `drainEventCh()` to process `eventCh` messages synchronously
- `EventChForTest()` exposes the internal channel for test draining
- `SetStreamingForTest(v bool)` flips streaming flag for tests that can't start real streams
