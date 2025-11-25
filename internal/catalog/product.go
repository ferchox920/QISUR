package catalog

import "time"

// TODO: define product entity.
type Product struct {
	ID          string
	Name        string
	Description string
	Price       int64 // stored in smallest currency unit
	Stock       int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
