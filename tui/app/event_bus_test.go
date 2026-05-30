package app

import (
	"sync"
	"testing"
)

func TestEventBusEmitAndSubscribe(t *testing.T) {
	eb := NewEventBus()

	var received []Event
	eb.Subscribe("test_event", func(e Event) {
		received = append(received, e)
	})

	eb.Emit("test_event", "hello")

	if len(received) != 1 {
		t.Fatalf("expected 1 event, got %d", len(received))
	}
	if received[0].Type != "test_event" {
		t.Errorf("expected type 'test_event', got %q", received[0].Type)
	}
	if received[0].Data.(string) != "hello" {
		t.Errorf("expected data 'hello', got %v", received[0].Data)
	}
}

func TestEventBusMultipleSubscribers(t *testing.T) {
	eb := NewEventBus()

	var count int
	eb.Subscribe("inc", func(Event) { count++ })
	eb.Subscribe("inc", func(Event) { count++ })
	eb.Subscribe("inc", func(Event) { count++ })

	eb.Emit("inc", nil)

	if count != 3 {
		t.Errorf("expected count=3, got %d", count)
	}
}

func TestEventBusDifferentEvents(t *testing.T) {
	eb := NewEventBus()

	var gotA, gotB bool
	eb.Subscribe("a", func(Event) { gotA = true })
	eb.Subscribe("b", func(Event) { gotB = true })

	eb.Emit("a", nil)
	if !gotA || gotB {
		t.Errorf("only handler 'a' should have been called")
	}

	eb.Emit("b", nil)
	if !gotA || !gotB {
		t.Errorf("both handlers should have been called")
	}
}

func TestEventBusNoSubscribers(t *testing.T) {
	eb := NewEventBus()
	eb.Emit("nobody_listening", "data")
}

func TestEventBusConcurrentSafety(t *testing.T) {
	eb := NewEventBus()

	var mu sync.Mutex
	var count int
	for i := 0; i < 10; i++ {
		eb.Subscribe("concurrent", func(Event) {
			mu.Lock()
			count++
			mu.Unlock()
		})
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eb.Emit("concurrent", nil)
		}()
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if count != 1000 {
		t.Errorf("expected 1000, got %d", count)
	}
}

type questionAnsweredData struct {
	QuestionID string
	Answer     bool
}

func TestEventBusQuestionAnsweredPayload(t *testing.T) {
	eb := NewEventBus()

	var result questionAnsweredData
	eb.Subscribe(EventQuestionAnswered, func(e Event) {
		result = e.Data.(questionAnsweredData)
	})

	eb.Emit(EventQuestionAnswered, questionAnsweredData{
		QuestionID: "q1",
		Answer:     true,
	})

	if result.QuestionID != "q1" {
		t.Errorf("expected QuestionID='q1', got %q", result.QuestionID)
	}
	if !result.Answer {
		t.Errorf("expected Answer=true")
	}
}
