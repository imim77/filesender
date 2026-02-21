package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/turn/v4"
)

type HelloMessage struct {
	Type       string          `json:"type"`
	Client     ClientInfo      `json:"client"`
	Peers      []ClientInfo    `json:"peers"`
	IceServers []IceServerInfo `json:"iceServers,omitempty"`
}

type IceServerInfo struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type JoinMessage struct {
	Type string     `json:"type"`
	Peer ClientInfo `json:"peer"`
}

type Config struct {
	Host         string
	Port         string
	PublicHost   string
	PublicIp     string
	TURNPort     string
	TURNRealm    string
	TURNSecret   string
	RelayPortMin uint16
	RelayPortMax uint16
}

type Server struct {
	cfg        *Config
	core       *Core
	turnServer *turn.Server
	httpServer *http.Server
}

func NewServer(cfg *Config) (*Server, error) {
	core := NewCore()

	turnServer, err := startTURN(cfg)
	if err != nil {
		log.Printf("warning: TURN disabled: %v", err)
	}

	srv := &Server{cfg: cfg, core: core, turnServer: turnServer}
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
	client := NewClient(uuid.New(), conn, s.core)
	log.Printf("Client connected: %s (%s)", client.ClientId, conn.RemoteAddr())
	res, err := s.core.Register(client)
	if err != nil {
		log.Printf("failed to register client %s: %v", client.ClientId, err)
		conn.Close()
		return
	}
	client.loop()
	hello := HelloMessage{
		Type:       "HELLO",
		Client:     client.GetPublicInfo(),
		Peers:      res.Peers,
		IceServers: buildIceServers(s.cfg, r, s.turnServer != nil),
	}
	if err := client.SendJSON(hello); err != nil {
		log.Printf("[%s] failed to send HELLO: %v", client.ClientId, err)
		_ = s.core.Unregister(client.ClientId)
		return
	}

	join := JoinMessage{Type: "JOIN", Peer: client.GetPublicInfo()}
	joinData, _ := json.Marshal(join)

	for _, peer := range res.Existing {
		if err := peer.Send(joinData); err != nil {
			log.Printf("[%s] failed to notify peer %s: %v", client.ClientId, peer.ClientId, err)
		}
	}
	log.Printf("client %s actor started, handling messages independently", client.ClientId)
}
