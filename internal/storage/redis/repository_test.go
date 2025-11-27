package redis

import (
	"context"
	"testing"
	"time"

	"catalog-api/internal/catalog"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestGetProductCachesAndHitsRedis(t *testing.T) {
	ctx := context.Background()
	mr := startMiniRedis(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	product := catalog.Product{ID: "p1", Name: "Phone"}
	repo := &fakeCatalogRepo{product: product}
	cached := NewCachedRepository(repo, repo, client, time.Minute)

	// First call should hit underlying repo and populate cache.
	got, err := cached.GetProduct(ctx, product.ID)
	if err != nil {
		t.Fatalf("GetProduct returned error: %v", err)
	}
	if got.Name != product.Name {
		t.Fatalf("expected product name %s, got %s", product.Name, got.Name)
	}
	if repo.getProductCalls != 1 {
		t.Fatalf("expected 1 repo call, got %d", repo.getProductCalls)
	}

	// Wait for async cache set.
	waitFor(func() bool { return mr.Exists("product:p1") }, time.Second)

	// Second call should use cache (no extra repo call).
	_, err = cached.GetProduct(ctx, product.ID)
	if err != nil {
		t.Fatalf("GetProduct returned error: %v", err)
	}
	if repo.getProductCalls != 1 {
		t.Fatalf("expected no additional repo call, got %d", repo.getProductCalls)
	}
}

func TestUpdateProductInvalidatesCache(t *testing.T) {
	ctx := context.Background()
	mr := startMiniRedis(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	product := catalog.Product{ID: "p1", Name: "Phone"}
	repo := &fakeCatalogRepo{product: product}
	cached := NewCachedRepository(repo, repo, client, time.Minute)

	// Populate cache
	if _, err := cached.GetProduct(ctx, product.ID); err != nil {
		t.Fatalf("GetProduct returned error: %v", err)
	}
	waitFor(func() bool { return mr.Exists("product:p1") }, time.Second)

	// Update product should invalidate cache key.
	product.Name = "New Phone"
	if _, err := cached.UpdateProduct(ctx, product); err != nil {
		t.Fatalf("UpdateProduct returned error: %v", err)
	}
	if mr.Exists("product:p1") {
		t.Fatalf("expected cache to be invalidated")
	}
}

func TestListCategoriesCachesAndInvalidatesOnWrite(t *testing.T) {
	ctx := context.Background()
	mr := startMiniRedis(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	categories := []catalog.Category{{ID: "c1", Name: "Phones"}}
	repo := &fakeCatalogRepo{categories: categories}
	cached := NewCachedRepository(repo, repo, client, time.Minute)

	// First call hits repo
	_, err := cached.ListCategories(ctx)
	if err != nil {
		t.Fatalf("ListCategories returned error: %v", err)
	}
	if repo.listCategoriesCalls != 1 {
		t.Fatalf("expected repo to be called once, got %d", repo.listCategoriesCalls)
	}
	waitFor(func() bool { return mr.Exists("categories:all") }, time.Second)

	// Second call uses cache
	_, err = cached.ListCategories(ctx)
	if err != nil {
		t.Fatalf("ListCategories returned error: %v", err)
	}
	if repo.listCategoriesCalls != 1 {
		t.Fatalf("expected repo not to be called on cache hit, got %d", repo.listCategoriesCalls)
	}

	// CreateCategory should invalidate cache
	_, err = cached.CreateCategory(ctx, catalog.Category{ID: "c2", Name: "Laptops"})
	if err != nil {
		t.Fatalf("CreateCategory returned error: %v", err)
	}
	if mr.Exists("categories:all") {
		t.Fatalf("expected categories cache to be invalidated after create")
	}

	// Next list should hit repo again
	_, err = cached.ListCategories(ctx)
	if err != nil {
		t.Fatalf("ListCategories returned error: %v", err)
	}
	if repo.listCategoriesCalls != 2 {
		t.Fatalf("expected repo to be called after invalidation, got %d", repo.listCategoriesCalls)
	}
}

// --- test helpers ---

type fakeCatalogRepo struct {
	product              catalog.Product
	products             []catalog.Product
	categories           []catalog.Category
	getProductCalls      int
	listCategoriesCalls  int
	createCategoryCalls  int
	updateProductCalls   int
	deleteProductCalls   int
	assignCategoryCalls  int
	listProductsCalls    int
	countProductsResult  int64
	countProductsCalls   int
	productHistory       []catalog.ProductHistory
	searchCategoriesHits int
}

func (f *fakeCatalogRepo) ListProducts(ctx context.Context, filter catalog.ProductFilter) ([]catalog.Product, error) {
	f.listProductsCalls++
	if f.products != nil {
		return f.products, nil
	}
	return []catalog.Product{}, nil
}

func (f *fakeCatalogRepo) CountProducts(ctx context.Context, filter catalog.ProductFilter) (int64, error) {
	f.countProductsCalls++
	return f.countProductsResult, nil
}

func (f *fakeCatalogRepo) GetProduct(ctx context.Context, id string) (catalog.Product, error) {
	f.getProductCalls++
	return f.product, nil
}

func (f *fakeCatalogRepo) CreateProduct(ctx context.Context, p catalog.Product) (catalog.Product, error) {
	f.product = p
	return p, nil
}

func (f *fakeCatalogRepo) UpdateProduct(ctx context.Context, p catalog.Product) (catalog.Product, error) {
	f.updateProductCalls++
	f.product = p
	return p, nil
}

func (f *fakeCatalogRepo) DeleteProduct(ctx context.Context, id string) error {
	f.deleteProductCalls++
	return nil
}

func (f *fakeCatalogRepo) ListProductHistory(ctx context.Context, id string, filter catalog.ProductHistoryFilter) ([]catalog.ProductHistory, error) {
	return f.productHistory, nil
}

func (f *fakeCatalogRepo) AssignProductCategory(ctx context.Context, productID, categoryID string) error {
	f.assignCategoryCalls++
	return nil
}

func (f *fakeCatalogRepo) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	f.listCategoriesCalls++
	if f.categories != nil {
		return f.categories, nil
	}
	return []catalog.Category{}, nil
}

func (f *fakeCatalogRepo) CreateCategory(ctx context.Context, cat catalog.Category) (catalog.Category, error) {
	f.createCategoryCalls++
	f.categories = append(f.categories, cat)
	return cat, nil
}

func (f *fakeCatalogRepo) UpdateCategory(ctx context.Context, cat catalog.Category) (catalog.Category, error) {
	return cat, nil
}

func (f *fakeCatalogRepo) DeleteCategory(ctx context.Context, id string) error {
	return nil
}

func (f *fakeCatalogRepo) SearchCategories(ctx context.Context, filter catalog.SearchFilter) ([]catalog.Category, int64, error) {
	f.searchCategoriesHits++
	return f.categories, int64(len(f.categories)), nil
}

func startMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	return mr
}

func waitFor(condition func() bool, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
