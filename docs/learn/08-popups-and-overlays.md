# 08 — Popups & Overlays

จะทำ UI ที่ "ลอยขึ้นมา" ทับ layout หลัก (เช่น auto-complete popup เหนือ
input, modal กลางจอ) ใน terminal ต้องเข้าใจก่อนว่า **ไม่มี z-index**
เราจะ "วาดทับ" ด้วยการคำนวณตำแหน่งเอง แล้วใช้ `lipgloss.Place` กับ
string composition

อ่านไฟล์นี้ก่อนทำ:
- [`03-layout-and-styling.md`](./03-layout-and-styling.md) — `Place`, `Join`
- [`04-components.md`](./04-components.md) — pattern component

---

## หลักการ "absolute" ใน terminal

terminal วาดเป็น cell grid 2D — ไม่มี layer. การทำ overlay ต้อง:

1. Render layout หลักเป็น string ปกติ (chat + input + sidebar + …)
2. Render popup/modal เป็น string แยก
3. **เขียนทับเอง** — แทนที่บางบรรทัด/บางช่วงของ string หลักด้วย popup

มี 2 pattern หลัก:

| Pattern | ใช้เมื่อ | เครื่องมือ |
|---|---|---|
| **Inline overlay** (เกาะ component) | auto-complete เหนือ input | line-by-line splice |
| **Centered modal** | command palette กลางจอ | `lipgloss.Place` + line splice |

---

## Pattern 1 — Auto-complete popup เหนือ input

**โจทย์:** พิมพ์ `/` ต้น input → popup โผล่ขึ้น **เหนือ** input
แสดง list ของ slash commands. ลูกศรขึ้น/ลง เลือก → Enter ยืนยัน → Esc ปิด

### Component skeleton

`components/autocomplete.go`:

```go
package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/styles"
)

// Suggestion = หนึ่ง entry ใน popup
type Suggestion struct {
	Label string // "/help"
	Desc  string // "Show help"
}

type AutoComplete struct {
	open        bool
	width       int
	items       []Suggestion
	filtered    []Suggestion
	selected    int
	maxVisible  int // จำนวนแถวสูงสุดที่โชว์ (ทำ scroll เมื่อเกิน)
	scrollStart int // index ของแถวแรกที่ visible
	query       string
}

func NewAutoComplete(items []Suggestion) AutoComplete {
	return AutoComplete{
		items:      items,
		filtered:   items,
		maxVisible: 6,
	}
}

func (a AutoComplete) IsOpen() bool { return a.open }

func (a *AutoComplete) Open(query string) {
	a.open = true
	a.SetQuery(query)
}

func (a *AutoComplete) Close() {
	a.open = false
	a.selected = 0
	a.scrollStart = 0
	a.query = ""
}

// Selected returns the currently-highlighted suggestion (or empty).
func (a AutoComplete) Selected() (Suggestion, bool) {
	if !a.open || len(a.filtered) == 0 {
		return Suggestion{}, false
	}
	return a.filtered[a.selected], true
}
```

### Filter + scroll logic

```go
func (a *AutoComplete) SetQuery(q string) {
	a.query = q
	a.filtered = a.filtered[:0]
	for _, s := range a.items {
		if strings.HasPrefix(s.Label, q) {
			a.filtered = append(a.filtered, s)
		}
	}
	a.selected = 0
	a.scrollStart = 0
}

func (a *AutoComplete) moveSelection(delta int) {
	if len(a.filtered) == 0 {
		return
	}
	a.selected = (a.selected + delta + len(a.filtered)) % len(a.filtered)
	// keep selected in visible window
	if a.selected < a.scrollStart {
		a.scrollStart = a.selected
	}
	if a.selected >= a.scrollStart+a.maxVisible {
		a.scrollStart = a.selected - a.maxVisible + 1
	}
}
```

### Key handling

```go
func (a AutoComplete) Update(msg tea.Msg) (AutoComplete, tea.Cmd) {
	if !a.open {
		return a, nil
	}
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "up":
			a.moveSelection(-1)
		case "down":
			a.moveSelection(+1)
		case "esc":
			a.Close()
		}
	}
	return a, nil
}
```

> Enter ไม่ handle ที่นี่ — ปล่อยให้ `app/update.go` จัดการ เพราะการ
> "เลือกแล้วยัดเข้า input" ต้องสื่อสารข้าม component (ดูส่วน Integration ล่าง)

### Rendering

```go
func (a AutoComplete) View() string {
	if !a.open || len(a.filtered) == 0 {
		return ""
	}

	end := a.scrollStart + a.maxVisible
	if end > len(a.filtered) {
		end = len(a.filtered)
	}
	visible := a.filtered[a.scrollStart:end]

	var rows []string
	for i, s := range visible {
		row := s.Label + "  " + s.Desc
		style := lipgloss.NewStyle().
			Background(styles.PanelBg).
			Foreground(styles.AssistantText).
			Padding(0, 1).
			Width(a.width)
		if a.scrollStart+i == a.selected {
			style = style.
				Background(styles.UserBorder).
				Bold(true)
		}
		rows = append(rows, style.Render(row))
	}

	// scroll indicator (ถ้า list ยาวเกิน)
	if len(a.filtered) > a.maxVisible {
		hint := lipgloss.NewStyle().
			Background(styles.PanelBg).
			Foreground(styles.SidebarLabel).
			Width(a.width).
			Padding(0, 1).
			Render("↑↓ " + posStr(a.selected+1, len(a.filtered)))
		rows = append(rows, hint)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func posStr(i, n int) string { return /* "3/12" */ ... }
```

