package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Config struct {
	Host string
	Port string
}

type Server struct {
	cfg        *Config
	httpServer *http.Server
}

func NewServer(cfg *Config) (*Server, error) {
	srv := &Server{cfg: cfg}
	srv.httpServer = &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		Handler: srv,
	}
	return srv, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ws":
		s.handleWebSocket(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server error: %w", err)
	}
	return nil
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	wsUpgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading to websocket: %v", err)
		http.Error(w, "failed to upgrade to websocket", http.StatusBadRequest)
		return
	}
	client := NewClient(uuid.New(), conn)
	log.Printf("Client connected: %s (%s)", client.ClientId, conn.RemoteAddr())

}
