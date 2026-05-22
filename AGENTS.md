# AGENTS.md

## Agent Workflow

Same as before — TDD is mandatory, and the same general process applies:

1. Read this file to understand the project.
2. Before changing code, understand which package / component the change belongs to (see the structure below).
3. **Never write production code without first writing a failing test.** The test must fail before the implementation is written, and pass after.
4. After a meaningful change, run `go build ./...` and `go test ./...` to confirm nothing regressed.
5. Update or create docs when behavior or architecture changes in a user-visible way.

The project no longer uses `docs/tui-plan.md` or `docs/tasks/*.md` as the
source of work — this is now a maintained, evolving codebase. Work is driven
by user requests.

---

## Project: AI Harness TUI

A terminal UI (TUI) for chatting with LLM providers, similar to OpenCode /
Claude Code in appearance. Built in Go.

This file is the short version that an agent must internalize before touching code.

> **Agent-specific docs** live in `docs/ai/`. Read them **only when necessary**
> — i.e. when the task at hand requires it. Do not read proactively.

---

## Architecture (read this in full)

### Layers

```
┌──────────────────────────────────────────────────────┐
│  main.go                                             │
│    Picks an LLMProvider, creates app.Model,          │
│    starts Bubble Tea with AltScreen + MouseCellMotion│
├──────────────────────────────────────────────────────┤
│  app/  (orchestration only)                          │
│    model.go   — root Model holds all components      │
│    update*.go — message routing by concern           │
│    view.go    — composes the final screen            │
├──────────────────────────────────────────────────────┤
│  components/  (standalone Bubble Tea sub-models)     │
│    chat*.go       — scrollable message history       │
│    input*.go      — textarea + Build/Plan footer     │
│    selection*.go  — drag-to-select primitive         │
│    sidebar.go     — right-side info panel            │
│    statusbar.go   — bottom status line               │
│    autocomplete.go — slash-command overlay           │
├──────────────────────────────────────────────────────┤
│  commands/  (slash command system)                   │
│    command*.go — registry + individual commands      │
├──────────────────────────────────────────────────────┤
│  memory/  (input history)                            │
│    input_mem_history.go — HistoryStorage interface   │
├──────────────────────────────────────────────────────┤
│  llm/  (provider abstraction)                        │
│    provider.go — LLMProvider interface + types       │
│    mock.go     — MockProvider (canned dev data)      │
├──────────────────────────────────────────────────────┤
│  styles/                                             │
│    theme.go      — every color used in the app       │
│    chat_style.go — glamour JSON theme generator      │
├──────────────────────────────────────────────────────┤
│  docs/ai/      — agent-specific guidance (read only  │
│                  when necessary)                     │
└──────────────────────────────────────────────────────┘
```

### Package responsibilities

| Package | Owns | Must not |
|---|---|---|
| `main` | Wiring concrete provider into `app.New`, starting Tea | Hold any business logic |
| `app` | Routing tea messages, streaming loop, layout math | Render component internals, define colors |
| `components` | One self-contained UI piece each | Know about other components or `llm.*` |
| `commands` | Slash command registry and execution | Touch UI / Tea directly |
| `memory` | Input history storage interface | Touch UI / Tea |
| `llm` | `LLMProvider` interface, types, mock | Touch UI / Tea |
| `styles` | The single source of truth for colors | Depend on anything else in the repo |

---

## Hard Rules (do not violate)

1. **Colors live only in `styles/theme.go`.** No hex codes anywhere else.
2. **All LLM access goes through `llm.LLMProvider`.** App code must never reference `MockProvider` or any future concrete type.
3. **Streaming uses `<-chan llm.Chunk`.** Do not change to a one-shot return — this pattern must work for both mock and real HTTP/SSE backends.
4. **Layout uses `lipgloss.JoinHorizontal` / `JoinVertical` / `Place` only.** No absolute positioning, no raw ANSI.
5. **`main.go` has no business logic.** It picks a provider and starts the program. Period.
6. **TDD always.** Failing test first, then implementation. No exceptions.
7. **Mouse events stay inside the program.** `tea.WithMouseCellMotion()` is set in `main.go` so the terminal never scrolls — scroll goes to the chat viewport.

