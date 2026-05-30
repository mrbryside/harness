package components

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestPermissionQuestionInitialState(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")
	if !q.Active() {
		t.Error("expected question to be active after creation")
	}
	if q.Type() != QuestionTypePermission {
		t.Errorf("expected type %q, got %q", QuestionTypePermission, q.Type())
	}
	if q.ID() != "q1" {
		t.Errorf("expected ID='q1', got %q", q.ID())
	}
}

func TestPermissionQuestionShowHide(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")
	q.Hide()
	if q.Active() {
		t.Error("expected question to be inactive after Hide")
	}

	q.Show("New question?")
	if !q.Active() {
		t.Error("expected question to be active after Show")
	}
}

func TestPermissionQuestionViewReturnsEmptyWhenInactive(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")
	q.Hide()
	view := q.OverlayView(80)
	if view != "" {
		t.Errorf("expected empty view when inactive, got %q", view)
	}
}

func TestPermissionQuestionViewContainsQuestion(t *testing.T) {
	q := NewPermissionQuestion("q1", "Execute rm -rf /?")

	view := q.OverlayView(80)
	if !strings.Contains(view, "Execute rm -rf /?") {
		t.Errorf("expected view to contain question text")
	}
	if !strings.Contains(view, "Yes") {
		t.Errorf("expected view to contain 'Yes'")
	}
	if !strings.Contains(view, "No") {
		t.Errorf("expected view to contain 'No'")
	}
}

func TestPermissionQuestionViewVerticalButtons(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")

	view := q.OverlayView(80)
	lines := strings.Split(view, "\n")

	var yesIdx, noIdx int
	foundYes := false
	foundNo := false
	for i, line := range lines {
		clean := stripAnsiForTest(line)
		if strings.Contains(clean, "Yes") {
			yesIdx = i
			foundYes = true
		}
		if strings.Contains(clean, "No") {
			noIdx = i
			foundNo = true
		}
	}
	if !foundYes || !foundNo {
		t.Fatal("expected both Yes and No in view")
	}
	if noIdx <= yesIdx {
		t.Errorf("expected Yes to appear before No (vertical layout), Yes=%d No=%d", yesIdx, noIdx)
	}
}

func TestPermissionQuestionMoveSelection(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")

	// Start at 0 (Yes)
	if q.SelectedIndex() != 0 {
		t.Errorf("expected initial index 0, got %d", q.SelectedIndex())
	}

	// Down → 1 (No)
	q.MoveSelection(1)
	if q.SelectedIndex() != 1 {
		t.Errorf("expected index 1 after Down, got %d", q.SelectedIndex())
	}

	// Down again → clamped to 1
	q.MoveSelection(1)
	if q.SelectedIndex() != 1 {
		t.Errorf("expected index clamped to 1, got %d", q.SelectedIndex())
	}

	// Up → 0 (Yes)
	q.MoveSelection(-1)
	if q.SelectedIndex() != 0 {
		t.Errorf("expected index 0 after Up, got %d", q.SelectedIndex())
	}

	// Up again → clamped to 0
	q.MoveSelection(-1)
	if q.SelectedIndex() != 0 {
		t.Errorf("expected index clamped to 0, got %d", q.SelectedIndex())
	}
}

func TestPermissionQuestionOptionCount(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")
	if q.OptionCount() != 2 {
		t.Errorf("expected 2 options, got %d", q.OptionCount())
	}
}

func TestPermissionQuestionOptionLabel(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")
	if q.OptionLabel(0) != "Yes, allow this permission." {
		t.Errorf("expected label 0, got %q", q.OptionLabel(0))
	}
	if q.OptionLabel(1) != "No, not allow." {
		t.Errorf("expected label 1, got %q", q.OptionLabel(1))
	}
	if q.OptionLabel(-1) != "" {
		t.Errorf("expected empty label for -1, got %q", q.OptionLabel(-1))
	}
	if q.OptionLabel(2) != "" {
		t.Errorf("expected empty label for 2, got %q", q.OptionLabel(2))
	}
}

func TestPermissionQuestionHandleKeyUpDown(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")

	// Down → index 1
	_, handled := q.HandleKey(tea.KeyPressMsg{Code: tea.KeyDown})
	if !handled {
		t.Fatal("expected Down to be handled")
	}
	if q.SelectedIndex() != 1 {
		t.Errorf("expected index 1 after Down, got %d", q.SelectedIndex())
	}

	// Up → index 0
	_, handled = q.HandleKey(tea.KeyPressMsg{Code: tea.KeyUp})
	if !handled {
		t.Fatal("expected Up to be handled")
	}
	if q.SelectedIndex() != 0 {
		t.Errorf("expected index 0 after Up, got %d", q.SelectedIndex())
	}
}

