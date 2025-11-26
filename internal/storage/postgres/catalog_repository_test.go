package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"catalog-api/internal/catalog"

	pgxmock "github.com/pashagolub/pgxmock/v3"
)

func TestCatalogRepository_ListCategoriesNilPool(t *testing.T) {
	repo := &CatalogRepository{}
	if _, err := repo.ListCategories(context.Background()); !errors.Is(err, catalog.ErrRepositoryNotConfigured) {
		t.Fatalf("expected ErrRepositoryNotConfigured, got %v", err)
	}
}

func TestCatalogRepository_ListCategories(t *testing.T) {
	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock: %v", err)
	}
	defer mock.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT id, name, description, created_at, updated_at FROM categories ORDER BY name`).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow("c1", "Books", "All", now, now))

	repo := &CatalogRepository{pool: mock}
	items, err := repo.ListCategories(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != "c1" {
		t.Fatalf("unexpected categories %+v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCatalogRepository_UpdateProductRecordsHistoryOnChange(t *testing.T) {
	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock: %v", err)
	}
	defer mock.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT price::bigint, stock FROM products WHERE id = \$1 FOR UPDATE`).
		WithArgs("p1").
		WillReturnRows(pgxmock.NewRows([]string{"price", "stock"}).AddRow(int64(10), int64(5)))

	mock.ExpectQuery(`UPDATE products\s+SET name = \$1, description = \$2, price = \$3, stock = \$4, updated_at = NOW\(\)\s+WHERE id = \$5\s+RETURNING id, name, description, price, stock, created_at, updated_at`).
		WithArgs("Pen", "Red", int64(12), int64(3), "p1").
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description", "price", "stock", "created_at", "updated_at"}).
			AddRow("p1", "Pen", "Red", int64(12), int64(3), time.Now(), time.Now()))

	mock.ExpectExec(`INSERT INTO product_history \(product_id, price, stock\)\s+VALUES \(\$1, \$2, \$3\)`).
		WithArgs("p1", int64(12), int64(3)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	repo := &CatalogRepository{pool: mock}
	updated, err := repo.UpdateProduct(ctx, catalog.Product{
		ID:          "p1",
		Name:        "Pen",
		Description: "Red",
		Price:       12,
		Stock:       3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Price != 12 || updated.Stock != 3 {
		t.Fatalf("unexpected product %+v", updated)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCatalogRepository_UpdateProductRollsBackOnHistoryFailure(t *testing.T) {
	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock: %v", err)
	}
	defer mock.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT price::bigint, stock FROM products WHERE id = \$1 FOR UPDATE`).
		WithArgs("p1").
		WillReturnRows(pgxmock.NewRows([]string{"price", "stock"}).AddRow(int64(10), int64(5)))

	mock.ExpectQuery(`UPDATE products\s+SET name = \$1, description = \$2, price = \$3, stock = \$4, updated_at = NOW\(\)\s+WHERE id = \$5\s+RETURNING id, name, description, price, stock, created_at, updated_at`).
		WithArgs("Pen", "Red", int64(12), int64(3), "p1").
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description", "price", "stock", "created_at", "updated_at"}).
			AddRow("p1", "Pen", "Red", int64(12), int64(3), time.Now(), time.Now()))

	mock.ExpectExec(`INSERT INTO product_history \(product_id, price, stock\)\s+VALUES \(\$1, \$2, \$3\)`).
		WithArgs("p1", int64(12), int64(3)).
		WillReturnError(errors.New("history fail"))
	mock.ExpectRollback()

	repo := &CatalogRepository{pool: mock}
	if _, err := repo.UpdateProduct(ctx, catalog.Product{
		ID:          "p1",
		Name:        "Pen",
		Description: "Red",
		Price:       12,
		Stock:       3,
	}); err == nil {
		t.Fatalf("expected error when history insert fails")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCatalogRepository_SearchCategoriesWithQuery(t *testing.T) {
	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery(`SELECT id, name, description, created_at, updated_at FROM categories WHERE \(name ILIKE \$1 OR description ILIKE \$1\) ORDER BY name LIMIT \$2 OFFSET \$3`).
		WithArgs("%bo%", 10, 5).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow("c1", "Books", "All", time.Now(), time.Now()))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM categories WHERE \(name ILIKE \$1 OR description ILIKE \$1\)`).
		WithArgs("%bo%").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

	repo := &CatalogRepository{pool: mock}
	items, total, err := repo.SearchCategories(ctx, catalog.SearchFilter{
		Query:  "bo",
		Limit:  10,
		Offset: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != "c1" {
		t.Fatalf("unexpected search result: total=%d items=%+v", total, items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
