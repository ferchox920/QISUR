package catalog

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type stubCategoryRepo struct {
	categories map[string]Category
	nextID     int
	errList    error
	errCreate  error
	errUpdate  error
	errDelete  error
}

func newStubRepo() *stubCategoryRepo {
	return &stubCategoryRepo{categories: make(map[string]Category)}
}

func (s *stubCategoryRepo) ListCategories(ctx context.Context) ([]Category, error) {
	if s.errList != nil {
		return nil, s.errList
	}
	out := make([]Category, 0, len(s.categories))
	for _, v := range s.categories {
		out = append(out, v)
	}
	return out, nil
}

func (s *stubCategoryRepo) SearchCategories(ctx context.Context, filter SearchFilter) ([]Category, int64, error) {
	// busqueda simplificada que ignora query/sort en los tests
	items, err := s.ListCategories(ctx)
	return items, int64(len(items)), err
}

func (s *stubCategoryRepo) CreateCategory(ctx context.Context, cat Category) (Category, error) {
	if s.errCreate != nil {
		return Category{}, s.errCreate
	}
	s.nextID++
	cat.ID = generateID(s.nextID)
	s.categories[cat.ID] = cat
	return cat, nil
}

func (s *stubCategoryRepo) UpdateCategory(ctx context.Context, cat Category) (Category, error) {
	if s.errUpdate != nil {
		return Category{}, s.errUpdate
	}
	if _, ok := s.categories[cat.ID]; !ok {
		return Category{}, errors.New("not found")
	}
	s.categories[cat.ID] = cat
	return cat, nil
}

func (s *stubCategoryRepo) DeleteCategory(ctx context.Context, id string) error {
	if s.errDelete != nil {
		return s.errDelete
	}
	if _, ok := s.categories[id]; !ok {
		return errors.New("not found")
	}
	delete(s.categories, id)
	return nil
}

func generateID(n int) string {
	return fmt.Sprintf("id-%d", n)
}

func TestCreateCategory_ValidatesName(t *testing.T) {
	svc, _ := NewService(ServiceDeps{CategoryRepo: newStubRepo(), ProductRepo: stubProductRepo{}})
	_, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Name: ""})
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("expected ErrInvalidCategory, got %v", err)
	}
}

func TestCreateCategory_Success(t *testing.T) {
	repo := newStubRepo()
	svc, _ := NewService(ServiceDeps{CategoryRepo: repo, ProductRepo: stubProductRepo{}})
	cat, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Name: "Books", Description: "All books"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat.ID == "" || cat.Name != "Books" {
		t.Fatalf("unexpected category %+v", cat)
	}
}

func TestListCategories(t *testing.T) {
	repo := newStubRepo()
	_, _ = repo.CreateCategory(context.Background(), Category{Name: "A"})
	_, _ = repo.CreateCategory(context.Background(), Category{Name: "B"})
	svc, _ := NewService(ServiceDeps{CategoryRepo: repo, ProductRepo: stubProductRepo{}})
	cats, err := svc.ListCategories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
}

func TestUpdateCategory_ValidatesID(t *testing.T) {
	svc, _ := NewService(ServiceDeps{CategoryRepo: newStubRepo(), ProductRepo: stubProductRepo{}})
	_, err := svc.UpdateCategory(context.Background(), UpdateCategoryInput{ID: "", Name: "New"})
	if !errors.Is(err, ErrInvalidCategoryID) {
		t.Fatalf("expected ErrInvalidCategoryID, got %v", err)
	}
}

func TestDeleteCategory_ValidatesID(t *testing.T) {
	svc, _ := NewService(ServiceDeps{CategoryRepo: newStubRepo(), ProductRepo: stubProductRepo{}})
	err := svc.DeleteCategory(context.Background(), "")
	if !errors.Is(err, ErrInvalidCategoryID) {
		t.Fatalf("expected ErrInvalidCategoryID, got %v", err)
	}
}

func TestUpdateCategory_Success(t *testing.T) {
	repo := newStubRepo()
	created, _ := repo.CreateCategory(context.Background(), Category{Name: "Old"})
	svc, _ := NewService(ServiceDeps{CategoryRepo: repo, ProductRepo: stubProductRepo{}})
	updated, err := svc.UpdateCategory(context.Background(), UpdateCategoryInput{ID: created.ID, Name: "New"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "New" {
		t.Fatalf("expected name to change, got %+v", updated)
	}
}

func TestDeleteCategory_Success(t *testing.T) {
	repo := newStubRepo()
	created, _ := repo.CreateCategory(context.Background(), Category{Name: "ToDelete"})
	svc, _ := NewService(ServiceDeps{CategoryRepo: repo, ProductRepo: stubProductRepo{}})
	if err := svc.DeleteCategory(context.Background(), created.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cats, _ := repo.ListCategories(context.Background())
	if len(cats) != 0 {
		t.Fatalf("expected empty repo after delete")
	}
}

type stubProductRepo struct{}

func (stubProductRepo) ListProducts(ctx context.Context, filter ProductFilter) ([]Product, error) {
	return nil, nil
}

func (stubProductRepo) CountProducts(ctx context.Context, filter ProductFilter) (int64, error) {
	return 0, nil
}

func (stubProductRepo) GetProduct(ctx context.Context, id string) (Product, error) {
	return Product{}, nil
}

func (stubProductRepo) CreateProduct(ctx context.Context, p Product) (Product, error) {
	return Product{}, nil
}

func (stubProductRepo) UpdateProduct(ctx context.Context, p Product) (Product, error) {
	return Product{}, nil
}

func (stubProductRepo) DeleteProduct(ctx context.Context, id string) error {
	return nil
}

func (stubProductRepo) ListProductHistory(ctx context.Context, id string, filter ProductHistoryFilter) ([]ProductHistory, error) {
	return nil, nil
}

func (stubProductRepo) AssignProductCategory(ctx context.Context, productID, categoryID string) error {
	return nil
}

func TestSearch_InvalidKind(t *testing.T) {
	svc, _ := NewService(ServiceDeps{CategoryRepo: newStubRepo(), ProductRepo: stubProductRepo{}})
	if _, err := svc.Search(context.Background(), SearchFilter{Kind: "unknown"}); !errors.Is(err, ErrInvalidSearchKind) {
		t.Fatalf("expected ErrInvalidSearchKind, got %v", err)
	}
}
