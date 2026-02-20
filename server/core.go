package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrClientNotFound = errors.New("client not found")
	ErrMailboxFull    = errors.New("mailbox full")
	ErrClientClosed   = errors.New("client closed")
	ErrCoreClosed     = errors.New("core closed")
)

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
		MessChan: make(chan coreMessage),
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
	default:
		return ErrMailboxFull

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
	cli.CloseCon()
	delete(c.clients, msg.ClientID)
}

func (c *Core) handleRoute(clientId uuid.UUID, msg WsClientMessage) error {
	switch msg.Type {
	case "UPDATE":
		return c.handleUpdate(clientId, msg)
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

func (c *Core) getClient(clientId uuid.UUID) (*Client, error) {
	client, ok := c.clients[clientId]
	if !ok {
		return nil, ErrClientNotFound
	}
	return client, nil
}

func (c *Core) getOthers(excludeId uuid.UUID) []*Client {
	others := make([]*Client, len(c.clients))
	for id, client := range c.clients {
		if id == excludeId {
			continue
		}
		others = append(others, client)
	}
	return others
}
