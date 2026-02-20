package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait  = 10 * time.Second
)

type ErrorMessage struct {
	Type string `json:"type"`
	Code int    `json:"code"`
}

type Client struct {
	ClientId  uuid.UUID
	Conn      *websocket.Conn
	CloseOnce sync.Once
	mu        sync.RWMutex
	info      ClientInfoWithoutId
	core      *Core
	Msgch     chan []byte
	Close     chan struct{}
}

func NewClient(id uuid.UUID, conn *websocket.Conn, core *Core) *Client {
	return &Client{
		ClientId: id,
		Conn:     conn,
		core:     core,
		Msgch:    make(chan []byte, 64),
		Close:    make(chan struct{}),
	}
}

func (c *Client) loop() {
	go c.writeLoop()
	go c.readLoop()
}

func (c *Client) readLoop() {
	defer func() {
		if c.core != nil {
			if err := c.core.Unregister(c.ClientId); err != nil && err != ErrCoreClosed {
				log.Printf("[%s] unregister failed: %v", c.ClientId, err)
			}
		}
		c.CloseCon()
	}()
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[%s] read error: %v", c.ClientId, err)
			}
			return
		}
		var msg WsClientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("[%s] invalid JSON: %v", c.ClientId, err)
			_ = c.SendJSON(ErrorMessage{Type: "ERROR", Code: http.StatusBadRequest})
			continue
		}
		if err := c.core.RouteMessage(c.ClientId, msg); err != nil {
			log.Printf("[%s] error routing message: %v", c.ClientId, err)
			_ = c.SendJSON(ErrorMessage{Type: "ERROR", Code: http.StatusBadRequest})
		}

	}
}

func (c *Client) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.CloseCon()
	}()
	for {
		select {
		case <-c.Close:
			return
		case msg := <-c.Msgch:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("[%s] write error: %v", c.ClientId, err)
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		}
	}
}

func (c *Client) Send(payload []byte) error {
	select {
	case <-c.Close:
		return ErrClientClosed
	case c.Msgch <- payload:
		return nil
	default:
		return ErrMailboxFull
	}
}

func (c *Client) SendJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Send(data)
}

func (c *Client) CloseCon() {
	c.CloseOnce.Do(func() {
		close(c.Close)
		_ = c.Conn.Close()
	})
}

func (c *Client) GetPublicInfo() ClientInfo {
	c.mu.RLock()
	info := c.info
	c.mu.RUnlock()

	return ClientInfo{
		Id:                  c.ClientId,
		ClientInfoWithoutId: info,
	}
}

func (c *Client) SetInfo(info ClientInfoWithoutId) {
	c.mu.Lock()
	c.info = info
	c.mu.Unlock()
}
