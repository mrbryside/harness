package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mrbryside/harness/llm"
)

func TestPermissionPromptIntegration(t *testing.T) {
	m := New(&llm.MockProvider{})

	// Initially not active
	if m.permissionPrompt.Active() {
		t.Fatal("expected prompt to be inactive initially")
	}

	// Show the prompt
	m.AskPermission("Execute command?", "test-q1")
	if !m.permissionPrompt.Active() {
		t.Fatal("expected prompt to be active after AskPermission")
	}

	// Simulate user pressing Enter (Yes is selected by default)
	model, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = model.(Model)

	// Cmd should have been returned
	if cmd == nil {
		t.Fatal("expected cmd to be non-nil")
	}

	// Process the answer message
	msg := cmd()
	model, _ = m.Update(msg)
	m = model.(Model)

	// Prompt should be hidden after answer is processed
	if m.permissionPrompt.Active() {
		t.Error("expected prompt to be hidden after answer")
	}
}

func TestPermissionAnswerEmitsEvent(t *testing.T) {
	m := New(&llm.MockProvider{})

	var receivedEvent *Event
	m.eventBus.Subscribe(EventQuestionAnswered, func(e Event) {
		receivedEvent = &e
	})

	m.AskPermission("Test?", "q1")

	// Simulate pressing Enter (No selected by default = false)
	model, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = model.(Model)

	// Execute the cmd to trigger PermissionAnswerMsg
	if cmd != nil {
		msg := cmd()
		model, _ := m.Update(msg)
		m = model.(Model)
	}

	if receivedEvent == nil {
		t.Fatal("expected event to be emitted")
	}
	if receivedEvent.Type != EventQuestionAnswered {
		t.Errorf("expected event type %q, got %q", EventQuestionAnswered, receivedEvent.Type)
	}

	data, ok := receivedEvent.Data.(struct {
		QuestionID string
		Answer     bool
	})
	if !ok {
		t.Fatalf("expected data struct, got %T", receivedEvent.Data)
	}
	if data.QuestionID != "q1" {
		t.Errorf("expected QuestionID='q1', got %q", data.QuestionID)
	}
	// Enter with default selection (Yes) should return true
	if !data.Answer {
		t.Errorf("expected Answer=true for Enter (Yes selected by default)")
	}
}

func TestPermissionAnswerYesEmitsTrue(t *testing.T) {
	m := New(&llm.MockProvider{})

	var answer bool
	m.eventBus.Subscribe(EventQuestionAnswered, func(e Event) {
		data := e.Data.(struct {
			QuestionID string
			Answer     bool
		})
		answer = data.Answer
	})

	m.AskPermission("Test?", "q2")

	model, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = model.(Model)

	if cmd != nil {
		msg := cmd()
		model, _ := m.Update(msg)
		m = model.(Model)
	}

	if !answer {
		t.Errorf("expected Answer=true for 'y' key")
	}
}

func TestPermissionPromptBlocksInputWhenActive(t *testing.T) {
	m := New(&llm.MockProvider{})

	m.AskPermission("Test?", "q1")

	// Type some text - should be ignored by input
	model, _ := m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	m = model.(Model)

	// Input should still be empty since prompt is active
	inputValue := m.input.Value()
	if inputValue != "" {
		t.Errorf("expected input to be empty while prompt is active, got %q", inputValue)
	}
}

func TestPermissionPromptTabTogglesSelection(t *testing.T) {
	m := New(&llm.MockProvider{})

	m.AskPermission("Test?", "q1")

	// Down should toggle selection, not answer
	model, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = model.(Model)

	// Prompt should still be active
	if !m.permissionPrompt.Active() {
		t.Error("expected prompt to still be active after Down")
	}

	// No cmd should be returned (just selection change)
	if cmd != nil {
		t.Error("expected nil cmd after Down (no answer yet)")
	}
}

func TestPermissionPromptEnterConfirmsSelected(t *testing.T) {
	m := New(&llm.MockProvider{})

	var answer bool
	m.eventBus.Subscribe(EventQuestionAnswered, func(e Event) {
		data := e.Data.(struct {
			QuestionID string
			Answer     bool
		})
		answer = data.Answer
	})

	m.AskPermission("Test?", "q1")

	// Down to No, then Enter
	model, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = model.(Model)
	model, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = model.(Model)

	if cmd != nil {
		msg := cmd()
		model, _ := m.Update(msg)
		m = model.(Model)
	}

	if answer {
		t.Errorf("expected Answer=false when No is selected and Enter pressed")
	}
}

func TestPermissionPromptEscCancels(t *testing.T) {
	m := New(&llm.MockProvider{})

	var answer bool
	var received bool
	m.eventBus.Subscribe(EventQuestionAnswered, func(e Event) {
		received = true
		data := e.Data.(struct {
			QuestionID string
			Answer     bool
		})
		answer = data.Answer
	})

	m.AskPermission("Test?", "q1")

	model, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = model.(Model)

	if cmd != nil {
		msg := cmd()
		model, _ := m.Update(msg)
		m = model.(Model)
	}

	if !received {
		t.Fatal("expected event to be emitted on Esc")
	}
	if answer {
		t.Errorf("expected Answer=false for Esc")
	}
}

func TestPermissionPromptCtrlCCancels(t *testing.T) {
	m := New(&llm.MockProvider{})

	var answer bool
	var received bool
	m.eventBus.Subscribe(EventQuestionAnswered, func(e Event) {
		received = true
		data := e.Data.(struct {
			QuestionID string
			Answer     bool
		})
		answer = data.Answer
	})

	m.AskPermission("Test?", "q1")

	model, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	m = model.(Model)

	if cmd != nil {
		msg := cmd()
		model, _ := m.Update(msg)
		m = model.(Model)
	}

	if !received {
		t.Fatal("expected event to be emitted on Ctrl+C")
	}
	if answer {
		t.Errorf("expected Answer=false for Ctrl+C")
	}
}
