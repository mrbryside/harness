# 10 — Input Patterns

textarea ของ `components/input.go` ทำหน้าที่หลัก. ไฟล์นี้รวบรวม
pattern ที่จะเจอเมื่อใส่ feature ใหม่ลงไป — เริ่มจาก slash commands
จนถึงการ scroll คลิป + edge cases

อ่านก่อน:
- [`04-components.md`](./04-components.md) — โครง `Input`
- [`08-popups-and-overlays.md`](./08-popups-and-overlays.md) — autocomplete

---

## Slash commands (`/help`, `/clear`, `/model …`)

**โจทย์:** พิมพ์ขึ้นด้วย `/` แล้วกด Enter → ไม่ส่งเข้า LLM แต่ run
เป็น command

### ตรวจที่ `app/update.go` ตอนรับ `input.SendMsg`

```go
case components.SendMsg:
	val := strings.TrimSpace(msg.Content)
	if strings.HasPrefix(val, "/") {
		// route ไป command handler — ไม่ผ่าน LLM
		return runCommand(m, val)
	}
	// ปกติ: append + stream
	m.chat.AppendMessage("user", val)
	return m, startStream(m, val)
```

### `runCommand` แยกออกมาเป็น dispatcher

```go
func runCommand(m Model, raw string) (Model, tea.Cmd) {
	parts := strings.Fields(raw)
	cmd := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	switch cmd {
	case "help":
		m.chat.AppendMessage("assistant", helpText())
	case "clear":
		m.chat = components.NewChat(m.chatW, m.chatH)
	case "model":
		if len(args) > 0 {
			m.modelName = args[0]
		}
	default:
		m.chat.AppendMessage("assistant",
			fmt.Sprintf("unknown command: /%s", cmd))
	}
	return m, nil
}
```

### Test pattern

```go
func TestSlashClearResetsChat(t *testing.T) {
	m := app.New(&llm.MockProvider{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m.Chat().AppendMessage("user", "hi")

	m, _ = m.Update(components.SendMsg{Content: "/clear"})
	if strings.Contains(m.Chat().View(), "hi") {
		t.Errorf("expected /clear to wipe chat")
	}
}
```

---

## Input scroll (เมื่อ content เกิน MaxHeight)

textarea ของ bubbles จะ **scroll ภายในตัวเอง** อัตโนมัติเมื่อจำนวน
soft-wrapped row > `MaxHeight`. ไม่ต้องทำอะไรเพิ่ม

แต่ถ้าอยาก override (เช่น scroll ด้วย `Alt+↑`/`Alt+↓` แทน):

```go
case "alt+up":
	i.textarea.LineUp(1)
case "alt+down":
	i.textarea.LineDown(1)
```

ดู `bubbles/v2/textarea` สำหรับ method ที่เปิดให้

### Scroll indicator (option)

ถ้าอยากแสดง "มี content เกิน" — render เครื่องหมายเล็กๆ ที่ขอบขวาของ
input:

```go
func (i Input) View() string {
	body := i.textarea.View()
	if i.textarea.TotalLineCount() > i.textarea.Height() {
		body = addScrollIndicator(body, i.textarea.Line(), i.textarea.TotalLineCount())
	}
	// ... footer ฯลฯ
}
```

---

## Multi-line paste

textarea จัดการ paste อัตโนมัติ — `\n` ใน clipboard กลายเป็น newline
ในตัวมัน. **แต่** ระวัง:

1. `Enter` ของเรา intercept เป็น "send" → paste ที่มี `\n` จะถูกตีว่า
   send ทันที (ถ้า terminal ไม่ใช้ bracketed paste)
2. แก้: ใช้ `tea.WithBracketedPaste()` ใน `main.go`
   ```go
   p := tea.NewProgram(m,
       tea.WithAltScreen(),
       tea.WithMouseCellMotion(),
       tea.WithBracketedPaste(),
   )
   ```
3. ใน `input.Update` เช็ค `tea.PasteMsg` แยกจาก `tea.KeyPressMsg`

---

## History (ลูกศรขึ้น/ลง ใน input ว่าง)

