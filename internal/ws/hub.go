package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// EventMessage is the JSON envelope sent over the socket.
type EventMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// Hub keeps track of all connected clients and broadcasts events to them.
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	upgrader websocket.Upgrader
}

// NewHub builds a hub ready to accept clients.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 64),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow all origins for now; frontends should still use auth if needed.
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// Run processes client lifecycle and fan-out broadcasts until the context is cancelled.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				addr := client.conn.RemoteAddr().String()
				delete(h.clients, client)
				close(client.send)
				_ = h.Publish(EventDisconnected, map[string]string{"id": addr})
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// client is not reading; drop it to avoid blocking the hub
					delete(h.clients, client)
					close(client.send)
					_ = client.conn.Close()
				}
			}
		case <-ctx.Done():
			h.shutdownClients()
			return
		}
	}
}

// ServeHTTP upgrades the connection and registers a new WebSocket client.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "websocket upgrade failed", http.StatusBadRequest)
		return
	}

	client := newClient(h, conn)
	select {
	case h.register <- client:
	default:
		// Hub is not running; reject connection.
		_ = conn.Close()
		return
	}

	go client.writePump()
	go client.readPump()

	// Send a lightweight hello so consumers can confirm the channel is alive.
	_ = h.Publish(EventConnected, map[string]string{"id": conn.RemoteAddr().String()})
}

// Publish pushes an event to every connected client. It is best-effort and will
// drop the payload if serialization fails or the hub is back-pressured.
func (h *Hub) Publish(event string, data interface{}) error {
	payload, err := json.Marshal(EventMessage{
		Event: event,
		Data:  data,
	})
	if err != nil {
		return err
	}

	select {
	case h.broadcast <- payload:
	default:
		// avoid blocking the API handler; client buffers are congested
	}
	return nil
}

func (h *Hub) shutdownClients() {
	for client := range h.clients {
		close(client.send)
		_ = client.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
		_ = client.conn.Close()
		delete(h.clients, client)
	}
}
