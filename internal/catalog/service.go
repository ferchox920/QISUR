package catalog

import "context"

// Service exposes catalog use cases.
type Service interface {
	ListCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, input CreateCategoryInput) (Category, error)
	UpdateCategory(ctx context.Context, input UpdateCategoryInput) (Category, error)
	DeleteCategory(ctx context.Context, id string) error
	ListProducts(ctx context.Context, filter ProductFilter) ([]Product, int64, error)
	GetProduct(ctx context.Context, id string) (Product, error)
	CreateProduct(ctx context.Context, input CreateProductInput) (Product, error)
	UpdateProduct(ctx context.Context, input UpdateProductInput) (Product, error)
	DeleteProduct(ctx context.Context, id string) error
}

// CreateCategoryInput captures creation fields.
type CreateCategoryInput struct {
	Name        string
	Description string
}

// CreateProductInput captures product creation fields.
type CreateProductInput struct {
	Name        string
	Description string
	Price       int64
	Stock       int64
}

// UpdateCategoryInput captures update fields.
type UpdateCategoryInput struct {
	ID          string
	Name        string
	Description string
}

// UpdateProductInput captures product update fields.
type UpdateProductInput struct {
	ID          string
	Name        string
	Description string
	Price       int64
	Stock       int64
}

// ServiceDeps wires dependencies into the catalog service.
type ServiceDeps struct {
	CategoryRepo CategoryRepository
	ProductRepo  ProductRepository
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
	return s.deps.CategoryRepo.ListCategories(ctx)
}

func (s *service) CreateCategory(ctx context.Context, input CreateCategoryInput) (Category, error) {
	if s.deps.CategoryRepo == nil {
		return Category{}, ErrRepositoryNotConfigured
	}
	if input.Name == "" {
		return Category{}, ErrInvalidCategory
	}
	return s.deps.CategoryRepo.CreateCategory(ctx, Category{
		Name:        input.Name,
		Description: input.Description,
	})
}

func (s *service) UpdateCategory(ctx context.Context, input UpdateCategoryInput) (Category, error) {
	if s.deps.CategoryRepo == nil {
		return Category{}, ErrRepositoryNotConfigured
	}
	if input.ID == "" {
		return Category{}, ErrInvalidCategoryID
	}
	if input.Name == "" {
		return Category{}, ErrInvalidCategory
	}
	return s.deps.CategoryRepo.UpdateCategory(ctx, Category{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
	})
}

func (s *service) DeleteCategory(ctx context.Context, id string) error {
	if s.deps.CategoryRepo == nil {
		return ErrRepositoryNotConfigured
	}
	if id == "" {
		return ErrInvalidCategoryID
	}
	return s.deps.CategoryRepo.DeleteCategory(ctx, id)
}

func (s *service) ListProducts(ctx context.Context, filter ProductFilter) ([]Product, int64, error) {
	if s.deps.ProductRepo == nil {
		return nil, 0, ErrRepositoryNotConfigured
	}
	// basic pagination defaults
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	items, err := s.deps.ProductRepo.ListProducts(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.deps.ProductRepo.CountProducts(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *service) GetProduct(ctx context.Context, id string) (Product, error) {
	if s.deps.ProductRepo == nil {
		return Product{}, ErrRepositoryNotConfigured
	}
	if id == "" {
		return Product{}, ErrInvalidProductID
	}
	return s.deps.ProductRepo.GetProduct(ctx, id)
}

func (s *service) CreateProduct(ctx context.Context, input CreateProductInput) (Product, error) {
	if s.deps.ProductRepo == nil {
		return Product{}, ErrRepositoryNotConfigured
	}
	if err := validateProductInput(input.Name, input.Price, input.Stock); err != nil {
		return Product{}, err
	}
	return s.deps.ProductRepo.CreateProduct(ctx, Product{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
	})
}

func (s *service) UpdateProduct(ctx context.Context, input UpdateProductInput) (Product, error) {
	if s.deps.ProductRepo == nil {
		return Product{}, ErrRepositoryNotConfigured
	}
	if input.ID == "" {
		return Product{}, ErrInvalidProductID
	}
	if err := validateProductInput(input.Name, input.Price, input.Stock); err != nil {
		return Product{}, err
	}
	return s.deps.ProductRepo.UpdateProduct(ctx, Product{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
	})
}

func (s *service) DeleteProduct(ctx context.Context, id string) error {
	if s.deps.ProductRepo == nil {
		return ErrRepositoryNotConfigured
	}
	if id == "" {
		return ErrInvalidProductID
	}
	return s.deps.ProductRepo.DeleteProduct(ctx, id)
}

func validateProductInput(name string, price, stock int64) error {
	if name == "" {
		return ErrInvalidProduct
	}
	if price < 0 {
		return ErrInvalidProduct
	}
	if stock < 0 {
		return ErrInvalidProduct
	}
	return nil
}
