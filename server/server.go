package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	Host               string
	Port               string
	ExternalIceServers []IceServerInfo
}

type Server struct {
	cfg        *Config
	httpServer *http.Server
}

func NewServer(cfg Config, wsHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, wsHandler)
	return mux
}

func addRoutes(mux *http.ServeMux, wsHandler http.Handler) {
	mux.Handle("/ws", wsHandler)
}

func parseExternalIceServers(raw string) ([]IceServerInfo, error) {
	if raw == "" {
		return nil, nil
	}

	var servers []IceServerInfo
	if err := json.Unmarshal([]byte(raw), &servers); err != nil {
		return nil, err
	}

	for i, server := range servers {
		if len(server.URLs) == 0 {
			return nil, fmt.Errorf("entry %d has empty urls", i)
		}
	}

	return servers, nil
}
