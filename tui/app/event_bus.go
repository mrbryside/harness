package app

import "sync"

const (
	EventQuestionAnswered = "question_answered"
)

type Event struct {
	Type string
	Data interface{}
}

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]func(Event)
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]func(Event)),
	}
}

func (eb *EventBus) Subscribe(eventType string, handler func(Event)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Emit(eventType string, data interface{}) {
	eb.mu.RLock()
	handlers := eb.handlers[eventType]
	eb.mu.RUnlock()
	for _, h := range handlers {
		h(Event{Type: eventType, Data: data})
	}
}
