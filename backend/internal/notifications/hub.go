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
	// Map UserID to their active WebSocket connection
	clients map[int]*websocket.Conn
	mutex   sync.Mutex
}

var GlobalHub = &Hub{
	clients: make(map[int]*websocket.Conn),
}

func (h *Hub) HandleWS(c *gin.Context) {
	// Retrieve UserID from the Gin context (set by AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	h.mutex.Lock()
	h.clients[userID.(int)] = conn
	h.mutex.Unlock()

	// Keep the connection alive and remove it when the client disconnects
	defer func() {
		h.mutex.Lock()
		delete(h.clients, userID.(int))
		h.mutex.Unlock()
		conn.Close()
	}()

	// Infinite loop to keep the connection open
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// SendToUser targets a specific individual (used for mentions)
func (h *Hub) SendToUser(userID int, message interface{}) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if conn, ok := h.clients[userID]; ok {
		conn.WriteJSON(message)
	}
}

// Broadcast still exists for global updates (like moving a card)
func (h *Hub) Broadcast(message interface{}) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	for _, conn := range h.clients {
		conn.WriteJSON(message)
	}
}
