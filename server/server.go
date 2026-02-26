package main

import (
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
	Host string
	Port string
}

type Server struct {
	cfg        *Config
	core       *Core
	httpServer *http.Server
}
