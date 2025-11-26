package catalog

import "time"

// Product representa un item en el catalogo. Los IDs son UUID.
type Product struct {
	ID          string
	Name        string
	Description string
	Price       int64 // almacenado en la unidad monetaria mas pequena
	Stock       int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProductHistory captura los cambios historicos de precio/stock.
type ProductHistory struct {
	ID        string
	ProductID string
	Price     int64
	Stock     int64
	ChangedAt time.Time
}