shell pattern: input ว่าง + กด ↑ → ใส่ message ก่อนหน้า

```go
type Input struct {
	// ...
	history    []string
	historyIdx int // -1 = ยังไม่ navigate
}

func (i *Input) PushHistory(s string) {
	i.history = append(i.history, s)
	i.historyIdx = -1
}

case "up":
	if i.textarea.Value() == "" || i.historyIdx >= 0 {
		if i.historyIdx < len(i.history)-1 {
			i.historyIdx++
			i.textarea.SetValue(i.history[len(i.history)-1-i.historyIdx])
		}
		return i, nil // กิน key ไม่ให้ลง textarea
	}
case "down":
	if i.historyIdx > 0 {
		i.historyIdx--
		i.textarea.SetValue(i.history[len(i.history)-1-i.historyIdx])
	} else if i.historyIdx == 0 {
		i.historyIdx = -1
		i.textarea.SetValue("")
	}
	return i, nil
```

**Gotcha:** ลูกศรขึ้น/ลงเดิมเอาไว้ scroll chat. ลำดับการ route ใน
`app/update.go`:

```go
if popupOpen { popup กิน }
else if inputEmpty { history กิน }
else { textarea / chat scroll }
```

---

## Placeholder / hint

```go
ta.Placeholder = "พิมพ์ข้อความ... (/ เพื่อดู commands)"
ta.SetPlaceholderStyle(lipgloss.NewStyle().
    Foreground(styles.SidebarLabel).
    Background(styles.PanelBg))
```

Placeholder โผล่เมื่อ value ว่าง — ใช้แนะนำ `/` กับ user ใหม่ได้ดี

---

## Validate ก่อนส่ง

ถ้าอยากกัน user ส่งข้อความว่าง / ยาวเกิน:

```go
case "enter":
	val := strings.TrimSpace(i.textarea.Value())
	if val == "" {
		return i, nil // เงียบ ๆ ไม่ทำอะไร
	}
	if utf8.RuneCountInString(val) > 4000 {
		return i, func() tea.Msg {
			return ErrMsg{Msg: "ข้อความยาวเกิน 4000 ตัวอักษร"}
		}
	}
	// ส่ง...
```

---

## Mode toggle ที่ input footer

ปัจจุบันมี `ModeBuild` / `ModePlan` ที่ตัด with `Shift+Tab` (สมมติ).
เพิ่มโหมดใหม่ตาม [Recipe 7](./06-recipes.md#recipe-7--เพิ่มโหมดใหม่-เช่น-review)

```go
case "shift+tab":
	switch i.mode {
	case ModeBuild: i.mode = ModePlan
	case ModePlan:  i.mode = ModeReview
	default:        i.mode = ModeBuild
	}
```

---

## Edge cases & gotchas

1. **`Ctrl+J` = LF ใน terminal ส่วนใหญ่** — ที่เราต้อง rewrite เป็น
   KeyEnter (ดู `components/input.go:94`). ถ้าไม่ rewrite — Shift+Enter
   จะกลายเป็น send แทน newline
2. **textarea กิน `tab` เป็น indent** — ถ้าอยากใช้ Tab สลับ mode ต้อง
   intercept ก่อนถึง textarea
3. **Multi-byte / combining marks (ไทย)** — bubbles textarea ใช้ rune
   count ไม่ใช่ grapheme cluster → cursor ตำแหน่งเพี้ยนกับ `ั`/`้`
   *เป็น upstream limitation* — รอ bubbles patch
4. **`SetValue()` ไม่ trigger event** — เซ็ตเอง ⇒ ต้อง call refilter
   ของ popup เองด้วยถ้าจำเป็น
5. **Width ของ input** ต้อง match กับ chat width — ไม่งั้น scroll bar
   ของ textarea จะไป overlap กับ accent bar ซ้าย
6. **Reflow หลังพิมพ์** — ทุกครั้งที่ input height เปลี่ยน chat ต้อง
   shrink ตาม. เรียก `reflowChat(&m)` หลัง `m.input.Update(msg)` เสมอ
