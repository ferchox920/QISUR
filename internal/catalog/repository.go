package catalog

import "context"

// CategoryRepository holds persistence contracts for categories.
type CategoryRepository interface {
	ListCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, cat Category) (Category, error)
	UpdateCategory(ctx context.Context, cat Category) (Category, error)
	DeleteCategory(ctx context.Context, id string) error
}

// ProductRepository holds persistence contracts for products.
type ProductRepository interface {
	ListProducts(ctx context.Context, filter ProductFilter) ([]Product, error)
	CountProducts(ctx context.Context, filter ProductFilter) (int64, error)
	GetProduct(ctx context.Context, id string) (Product, error)
	CreateProduct(ctx context.Context, p Product) (Product, error)
	UpdateProduct(ctx context.Context, p Product) (Product, error)
	DeleteProduct(ctx context.Context, id string) error
}

// ProductFilter supports pagination and future filtering.
type ProductFilter struct {
	Limit  int
	Offset int
}
