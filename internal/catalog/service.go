package catalog

import "context"

// Service expone casos de uso del catalogo.
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
	Search(ctx context.Context, filter SearchFilter) (SearchResult, error)
	GetProductHistory(ctx context.Context, id string, filter ProductHistoryFilter) ([]ProductHistory, error)
	AssignProductCategory(ctx context.Context, productID, categoryID string) error
}

// CreateCategoryInput encapsula campos de creacion.
type CreateCategoryInput struct {
	Name        string
	Description string
}

// CreateProductInput encapsula campos para crear producto.
type CreateProductInput struct {
	Name        string
	Description string
	Price       int64
	Stock       int64
}

// UpdateCategoryInput encapsula campos de actualizacion de categoria.
type UpdateCategoryInput struct {
	ID          string
	Name        string
	Description string
}

// UpdateProductInput encapsula campos de actualizacion de producto.
type UpdateProductInput struct {
	ID          string
	Name        string
	Description string
	Price       int64
	Stock       int64
}

// SearchResult envuelve las respuestas de busqueda.
type SearchResult struct {
	Products   []Product
	Categories []Category
	Total      int64
}

// ServiceDeps cablea las dependencias en el servicio de catalogo.
type ServiceDeps struct {
	CategoryRepo CategoryRepository
	ProductRepo  ProductRepository
}

type service struct {
	deps ServiceDeps
}

// NewService construye un servicio de catalogo.
func NewService(deps ServiceDeps) (Service, error) {
	if deps.CategoryRepo == nil || deps.ProductRepo == nil {
		return nil, ErrRepositoryNotConfigured
	}
	return &service{deps: deps}, nil
}

func (s *service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.deps.CategoryRepo.ListCategories(ctx)
}

func (s *service) CreateCategory(ctx context.Context, input CreateCategoryInput) (Category, error) {
	if input.Name == "" {
		return Category{}, ErrInvalidCategory
	}
	return s.deps.CategoryRepo.CreateCategory(ctx, Category{
		Name:        input.Name,
		Description: input.Description,
	})
}

func (s *service) UpdateCategory(ctx context.Context, input UpdateCategoryInput) (Category, error) {
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
	if id == "" {
		return ErrInvalidCategoryID
	}
	return s.deps.CategoryRepo.DeleteCategory(ctx, id)
}

func (s *service) ListProducts(ctx context.Context, filter ProductFilter) ([]Product, int64, error) {
	// defaults basicos de paginacion
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
	if id == "" {
		return Product{}, ErrInvalidProductID
	}
	return s.deps.ProductRepo.GetProduct(ctx, id)
}

func (s *service) CreateProduct(ctx context.Context, input CreateProductInput) (Product, error) {
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
	if id == "" {
		return ErrInvalidProductID
	}
	return s.deps.ProductRepo.DeleteProduct(ctx, id)
}

func (s *service) GetProductHistory(ctx context.Context, id string, filter ProductHistoryFilter) ([]ProductHistory, error) {
	if id == "" {
		return nil, ErrInvalidProductID
	}
	if !filter.Start.IsZero() && !filter.End.IsZero() && filter.End.Before(filter.Start) {
		return nil, ErrInvalidProduct
	}
	return s.deps.ProductRepo.ListProductHistory(ctx, id, filter)
}

func (s *service) AssignProductCategory(ctx context.Context, productID, categoryID string) error {
	if productID == "" {
		return ErrInvalidProductID
	}
	if categoryID == "" {
		return ErrInvalidCategoryID
	}
	return s.deps.ProductRepo.AssignProductCategory(ctx, productID, categoryID)
}

// Search maneja la busqueda combinada de productos o categorias.
func (s *service) Search(ctx context.Context, filter SearchFilter) (SearchResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	switch filter.Kind {
	case "product":
		pf := ProductFilter{
			Query:   filter.Query,
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			SortBy:  filter.SortBy,
			SortDir: filter.SortDir,
		}
		items, total, err := s.ListProducts(ctx, pf)
		if err != nil {
			return SearchResult{}, err
		}
		return SearchResult{Products: items, Total: total}, nil
	case "category":
		items, total, err := s.deps.CategoryRepo.SearchCategories(ctx, filter)
		if err != nil {
			return SearchResult{}, err
		}
		return SearchResult{Categories: items, Total: total}, nil
	default:
		return SearchResult{}, ErrInvalidSearchKind
	}
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
