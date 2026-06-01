package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jrperin/copilot-go-proxy/internal/copilot"
)

type Server struct {
	port    int
	client  *copilot.Client
	mux     *http.ServeMux
	httpSrv *http.Server
}

func New(port int, client *copilot.Client) *Server {
	s := &Server{
		port:   port,
		client: client,
		mux:    http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/v1/messages", s.handleMessages)
	s.mux.HandleFunc("/v1/chat/completions", s.handleChatCompletions)
	s.mux.HandleFunc("/chat/completions", s.handleChatCompletions)
	s.mux.HandleFunc("/v1/models", s.handleModels)
	s.mux.HandleFunc("/models", s.handleModels)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/", s.handleHealth)
}

func (s *Server) Start() error {
	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler: s.mux,
	}

	log.Printf("copilot-proxy listening on http://127.0.0.1:%d", s.port)
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.httpSrv != nil {
		return s.httpSrv.Close()
	}
	return nil
}
