package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/eventbus"
	"github.com/mrbryside/harness/tui/components"
)

func newTestModel() Model {
	eb := eventbus.NewEventBus()
	return New(eb)
}

func TestHandleQuestionIntegration(t *testing.T) {
	m := newTestModel()

	if m.activeQuestion != nil && m.activeQuestion.question != nil && m.activeQuestion.question.Active() {
		t.Fatal("expected question to be inactive initially")
	}

	m.AskPermission("Execute command?", "test-q1")
	if m.activeQuestion == nil || m.activeQuestion.question == nil || !m.activeQuestion.question.Active() {
		t.Fatal("expected question to be active after AskPermission")
	}

	// Simulate the question's HandleKey returning a QuestionAnswerMsg.
	questionCmd, handled := m.activeQuestion.question.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !handled {
		t.Fatal("expected Enter to be handled")
	}
	answerMsg := questionCmd().(components.QuestionAnswerMsg)

	// Process the answer through the handler directly.
	model, _ := m.handleQuestionAnswer(answerMsg)
	m = model.(Model)

	if m.activeQuestion != nil && m.activeQuestion.question != nil && m.activeQuestion.question.Active() {
		t.Error("expected question to be hidden after answer")
	}
}

func TestHandleQuestionAnswerEmitsEvent(t *testing.T) {
	eb := eventbus.NewEventBus()
	m := New(eb)

	var receivedEvent *eventbus.Event
	eb.Subscribe(eventbus.EventQuestionAnswered, func(e eventbus.Event) {
		receivedEvent = &e
	})

	m.AskPermission("Test?", "q1")

	questionCmd, _ := m.activeQuestion.question.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	answerMsg := questionCmd().(components.QuestionAnswerMsg)

	model, _ := m.handleQuestionAnswer(answerMsg)
	m = model.(Model)

	if receivedEvent == nil {
		t.Fatal("expected event to be emitted")
	}
	if receivedEvent.Type != eventbus.EventQuestionAnswered {
		t.Errorf("expected event type %q, got %q", eventbus.EventQuestionAnswered, receivedEvent.Type)
	}

	data, ok := receivedEvent.Data.(struct {
		QuestionID string
		Answer     components.QuestionChoice
	})
	if !ok {
		t.Fatalf("expected data struct, got %T", receivedEvent.Data)
	}
	if data.QuestionID != "q1" {
		t.Errorf("expected QuestionID='q1', got %q", data.QuestionID)
	}
	if data.Answer.Label != "Yes, allow this permission." {
		t.Errorf("expected Label='Yes, allow this permission.', got %q", data.Answer.Label)
	}
	if data.Answer.Index != 0 {
		t.Errorf("expected Index=0, got %d", data.Answer.Index)
	}
}

func TestHandleQuestionAnswerYesEmitsTrue(t *testing.T) {
	eb := eventbus.NewEventBus()
	m := New(eb)

	var choice components.QuestionChoice
	eb.Subscribe(eventbus.EventQuestionAnswered, func(e eventbus.Event) {
		data := e.Data.(struct {
			QuestionID string
			Answer     components.QuestionChoice
		})
		choice = data.Answer
	})

	m.AskPermission("Test?", "q2")

	questionCmd, _ := m.activeQuestion.question.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	answerMsg := questionCmd().(components.QuestionAnswerMsg)
	model, _ := m.handleQuestionAnswer(answerMsg)
	m = model.(Model)

	if choice.Index != 0 {
		t.Errorf("expected Index=0, got %d", choice.Index)
	}
}

func TestHandleQuestionBlocksInputWhenActive(t *testing.T) {
	m := newTestModel()

	m.AskPermission("Test?", "q1")

	model, _ := m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	m = model.(Model)

	inputValue := m.input.Value()
	if inputValue != "" {
		t.Errorf("expected input to be empty while question is active, got %q", inputValue)
	}
}

