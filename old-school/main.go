package main

import (
	"database/sql"
	"net"
	"sync"
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

func NewServer() Server {

	return &server{}
}
