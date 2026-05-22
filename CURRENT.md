# Current Work Status

## Active Task
Implementing code diff rendering with proper syntax highlighting, line numbers, and full-width backgrounds that don't break on soft-wrap.

## Problem Being Solved
The TUI needs to display code diffs (like git diff) when agents make file changes. Previous implementation had several issues:

1. **Background gaps**: Highlight colors (green/red) didn't extend full width when content wrapped
2. **Line number collision**: `+`/`-` markers were too close to line numbers (e.g., "100+8.")
3. **No syntax highlighting**: Diff lines were plain text without code highlighting
4. **Wrong line numbers**: Couldn't show actual file line numbers from partial snippets

## Solution Approach
- Read actual file to locate oldContent and determine real line numbers
- Grab 5 lines of context above/below the change
- Use chroma for syntax highlighting (not glamour - avoids padding artifacts)
- Manual ANSI-aware word wrap with preserved color sequences
- Apply full-width backgrounds via lipgloss `Width(c.width)` 
- Layout: `[4 char line num] [space] [marker] [space] [content]`

## Files Being Modified
- `tui/components/chat_message_code_diff.go` - Main diff rendering logic
- `tui/app/model.go` - Demo diff in startup (AGENTS.md example)
- `AGENTS.md` - Architecture documentation

## Current Status
- [x] File lookup with real line numbers
- [x] Context lines (5 above/below)
- [x] Syntax highlighting via chroma
- [x] Manual ANSI wrapping
- [x] Green/red backgrounds for add/remove
- [x] Proper spacing between line num and marker
- [ ] Final width/padding calibration (in progress)

## Known Issues
- Need to verify wrapped continuation lines have correct indentation
- Need to ensure all wrapped segments have consistent background color
- Content width calculation may need adjustment

## Testing
Run `go run .` to see the AGENTS.md diff demo showing line 99-100 addition.
