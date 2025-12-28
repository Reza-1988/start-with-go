package main

import "net"

type Server struct {
	Port string // TCP port to listen on, e.g. "4001"
}

// Response is the top-level message sent from server to client (tests).
type Response struct {
	Result string `json:"result""`
	Error  string `json:"error"`
}

// Request is the top-level message sent from client (tests) to the TCP server.
// Protocol notes:
//   - Communication is JSON over raw TCP (not HTTP).
//   - Client uses json.Encoder.Encode(...) which writes one JSON object per request.
type Request struct {
	Action  string `json:"action,omitempty"`
	Numbers string `json:"numbers,omitempty"`
}

func NewServer(port string) *Server {
	srv := &Server{
		Port: port,
	}
	return srv
}

func (s *Server) handleAdd(data Response)

func (s *Server) Start() {
	ln, err := net.Listen("tcp", ":"+port)
}
