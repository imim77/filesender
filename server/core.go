package main

import (
	"encoding/json"
	"fmt"
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
	Type   string    `json:"type"`
	PeerID uuid.UUID `json:"peerId"`
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
				if client.conn != nil {
					fmt.Printf("Delete client %v from the map\n", client.conn.RemoteAddr())
				}

				leftMessage := LeftMessage{
					Type:   "LEFT",
					PeerID: client.info.Id,
				}
				for _, peer := range c.clients {
					select {
					case peer.send <- leftMessage:
					default:
						close(peer.send)
						delete(c.clients, peer.info.Id)
					}
				}
			}
		case message := <-c.broadcast:
			for _, client := range c.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(c.clients, client.info.Id)
				}
			}
		case dm := <-c.sendToCh:
			id := dm.getId()
			msg := dm.getMsg()
			if client, ok := c.clients[id]; ok {
				select {
				case client.send <- msg:
					fmt.Println(msg)
					fmt.Printf("sending direct message: %s", msg)
				default:
					close(client.send)
					delete(c.clients, id)
				}
			}

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

func (c Core) getPeers(excludeId uuid.UUID) ([]ClientInfo, []*Client) {
	peers := make([]ClientInfo, len(c.clients))
	result := make([]*Client, len(c.clients))
	for id, client := range c.clients {
		if id == excludeId {
			continue
		}
		peers = append(peers, client.info)
		result = append(result, client)
	}
	return peers, result
}
