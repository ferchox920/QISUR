package catalog

import (
	"context"
	"time"
)

// CategoryRepository define contratos de persistencia para categorias.
type CategoryRepository interface {
	ListCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, cat Category) (Category, error)
	UpdateCategory(ctx context.Context, cat Category) (Category, error)
	DeleteCategory(ctx context.Context, id string) error
	SearchCategories(ctx context.Context, filter SearchFilter) ([]Category, int64, error)
}

// ProductRepository define contratos de persistencia para productos.
type ProductRepository interface {
	ListProducts(ctx context.Context, filter ProductFilter) ([]Product, error)
	CountProducts(ctx context.Context, filter ProductFilter) (int64, error)
	GetProduct(ctx context.Context, id string) (Product, error)
	CreateProduct(ctx context.Context, p Product) (Product, error)
	UpdateProduct(ctx context.Context, p Product) (Product, error)
	DeleteProduct(ctx context.Context, id string) error
	ListProductHistory(ctx context.Context, id string, filter ProductHistoryFilter) ([]ProductHistory, error)
	AssignProductCategory(ctx context.Context, productID, categoryID string) error
}

// ProductFilter soporta paginacion y futuros filtros.
type ProductFilter struct {
	Query   string
	Limit   int
	Offset  int
	SortBy  string
	SortDir string
}

// SearchFilter supports combined search for products or categories.
type SearchFilter struct {
	Kind    string // "product" o "category"
	Query   string
	Limit   int
	Offset  int
	SortBy  string
	SortDir string
}

// ProductHistoryFilter filtra consultas de historial.
type ProductHistoryFilter struct {
	Start time.Time
	End   time.Time
}
