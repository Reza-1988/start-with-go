package main

type Stack struct {
}

func (s Stack) Top() (rune, error) {
	return ' ', nil
}

type TwoStackEditor struct {
	LeftStack  *Stack
	RightStack *Stack
	// rest
}

func NewTwoStackEditor(initialText string) *TwoStackEditor {
	return nil
}

func (e *TwoStackEditor) InsertRune(r rune) {
}

func (e *TwoStackEditor) InsertString(text string) {
}

func (e *TwoStackEditor) CursorLeft() {
}
func (e *TwoStackEditor) CursorRight() {
}

func (e *TwoStackEditor) DeleteForward() {
}
func (e *TwoStackEditor) DeleteBackward() {
}

func (e *TwoStackEditor) WholeText() string {
	return ""
}
