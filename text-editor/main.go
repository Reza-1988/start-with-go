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
	newTwoStack := &TwoStackEditor{
		LeftStack:  NewStack(),
		RightStack: NewStack(),
	}
	runes := []rune(initialText)
	for i := len(runes) - 1; i >= 0; i-- {
		newTwoStack.RightStack.Push(runes[i])
	}
	return newTwoStack

}

func (e *TwoStackEditor) InsertRune(r rune) {
	e.LeftStack.Push(r)
}

func (e *TwoStackEditor) InsertString(text string) {
	for _, r := range text {
		e.LeftStack.Push(r)
	}
}

func (e *TwoStackEditor) CursorLeft() {
	if len(e.LeftStack.text) == 0 {
		return
	}
	r, _ := e.LeftStack.Pop()
	e.RightStack.Push(r)

}
func (e *TwoStackEditor) CursorRight() {
	if len(e.RightStack.text) == 0 {
		return
	}
	r, _ := e.RightStack.Pop()
	e.LeftStack.Push(r)
}

func (e *TwoStackEditor) DeleteForward() {
	if len(e.RightStack.text) == 0 {
		return
	}
	_, _ = e.RightStack.Pop()
}
func (e *TwoStackEditor) DeleteBackward() {
	if len(e.LeftStack.text) == 0 {
		return
	}
	_, _ = e.LeftStack.Pop()
}

func (e *TwoStackEditor) WholeText() string {
	left := e.LeftStack.text
	right := e.RightStack.text

	result := make([]rune, 0, len(left)+len(right))

	for i := 0; i < len(left); i++ {
		result = append(result, left[i])
	}

	for i := len(right) - 1; i >= 0; i-- {
		result = append(result, right[i])
	}

	return string(result)
}
