package main

type Engine struct {
	ID int
}
type Wing struct {
	ID     int
	Engine *Engine
}
type Wheel struct {
	ID int
}
type Cabin struct {
	ID int
}

type Aircraft struct {
	ID     int
	Wings  []*Wing
	Cabin  *Cabin
	Wheels []*Wheel
}

type Nest struct {
	ID        int
	Aircrafts []*Aircraft
}
