package http

import (
	"catalog-api/internal/ws"

	"github.com/gin-gonic/gin"
)

// EventEmitter difunde eventos a clientes conectados.
type EventEmitter interface {
	Emit(event string, data interface{})
}

type socketEmitter struct {
	hub *ws.Hub
}

// NewSocketEmitter construye un emisor de eventos apoyado en el hub WebSocket.
func NewSocketEmitter(hub *ws.Hub) EventEmitter {
	return &socketEmitter{hub: hub}
}

func (e *socketEmitter) Emit(event string, data interface{}) {
	if e == nil || e.hub == nil {
		return
	}
	_ = e.hub.Publish(event, data)
}

var catalogEvents = []EventInfo{
	{Name: ws.EventCategoryCreated, Description: "Category created", Payload: `{"id","name","description"}`},
	{Name: ws.EventCategoryUpdated, Description: "Category updated", Payload: `{"id","name","description"}`},
	{Name: ws.EventCategoryDeleted, Description: "Category deleted", Payload: `{"id"}`},
	{Name: ws.EventProductCreated, Description: "Product created", Payload: `{"id","name","description","price","stock"}`},
	{Name: ws.EventProductUpdated, Description: "Product updated", Payload: `{"id","name","description","price","stock"}`},
	{Name: ws.EventProductDeleted, Description: "Product deleted", Payload: `{"id"}`},
	{Name: ws.EventProductCategoryAssigned, Description: "Product assigned to category", Payload: `{"product_id","category_id"}`},
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