### Integration กับ app

ใน `app/model.go`:

```go
type Model struct {
	chat     components.Chat
	input    components.Input
	popup    components.AutoComplete
	// ...
}
```

ใน `app/update.go`:

```go
case tea.KeyPressMsg:
	// 1) ถ้า popup เปิด — ส่ง key ให้ popup ก่อน
	if m.popup.IsOpen() {
		switch msg.String() {
		case "enter":
			if s, ok := m.popup.Selected(); ok {
				m.input.SetValue(s.Label + " ")
				m.popup.Close()
			}
			return m, nil
		case "esc":
			m.popup.Close()
			return m, nil
		case "up", "down":
			m.popup, _ = m.popup.Update(msg)
			return m, nil
		}
	}

	// 2) ส่ง key ลง input ตามปกติ
	m.input, cmd = m.input.Update(msg)
	reflowChat(&m)

	// 3) หลัง input อัปเดต — เช็คว่าควรเปิด/อัปเดต popup มั้ย
	val := m.input.Value()
	if strings.HasPrefix(val, "/") {
		m.popup.Open(val)
	} else if m.popup.IsOpen() {
		m.popup.Close()
	}
```

### วาง popup เหนือ input (โผล่ลอย)

ใน `app/view.go` — pop ขึ้นแทนที่บางบรรทัดของ chat:

```go
chatView := m.chat.View()
inputView := m.input.View()
popupView := m.popup.View()

if popupView != "" {
	// popup height = จำนวนบรรทัดของมัน
	ph := lipgloss.Height(popupView)
	// แทนที่ ph บรรทัดล่างสุดของ chat ด้วย popup
	chatLines := strings.Split(chatView, "\n")
	popupLines := strings.Split(popupView, "\n")
	start := len(chatLines) - ph
	if start < 0 {
		start = 0
	}
	for i, p := range popupLines {
		if start+i < len(chatLines) {
			chatLines[start+i] = p
		}
	}
	chatView = strings.Join(chatLines, "\n")
}

return lipgloss.JoinVertical(lipgloss.Left, chatView, inputView, ...)
```

> ทำไม "แทน" ไม่ใช่ "แทรก"? — เพราะถ้าแทรกบรรทัด chat จะถูกดันลง
> ทำให้ input ขยับ ไม่ใช่ overlay ที่ดี

### Test ที่ควรมี

```go
func TestAutoCompleteScrollsLongList(t *testing.T) {
	items := make([]components.Suggestion, 20)
	for i := range items {
		items[i] = components.Suggestion{Label: fmt.Sprintf("/cmd%02d", i)}
	}
	a := components.NewAutoComplete(items)
	a.Open("/")
	// เลื่อนลง 10 ครั้ง — selected ต้อง = 10, scrollStart ต้องเลื่อนตาม
	for i := 0; i < 10; i++ {
		a, _ = a.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	view := a.View()
	if !strings.Contains(view, "/cmd10") {
		t.Errorf("expected /cmd10 visible after scrolling")
	}
}
```

---

## Pattern 2 — Centered modal (command palette)

ดู [`09-views-and-routing.md`](./09-views-and-routing.md) — modal กลางจอ
มี logic วาดทับและ routing คล้ายกัน แต่ใช้ `lipgloss.Place` เพื่อจัด
กลางหน้าจอ

---

## Gotchas

1. **อย่าให้ popup ทำให้ chat ขยับ** — ใช้ "splice ทับ" ไม่ใช่
   `JoinVertical` ที่ดันบรรทัดถัดไป
2. **คำนวณ height ของ popup ก่อน splice** — `lipgloss.Height(popupView)`
   จะถูกต้องเสมอ
3. **width ของ popup = width ของ input** เพื่อ alignment สวยๆ —
   ส่ง `WindowSizeMsg` ให้ popup ด้วยใน `app/update.go`
4. **กัน popup โผล่ตอนพิมพ์ปกติ** — เช็ค `strings.HasPrefix(val, "/")`
   ทุกครั้งหลัง input อัปเดต
5. **Routing precedence** — ถ้า popup เปิด keys สำคัญ
   (up/down/enter/esc) ต้องไปที่ popup ก่อน ไม่ลง chat/input ไม่งั้น
   ลูกศรจะ scroll chat แทนที่จะเลื่อน selection
6. **ANSI bleed รอบ popup** — ถ้า popup ใช้ `PanelBg` แต่ chat ใช้
   `Background` — ขอบรอบ popup อาจมี seam สีต่าง. แก้โดยให้ popup กว้าง
   เต็มความกว้างของ chat column (`Width(c.width)`)
