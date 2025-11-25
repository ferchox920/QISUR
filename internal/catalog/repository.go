package catalog

import "context"

// CategoryRepository holds persistence contracts for categories.
type CategoryRepository interface {
	ListCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, cat Category) (Category, error)
	UpdateCategory(ctx context.Context, cat Category) (Category, error)
	DeleteCategory(ctx context.Context, id string) error
}