func TestPermissionQuestionHandleKeyEnter(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")

	cmd, handled := q.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !handled {
		t.Fatal("expected key to be handled")
	}

	msg := cmd()
	answer, ok := msg.(QuestionAnswerMsg)
	if !ok {
		t.Fatalf("expected QuestionAnswerMsg, got %T", msg)
	}
	if answer.QuestionID != "q1" {
		t.Errorf("expected QuestionID='q1', got %q", answer.QuestionID)
	}
	if answer.Data.Index != 0 {
		t.Errorf("expected Index=0 for Enter (Yes selected), got %d", answer.Data.Index)
	}
	if answer.Data.Label != "Yes, allow this permission." {
		t.Errorf("expected Label='Yes, allow this permission.', got %q", answer.Data.Label)
	}
}

func TestPermissionQuestionHandleKeyEnterOnNo(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")
	q.selectedIndex = 1

	cmd, handled := q.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !handled {
		t.Fatal("expected key to be handled")
	}

	msg := cmd()
	answer := msg.(QuestionAnswerMsg)
	if answer.Data.Index != 1 {
		t.Errorf("expected Index=1 for Enter when No selected, got %d", answer.Data.Index)
	}
	if answer.Data.Label != "No, not allow." {
		t.Errorf("expected Label='No, not allow.', got %q", answer.Data.Label)
	}
}

func TestPermissionQuestionHandleKeyIgnoredWhenInactive(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")
	q.Hide()

	cmd, handled := q.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if handled {
		t.Error("expected key to be ignored when inactive")
	}
	if cmd != nil {
		t.Error("expected nil cmd when inactive")
	}
}

func TestPermissionQuestionHandleKeyEscCancels(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")

	cmd, handled := q.HandleKey(tea.KeyPressMsg{Code: tea.KeyEsc})
	if !handled {
		t.Fatal("expected Esc to be handled")
	}

	msg := cmd()
	answer := msg.(QuestionAnswerMsg)
	if answer.Data.Index != -1 {
		t.Errorf("expected Index=-1 for Esc, got %d", answer.Data.Index)
	}
	if answer.Data.Label != "cancelled" {
		t.Errorf("expected Label='cancelled', got %q", answer.Data.Label)
	}
}

func TestPermissionQuestionHandleKeySpaceConfirms(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")

	cmd, handled := q.HandleKey(tea.KeyPressMsg{Code: ' ', Text: " "})
	if !handled {
		t.Fatal("expected Space to be handled")
	}

	msg := cmd()
	answer := msg.(QuestionAnswerMsg)
	if answer.Data.Index != 0 {
		t.Errorf("expected Index=0 for Space (Yes selected), got %d", answer.Data.Index)
	}
	if answer.Data.Label != "Yes, allow this permission." {
		t.Errorf("expected Label='Yes, allow this permission.', got %q", answer.Data.Label)
	}
}

func TestPermissionQuestionCancel(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")

	cmd := q.Cancel()
	if cmd == nil {
		t.Fatal("expected cmd to be non-nil")
	}

	msg := cmd()
	answer := msg.(QuestionAnswerMsg)
	if answer.Data.Index != -1 {
		t.Errorf("expected Index=-1 for Cancel, got %d", answer.Data.Index)
	}
	if answer.Data.Label != "cancelled" {
		t.Errorf("expected Label='cancelled', got %q", answer.Data.Label)
	}
}

func TestPermissionQuestionCancelWhenInactive(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")
	q.Hide()

	cmd := q.Cancel()
	if cmd != nil {
		t.Error("expected nil cmd when inactive")
	}
}

func TestPermissionQuestionHeight(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")

	h := q.Height(80)
	if h < 4 {
		t.Errorf("expected height >= 4, got %d", h)
	}

	q.Hide()
	h = q.Height(80)
	if h != 0 {
		t.Errorf("expected height=0 when inactive, got %d", h)
	}
}

func TestPermissionQuestionString(t *testing.T) {
	q := NewPermissionQuestion("q1", "Test?")
	s := q.String()
	if !strings.Contains(s, "Test?") {
		t.Errorf("expected string to contain question")
	}
	if !strings.Contains(s, "q1") {
		t.Errorf("expected string to contain questionID")
	}

	q.Hide()
	s = q.String()
	if s != "PermissionQuestion{inactive}" {
		t.Errorf("expected inactive string, got %q", s)
	}
}

func TestQuestionRegistryCreate(t *testing.T) {
	q := CreateQuestion(QuestionTypePermission, "test-id", "Test question?")
	if q == nil {
		t.Fatal("expected question to be created")
	}
	if q.Type() != QuestionTypePermission {
		t.Errorf("expected type %q, got %q", QuestionTypePermission, q.Type())
	}
	if q.ID() != "test-id" {
		t.Errorf("expected ID='test-id', got %q", q.ID())
	}
}

func TestQuestionRegistryUnknownType(t *testing.T) {
	q := CreateQuestion("unknown_type", "test-id", "Test?")
	if q != nil {
		t.Error("expected nil for unknown question type")
	}
}

func stripAnsiForTest(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(s, "")
}
