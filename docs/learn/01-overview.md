# 01 — Overview

## โปรเจกต์คืออะไร

Harness เป็น **TUI (Terminal User Interface)** สำหรับคุยกับ LLM provider
หน้าตาคล้าย OpenCode / Claude Code เขียนด้วย Go บน Bubble Tea v2

โครงหน้าจอ:

```
┌─────────────────────────────────────────────┬──────────┐
│                                             │          │
│   chat (scrollable message history)         │ sidebar  │
│                                             │          │
│                                             │ - model  │
│                                             │ - tokens │
│   input (textarea, dynamic height)          │ - status │
├─────────────────────────────────────────────┤          │
│ statusbar (mode · hints)                    │          │
└─────────────────────────────────────────────┴──────────┘
```

## โครงสร้าง package

```
main.go               เลือก provider + start tea.Program
app/                  orchestration: routing, layout, streaming loop
  model.go            root Model (ถือทุก sub-component)
  update.go           message router + reflowChat helper
  view.go             ประกอบ layout ด้วย lipgloss
components/           UI ชิ้น standalone (ตัวละไฟล์)
  chat.go             viewport ของ messages
  input.go            textarea + footer (mode/model)
  sidebar.go          panel ขวา
  statusbar.go        แถวล่าง
llm/                  provider abstraction
  provider.go         interface LLMProvider + types
  mock.go             MockProvider สำหรับ dev/test
styles/theme.go       สีทั้งหมดของแอป
docs/learn/           เอกสารชุดนี้
```

## ความรับผิดชอบของแต่ละ layer

| Layer | ทำอะไร | ห้ามทำ |
|---|---|---|
| `main` | wiring provider + start program | มี business logic |
| `app` | routing tea msg, streaming loop, layout math | render เนื้อหา component, นิยามสี |
| `components` | UI ชิ้นเดียวจบในตัว | รู้จัก component อื่น, รู้จัก `llm.*` |
| `llm` | interface + types + mock | แตะ UI / Bubble Tea |
| `styles` | สีอย่างเดียว | depend อะไรใน repo |

## กฎเหล็ก (ห้ามฝ่าฝืน)

1. **สี** อยู่ใน `styles/theme.go` ที่เดียว ห้าม hardcode hex ที่อื่น
2. **LLM access** ผ่าน interface `llm.LLMProvider` เท่านั้น app ห้าม import `MockProvider` หรือ provider concrete ตรงๆ
3. **Streaming** ต้องคืน `<-chan llm.Chunk` เสมอ — ห้ามเปลี่ยนเป็น one-shot
4. **Layout** ใช้ `lipgloss.JoinHorizontal/JoinVertical/Place` เท่านั้น ห้ามใช้ ANSI ดิบหรือ absolute positioning
5. **`main.go`** ไม่มี business logic — เลือก provider + start program จบ
6. **TDD** เขียน test ที่ fail ก่อนเขียน implementation เสมอ
7. **Mouse events** อยู่ใน program (`tea.MouseModeCellMotion` ใน `app/view.go`) — terminal scrollback ถูก capture ไว้

## Dev commands

```bash
go run .             # รัน TUI
go build -o harness  # build binary
go test ./...        # รัน test ทั้งหมด
go mod tidy          # หลังเพิ่ม/ลบ dep
```

Module path: `github.com/mrbryside/harness` (อย่าเปลี่ยน)
