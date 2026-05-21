package components

// SelectStart begins a new selection at (line, col) inside the input body.
func (i *Input) SelectStart(line, col int) { i.sel.Start(line, col) }

// SelectExtend moves the selection's end anchor.
func (i *Input) SelectExtend(line, col int) { i.sel.Extend(line, col) }

// SelectClear drops any in-flight selection.
func (i *Input) SelectClear() { i.sel.Clear() }

// HasSelection reports whether a selection is active.
func (i Input) HasSelection() bool { return i.sel.Active() }

// SelectedText returns the plain text covered by the current selection.
func (i Input) SelectedText() string { return i.sel.Text(StripANSI(i.textarea.View())) }

// AddHistory stores a sent message in the history buffer.
func (i *Input) AddHistory(text string) {
	if i.history != nil {
		i.history.Add(text)
	}
}
