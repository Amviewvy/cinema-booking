package websocket

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Hub struct {
	clients   map[*websocket.Conn]bool
	broadcast chan interface{}
	mu        sync.RWMutex
}

var hub = &Hub{
	clients:   make(map[*websocket.Conn]bool),
	broadcast: make(chan interface{}),
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	hub.mu.Lock()
	hub.clients[conn] = true
	hub.mu.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			hub.mu.Lock()
			delete(hub.clients, conn)
			hub.mu.Unlock()
			conn.Close()
			break
		}
	}
}

func StartBroadcast() {
	go func() {
		for {
			msg := <-hub.broadcast

			hub.mu.RLock()
			for client := range hub.clients {
				err := client.WriteJSON(msg)
				if err != nil {
					client.Close()
					hub.mu.RUnlock()
					hub.mu.Lock()
					delete(hub.clients, client)
					hub.mu.Unlock()
					hub.mu.RLock()
				}
			}
			hub.mu.RUnlock()
		}
	}()
}

func SendUpdate(message interface{}) {
	hub.broadcast <- message
}
