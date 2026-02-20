package main

import (
	"encoding/json"
	"errors"
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

type coreMessage interface {
	isCoreMessage()
}

type registerMsg struct {
	Client   *Client
	Response chan RegisterResult
}

func (m registerMsg) isCoreMessage() {}

type RegisterMsg struct {
	Client *ClientInfo
	Resp   chan RegisterResult
}

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
			for i, _ := range c.clients {
				c.clients[i].CloseCon()
			}
			return
		case msg := <-c.MessChan:
			switch m := msg.(type) {
			case registerMsg:
				c.handleRegister(m)
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

func (c *Core) RegisterPeer(cli *Client) (RegisterResult, error) {
	resp := make(chan RegisterResult, 1)
	if err := c.Enqueue(registerMsg{Client: cli, Response: resp}); err != nil {
		return RegisterResult{}, nil
	}
	select {
	case result := <-resp:
		return result, nil
	case <-c.closed:
		return RegisterResult{}, ErrCoreClosed
	}

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
