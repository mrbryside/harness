# 04 — Components

ทุก component อยู่ใน `components/` ไฟล์ละตัว มาตรฐาน 4 อย่าง:

- มี `Init() tea.Cmd`, `Update(msg) (Self, tea.Cmd)`, `View() string`
- คำนวณ width/height จาก `tea.WindowSizeMsg` ของตัวเอง
- ไม่ import component อื่น และไม่ import `llm.*`
- มี `xxx_test.go` คู่กันเสมอ

---

## Chat (`components/chat.go`)

Scrollable message history. ห่อ `viewport.Model` ของ bubbles

**โครงสร้าง:**
```go
type Chat struct {
    messages     []chatMessage    // role + content
    viewport     viewport.Model
    width, height int
    userScrolled bool             // scroll-lock
    renderer     *glamour.TermRenderer
}
```

**API หลัก:**
- `NewChat(width, height) Chat`
- `AppendMessage(role, content)` — เพิ่มข้อความใหม่ (user reset scroll-lock)
- `AppendChunk(chunk)` — ต่อข้อความเข้า message ตัวสุดท้าย (สำหรับ streaming)

**Rendering rules** (ดู `renderMessage` ที่ `components/chat.go:115`):
- **User** → กล่อง `PanelBg` + accent bar `UserBorder` ซ้าย + padding(1,2)
- **Assistant** → glamour render → patch SGR reset → width-pad ทุกบรรทัด

**Scroll-lock:** เมื่อ user เลื่อนขึ้น (`up/pgup/wheel-up`) → `userScrolled=true`
→ `refresh()` ไม่ auto-scroll. กลับมา bottom เมื่อไหร่ → unlock อัตโนมัติ
หรือเริ่ม user turn ใหม่

**Markdown renderer** rebuild ทุก `WindowSizeMsg` เพราะ word-wrap ผูกกับ width

---

## Input (`components/input.go`)

Textarea + footer แสดง mode/model

**โครงสร้าง:**
```go
type Input struct {
    textarea textarea.Model
    width    int
    mode     Mode      // Build | Plan
    model    string
}
```

**Dynamic height:**
```go
ta.DynamicHeight = true
ta.MinHeight = 1
ta.MaxHeight = 10
```

Textarea ขยายเองตาม soft-wrapped row จนถึง MaxHeight แล้วเริ่ม scroll
ในตัว → chat reflow ตามผ่าน `reflowChat()` ใน app

**Enter handling** (`components/input.go:94`):
- `Ctrl+J` → rewrite เป็น `KeyEnter` แล้วยิงเข้า textarea (newline)
  เพราะ terminal ส่วนมากส่ง LF เมื่อ Shift+Enter (Shift bit หาย)
- `Enter + Shift/Alt` → newline (fall through ลง textarea)
- `Enter` เปล่า → ยิง `SendMsg{Content: value}` แล้ว reset

**Styling:** ทุก surface ของ textarea (Base, CursorLine, EndOfBuffer,
Placeholder) ถูก override ด้วย `PanelBg` เพื่อไม่ให้ default grey ของ
bubbles โผล่

**Layout:** body + spacer + footer แนวตั้ง → padding(1,1) → ติด accent bar
ซ้าย (height เท่ากับ padded)

---

## Sidebar (`components/sidebar.go`)

Panel ขวาแสดง model / tokens / cost / status. กว้างคงที่
`SidebarWidth = 40`

**API:**
- `NewSidebar(modelName) Sidebar`
- `SetTokens(n)` — อัปเดต token counter จาก streaming

**Layout:** label ตัวจางบน, value ตัวสว่างล่าง, คั่นด้วย blank line.
ทุกบรรทัด render ด้วย `Width(innerWidth)` เพื่อ pad ขวาให้พื้นเต็ม

**Connection status:** ตอนนี้ hardcode "Connected" + dot สีเขียว
(`ConnectedDot`) — ยังไม่ได้ผูก provider จริง

---

## StatusBar (`components/statusbar.go`)

แถวล่างสุด — provider/model ซ้าย, hint ขวา. version constant
`v0.1.0`

**Layout trick:** อยู่บน Background (เข้ม) ไม่ใช่ StatusBarBg เพราะ
ต้อง blend กับ chat column. Padding(1,2) เพื่อ align กับ input
(input มี accent bar 1 col + padding 1 col = ขอบเริ่มที่ col 2)

---

## Pattern เพิ่ม component ใหม่

1. สร้างไฟล์ `components/myname.go` + `myname_test.go`
2. Struct + `New(...)` constructor
3. Implement `Init() tea.Cmd`, `Update(msg) (MyName, tea.Cmd)`, `View() string`
4. ถ้าต้องรู้ขนาดจอ → handle `tea.WindowSizeMsg` ใน Update
5. เพิ่ม field ใน `app.Model`, wire ใน `app.New(...)`, route message ที่
   `app/update.go`, ประกอบลง layout ที่ `app/view.go`

ดูตัวอย่างยาวที่ [`06-recipes.md`](./06-recipes.md)
