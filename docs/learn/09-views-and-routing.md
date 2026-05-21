# 09 — Views & Routing

ตอนนี้ Harness มี "view" เดียวคือ chat. พอจะเพิ่มหน้าอื่น (modal,
settings page, help screen) ต้องมี **view router** เพื่อบอกว่า "ตอนนี้
ผู้ใช้กำลังดูอะไร" และ route message/render ให้ถูก

ไฟล์นี้อธิบาย:
- การออกแบบ view enum
- การทำ command palette modal (Ctrl+P)
- การสลับไปหน้าอื่น

---

## หลักการ

```
┌─────────────────────────────────────┐
│ activeView  (enum: chat | palette)  │  ← state เดียวที่บอกว่าโชว์อะไร
└─────────────────────────────────────┘
              │
              ├── Update routing — key/msg ไป view ที่ active
              │
              └── View routing  — render layer ตาม active
```

**กฎ:** มีแค่ **หนึ่ง active view** ที่จะกินทุก key. modal/popup
ถือเป็น "view ซ้อน" — เปิดเมื่อไหร่ก็ block key ของหน้าหลัก

---

## View enum

`app/model.go`:

```go
type View int

const (
	ViewChat View = iota
	ViewPalette
	// ViewHelp, ViewSettings, ...
)

type Model struct {
	activeView View
	chat       components.Chat
	input      components.Input
	palette    components.Palette
	// ...
}
```

---

## Command Palette (Ctrl+P)

**โจทย์:** กด Ctrl+P → modal กลางจอ list ของ actions
- Type → filter
- ↑/↓ → เลือก
- Enter → execute action แล้วปิด
- Esc → ปิด

### Component skeleton

`components/palette.go`:

```go
package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mrbryside/harness/styles"
)

// Action = หนึ่ง command ที่ palette สามารถยิงได้
type Action struct {
	ID    string // "clear-chat"
	Label string // "Clear chat history"
	Desc  string
}

// ExecuteMsg ถูกส่งกลับเข้า tea loop เมื่อ user กด Enter
type ExecuteMsg struct{ ID string }

type Palette struct {
	width, height int
	actions       []Action
	filtered      []Action
	query         string
	selected      int
	scrollStart   int
	maxVisible    int
}

func NewPalette(actions []Action) Palette {
	return Palette{
		actions:    actions,
		filtered:   actions,
		maxVisible: 10,
	}
}
```

### Key handling + filter

```go
func (p Palette) Update(msg tea.Msg) (Palette, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}
	switch key.String() {
	case "up":
		p.move(-1)
	case "down":
		p.move(+1)
	case "enter":
		if a, ok := p.current(); ok {
			return p, func() tea.Msg { return ExecuteMsg{ID: a.ID} }
		}
	case "backspace":
		if len(p.query) > 0 {
			p.query = p.query[:len(p.query)-1]
			p.refilter()
		}
	default:
		// printable char → append to query
		if r := key.String(); len(r) == 1 {
			p.query += r
			p.refilter()
		}
	}
	return p, nil
}

func (p *Palette) refilter() {
	p.filtered = p.filtered[:0]
	q := strings.ToLower(p.query)
	for _, a := range p.actions {
		if q == "" || strings.Contains(strings.ToLower(a.Label), q) {
			p.filtered = append(p.filtered, a)
		}
	}
	p.selected = 0
	p.scrollStart = 0
}
```

### Rendering กลางจอ

```go
func (p Palette) View() string {
	// modal width = 60% of screen, clamp
	modalW := p.width * 6 / 10
	if modalW < 40 {
		modalW = 40
	}

	// 1) query input
	queryLine := lipgloss.NewStyle().
		Background(styles.PanelBg).
		Foreground(styles.AssistantText).
		Width(modalW).
		Padding(0, 2).
		Render("> " + p.query + "▎")

	// 2) results (visible window only)
	end := p.scrollStart + p.maxVisible
	if end > len(p.filtered) {
		end = len(p.filtered)
	}
	var rows []string
	for i, a := range p.filtered[p.scrollStart:end] {
		style := lipgloss.NewStyle().
			Background(styles.PanelBg).
			Foreground(styles.AssistantText).
			Width(modalW).
			Padding(0, 2)
		if p.scrollStart+i == p.selected {
			style = style.Background(styles.UserBorder).Bold(true)
		}
		rows = append(rows, style.Render(a.Label))
	}

	// 3) ประกอบ + border
	box := lipgloss.JoinVertical(lipgloss.Left,
		append([]string{queryLine, ""}, rows...)...)
	box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.UserBorder).
		Background(styles.PanelBg).
		Render(box)

	// 4) place กลางจอ — Place pad ทุกรอบด้วย transparent (chat ทะลุได้)
	return lipgloss.Place(p.width, p.height,
		lipgloss.Center, lipgloss.Center,
		box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(styles.Background)),
	)
}
```

