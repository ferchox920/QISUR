package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// EventMessage es el sobre JSON enviado por el socket.
type EventMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// Hub registra los clientes conectados y les difunde eventos.
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	upgrader websocket.Upgrader
}

// NewHub construye un hub listo para aceptar clientes.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 64),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Permitimos todos los origenes por ahora; los frontends igual deben usar auth si hace falta.
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// Run procesa el ciclo de vida de clientes y difunde eventos hasta que el contexto se cancele.
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
					// el cliente no esta leyendo; lo descartamos para no bloquear el hub
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

// ServeHTTP actualiza la conexion y registra un cliente WebSocket.
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
		// El hub no esta corriendo; se rechaza la conexion.
		_ = conn.Close()
		return
	}

	go client.writePump()
	go client.readPump()

	// Se envia un saludo ligero para confirmar que el canal esta vivo.
	_ = h.Publish(EventConnected, map[string]string{"id": conn.RemoteAddr().String()})
}

// Publish envia un evento a cada cliente conectado. Es best-effort y descarta
// el payload si falla la serializacion o el hub esta congestionado.
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
		// evitar bloquear el handler de la API; los buffers de clientes estan llenos
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
