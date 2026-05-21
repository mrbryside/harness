# Theme Guide

Read this **only when** you are changing colors or fixing background/visual
issues. Do not read proactively.

## Where Colors Live

**`styles/theme.go`** is the single source of truth. No hex codes anywhere
else. Ever.

## Color Variables

There are two kinds:

1. **`lipgloss.Color` variables** — used by `lipgloss.NewStyle()`
2. **ANSI SGR string constants** — injected after resets to prevent background
   bleed (e.g. `ChatBgSGR`, `PanelBgSGR`, `SelectionBgSGR`)

### When you change a color, you MUST update BOTH

Example:

```go
ChatBackground = lipgloss.Color("#0a0a0a")   // used by lipgloss styles
ChatBgSGR      = hexToAnsiBg("#0a0a0a")      // injected after resets
```

If you forget the SGR constant, you'll get grey strips / background bleed.

## Background Color Mapping

| Area | lipgloss.Color | SGR Constant | Hex |
|---|---|---|---|
| Chat area (main) | `ChatBackground` | `ChatBgSGR` | `#0a0a0a` |
| Sidebar, user msgs | `Background` | `BackgroundSGR` | `#141414` |
| Input area | `PanelBg` | `PanelBgSGR` | `#1e1e1e` |
| Selection highlight | `SelectionBg` | `SelectionBgSGR` | `#49483e` |

## Code Block Background

The glamour markdown renderer has its own JSON theme in
`styles/chat_style.go`. The `code_block` background must align with
`ChatBackground`:

```go
"code_block": map[string]any{
    "background_color": hexFromColor(ChatBackground),
    // ...
}
```

The syntax highlighting colors inside `chroma` can stay fixed (not linked to
theme) — they are independent syntax colors.

## How Background Bleed Happens

Three layers emit bare SGR resets (`\x1b[m` or `\x1b[0m`):

1. **glamour** — after every styled token in rendered markdown
2. **lipgloss** — when padding a styled string to fixed width
3. **viewport** — when padding short lines / empty rows

These resets strip the background color, so the terminal renders spaces with
its default grey background.

### The Fix

In `chat_view.go`, we patch resets at the outermost layer:

```go
out = strings.ReplaceAll(out, "\x1b[m",  "\x1b[m"+styles.ChatBgSGR)
out = strings.ReplaceAll(out, "\x1b[0m", "\x1b[0m"+styles.ChatBgSGR)
```

This covers all three layers in one pass.

### Selection Overlay

When the user drags to select text, `WrapLineRange` in
`selection_overlay.go` paints `SelectionBgSGR` over selected cells. After the
selection ends, it restores the background using the caller-supplied `bgSGR`
(typically `styles.ChatBgSGR` or `styles.PanelBgSGR`).

## Testing Theme Changes

1. Run `go build ./...`
2. Run `go test ./...`
3. Visually verify in the TUI:
   - Chat area background is uniform (no grey strips)
   - Code blocks have correct background
   - Selection highlight works
   - Input area has correct background
   - Sidebar has correct background
