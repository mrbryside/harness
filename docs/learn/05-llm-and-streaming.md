# 05 — LLM & Streaming

LLM ทุกตัวเข้าผ่าน interface เดียว — app ไม่รู้จัก provider ตัวจริง

## Interface (`llm/provider.go`)

```go
type LLMProvider interface {
    Name() string
    ChatCompletion(ctx context.Context, messages []Message, opts Options) (<-chan Chunk, error)
}

type Message struct { Role, Content string }
type Chunk   struct { Content string; TokensUsed int; Done bool; Err error }
type Options struct { Model string; Temperature float64 }
```

**กฎ:**
- `ChatCompletion` คืน **read-only channel** — caller drain จนปิด หรือเจอ
  `Done=true`
- ถ้า error ระดับ stream → ใส่ `Err` ในตัว Chunk (พร้อม `Done=true`)
- ถ้า error ก่อนเริ่มเลย → คืน `(nil, error)`
- Provider ต้อง honour `ctx.Done()` ใน goroutine ตัวเอง

---

## Mock provider (`llm/mock.go`)

มีไว้สำหรับ dev/test โดยไม่ต้องเชื่อมต่อจริง

**สิ่งที่ทำ:**
1. สุ่ม response จาก `mockResponses[]` (มีหลายแบบ: plain, code block,
   table, list, blockquote, multi-language code)
2. ตัดเป็น token ด้วย regex `\s+|\S+` — เก็บ whitespace ไว้ครบ
   (สำคัญ! ไม่งั้น markdown rendering พังตอน streaming)
3. ส่ง chunk ละตัวทุก 40ms ด้วย goroutine
4. คำนวณ token estimate = `round(wordCount * 1.3)`

**Mock ใช้เป็นแม่แบบของ provider จริง** — ทุก provider ใหม่ต้องเลียน
แบบ pattern นี้ (goroutine + channel + ctx.Done check)

---

## Streaming flow ใน app

ดู `app/update.go` — มี 2 จุด:

### 1. User กด Enter → เริ่ม stream

```go
case components.SendMsg:
    m.messages = append(m.messages, llm.Message{Role: "user", Content: msg.Content})
    m.chat.AppendMessage("user", msg.Content)
    m.chat.AppendMessage("assistant", "")    // placeholder ว่าง
    m.streaming = true

    ch, err := m.provider.ChatCompletion(ctx, m.messages, llm.Options{})
    if err != nil { ... }
    m.streamCh = ch
    cmds = append(cmds, nextChunk(ch))       // schedule read แรก
```

### 2. Chunk เข้ามา → ต่อข้อความ + schedule ตัวต่อไป

```go
case chunkMsg:
    c := msg.chunk
    if c.Err != nil { ... }
    m.chat.AppendChunk(c.Content)            // ต่อท้าย assistant message
    m.sidebar.SetTokens(c.TokensUsed)
    if c.Done {
        m.streaming = false; m.streamCh = nil
        return m, nil
    }
    cmds = append(cmds, nextChunk(m.streamCh))  // อ่านตัวถัดไป
```

`nextChunk` (`app/update.go:18`) คือ Cmd ที่ block อ่าน channel 1 ตัว
แล้วคืนเป็น `chunkMsg`:

```go
func nextChunk(ch <-chan llm.Chunk) tea.Cmd {
    return func() tea.Msg {
        chunk, ok := <-ch
        if !ok { return chunkMsg{chunk: llm.Chunk{Done: true}} }
        return chunkMsg{chunk: chunk}
    }
}
```

**ทำไมต้อง re-schedule แต่ละ chunk?**
Bubble Tea Cmd รันครั้งเดียวจบ. ถ้าจะอ่านยาวต้อง yield message กลับ
loop แล้วค่อย dispatch Cmd ใหม่ — แบบนี้ UI render ได้ระหว่าง chunk
(ถ้า loop ใน goroutine เดียวจะฟรีซ)

---

## เพิ่ม provider ใหม่

ตัวอย่าง — provider จริงที่ stream ผ่าน HTTP/SSE:

```go
// llm/openai.go
package llm

type OpenAIProvider struct {
    apiKey string
    client *http.Client
}

func (p *OpenAIProvider) Name() string { return "gpt-4o" }

func (p *OpenAIProvider) ChatCompletion(ctx context.Context, msgs []Message, opts Options) (<-chan Chunk, error) {
    req, err := p.buildRequest(ctx, msgs, opts)
    if err != nil { return nil, err }

    resp, err := p.client.Do(req)
    if err != nil { return nil, err }

    ch := make(chan Chunk)
    go func() {
        defer close(ch)
        defer resp.Body.Close()
        // อ่าน SSE event-by-event
        for {
            select {
            case <-ctx.Done():
                ch <- Chunk{Err: ctx.Err(), Done: true}
                return
            default:
            }
            event, err := readSSE(resp.Body)
            if errors.Is(err, io.EOF) {
                ch <- Chunk{Done: true}
                return
            }
            if err != nil {
                ch <- Chunk{Err: err, Done: true}
                return
            }
            ch <- Chunk{Content: event.Delta, TokensUsed: event.Tokens}
        }
    }()
    return ch, nil
}
```

แล้วใน `main.go` แค่สลับ:
```go
provider := llm.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"))
model := app.New(provider)
```

**app ไม่ต้องแก้บรรทัดเดียว** — นั่นคือจุดประสงค์ของ interface

---

## Cancellation

ตอนนี้ app **ยังไม่ได้** cancel context เมื่อ user ปิดหรือเริ่ม stream
ใหม่ — เป็นหัวข้อที่ต้องเพิ่มในอนาคต (เก็บ `context.CancelFunc` ใน Model
แล้วเรียกตอน `tea.QuitMsg` หรือเริ่ม `SendMsg` ใหม่)

---

## Gotchas

- **อย่าลืม close channel** ใน provider — caller จะรอตลอดกาล
- **อย่าลืม `defer close`** ใน goroutine — panic ทำให้ leak ได้
- Chunk **สุดท้าย** ของ stream ต้องมี `Done=true` (หรือ channel closed)
- Whitespace สำคัญสำหรับ markdown — ตัดเป็น token อย่าทิ้ง `\n` หรือ `  `
- TokensUsed เป็น **running total** ไม่ใช่ delta — sidebar ใช้ค่าล่าสุด
