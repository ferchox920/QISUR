package ws

// Event names for catalog notifications.
const (
	EventConnected               = "socket.connected"
	EventDisconnected            = "socket.disconnected"
	EventCategoryCreated         = "category.created"
	EventCategoryUpdated         = "category.updated"
	EventCategoryDeleted         = "category.deleted"
	EventProductCreated          = "product.created"
	EventProductUpdated          = "product.updated"
	EventProductDeleted          = "product.deleted"
	EventProductCategoryAssigned = "product.category_assigned"
)
