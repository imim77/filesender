package main

import (
	"encoding/json"

	"github.com/google/uuid"
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

type MessageHandler func(m *Manager, clientID uuid.UUID) error

type Manager struct {
	MessChan chan struct{}
	clients  map[uuid.UUID]*Client
	handlers map[string]MessageHandler
}

func NewManager() *Manager {
	return &Manager{
		MessChan: make(chan struct{}),
		clients:  make(map[uuid.UUID]*Client),
	}
}