---

## Routing ใน app

### Update — ส่ง key ไปยัง active view

`app/update.go`:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		// route ลงทุก view (chat, input, sidebar, palette)
		m.chat, _    = m.chat.Update(msg)
		m.input, _   = m.input.Update(msg)
		m.palette, _ = m.palette.Update(msg)
		return m, nil

	case tea.KeyPressMsg:
		// 1) ESC ใน modal → กลับ chat
		if m.activeView == ViewPalette && msg.String() == "esc" {
			m.activeView = ViewChat
			return m, nil
		}

		// 2) Ctrl+P เปิด palette (จาก view ไหนก็ได้)
		if msg.String() == "ctrl+p" {
			m.activeView = ViewPalette
			m.palette = components.NewPalette(allActions)
			return m, nil
		}

		// 3) route ไป active view
		switch m.activeView {
		case ViewPalette:
			var cmd tea.Cmd
			m.palette, cmd = m.palette.Update(msg)
			return m, cmd
		case ViewChat:
			// route ลง input/chat ปกติ (ดู Recipe 1)
		}

	case components.ExecuteMsg:
		// palette ยิงกลับมา → ทำ action + ปิด modal
		m = executeAction(m, msg.ID)
		m.activeView = ViewChat
		return m, nil
	}

	return m, nil
}

func executeAction(m Model, id string) Model {
	switch id {
	case "clear-chat":
		m.chat = components.NewChat(m.chatW, m.chatH)
	case "switch-model":
		// ...
	}
	return m
}
```

### View — overlay เมื่อ modal เปิด

`app/view.go`:

```go
func (m Model) View() string {
	base := m.composeChatScreen() // chat + input + sidebar ปกติ

	if m.activeView == ViewPalette {
		// modal วาดทับ base ทั้งหมด (มันจัดกลางเองด้วย Place)
		modal := m.palette.View()
		return overlayCenter(base, modal)
	}
	return base
}

// overlayCenter splice modal บรรทัดที่ตรงกลางของ base แทน
func overlayCenter(base, overlay string) string {
	baseLines    := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")
	// palette ใช้ Place เต็มจอแล้ว — ก็แค่คืน overlay เลย
	if len(overlayLines) >= len(baseLines) {
		return overlay
	}
	start := (len(baseLines) - len(overlayLines)) / 2
	for i, line := range overlayLines {
		if strings.TrimSpace(stripANSI(line)) == "" {
			continue // เซลล์โปร่ง — ปล่อย base โผล่
		}
		baseLines[start+i] = line
	}
	return strings.Join(baseLines, "\n")
}
```

---

## เพิ่มหน้าใหม่ (เช่น Settings)

1. เพิ่ม const ใน View enum:
   ```go
   ViewSettings
   ```
2. สร้าง `components/settings.go` (Init/Update/View)
3. เพิ่ม field ใน `Model` + `New(...)`
4. route key ใน switch `m.activeView`
5. compose ใน `View()` — เลือกระหว่าง chat / settings ตาม activeView

```go
func (m Model) View() string {
	switch m.activeView {
	case ViewSettings:
		return m.settings.View() // เต็มจอ แทน chat
	default:
		return m.composeChatScreen()
	}
}
```

---

## Gotchas

1. **ESC ต้องอยู่บนสุดของ route** — ไม่งั้น textarea จะกินก่อน
2. **อย่ารัน Ctrl+P จากใน modal** — เช็ค `m.activeView != ViewPalette`
   ก่อน toggle เพื่อกัน flicker
3. **WindowSizeMsg ต้อง broadcast ทุก view** — ไม่งั้นเปิด modal แล้ว
   resize จอ modal จะค้างไซส์เดิม
4. **action execution ผ่าน Msg** — palette ไม่ควรเรียก `m.chat.Clear()`
   ตรงๆ (ละเมิดกฎ component-isolation) ส่ง `ExecuteMsg{ID: ...}` กลับ
   ให้ `app` จัดการ
5. **state ของ modal reset ทุกครั้งที่เปิด** — `m.palette = NewPalette(...)`
   ใน Ctrl+P handler ไม่งั้น query เก่าค้าง
