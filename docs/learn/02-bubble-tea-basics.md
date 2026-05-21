# 02 — Bubble Tea Basics

ก่อนแตะโค้ดต้องเข้าใจ loop ของ Bubble Tea v2 — ทั้ง app เดินอยู่บน
pattern นี้ทุกที่ ทุก component

## The Elm loop

```
            ┌──────────────┐
   Msg ───▶ │   Update     │ ───▶ Cmd ───▶ runtime ──┐
            │ (Model, Msg) │                          │
            │ → (Model,Cmd)│                          │
            └──────┬───────┘                          │
                   │                                  │
                   ▼                                  │
            ┌──────────────┐                          │
            │     View     │ ───▶ string ──▶ screen   │
            │   (Model)    │                          │
            └──────────────┘                          ▼
                   ▲                              produces
                   └──────────────────────────────── Msg
```

3 อย่างที่ทุก model ต้องมี:

```go
type Model interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (tea.Model, tea.Cmd)
    View() tea.View
}
```

- **Model** = state ทั้งหมด (immutable per tick — `Update` คืน Model ใหม่เสมอ)
- **Msg**   = event เช่น keypress, window resize, ผลของ Cmd
- **Cmd**   = side effect (I/O, timer, channel read) — runtime จะรันแล้ว
  ส่งผลกลับมาเป็น Msg ตัวใหม่

## ตัวอย่างจากโค้ดจริง

`app/update.go:28` — root Update ของเรา:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:    // terminal resized
        ...
    case tea.KeyPressMsg:      // keyboard
        ...
    case tea.MouseWheelMsg:    // scroll
        ...
    case components.SendMsg:   // user hit Enter in input
        ...
    case chunkMsg:             // LLM streamed a token
        ...
    }
}
```

ทุก message มาที่นี่ — เรา route ต่อไป component ที่เกี่ยวข้อง

## Msg ที่เจอบ่อย

| Msg | จากไหน | ใช้ตอนไหน |
|---|---|---|
| `tea.WindowSizeMsg` | runtime | resize layout |
| `tea.KeyPressMsg` | runtime | keyboard input |
| `tea.MouseWheelMsg` | runtime | mouse scroll (ต้องเปิด `MouseMode`) |
| `tea.QuitMsg` | runtime / `tea.Quit` | กำลังจะปิด program |
| custom struct | Cmd ของเราเอง | dispatch ผลของ async งาน |

ใน Harness เรามี custom 2 ตัว:
- `components.SendMsg` — input เปล่ง เมื่อกด Enter
- `chunkMsg` — LLM streaming chunk (ใน `app/update.go:13`)

## Cmd

`Cmd` คือ `func() tea.Msg` — function ที่ runtime จะรันใน goroutine แยก
แล้วเอา return มา dispatch กลับเข้า `Update`

ตัวอย่างจาก `app/update.go:18` — อ่าน channel แล้วส่งเข้า loop:

```go
func nextChunk(ch <-chan llm.Chunk) tea.Cmd {
    return func() tea.Msg {
        chunk, ok := <-ch
        if !ok {
            return chunkMsg{chunk: llm.Chunk{Done: true}}
        }
        return chunkMsg{chunk: chunk}
    }
}
```

แบบนี้ blocking read ไม่ฟรีซ UI เพราะรันใน goroutine แยก

### Batch / Sequence

- `tea.Batch(cmds...)` — รันทุก cmd ขนานกัน (order ไม่ค้ำ)
- `tea.Sequence(cmds...)` — รัน cmd ทีละตัวตามลำดับ
- `nil` — ไม่มี side effect

## Sub-models

ทุก component (`Chat`, `Input`, `Sidebar`, `StatusBar`) คือ Bubble Tea
model ของตัวเอง — มี `Init/Update/View` ครบ. Root model ใน `app/`
แค่ "ส่งต่อ" message ลงไปแล้วเก็บผลกลับ:

```go
m.input, cmd = m.input.Update(msg)
```

Pattern นี้ทำให้ component test ได้แยก และ swap/reuse ได้

## tea.View vs string

v2 แยก `View()` คืน `tea.View` (struct) ไม่ใช่ string ดิบเหมือน v1
ใน `tea.View` เก็บ feature flag ของ terminal เช่น `AltScreen`,
`MouseMode`, `KeyboardEnhancements` (ดู `app/view.go:19`)

```go
v := tea.NewView(m.render())
v.AltScreen = true
v.MouseMode = tea.MouseModeCellMotion
return v
```

v1 ใช้ `tea.WithAltScreen()` เป็น program option — v2 ย้ายมาที่ View

## Gotchas

- `Update` ต้องคืน Model ตัวใหม่ทุกครั้ง — **อย่า mutate receiver แล้วลืม return**
- `Cmd` รันใน goroutine — อย่า touch Model ตรงๆ ส่ง Msg กลับมาเสมอ
- ค่า zero-value ของ `tea.Cmd` คือ `nil` ใช้ได้เลย ไม่ต้อง guard
- Bubble Tea **ไม่ promise ลำดับ** ของ Cmd → ถ้าจำเป็นต้องเรียงใช้ `tea.Sequence`
