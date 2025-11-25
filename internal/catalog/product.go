package catalog

import "time"

// Product represents an item in the catalog. IDs are UUID strings.
type Product struct {
	ID          string
	Name        string
	Description string
	Price       int64 // stored in smallest currency unit
	Stock       int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProductHistory captures historical price/stock snapshots.
type ProductHistory struct {
	ID        string
	ProductID string
	Price     int64
	Stock     int64
	ChangedAt time.Time
}
