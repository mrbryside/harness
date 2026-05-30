# Current Work Status

## Completed Tasks

### 1. Code Diff Line Numbers (COMPLETE)
Fixed line number display to match git diff format:
- **Context/add lines**: use new file line numbers (current state)
- **Remove lines**: use old file line numbers (where they existed before)
- **No sorting**: lines appear in chunk order (old → new) like real git diff output
- **Continuous numbering**: no jumps or gaps in displayed line numbers

### 2. Selection Overlay Background Fix (COMPLETE)
Fixed selection highlight destroying code diff backgrounds:
- **Before**: selection painted over backgrounds with `SelectionBgSGR` (#49483e), hiding add/remove colors
- **After**: uses **reverse video** (`\x1b[7m`) which swaps fg/bg colors without painting over them
- **Result**: code diff add/remove backgrounds remain visible when selected
- **Partial selections**: no longer pad to full width (prevents black bars on the right)
- **Full-line selections**: still pad to right edge with reverse video

### 3. Code Diff Container Width (COMPLETE)
Fixed add/remove line backgrounds not filling container:
- **Root cause**: `wrapAndRenderLine` was using full `c.width` without accounting for container border+padding
- **Fix**: subtracted `diffContainerOverhead = 3` (border 1 + padding 2) from `totalWidth` and `contentWidth`
- **Result**: backgrounds now fill edge-to-edge inside the diff container

### 4. Selection Auto-Scroll with Smooth Scrolling (COMPLETE)
Fixed dragging selection past viewport edges to auto-scroll:
- **Timer-based**: `tea.Tick` every 60ms for smooth continuous scrolling
- **3 lines per tick**: ~50 lines/second scroll speed
- **Content-relative coordinates**: selection anchors use content lines (not viewport-relative)
- **Result**: drag past top/bottom edge → viewport scrolls smoothly, selection extends automatically

### 5. Selection Copy Includes Scrolled Content (COMPLETE)
Fixed copy only getting visible portion:
- **Before**: `SelectedText()` used `viewport.View()` which only returns visible lines
- **After**: uses `viewport.GetContent()` which returns full content
- **Result**: copy includes text scrolled out of view

### 6. Selection Coordinate System Fix (COMPLETE)
Fixed selection drift when scrolling:
- **Before**: coordinates were viewport-relative → selection "moved" when viewport scrolled
- **After**: coordinates are content-relative throughout the system
- **Changes**:
  - `chatContentCoord()` and `chatContentCoordClamped()` add `YOffset` to return content lines
  - `ScrollUpAndExtend` / `ScrollDownAndExtend` use content lines for Extend
  - `Overlay` uses `yoff = viewport.YOffset()` to map content lines to visible lines
  - Removed `Shift()` method (no longer needed)
- **Result**: selection stays locked to content, doesn't drift when scrolling

## Testing
All tests pass:
```bash
go test ./...
go build ./...
```

## Current Architecture Decisions
- `diffLine.lineNum`: context/add use `newIdx+1`, remove uses `oldIdx+1` (git diff semantics)
- `diffContainerOverhead = 3`: accounts for `renderDiffContainer` border+padding
- Selection: reverse video (`\x1b[7m`) instead of bg color painting — preserves underlying backgrounds
- Selection coordinates: **content-relative** throughout (app layer + components)
- `RenderCodeDiffV2` takes `StartLine` only, computes `EndLine` internally
- `AppendCodeDiff(CodeDiff{...})` — no EndLine field in struct

## Next Steps
(None — all reported issues resolved)