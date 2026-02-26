package main

import (
	"encoding/json"
	"log/slog"

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
	sendToCh   chan Messeger
}

func newCore() *Core {
	return &Core{
		clients:    make(map[uuid.UUID]*Client),
		broadcast:  make(chan any),
		register:   make(chan Client),
		unregister: make(chan Client),
		sendToCh:   make(chan Messeger),
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
			//case message := <-c.broadcast:
			//case message := <-c.sendToChan:

		}
	}
}

type Messeger interface {
	getId() uuid.UUID
	getMsg() any
}

type targetWsServerSDPMessage struct {
	id      uuid.UUID
	message WsServerSdpMessage
}

func (t targetWsServerSDPMessage) getId() uuid.UUID {
	return t.id
}

func (t targetWsServerSDPMessage) getMsg() any {
	return t.message
}

type targetWsServerCandidateMessage struct {
	id      uuid.UUID
	message WsServerCandidateMessage
}

func (t targetWsServerCandidateMessage) getId() uuid.UUID {
	return t.id
}

func (t targetWsServerCandidateMessage) getMsg() any {
	return t.message
}

func (c Core) sendTo(targetId string, msg WsClientMessage, cli *Client) {
	id, err := uuid.Parse(targetId)
	if err != nil {
		slog.Error("error while parsing targetId", "error", err)
	}
	switch msg.Type {
	case "OFFER":
		c.sendToCh <- targetWsServerSDPMessage{id: id,
			message: WsServerSdpMessage{Type: "OFFER", Peer: cli.info, SessionID: msg.SessionID, SDP: msg.SDP}}
	case "ANSWER":
		c.sendToCh <- targetWsServerSDPMessage{id: id,
			message: WsServerSdpMessage{Type: "ANSWER", Peer: cli.info, SessionID: msg.SessionID, SDP: msg.SDP}}
	case "CANDIDATE":
		c.sendToCh <- targetWsServerCandidateMessage{id: id,
			message: WsServerCandidateMessage{Type: "CANDIDATE", Peer: cli.info, SessionID: msg.SessionID, Candidate: msg.Candidate}}
	default:
		slog.Error("unknown message type", "error", err)
	}
}
