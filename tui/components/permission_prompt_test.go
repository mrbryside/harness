package components

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestPermissionPromptInitialState(t *testing.T) {
	p := NewPermissionPrompt()
	if p.Active() {
		t.Error("expected prompt to be inactive initially")
	}
}

func TestPermissionPromptShowHide(t *testing.T) {
	p := NewPermissionPrompt()

	p.Show("Execute command?", "q1")
	if !p.Active() {
		t.Fatal("expected prompt to be active after Show")
	}
	if p.selectedIndex != 0 {
		t.Errorf("expected selectedIndex=0, got %d", p.selectedIndex)
	}

	p.Hide()
	if p.Active() {
		t.Error("expected prompt to be inactive after Hide")
	}
}

func TestPermissionPromptViewReturnsEmptyWhenInactive(t *testing.T) {
	p := NewPermissionPrompt()
	view := p.View(80)
	if view != "" {
		t.Errorf("expected empty view when inactive, got %q", view)
	}
}

func TestPermissionPromptViewContainsQuestion(t *testing.T) {
	p := NewPermissionPrompt()
	p.Show("Execute rm -rf /?", "dangerous")

	view := p.View(80)
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

func TestPermissionPromptViewVerticalButtons(t *testing.T) {
	p := NewPermissionPrompt()
	p.Show("Test?", "q1")

	view := p.View(80)
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

func TestPermissionPromptHandleKeyUpDownToggles(t *testing.T) {
	p := NewPermissionPrompt()
	p.Show("Test?", "q1")

	_, handled := p.HandleKey(tea.KeyPressMsg{Code: tea.KeyDown})
	if !handled {
		t.Fatal("expected Down to be handled")
	}
	if p.selectedIndex != 1 {
		t.Errorf("expected selectedIndex=1 after Down, got %d", p.selectedIndex)
	}

	_, handled = p.HandleKey(tea.KeyPressMsg{Code: tea.KeyUp})
	if !handled {
		t.Fatal("expected Up to be handled")
	}
	if p.selectedIndex != 0 {
		t.Errorf("expected selectedIndex=0 after Up, got %d", p.selectedIndex)
	}
}

func TestPermissionPromptHandleKeyEnter(t *testing.T) {
	p := NewPermissionPrompt()
	p.Show("Test?", "q1")

	cmd, handled := p.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !handled {
		t.Fatal("expected key to be handled")
	}

	msg := cmd()
	answer := msg.(PermissionAnswerMsg)
	if !answer.Answer {
		t.Errorf("expected Answer=true for Enter (Yes selected)")
	}
}

func TestPermissionPromptHandleKeyEnterOnNo(t *testing.T) {
	p := NewPermissionPrompt()
	p.Show("Test?", "q1")
	p.selectedIndex = 1

	cmd, handled := p.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !handled {
		t.Fatal("expected key to be handled")
	}

	msg := cmd()
	answer := msg.(PermissionAnswerMsg)
	if answer.Answer {
		t.Errorf("expected Answer=false for Enter when No selected")
	}
}

func TestPermissionPromptHandleKeyIgnoredWhenInactive(t *testing.T) {
	p := NewPermissionPrompt()

	cmd, handled := p.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if handled {
		t.Error("expected key to be ignored when inactive")
	}
	if cmd != nil {
		t.Error("expected nil cmd when inactive")
	}
}

func TestPermissionPromptHandleKeyEscCancels(t *testing.T) {
	p := NewPermissionPrompt()
	p.Show("Test?", "q1")

	cmd, handled := p.HandleKey(tea.KeyPressMsg{Code: tea.KeyEsc})
	if !handled {
		t.Fatal("expected Esc to be handled")
	}

	msg := cmd()
	answer := msg.(PermissionAnswerMsg)
	if answer.Answer {
		t.Errorf("expected Answer=false for Esc")
	}
}

func TestPermissionPromptHandleKeySpaceConfirms(t *testing.T) {
	p := NewPermissionPrompt()
	p.Show("Test?", "q1")

	cmd, handled := p.HandleKey(tea.KeyPressMsg{Code: ' ', Text: " "})
	if !handled {
		t.Fatal("expected Space to be handled")
	}

	msg := cmd()
	answer := msg.(PermissionAnswerMsg)
	if !answer.Answer {
		t.Errorf("expected Answer=true for Space (Yes selected)")
	}
}

func TestPermissionPromptOverlayReturnsOriginalWhenInactive(t *testing.T) {
	p := NewPermissionPrompt()
	original := "line1\nline2\nline3"
	result := p.Overlay(original, 80)
	if result != original {
		t.Errorf("expected original view when inactive")
	}
}

func TestPermissionPromptOverlayReplacesLines(t *testing.T) {
	p := NewPermissionPrompt()
	p.Show("Test?", "q1")

	chatView := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10"
	result := p.Overlay(chatView, 80)

	if !strings.Contains(result, "Yes") {
		t.Error("expected overlay to contain 'Yes'")
	}
	if strings.Contains(result, "line10") {
		t.Error("expected last lines to be replaced by overlay")
	}
}

func TestPermissionPromptString(t *testing.T) {
	p := NewPermissionPrompt()
	s := p.String()
	if s != "PermissionPrompt{inactive}" {
		t.Errorf("expected inactive string, got %q", s)
	}

	p.Show("Test?", "q1")
	s = p.String()
	if !strings.Contains(s, "Test?") {
		t.Errorf("expected string to contain question")
	}
	if !strings.Contains(s, "q1") {
		t.Errorf("expected string to contain questionID")
	}
}

func stripAnsiForTest(s string) string {
	re := regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")
	return re.ReplaceAllString(s, "")
}
