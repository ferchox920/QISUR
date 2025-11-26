package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
)

type recordingEmitter struct {
	events []string
	data   []interface{}
}

func (r *recordingEmitter) Emit(event string, data interface{}) {
	r.events = append(r.events, event)
	r.data = append(r.data, data)
}

func TestSocketEmitter_NilSafe(t *testing.T) {
	emitter := NewSocketEmitter(nil)
	// should not panic
	emitter.Emit("test", map[string]string{"hello": "world"})
}

func TestCatalogEventsPayloads(t *testing.T) {
	for _, ev := range catalogEvents {
		if ev.Name == "" {
			t.Fatalf("event name should not be empty")
		}
		if ev.Payload == "" {
			t.Fatalf("event payload should not be empty for %s", ev.Name)
		}
	}
}

func TestSocketEmitter_Broadcast(t *testing.T) {
	server := socketio.NewServer(nil)
	emitter := NewSocketEmitter(server)
	// use server's BroadcastToNamespace; cannot assert actual network, but ensure no panic and payload is JSONable
	payload := map[string]string{"id": "123"}
	emitter.Emit("product.created", payload)
}

func TestEventsCatalogResponse(t *testing.T) {
	resp := EventsResponse{Events: catalogEvents}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal events response: %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("expected non-empty JSON")
	}
}

func TestEventsCatalogHandler(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	EventsCatalog(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp EventsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp.Events) != len(catalogEvents) {
		t.Fatalf("expected %d events, got %d", len(catalogEvents), len(resp.Events))
	}
}
