package notifications

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	clients map[*websocket.Conn]bool
	mutex   sync.Mutex
}

var GlobalHub = &Hub{
	clients: make(map[*websocket.Conn]bool),
}

func (h *Hub) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.mutex.Lock()
	h.clients[conn] = true
	h.mutex.Unlock()
}

func (h *Hub) Broadcast(message interface{}) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	for client := range h.clients {
		client.WriteJSON(message)
	}
}
