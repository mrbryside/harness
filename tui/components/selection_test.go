package components

import "testing"

func TestSelectionExtend(t *testing.T) {
	var s Selection
	s.Start(10, 5)
	s.Extend(15, 20)

	sl, sc, el, ec := s.Normalised()
	if sl != 10 || sc != 5 || el != 15 || ec != 20 {
		t.Fatalf("got (%d,%d,%d,%d), want (10,5,15,20)", sl, sc, el, ec)
	}
}

func TestSelectionInactiveExtendNoOp(t *testing.T) {
	var s Selection
	s.Extend(10, 5)
	if s.Active() {
		t.Fatal("extend on inactive selection should not activate it")
	}
}