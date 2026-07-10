package websocket

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn

	mu             sync.RWMutex
	checkinIsOpen  bool
	sessionID      int
	checkedInUsers map[int]bool
}

func NewHub() *Hub {
	return &Hub{
		broadcast:      make(chan []byte),
		register:       make(chan *websocket.Conn),
		unregister:     make(chan *websocket.Conn),
		clients:        make(map[*websocket.Conn]bool),
		checkinIsOpen:  false,
		sessionID:      0,
		checkedInUsers: make(map[int]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			// Send current status immediately upon connection
			status := h.GetStatusMessage()
			client.WriteMessage(websocket.TextMessage, status)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Register(conn *websocket.Conn) {
	h.register <- conn
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.unregister <- conn
}

func (h *Hub) SetCheckinStatus(isOpen bool) {
	h.mu.Lock()
	if isOpen && !h.checkinIsOpen {
		h.sessionID++
		h.checkedInUsers = make(map[int]bool)
	}
	h.checkinIsOpen = isOpen
	h.mu.Unlock()
	h.broadcast <- h.GetStatusMessage()
}

func (h *Hub) GetStatusMessage() []byte {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.checkinIsOpen {
		return []byte(fmt.Sprintf(`{"status":"open", "session_id": %d}`, h.sessionID))
	}
	return []byte(`{"status":"closed"}`)
}

func (h *Hub) IsOpen() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.checkinIsOpen
}

func (h *Hub) HasCheckedIn(userID int) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.checkedInUsers == nil {
		return false
	}
	return h.checkedInUsers[userID]
}

func (h *Hub) MarkCheckedIn(userID int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.checkedInUsers == nil {
		h.checkedInUsers = make(map[int]bool)
	}
	h.checkedInUsers[userID] = true
}

func (h *Hub) CurrentSessionID() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sessionID
}
