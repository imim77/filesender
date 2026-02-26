package main

import (
	"encoding/json"

	"github.com/google/uuid"
)

type WsClientMessage struct {
	Type      string               `json:"type"`
	SessionID string               `json:"sessionId,omitempty"`
	Target    string               `json:"target,omitempty"`
	SDP       string               `json:"sdp,omitempty"`
	Candidate json.RawMessage      `json:"candidate,omitempty"`
	Info      *ClientInfoWithoutId `json:"info,omitempty"`
}

type UpdateMessage struct {
	Type string     `json:"type"`
	Peer ClientInfo `json:"peer"`
}

type LeftMessage struct {
	Type   string `json:"type"`
	PeerID string `json:"peerId"`
}

type WsServerSdpMessage struct {
	Type      string     `json:"type"`
	Peer      ClientInfo `json:"peer"`
	SessionID string     `json:"sessionId"`
	SDP       string     `json:"sdp"`
}

type WsServerCandidateMessage struct {
	Type      string          `json:"type"`
	Peer      ClientInfo      `json:"peer"`
	SessionID string          `json:"sessionId"`
	Candidate json.RawMessage `json:"candidate"`
}

type Core struct {
	clients    map[uuid.UUID]*Client
	broadcast  chan any
	register   chan Client
	unregister chan Client
}

func newCore() *Core {
	return &Core{
		clients:    make(map[uuid.UUID]*Client),
		broadcast:  make(chan any),
		register:   make(chan Client),
		unregister: make(chan Client),
	}
}

func (c Core) run() {
	for {
		select {
		case client := <-c.register:
			c.clients[client.info.Id] = &client
		case client := <-c.unregister:
			if _, ok := c.clients[client.info.Id]; ok {
				delete(c.clients, client.info.Id)
				close(client.send)
			}
		case message := <-c.broadcast:

		}
	}
}
