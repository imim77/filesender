package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait  = 10 * time.Second
)

var upgarder = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

type ErrorMessage struct {
	Type string `json:"type"`
	Code int    `json:"code"`
}

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

type Client struct {
	core *Core
	conn *websocket.Conn
	send chan any
	info ClientInfo
}

func newClient(connection *websocket.Conn, core *Core) *Client {
	return &Client{core: core, conn: connection, send: make(chan any, 256)}
}

func (c *Client) readPump() {
	defer func() {
		c.core.unregister <- c
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			return
		}
		var msg WsClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("[%s] invalid JSON: %v", c.conn.RemoteAddr(), err)
			continue
		}
		switch msg.Type {
		case "UPDATE":
			if msg.Info != nil {
				c.info.ClientInfoWithoutId = *msg.Info
				c.core.broadcast <- UpdateMessage{Type: "UPDATE", Peer: c.info}
			}
		case "OFFER", "ANSWER", "CANDIDATE":
			if msg.Target != "" {
				c.core.sendTo(msg.Target, msg, c)
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(core *Core, w http.ResponseWriter, r *http.Request) {
	if err := godotenv.Load(); err != nil {
		slog.Error("error loading .env file", "error", err)
	}
	conn, err := upgarder.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("error upgrading a connection", "error", err)
		return
	}
	client := newClient(conn, core)
	client.info.Id = uuid.New()
	externalIceServers, err := parseExternalIceServers(os.Getenv("EXTERNAL_ICE_SERVERS_JSON"))
	if err != nil {
		slog.Error("invalid EXTERNAL_ICE_SERVERS_JSON", "error", err)
		externalIceServers = nil
	}

	client.core.register <- client
	peers, _ := core.getPeers(client.info.Id)
	client.send <- HelloMessage{
		Type:       "HELLO",
		Client:     client.info,
		Peers:      peers,
		IceServers: externalIceServers,
	}
	client.core.broadcast <- JoinMessage{Type: "JOIN", Peer: client.info}

	go client.writePump()
	go client.readPump()

}
