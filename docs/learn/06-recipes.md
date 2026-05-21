# 06 — Recipes

สูตรสำเร็จสำหรับงานที่เจอบ่อย. ทุก recipe เริ่มด้วย **เขียน test ที่ fail
ก่อน** ตามกฎ TDD

---

## Recipe 1 — เพิ่ม keybind ใหม่

**โจทย์:** กด `Ctrl+L` แล้ว clear chat

1. เขียน test ใน `app/update_test.go`:
```go
func TestCtrlLClearsChat(t *testing.T) {
    m := app.New(&llm.MockProvider{})
    m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
    // ส่งข้อความเข้าไปก่อน
    ...
    updated, _ := m.Update(tea.KeyPressMsg{Code: 'l', Mod: tea.ModCtrl})
    if strings.Contains(updated.(app.Model).ChatView(), "user message") {
        t.Errorf("expected chat to be cleared")
    }
}
```

2. รัน → fail
3. เพิ่ม case ใน `app/update.go`:
```go
case "ctrl+l":
    m.chat = components.NewChat(m.width, ...)  // reset
    m.messages = nil
    return m, nil
```
4. รัน → pass

---

## Recipe 2 — เพิ่ม component ใหม่ (เช่น TitleBar)

**โจทย์:** แถวบนสุดแสดงชื่อ session

1. `components/titlebar.go`:
```go
package components

type TitleBar struct {
    title string
    width int
}

func NewTitleBar(title string) TitleBar { return TitleBar{title: title} }

func (t TitleBar) Init() tea.Cmd { return nil }

func (t TitleBar) Update(msg tea.Msg) (TitleBar, tea.Cmd) {
    if m, ok := msg.(tea.WindowSizeMsg); ok {
        t.width = m.Width
    }
    return t, nil
}

func (t TitleBar) View() string {
    return lipgloss.NewStyle().
        Background(styles.PanelBg).
        Foreground(styles.AssistantText).
        Width(t.width).
        Padding(0, 2).
        Render(t.title)
}
```

2. เขียน test คู่ — `components/titlebar_test.go`
3. ใน `app/model.go` เพิ่ม field + ใน `New(...)`:
```go
type Model struct {
    titlebar components.TitleBar
    ...
}
```

4. ใน `app/update.go` route `WindowSizeMsg` ลง:
```go
m.titlebar, _ = m.titlebar.Update(msg)
```

5. ใน `app/view.go` ใส่ไว้บนสุด:
```go
leftStack := lipgloss.JoinVertical(lipgloss.Left,
    m.titlebar.View(),   // ← เพิ่ม
    topMarginLeft, chatBlock, chatInputSpacer, inputView,
)
```

6. อัปเดต `reflowChat()` ให้ลบ title height ออกจาก chatHeight ด้วย

---

## Recipe 3 — เปลี่ยน theme

แก้ไฟล์ **เดียว** — `styles/theme.go`. ทุก component pull จากตรงนั้น
อัตโนมัติ. **อย่าลืมอัปเดต `chatBgSGR`** ใน `components/chat.go:169`
ให้ตรงกับ `styles.Background` ตัวใหม่ (ถ้าเปลี่ยนสี chat bg)

---

## Recipe 4 — เพิ่ม provider จริง (เช่น Anthropic)

ดูตัวอย่างเต็มใน [`05-llm-and-streaming.md`](./05-llm-and-streaming.md)
โดยย่อ:

1. `llm/anthropic.go` — struct + implement `LLMProvider`
2. Mock test ก่อน (HTTP stub) แล้ว implement จริง
3. ใน `main.go` สลับ:
```go
provider := llm.NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"))
```
ไม่ต้องแตะ `app/` หรือ `components/` เลย

---

## Recipe 5 — debug "ทำไมพื้นเทา?"

1. เปิด `components/chat.go` — assistant ใช้ glamour → ดู SGR reset patch
   ครบทั้ง `\x1b[m` และ `\x1b[0m` มั้ย
2. ถ้าเทาในที่อื่น — ดู `Place(..., WithWhitespaceStyle(...))` หรือ
   `Join` ที่ height ไม่เท่า (`03-layout-and-styling.md`)
3. dump ANSI ดูตรงๆ ก็ได้:
```go
v := chat.View()
for _, r := range v {
    if r == 0x1b { fmt.Print("\\x1b") } else { fmt.Print(string(r)) }
}
```
ถ้าเจอ `\x1b[48;...` แปลกๆ (ไม่ใช่ `48;2;0;0;0`) — นั่นแหละต้นเหตุ

---

## Recipe 6 — เพิ่ม unit test สำหรับ streaming

ใช้ stub provider ในเทสได้เลย:

```go
type stubProvider struct{ chunks []string }

func (s *stubProvider) Name() string { return "stub" }
func (s *stubProvider) ChatCompletion(ctx context.Context, _ []llm.Message, _ llm.Options) (<-chan llm.Chunk, error) {
    ch := make(chan llm.Chunk, len(s.chunks))
    for i, c := range s.chunks {
        ch <- llm.Chunk{Content: c, Done: i == len(s.chunks)-1}
    }
    close(ch)
    return ch, nil
}

func TestStreaming(t *testing.T) {
    m := app.New(&stubProvider{chunks: []string{"Hel", "lo"}})
    // ...ส่ง SendMsg แล้ว drain chunkMsg ทีละตัว
}
```

---

## Recipe 7 — เพิ่มโหมดใหม่ (เช่น "Review")

1. `components/input.go` เพิ่ม const:
```go
ModeReview Mode = "Review"
```
2. `styles/theme.go` เพิ่มสี:
```go
ModeReviewColor = lipgloss.Color("#xxxxxx")
```
3. แก้ `Input.View()` ให้เลือกสีตามโหมด:
```go
switch i.mode {
case ModePlan:   modeColor = styles.ModePlanColor
case ModeReview: modeColor = styles.ModeReviewColor
default:         modeColor = styles.ModeBuildColor
}
```
4. (option) เพิ่ม keybind สลับโหมดใน `app/update.go`
