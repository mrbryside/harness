# 03 — Layout & Styling

ทุกพิกเซลในจอ Harness ถูกประกอบด้วย **lipgloss v2** — ไม่มี ANSI ดิบ
ไม่มี absolute positioning เลย

## หลักการ

1. **สีอยู่ที่เดียว** — `styles/theme.go` เป็น source of truth ห้าม
   hardcode hex ที่อื่น
2. **ประกอบด้วย Join/Place** — `lipgloss.JoinHorizontal/JoinVertical/Place`
   เท่านั้น ไม่ใช้ cursor escape
3. **เผื่อ width/height ผ่าน `WindowSizeMsg`** — ทุก component คำนวณ
   ขนาดตัวเองจากที่ root ส่งมา
4. **ทุกพื้นที่ที่ไม่มี content ต้อง paint สี** — ไม่งั้น terminal default
   จะโผล่มา (เห็นเป็นแถบเทา/ใส)

## Theme (`styles/theme.go`)

Two-tone palette:
- `Background` (`#000000`) — chat area, status bar (โทนเข้ม)
- `PanelBg` (`#1a1b26`) — sidebar, input, user message panel (โทนอ่อนกว่า)

Accent:
- `UserBorder` — แถบสีฟ้าซ้าย user message + input
- `AssistantText`, `SidebarLabel`, `SidebarValue` — สีตัวอักษร
- `StatusBarAccent`, `ConnectedDot` — accent ใน status bar / sidebar
- `ModeBuildColor` (ฟ้า), `ModePlanColor` (ม่วง) — label โหมดใน input footer

```go
import "github.com/mrbryside/harness/styles"

style := lipgloss.NewStyle().Background(styles.PanelBg)
```

## lipgloss building blocks

### Style — describe ไม่ใช่ apply

```go
s := lipgloss.NewStyle().
    Background(styles.Background).
    Foreground(styles.AssistantText).
    Width(40).
    Padding(1, 2).
    Bold(true)
out := s.Render("hello")
```

Style **immutable** — chain method คืน Style ใหม่เสมอ

### Join — ต่อแบบ box

```go
lipgloss.JoinHorizontal(lipgloss.Top, left, right)
lipgloss.JoinVertical(lipgloss.Left, top, bottom)
```

- ทุก row ใน vertical join ต้อง width เท่ากัน ไม่งั้น layout เพี้ยน
- ทุก column ใน horizontal join ต้อง height เท่ากัน

ถ้าไม่เท่า — pad เอง ด้วย style `Width(N).Height(M).Render("")`

### Place — วาง content ในกล่อง

```go
lipgloss.Place(
    width, height,
    lipgloss.Left, lipgloss.Top,   // horizontal, vertical alignment
    content,
    lipgloss.WithWhitespaceStyle(
        lipgloss.NewStyle().Background(styles.Background),
    ),
)
```

`WithWhitespaceStyle` สำคัญ — ถ้าไม่ใส่ พื้นที่ว่างจะเป็นสี default

## Layout ใน Harness

```
┌──┬──────────────────────────────┬──┬────────┐
│  │ topMargin (chat width)       │  │        │
│LM│ chatBlock                    │GC│sidebar │
│  │ chatInputSpacer              │  │        │
│  │ inputView                    │  │        │
├──┴──────────────────────────────┴──┤        │
│ statusBar (across left + gap)      │        │
└────────────────────────────────────┴────────┘
LM = leftMargin (outerMarginX cols of Background)
GC = gapCol      (innerGap cols of Background)
```

ที่ `app/view.go:36` (`render()`) เห็นการประกอบ:

1. คำนวณ chatWidth, chatHeight จาก window size
2. `Place` chat + sidebar ลงในกล่องของตัวเอง (paint whitespace)
3. ประกอบ leftStack = topMargin + chat + spacer + input ด้วย `JoinVertical`
4. ประกอบ leftCol = leftMargin + leftStack ด้วย `JoinHorizontal`
5. ต่อ gapCol แล้วแปะ statusBar ใต้สุด
6. `JoinHorizontal(leftFull, sidebarBlock)`

### Constants ที่คุม layout

ใน `app/view.go:12`:
```go
outerMarginX = 2  // ขอบซ้าย/ขวา
outerMarginY = 1  // ขอบบน
innerGap     = 2  // ช่องว่างระหว่าง chat กับ sidebar
chatInputGap = 1  // ช่องว่างระหว่าง chat กับ input
```

อยากปรับ spacing ทั้งแอป — แก้ที่นี่จุดเดียว

## การคำนวณ chatHeight

Input ของเรา **DynamicHeight** (ขยายเองตามจำนวนบรรทัด) → chat ต้องหด/
ขยายตามเสมอ. หัวใจอยู่ที่ `reflowChat()` ใน `app/update.go:130`:

```go
inputLines := lipgloss.Height(m.input.View())
statusLines := lipgloss.Height(m.statusbar.View())
chatHeight := m.height - inputLines - statusLines - outerMarginY - chatInputGap
```

เรียก `reflowChat()` ทุกครั้งที่ key forward เข้า input → ถ้า input โต
ขึ้น chat viewport หดทันที (ไม่ทำให้ UI ดันลง)

## SGR reset trap (Glamour)

Glamour และ chroma syntax highlighter ทิ้ง SGR reset (`\x1b[m` หรือ
`\x1b[0m`) ระหว่าง token → BG หาย → terminal default โผล่ระหว่าง token
(เห็นเป็นแถบเทาแคบๆ)

แก้ใน `components/chat.go:153`:
```go
content = strings.ReplaceAll(content, "\x1b[m",  "\x1b[m"+chatBgSGR)
content = strings.ReplaceAll(content, "\x1b[0m", "\x1b[0m"+chatBgSGR)
```

`chatBgSGR = "\x1b[48;2;0;0;0m"` ตรงกับ `styles.Background = #000000`
ถ้าจะเปลี่ยนสี Background ต้องอัปเดต `chatBgSGR` ให้ตรงด้วย

## Width math gotchas

- `lipgloss.Width(s)` / `lipgloss.Height(s)` — measure ที่ render แล้ว
  (ไม่ใช่ rune count)
- Style.Width(N) คือ **target** — content ที่ยาวกว่าจะ wrap, สั้นกว่า pad ขวา
- Padding กิน width — ถ้า outer Width=40 + Padding(1,2) → inner = 36
- ทุก component ต้อง deal กับเคส width=0 (ก่อน `WindowSizeMsg` ตัวแรก)
  → guard `if w < 1 { w = 1 }`

## Whitespace bg checklist

ทุกครั้งที่เห็น "แถบสีไม่ใช่" บนจอ ให้เช็คตามนี้:

1. ใช้ `Place` แล้วลืม `WithWhitespaceStyle` มั้ย
2. `Join` องค์ประกอบที่ height ไม่เท่า → ช่วงต่างเป็น default bg
3. Content สั้นกว่า `Width(N)` แต่ style ไม่ได้ set Background
4. Glamour render โดยไม่ patch SGR reset (ดูข้างบน)
