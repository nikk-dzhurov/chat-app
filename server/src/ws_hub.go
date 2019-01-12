package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"./model"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 2 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 2 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type BroadcastData struct {
	BroadcastToAll bool
	UserIDs []string
	Data    []byte
}

type WSHub struct {
	clients map[string][]*WsClient

	broadcast chan *BroadcastData

	register   chan *WsClient
	unregister chan *WsClient

	upgrader *websocket.Upgrader
}

func newWsHub() *WSHub {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		Subprotocols:    []string{"access_token"},
		CheckOrigin: func(r *http.Request) bool {
			// Origin is checked by CORS middleware
			return true
		},
	}

	return &WSHub{
		broadcast:  make(chan *BroadcastData, 1000),
		register:   make(chan *WsClient, 100),
		unregister: make(chan *WsClient, 100),
		clients:    make(map[string][]*WsClient),
		upgrader:   upgrader,
	}
}

func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			log.Printf("%+v\n", len(h.clients[client.userID]))
			prevLen := 0
			if _, ok := h.clients[client.userID]; !ok {
				h.clients[client.userID] = []*WsClient{client}
			} else {
				prevLen = len(h.clients[client.userID])
				h.clients[client.userID] = append(h.clients[client.userID], client)
			}

			if prevLen == 0 {
				h.broadcastDataToAll(map[string]string{
					"type": WSTypeUserStatusChange,
				})
			}

			client.send <- []byte("connected:" + client.userID)
		case client := <-h.unregister:
			if _, ok := h.clients[client.userID]; ok {
				prevLen := len(h.clients[client.userID])
				for i := range h.clients[client.userID] {
					if h.clients[client.userID][i] == client {
						h.clients[client.userID] = append(h.clients[client.userID][:i], h.clients[client.userID][i+1:]...)
						break
					}
				}

				if prevLen > 0 && len(h.clients[client.userID]) == 0 {
					h.broadcastDataToAll(map[string]string{
						"type": WSTypeUserStatusChange,
					})
				}

				close(client.send)
			}
		case data := <-h.broadcast:
			if data.BroadcastToAll {
				for userID := range h.clients {
					for i := len(h.clients[userID]) - 1; i >= 0; i-- {
						select {
						case h.clients[userID][i].send <- data.Data:
						default:
							close(h.clients[userID][i].send)
							h.clients[userID] = append(h.clients[userID][:i], h.clients[userID][i+1:]...)
						}
					}
				}
			} else {
				for _, userID := range data.UserIDs {
					if _, ok := h.clients[userID]; ok {
						for i := len(h.clients[userID]) - 1; i >= 0; i-- {
							select {
							case h.clients[userID][i].send <- data.Data:
							default:
								close(h.clients[userID][i].send)
								h.clients[userID] = append(h.clients[userID][:i], h.clients[userID][i+1:]...)
							}
						}
					}
				}
			}
		}
	}
}

func (h *WSHub) listActiveUserIDs() []string {
	result := []string{}
	for userID := range h.clients {
		if len(h.clients[userID]) > 0 {
			result = append(result, userID)
		}
	}

	return result
}

func (h *WSHub) broadcastData(userIDs []string, data interface{}) {
	if data == nil {
		log.Println("Data is nil")
		return
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Failed to marshal the data")
		return
	}

	broadcastData := &BroadcastData{
		UserIDs: userIDs,
		Data:    bytes,
	}

	h.broadcast <- broadcastData
}

func (h *WSHub) broadcastDataToAll(data interface{}) {
	if data == nil {
		log.Println("Data is nil")
		return
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Failed to marshal the data")
		return
	}

	broadcastData := &BroadcastData{
		BroadcastToAll: true,
		Data:    bytes,
	}

	h.broadcast <- broadcastData
}

type WsClient struct {
	userID      string
	accessToken *model.AccessToken

	hub  *WSHub
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *WsClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
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
