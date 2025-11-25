package catalog

import (
	"context"
	"time"
)

// CategoryRepository holds persistence contracts for categories.
type CategoryRepository interface {
	ListCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, cat Category) (Category, error)
	UpdateCategory(ctx context.Context, cat Category) (Category, error)
	DeleteCategory(ctx context.Context, id string) error
	SearchCategories(ctx context.Context, filter SearchFilter) ([]Category, int64, error)
}

// ProductRepository holds persistence contracts for products.
type ProductRepository interface {
	ListProducts(ctx context.Context, filter ProductFilter) ([]Product, error)
	CountProducts(ctx context.Context, filter ProductFilter) (int64, error)
	GetProduct(ctx context.Context, id string) (Product, error)
	CreateProduct(ctx context.Context, p Product) (Product, error)
	UpdateProduct(ctx context.Context, p Product) (Product, error)
	DeleteProduct(ctx context.Context, id string) error
	ListProductHistory(ctx context.Context, id string, filter ProductHistoryFilter) ([]ProductHistory, error)
}

// ProductFilter supports pagination and future filtering.
type ProductFilter struct {
	Query   string
	Limit   int
	Offset  int
	SortBy  string
	SortDir string
}

// SearchFilter supports combined search for products or categories.
type SearchFilter struct {
	Kind    string // "product" or "category"
	Query   string
	Limit   int
	Offset  int
	SortBy  string
	SortDir string
}

// ProductHistoryFilter filters history queries.
type ProductHistoryFilter struct {
	Start time.Time
	End   time.Time
}
