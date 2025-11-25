package catalog

import "context"

// Service exposes catalog use cases.
type Service interface {
	ListCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, input CreateCategoryInput) (Category, error)
	UpdateCategory(ctx context.Context, input UpdateCategoryInput) (Category, error)
	DeleteCategory(ctx context.Context, id string) error
}

// CreateCategoryInput captures creation fields.
type CreateCategoryInput struct {
	Name        string
	Description string
}

// UpdateCategoryInput captures update fields.
type UpdateCategoryInput struct {
	ID          string
	Name        string
	Description string
}

// ServiceDeps wires dependencies into the catalog service.
type ServiceDeps struct {
	CategoryRepo CategoryRepository
}

type service struct {
	deps ServiceDeps
}

// NewService builds a catalog service.
func NewService(deps ServiceDeps) Service {
	return &service{deps: deps}
}

func (s *service) ListCategories(ctx context.Context) ([]Category, error) {
	if s.deps.CategoryRepo == nil {
		return nil, ErrRepositoryNotConfigured
	}
	return nil, ErrNotImplemented
}

func (s *service) CreateCategory(ctx context.Context, input CreateCategoryInput) (Category, error) {
	if s.deps.CategoryRepo == nil {
		return Category{}, ErrRepositoryNotConfigured
	}
	return Category{}, ErrNotImplemented
}

func (s *service) UpdateCategory(ctx context.Context, input UpdateCategoryInput) (Category, error) {
	if s.deps.CategoryRepo == nil {
		return Category{}, ErrRepositoryNotConfigured
	}
	return Category{}, ErrNotImplemented
}

func (s *service) DeleteCategory(ctx context.Context, id string) error {
	if s.deps.CategoryRepo == nil {
		return ErrRepositoryNotConfigured
	}
	return ErrNotImplemented
}
