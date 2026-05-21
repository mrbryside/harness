# Harness — Learn

เอกสารชุดนี้เขียนเพื่อให้คนใหม่อ่านจบแล้วลงมือพัฒนา UI ของ Harness ได้
ทันที โดยไม่ต้องไปอ่านโค้ดทั้ง repo ก่อน อ่านตามลำดับเล่มด้านล่าง

## ลำดับการอ่าน

1. [`01-overview.md`](./01-overview.md) — ภาพรวมโปรเจกต์, โครงสร้าง, กฎเหล็ก
2. [`02-bubble-tea-basics.md`](./02-bubble-tea-basics.md) — Elm-style loop, Model/Update/View, Cmd/Msg
3. [`03-layout-and-styling.md`](./03-layout-and-styling.md) — lipgloss, Place/Join, การคำนวณ layout, theme
4. [`04-components.md`](./04-components.md) — chat / input / sidebar / statusbar
5. [`05-llm-and-streaming.md`](./05-llm-and-streaming.md) — `LLMProvider` interface และ streaming pattern
6. [`06-recipes.md`](./06-recipes.md) — สูตรสำเร็จสำหรับงานที่เจอบ่อย
7. [`07-glossary.md`](./07-glossary.md) — คำศัพท์ + gotchas
8. [`08-popups-and-overlays.md`](./08-popups-and-overlays.md) — auto-complete popup, overlay เหนือ layout
9. [`09-views-and-routing.md`](./09-views-and-routing.md) — view enum, Ctrl+P command palette, สลับหน้า
10. [`10-input-patterns.md`](./10-input-patterns.md) — slash commands, history, scroll, validate, edge cases

## TL;DR

- **Bubble Tea v2** เป็น framework แบบ Elm: `Update(Model, Msg) → (Model, Cmd)` แล้ว `View(Model) → string`
- ทุก component เป็น standalone Bubble Tea sub-model (มี `Init`/`Update`/`View` ของตัวเอง)
- App layer (`app/`) ทำหน้าที่ route message + ประกอบ layout เท่านั้น — ไม่ render เนื้อหา component เอง
- LLM access ทุกตัวผ่าน interface `llm.LLMProvider` + streaming ผ่าน `<-chan llm.Chunk`
- สีทุกสีอยู่ใน `styles/theme.go` ที่เดียว
- TDD บังคับ: ทดสอบล้มก่อน → เขียน implementation → ทดสอบผ่าน