---

## LLMProvider Interface

Do not change this signature without updating the mock and every caller.

```go
type LLMProvider interface {
    Name() string
    ChatCompletion(ctx context.Context, messages []Message, opts Options) (<-chan Chunk, error)
}

type Message struct { Role, Content string }
type Chunk   struct { Content string; TokensUsed int; Done bool; Err error }
type Options struct { Model string; Temperature float64 }
```

The streaming contract: callers drain the channel until it's closed or a
chunk arrives with `Done: true`. `Err` on a chunk is a stream-level error.

---

## Dev Commands

```bash
go run .             # run the TUI
go build -o harness  # build the binary
go test ./...        # run all tests
go mod tidy          # after adding/removing dependencies
```

Module path is `github.com/mrbryside/harness` — keep it that way.

---

## Theme

All colors are defined in `styles/theme.go`. Palette:

- `ChatBackground` — main chat area + status bar (#0a0a0a)
- `Background`     — sidebar + user-message panels (#141414)
- `PanelBg`        — input area (#1e1e1e)

Plus accent colors (`UserBorder`, `StatusBarAccent`, `ConnectedDot`,
`ModeBuildColor`, `ModePlanColor`, `SidebarLabel`, `SidebarValue`,
`AssistantText`).

### Changing theme colors

When modifying colors, you must update **both** the `lipgloss.Color` variable
and its matching ANSI SGR constant (e.g. `ChatBackground` + `ChatBgSGR`).
The SGR constants are injected after glamour/lipgloss resets to prevent
background bleed. Read `docs/ai/theme-guide.md` before making theme changes.

Do not hardcode hex anywhere else.

---

## Keybinds (defined in `app/update.go`)

| Key | Action |
|---|---|
| `Enter` | Send message |
| `Alt+Enter` (`Shift+Enter` on most terminals) | Newline in input |
| `Ctrl+C` / `Esc` | Quit |
| `↑ / ↓ / PgUp / PgDn` | Scroll chat |
| Mouse wheel | Scroll chat (terminal scrollback is captured) |

---

## Debugging ANSI Background Bleed

If you see **grey strips, gaps, or seams** inside the chat area (most
visible on empty lines inside fenced code blocks, or on padding rows
below the conversation), the cause is almost always the same:

> A bare SGR reset (`\x1b[m` or `\x1b[0m`) is emitted before plain space
> characters, so the terminal renders those spaces with its **default
> background** instead of `styles.Background`.

Three layers can each emit those resets independently:

| Layer | When it emits a reset |
|---|---|
| `glamour` | After every styled token in rendered markdown |
| `lipgloss` | When padding a styled string to a fixed `Width(...)` |
| `viewport` | When padding short lines + when `FillHeight=true` for empty rows |

### How to debug it

1. **Reproduce in isolation.** Write a small `main` (e.g. in `/tmp`)
   that builds the affected component and dumps `View()` directly.
   Don't trust what the terminal shows you until you've inspected the
   raw bytes.

2. **Inspect the raw bytes.** Print the output with `\x1b` escaped so
   it's readable. Split on `\n`, then for each suspect line check:
   - the visible width (`lipgloss.Width(line)`)
   - all unique SGR escapes on that line (`regexp.MustCompile(\x1b\[[0-9;]*m).FindAllString`)
   - the raw text after stripping ANSI

3. **Find the smoking gun.** Scan for the offending pattern directly:

   ```go
   strings.Contains(view, "\x1b[m    ")   // reset + plain padding
   strings.Contains(view, "\x1b[0m    ")
   ```

   If those patterns exist, that's the grey strip.

4. **Trace which layer emits it.** Render each layer in isolation
   (e.g. just `glamour.Render(md)`, then wrap with `lipgloss.NewStyle().
   Width(w).Render(...)`, then push through a viewport) and re-run the
   scan. The layer that first introduces the offending pattern is the
   culprit. In practice viewport is the outermost emitter, which is
   why the fix lives in `Chat.View()`.

5. **Patch at the outermost layer.** Insert `ChatBgSGR` right after
   every reset:

   ```go
   out = strings.ReplaceAll(out, "\x1b[m",  "\x1b[m"+styles.ChatBgSGR)
   out = strings.ReplaceAll(out, "\x1b[0m", "\x1b[0m"+styles.ChatBgSGR)
   ```

   Doing this at `View()` covers resets from all three layers in one
   pass. Patching earlier (e.g. inside `renderMessage`) misses the
   resets that viewport adds afterwards.

6. **Confirm with TDD.** Add a regression test that fails before the
   fix and passes after — assert that the rendered view does **not**
   contain `"\x1b[m    "` or `"\x1b[0m    "`. See
   `TestChatFencedCodeBlockEmptyLinesHaveBlackBg` for the pattern.

### Inline-code grey (different problem, same family)

Glamour's dark style paints inline `` `code` `` with `\x1b[…;48;5;236m`
(256-color BG 236). We strip just that one BG segment with a regex
(`stripInlineCodeBg` in `components/chat.go`) while preserving the
foreground colour and any other attributes in the same SGR. Fenced
code blocks use a 24-bit BG (`48;2;...`), so they're untouched.

Test: `TestChatStripsInlineCodeBackground`.

---

## Code Diff Full-Width Backgrounds

When rendering syntax-highlighted diff lines (via chroma) with a solid
background that must span the full viewport width, **do not manually pad
with spaces** — the viewport's soft-wrap will split the padding onto
wrapped continuation lines and the background will break.

Instead, let lipgloss handle width + padding:

```go
style := lipgloss.NewStyle().
    Background(lipgloss.Color(bgHex)).
    Width(targetWidth)
return style.Render(content)
```

Lipgloss correctly measures the visible width of strings that already
contain ANSI escape sequences (e.g. chroma output) and pads with spaces
that inherit the style's background colour. This avoids wrapping
artifacts and keeps the background flush to the right edge.

Used in: `components/chat_message_code_diff.go:applyLineBackground`.

---

## Code Diff Component Architecture

### Overview
The code diff component (`components/chat_message_code_diff.go`) renders unified diffs with syntax highlighting and full-width line backgrounds. It supports both full-file diffs and snippet-based diffs with context lines.

### Key Features
- **Real line numbers**: Reads the actual file to display correct line numbers
- **Context lines**: Shows 5 lines of context above and below changes
- **Syntax highlighting**: Uses chroma with a custom theme (`diffChromaStyle`)
- **Full-width backgrounds**: Line backgrounds extend to viewport edge via lipgloss `Width()`
- **Soft-wrap handling**: Long lines are manually wrapped with ANSI preservation
- **Git-style markers**: `+` for additions (green), `-` for removals (red)

### Layout
```
[4 char line num] [1 space] [1 char marker] [1 space] [content...]
```
- Total prefix: 7 characters
- Content width: `c.width - 7`
- Continuation lines indent 7 spaces to align with content column

### Color Scheme
- **Context lines**: `styles.Background` (#141414) bg, white text
- **Add lines**: `#2c3c2c` bg (dark green), `#6bff6b` text (green)
- **Remove lines**: `#3c2c2c` bg (dark red), `#ff6b6b` text (red)
- **Line numbers**: Context uses `styles.SidebarValue`, add/remove use matching colors

### API
```go
// Full file diff with context
chat.AppendCodeDiff("path/to/file.go", oldContent, newContent)

// Direct render (used internally)
chat.RenderCodeDiff(path, oldContent, newContent)
```

### Important Notes
- The component reads the actual file to determine real line numbers
- If file not found or oldContent not matched, falls back to simple diff (no line numbers)
- Background colors are applied via lipgloss `Width(c.width)` to ensure full coverage
- Manual ANSI-aware wrapping prevents viewport soft-wrap from breaking backgrounds

---
