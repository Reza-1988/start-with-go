package main

type TikTak struct {
	n int
}

func NewTikTak(n int) *TikTak {
	return &TikTak{
		n: n,
	}
}

func (t *TikTak) Tik() {
	// TODO
}

func (t *TikTak) Tak() {
	// TODO
}
