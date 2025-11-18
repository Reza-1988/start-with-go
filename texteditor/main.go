package main

import "fmt"

type Stack struct {
	text []rune
}

func NewStack() *Stack {
	return &Stack{
		text: make([]rune, 0),
	}
}

func (s *Stack) Top() (rune, error) {
	if len(s.text) == 0 {
		return ' ', fmt.Errorf("stack is empty")
	}
	topText := s.text[len(s.text)-1]
	return topText, nil
}

func (s *Stack) Push(r rune) {
	s.text = append(s.text, r)
}
func (s *Stack) Pop() (rune, error) {
	if len(s.text) == 0 {
		return ' ', fmt.Errorf("stack is empty")
	}
	lastText := s.text[len(s.text)-1]
	s.text = s.text[:len(s.text)-1]
	return lastText, nil
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
