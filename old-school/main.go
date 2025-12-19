package main

import (
	"net"
)

const (
	CreateSchoolMethod      = "/school/create"
	CreateClassMethod       = "/class/create"
	CreatePersonMethod      = "/person/create"
	AddStudentToClassMethod = "/class/add/student"
	WhoAmIMethod            = "/who/am/i"
)

func main() {
}

type Server interface {
	Start(port string) error
	Stop() error
}

type server struct {
	listener net.Listener
}

func NewServer() Server {
	return &server{}
}

func (s *server) Start(port string) error {
	// TODO
	return nil
}

func (s *server) Stop() error {
	// TODO
	return nil
}
