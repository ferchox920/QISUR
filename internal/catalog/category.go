package catalog

import "time"

// Category represents a product grouping.
type Category struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
