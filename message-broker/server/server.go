package server

import (
	"net"
)

type Server struct {
	Addr string
}

func NewServer(address string) *Server {
	return &Server{
		Addr: address,
	}
}

func (s *Server) Run() error

func (s *Server) Stop()

func (s *Server) GetTopic(topicName string) (*Topic, bool)

func (s *Server) GetClientConnections() []net.Conn
