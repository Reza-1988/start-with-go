package main

type TikTak struct {
	n     int
	tikCh chan struct{}
	takCh chan struct{}
}

func NewTikTak(n int) *TikTak {
	t := &TikTak{
		n:     n,
		tikCh: make(chan struct{}, 1),
		takCh: make(chan struct{}),
	}
	t.tikCh <- struct{}{}
	return t
}

func (t *TikTak) Tik() {
	for i := 0; i < t.n; i++ {
		<-t.tikCh
		SayTik()
		t.takCh <- struct{}{}
	}
}

func (t *TikTak) Tak() {
	for i := 0; i < t.n; i++ {
		<-t.takCh
		SayTak()
		t.tikCh <- struct{}{}
	}
}
