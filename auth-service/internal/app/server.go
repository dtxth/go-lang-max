package app

import (
    "log"
    "net/http"
)

type Server struct {
    Handler http.Handler
    Port    string
}

func (s *Server) Run() {
    log.Println("Starting server on port", s.Port)
    log.Fatal(http.ListenAndServe(":"+s.Port, s.Handler))
}