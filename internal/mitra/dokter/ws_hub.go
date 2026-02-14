package dokter

import (
	"sync"

	"github.com/gofiber/websocket/v2"
)

type OrderHub struct {
	mu    sync.RWMutex
	rooms map[int]map[*websocket.Conn]bool
}

func NewOrderHub() *OrderHub {
	return &OrderHub{
		rooms: make(map[int]map[*websocket.Conn]bool),
	}
}

func (h *OrderHub) Join(orderID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[orderID] == nil {
		h.rooms[orderID] = make(map[*websocket.Conn]bool)
	}
	h.rooms[orderID][conn] = true
}

func (h *OrderHub) Leave(orderID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[orderID] != nil {
		delete(h.rooms[orderID], conn)
		if len(h.rooms[orderID]) == 0 {
			delete(h.rooms, orderID)
		}
	}
}

func (h *OrderHub) Broadcast(orderID int, message interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.rooms[orderID] {
		_ = conn.WriteJSON(message)
	}
}
