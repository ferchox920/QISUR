package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
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

	upgrader       websocket.Upgrader
	allowedOrigins map[string]struct{}
	logr           *slog.Logger
}

// NewHub construye un hub listo para aceptar clientes.
func NewHub(allowedOrigins []string, logr *slog.Logger) *Hub {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if host := normalizeOriginHost(o); host != "" {
			originSet[host] = struct{}{}
		}
	}
	if logr == nil {
		logr = slog.Default()
	}
	h := &Hub{
		clients:        make(map[*Client]bool),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan []byte, 64),
		allowedOrigins: originSet,
		logr:           logr,
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	return h
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
		// no bloqueamos peticion, pero avisamos si el buffer esta lleno y se pierde el evento
		if h.logr != nil {
			h.logr.Warn("websocket event dropped: broadcast queue full", "event", event)
		}
		return errors.New("broadcast queue full")
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

func (h *Hub) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// solicitudes sin origen (CLI/servicios) son aceptadas; el control principal es auth.
		return true
	}
	originHost := normalizeOriginHost(origin)
	if originHost == "" {
		return false
	}
	reqHost := strings.ToLower(r.Host)
	if originHost == reqHost {
		return true
	}
	if _, ok := h.allowedOrigins[originHost]; ok {
		return true
	}
	return false
}

func normalizeOriginHost(origin string) string {
	if origin == "" {
		return ""
	}
	u, err := url.Parse(origin)
	if err == nil && u.Host != "" {
		return strings.ToLower(u.Host)
	}
	return strings.ToLower(origin)
}