func TestHandleQuestionDownTogglesSelection(t *testing.T) {
	m := newTestModel()

	m.AskPermission("Test?", "q1")

	_, handled := m.activeQuestion.question.HandleKey(tea.KeyPressMsg{Code: tea.KeyDown})
	if !handled {
		t.Fatal("expected Down to be handled")
	}

	if m.activeQuestion == nil || m.activeQuestion.question == nil || !m.activeQuestion.question.Active() {
		t.Error("expected question to still be active after Down")
	}
	if m.activeQuestion.question.SelectedIndex() != 1 {
		t.Errorf("expected index 1 after Down, got %d", m.activeQuestion.question.SelectedIndex())
	}
}

func TestHandleQuestionEnterConfirmsSelected(t *testing.T) {
	eb := eventbus.NewEventBus()
	m := New(eb)

	var choice components.QuestionChoice
	eb.Subscribe(eventbus.EventQuestionAnswered, func(e eventbus.Event) {
		data := e.Data.(struct {
			QuestionID string
			Answer     components.QuestionChoice
		})
		choice = data.Answer
	})

	m.AskPermission("Test?", "q1")

	m.activeQuestion.question.HandleKey(tea.KeyPressMsg{Code: tea.KeyDown})
	questionCmd, _ := m.activeQuestion.question.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	answerMsg := questionCmd().(components.QuestionAnswerMsg)
	model, _ := m.handleQuestionAnswer(answerMsg)
	m = model.(Model)

	if choice.Index != 1 {
		t.Errorf("expected Index=1 (No selected), got %d", choice.Index)
	}
	if choice.Label != "No, not allow." {
		t.Errorf("expected Label='No, not allow.', got %q", choice.Label)
	}
}

func TestHandleQuestionEscCancels(t *testing.T) {
	eb := eventbus.NewEventBus()
	m := New(eb)

	var choice components.QuestionChoice
	var received bool
	eb.Subscribe(eventbus.EventQuestionAnswered, func(e eventbus.Event) {
		received = true
		data := e.Data.(struct {
			QuestionID string
			Answer     components.QuestionChoice
		})
		choice = data.Answer
	})

	m.AskPermission("Test?", "q1")

	questionCmd, _ := m.activeQuestion.question.HandleKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	answerMsg := questionCmd().(components.QuestionAnswerMsg)
	model, _ := m.handleQuestionAnswer(answerMsg)
	m = model.(Model)

	if !received {
		t.Fatal("expected event to be emitted on Esc")
	}
	if choice.Index != -1 {
		t.Errorf("expected Index=-1 for cancel, got %d", choice.Index)
	}
	if choice.Label != "cancelled" {
		t.Errorf("expected Label='cancelled', got %q", choice.Label)
	}
}

func TestHandleQuestionCtrlCCancels(t *testing.T) {
	eb := eventbus.NewEventBus()
	m := New(eb)

	var choice components.QuestionChoice
	var received bool
	eb.Subscribe(eventbus.EventQuestionAnswered, func(e eventbus.Event) {
		received = true
		data := e.Data.(struct {
			QuestionID string
			Answer     components.QuestionChoice
		})
		choice = data.Answer
	})

	m.AskPermission("Test?", "q1")

	cancelCmd := m.activeQuestion.question.Cancel()
	answerMsg := cancelCmd().(components.QuestionAnswerMsg)
	model, _ := m.handleQuestionAnswer(answerMsg)
	m = model.(Model)

	if !received {
		t.Fatal("expected event to be emitted on Ctrl+C")
	}
	if choice.Index != -1 {
		t.Errorf("expected Index=-1 for cancel, got %d", choice.Index)
	}
}

func TestHandleQuestionRegistryCreate(t *testing.T) {
	q := components.CreateQuestion(components.QuestionTypePermission, "test-id", "Test question?")
	if q == nil {
		t.Fatal("expected question to be created")
	}
	if q.Type() != components.QuestionTypePermission {
		t.Errorf("expected type %q, got %q", components.QuestionTypePermission, q.Type())
	}
	if q.ID() != "test-id" {
		t.Errorf("expected ID='test-id', got %q", q.ID())
	}
}

func TestHandleQuestionRegistryUnknownType(t *testing.T) {
	q := components.CreateQuestion("unknown_type", "test-id", "Test?")
	if q != nil {
		t.Error("expected nil for unknown question type")
	}
}
