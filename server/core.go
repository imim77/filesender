package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrClientNotFound = errors.New("client not found")
	ErrMailboxFull    = errors.New("mailbox full")
	ErrClientClosed   = errors.New("client closed")
	ErrCoreClosed     = errors.New("core closed")
)

const coreMailboxSize = 256

type ClientInfoWithoutId struct {
	Alias       string `json:"alias,omitempty"`
	DeviceModel string `json:"deviceModel,omitempty"`
	DeviceType  string `json:"deviceType,omitempty"`
	Token       string `json:"token,omitempty"`
}

type ClientInfo struct {
	Id uuid.UUID `json:"id"`
	ClientInfoWithoutId
}

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

// pretty ugly tbh
type coreMessage interface {
	isCoreMessage()
}

type registerMsg struct {
	Client   *Client
	Response chan RegisterResult
}

func (m registerMsg) isCoreMessage() {}

type unregisterMsg struct {
	ClientID uuid.UUID
}

func (m unregisterMsg) isCoreMessage() {}

type routeMsg struct {
	ClientId uuid.UUID
	Message  WsClientMessage
	Response chan error
}

func (m routeMsg) isCoreMessage() {}

type RegisterResult struct {
	Peers    []ClientInfo
	Existing []*Client
}

type Core struct {
	MessChan   chan coreMessage
	clients    map[uuid.UUID]*Client
	closed     chan struct{}
	closedOnce sync.Once
}

func NewCore() *Core {
	c := &Core{
		MessChan: make(chan coreMessage, coreMailboxSize),
		clients:  map[uuid.UUID]*Client{},
		closed:   make(chan struct{}),
	}
	go c.run()
	return c
}

func (c *Core) run() {
	for {
		select {
		case <-c.closed:
			for i := range c.clients {
				c.clients[i].CloseCon()
			}
			return
		case msg := <-c.MessChan:
			switch m := msg.(type) {
			case registerMsg:
				c.handleRegister(m)
			case unregisterMsg:
				c.handleUnregister(m)
			case routeMsg:
				m.Response <- c.handleRoute(m.ClientId, m.Message)
			}

		}
	}
}

func (c *Core) Enqueue(msg coreMessage) error {
	select {
	case <-c.closed:
		return ErrCoreClosed
	case c.MessChan <- msg:
		return nil
	}
}

func (c *Core) Register(cli *Client) (RegisterResult, error) {
	resp := make(chan RegisterResult, 1)
	if err := c.Enqueue(registerMsg{Client: cli, Response: resp}); err != nil {
		return RegisterResult{}, err
	}
	select {
	case result := <-resp:
		return result, nil
	case <-c.closed:
		return RegisterResult{}, ErrCoreClosed
	}

}

func (c *Core) Unregister(clientId uuid.UUID) error {
	return c.Enqueue(unregisterMsg{ClientID: clientId})
}

func (c *Core) handleRegister(msg registerMsg) {
	peers := make([]ClientInfo, 0, len(c.clients))
	existing := make([]*Client, 0, len(c.clients))
	for _, client := range c.clients {
		peers = append(peers, client.GetPublicInfo())
		existing = append(existing, client)
	}
	c.clients[msg.Client.ClientId] = msg.Client
	msg.Response <- RegisterResult{Peers: peers, Existing: existing}
}

func (c *Core) handleUnregister(msg unregisterMsg) {
	cli, ok := c.clients[msg.ClientID]
	if !ok {
		return
	}

	leftData, err := json.Marshal(LeftMessage{Type: "LEFT", PeerID: msg.ClientID.String()})
	if err == nil {
		for _, peer := range c.getOthers(msg.ClientID) {
			if sendErr := peer.Send(leftData); sendErr != nil {
				log.Printf("[%s] failed to send LEFT to %s: %v", msg.ClientID, peer.ClientId, sendErr)
			}
		}
	}

	cli.CloseCon()
	delete(c.clients, msg.ClientID)
}

func (c *Core) handleRoute(clientId uuid.UUID, msg WsClientMessage) error {
	switch msg.Type {
	case "UPDATE":
		return c.handleUpdate(clientId, msg)
	case "OFFER", "ANSWER", "CANDIDATE":
		return c.handleSignaling(clientId, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}

}

func (c *Core) handleUpdate(clientId uuid.UUID, msg WsClientMessage) error {
	cli, err := c.getClient(clientId)
	if err != nil {
		return err
	}
	if msg.Info == nil {
		return nil
	}
	cli.SetInfo(*msg.Info)

	update := UpdateMessage{Type: "UPDATE", Peer: cli.GetPublicInfo()}
	updateData, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to encode UPDATE message: %w", err)
	}

	for _, peer := range c.getOthers(clientId) {
		if err := peer.Send(updateData); err != nil {
			log.Printf("[%s] failed to send UPDATE to %s: %v", clientId, peer.ClientId, err)
		}
	}

	return nil
}

func (c *Core) handleSignaling(clientId uuid.UUID, msg WsClientMessage) error {
	cli, err := c.getClient(clientId)
	if err != nil {
		return err
	}
	if msg.Target == "" || msg.SessionID == "" {
		return cli.SendJSON(ErrorMessage{Type: "ERROR", Code: http.StatusBadRequest})
	}

	targetId, err := uuid.Parse(msg.Target)
	if err != nil {
		return err
	}

	var payload any
	switch msg.Type {
	case "OFFER", "ANSWER":
		if msg.SDP == "" {
			return cli.SendJSON(ErrorMessage{Type: "ERROR", Code: http.StatusBadRequest})
		}
		payload = WsServerSdpMessage{
			Type:      msg.Type,
			Peer:      cli.GetPublicInfo(),
			SessionID: msg.SessionID,
			SDP:       msg.SDP,
		}
	case "CANDIDATE":
		if len(msg.Candidate) == 0 {
			return cli.SendJSON(ErrorMessage{Type: "ERROR", Code: http.StatusBadRequest})
		}
		payload = WsServerCandidateMessage{
			Type:      msg.Type,
			Peer:      cli.GetPublicInfo(),
			SessionID: msg.SessionID,
			Candidate: msg.Candidate,
		}
	}
	if err := c.forward(targetId, payload); err != nil {
		return cli.SendJSON(ErrorMessage{Type: "ERROR", Code: http.StatusNotFound})
	}
	log.Printf("[SENDING PEER][%s] -> [RECIEVING PEER][%s] %s", clientId, msg.Target, msg.Type)
	return nil

}

func (c *Core) getClient(clientId uuid.UUID) (*Client, error) {
	client, ok := c.clients[clientId]
	if !ok {
		return nil, ErrClientNotFound
	}
	return client, nil
}

func (c *Core) getOthers(excludeId uuid.UUID) []*Client {
	others := make([]*Client, 0, len(c.clients))
	for id, client := range c.clients {
		if id == excludeId {
			continue
		}
		others = append(others, client)
	}
	return others
}

func (c *Core) forward(recieverPeer uuid.UUID, payload any) error {
	target, err := c.getClient(recieverPeer)
	if err != nil {
		return err
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error")
	}
	return target.Send(data)
}

func (c *Core) RouteMessage(clientId uuid.UUID, msg WsClientMessage) error {
	resp := make(chan error, 1)
	if err := c.Enqueue(routeMsg{ClientId: clientId, Message: msg, Response: resp}); err != nil {
		return err
	}
	select {
	case err := <-resp:
		return err
	case <-c.closed:
		return ErrCoreClosed
	}
}
