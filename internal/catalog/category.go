package catalog

import "time"

// Category representa un agrupamiento de productos.
type Category struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
