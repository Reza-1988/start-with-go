package main

type Editor interface {
	InsertRune(rune)
	InsertString(string)

	CursorLeft()
	CursorRight()

	DeleteForward()
	DeleteBackward()

	WholeText() string
}
