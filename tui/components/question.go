package components

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// QuestionAnswerMsg is emitted when any question is answered.
// Data is always a QuestionChoice struct.
type QuestionAnswerMsg struct {
	QuestionID string
	Data       QuestionChoice
}

// QuestionChoice represents the user's answer to a question.
type QuestionChoice struct {
	Index int    // selected option index (-1 = cancelled)
	Label string // display label of the selected option
}

// Question is the interface all question types must implement.
// To add a new question type:
//  1. Create a new file (e.g. question_yourtype.go)
//  2. Implement this interface
//  3. Register it with QuestionRegistry
type Question interface {
	// Type returns the question type identifier (e.g. "permission").
	Type() string

	// ID returns the unique question ID.
	ID() string

	// Active reports whether this question is currently shown.
	Active() bool

	// Show displays the question with the given text.
	Show(question string)

	// Hide removes the question from view.
	Hide()

	// HandleKey processes a key event. Returns a cmd and whether the key was handled.
	// If the key confirms/cancels the question, cmd returns a QuestionAnswerMsg.
	HandleKey(msg tea.Msg) (tea.Cmd, bool)

	// Cancel emits a cancel answer (equivalent to answering No).
	Cancel() tea.Cmd

	// OverlayView renders the question for display in the input area.
	OverlayView(width int) string

	// Height returns the number of lines the rendered question occupies.
	Height(width int) int

	// OptionCount returns the number of selectable options.
	OptionCount() int

	// SelectedIndex returns the currently selected option index.
	SelectedIndex() int

	// MoveSelection moves the selection by dir (+1 down, -1 up), clamped to valid range.
	MoveSelection(dir int)

	// OptionLabel returns the display label for the option at the given index.
	OptionLabel(index int) string

	fmt.Stringer
}

// QuestionRegistry holds factories for creating question types.
type QuestionRegistry struct {
	factories map[string]func(id, question string) Question
}

// NewQuestionRegistry creates an empty registry.
func NewQuestionRegistry() *QuestionRegistry {
	return &QuestionRegistry{
		factories: make(map[string]func(id, question string) Question),
	}
}

// Register adds a question type factory to the registry.
func (r *QuestionRegistry) Register(qtype string, factory func(id, question string) Question) {
	r.factories[qtype] = factory
}

// Create instantiates a question of the given type.
func (r *QuestionRegistry) Create(qtype, id, question string) Question {
	factory, ok := r.factories[qtype]
	if !ok {
		return nil
	}
	return factory(id, question)
}

// DefaultRegistry is the global registry with built-in question types.
var DefaultRegistry = NewQuestionRegistry()

func init() {
	DefaultRegistry.Register(QuestionTypePermission, func(id, question string) Question {
		return NewPermissionQuestion(id, question)
	})
}

// Question type constants.
const (
	QuestionTypePermission = "permission"
)

// CreateQuestion creates a question using the default registry.
func CreateQuestion(qtype, id, question string) Question {
	return DefaultRegistry.Create(qtype, id, question)
}

func countLines(s string) int {
	return strings.Count(s, "\n") + 1
}
