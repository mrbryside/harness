# 07 — Glossary & Gotchas

## คำศัพท์

| คำ | คือ |
|---|---|
| **Bubble Tea** | TUI framework Elm-style ที่ใช้: Model + Update + View |
| **bubbles** | library ที่รวม widget สำเร็จ (textarea, viewport, ...) |
| **lipgloss** | DSL สำหรับ style + layout (Width/Height/Padding/Join/Place) |
| **glamour** | markdown → ANSI renderer (ใช้ chroma ภายในสำหรับ code) |
| **chroma** | syntax highlighter ที่ใส่สี foreground ให้ token แต่ละตัว |
| **Model** | struct state ทั้งหมดของ component หรือ app |
| **Msg** | event ที่ runtime ส่งเข้า `Update` |
| **Cmd** | `func() tea.Msg` — side effect ที่ runtime จะรันใน goroutine |
| **tea.View** | wrapper ของ string ที่ถือ terminal feature flags ใน v2 |
| **SGR** | "Select Graphic Rendition" — escape sequence ของสี/style ใน ANSI |
| **viewport** | bubbles widget แสดง content แบบ scrollable |
| **soft wrap** | wrap ที่จอ ไม่แทรก `\n` จริงเข้า buffer |
| **DynamicHeight** | textarea option ให้ขยาย/หดเองตาม content (v2) |
| **MouseMode** | flag เปิดรับ mouse event (ถ้าไม่เปิด terminal scrollback กิน) |
| **AltScreen** | mode ที่ใช้ buffer แยก ไม่กวน scrollback ของ terminal |
| **Kitty protocol** | spec ให้ terminal ส่ง modifier key (Shift+Enter ฯลฯ) ครบ |
| **chatBgSGR** | constant `"\x1b[48;2;0;0;0m"` ใช้ patch glamour reset |
| **reflowChat** | helper recompute chat width/height หลัง input resize |

## Gotchas (เจอแล้วเสียเวลา)

### Bubble Tea
- **Update ต้อง return Model ใหม่** — ลืม return = state หาย
- **`Cmd` รันใน goroutine** — อย่า touch Model จากใน Cmd ส่ง Msg กลับเสมอ
- **Cmd ไม่ promise ลำดับ** — ถ้าต้องเรียง ใช้ `tea.Sequence`
- **v2 API ต่างจาก v1** — `tea.WithAltScreen()` ย้ายไปอยู่บน `tea.View` แล้ว
- v2 init **ไม่ probe background** เหมือน v1 (เคยทำ OSC 11 reply leak ใน Zed/iTerm)

### lipgloss
- **Style immutable** — chain method คืน Style ใหม่ ต้อง assign กลับเสมอ
- **`Width(N)` คือ target** — content ยาวกว่า wrap, สั้นกว่า pad
- **Padding กิน width** — Width(40) + Padding(1,2) → content เหลือ 36 cols
- **Place ไม่ paint whitespace** ถ้าไม่ใส่ `WithWhitespaceStyle`
- **JoinHorizontal ต้อง height เท่า** ไม่งั้นช่วงต่างเป็น default bg
- **v2 ตัด `WithWhitespaceBackground`** → ใช้ `WithWhitespaceStyle(style)` แทน

### Glamour / chroma
- **SGR reset 2 รูปแบบ:** `\x1b[m` (glamour) และ `\x1b[0m` (chroma) — patch ครบทั้งคู่
- **Rebuild renderer ทุก resize** ถ้าใช้ `WithWordWrap(width)` ไม่งั้น wrap ผิด
- **Streaming + markdown:** ต้องส่ง whitespace/newline ครบใน chunk
  ไม่งั้น glamour parse ผิด (mock ใช้ regex `\s+|\S+` รักษา whitespace)
- **`WithStandardStyle("dark")` ไม่ใช่ `WithAutoStyle()`** — v2 ตัด auto ออก
- ไม่ต้อง strip ทุก BG ของ glamour ออก — patch reset อย่างเดียวพอ (เร็วและเสถียรกว่า)

### bubbles textarea
- **DynamicHeight ใช้กับ MinHeight/MaxHeight** — ไม่งั้นจะใหญ่กว่า MaxHeight
- **`LineCount()` คือ logical line** ไม่ใช่ visual row (อย่าใช้นับ row บนจอ)
- **CursorLine มี BG เฉพาะของมัน** — ต้อง override ด้วย PanelBg
- **Shift+Enter:** terminal มาก/ไม่ส่ง modifier — fallback คือ rewrite `Ctrl+J` → KeyEnter

### Mouse / terminal
- **เปิด MouseMode = terminal scrollback หาย** — ผู้ใช้เลื่อนได้แค่ใน viewport
- **Mouse scroll = `tea.MouseWheelMsg`** ต้อง route ไป chat โดยเฉพาะ
- **Esc ออกจาก program** — ถ้าไม่ต้องการ ปิด case ใน `app/update.go:58`

### Thai / multi-byte
- **bubbles textarea ใช้ rune count** ไม่ใช่ grapheme cluster width
  → cursor เพี้ยนกับ Thai combining marks (เช่น ไม้โท `้`)
  — เป็น upstream limitation, ยังไม่มี fix ฝั่งเรา

## ไฟล์อ้างอิงเร็ว

| ต้องการรู้ | อ่านที่ |
|---|---|
| สีอะไรบ้าง | `styles/theme.go` |
| layout math | `app/view.go:36` (`render()`) |
| message routing | `app/update.go:28` (`Update`) |
| streaming pattern | `app/update.go:18` (`nextChunk`), `app/update.go:81` (SendMsg case) |
| markdown rendering | `components/chat.go:115` (`renderMessage`) |
| SGR patch | `components/chat.go:153` |
| keybind handling | `app/update.go:56`, `components/input.go:94` |
| reflow logic | `app/update.go:130` (`reflowChat`) |
| terminal flags | `app/view.go:19` (`View`) |
| provider interface | `llm/provider.go` |
| mock pattern | `llm/mock.go` |
