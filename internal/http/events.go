package http

import (
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
)

// EventEmitter broadcasts events to connected clients.
type EventEmitter interface {
	Emit(event string, data interface{})
}

type socketEmitter struct {
	server *socketio.Server
}

// NewSocketEmitter builds an event emitter backed by a socket.io server.
func NewSocketEmitter(server *socketio.Server) EventEmitter {
	return &socketEmitter{server: server}
}

func (e *socketEmitter) Emit(event string, data interface{}) {
	if e.server == nil {
		return
	}
	// broadcast to default namespace
	_ = e.server.BroadcastToNamespace("/", event, data)
}

var catalogEvents = []EventInfo{
	{Name: "category.created", Description: "Category created", Payload: `{"id","name","description"}`},
	{Name: "category.updated", Description: "Category updated", Payload: `{"id","name","description"}`},
	{Name: "category.deleted", Description: "Category deleted", Payload: `{"id"}`},
	{Name: "product.created", Description: "Product created", Payload: `{"id","name","description","price","stock"}`},
	{Name: "product.updated", Description: "Product updated", Payload: `{"id","name","description","price","stock"}`},
	{Name: "product.deleted", Description: "Product deleted", Payload: `{"id"}`},
	{Name: "product.category_assigned", Description: "Product assigned to category", Payload: `{"product_id","category_id"}`},
}

// EventsCatalogDoc godoc
// @Summary WebSocket events catalog
// @Tags Events
// @Produce json
// @Success 200 {object} EventsResponse
// @Router /events [get]
func EventsCatalog(c *gin.Context) {
	c.JSON(200, EventsResponse{Events: catalogEvents})
}
