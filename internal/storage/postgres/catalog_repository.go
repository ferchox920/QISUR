package postgres

import (
	"context"
	"fmt"
	"strings"

	"catalog-api/internal/catalog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgxPool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// CatalogRepository implementa repositorios de catalogo sobre Postgres.
type CatalogRepository struct {
	pool pgxPool
}

// NewCatalogRepository construye un repo de catalogo respaldado por un pool pgx.
func NewCatalogRepository(pool pgxPool) *CatalogRepository {
	return &CatalogRepository{pool: pool}
}

// ListCategories devuelve todas las categorias ordenadas por nombre.
func (r *CatalogRepository) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	if r.pool == nil {
		return nil, catalog.ErrRepositoryNotConfigured
	}
	rows, err := r.pool.Query(ctx, `SELECT id, name, description, created_at, updated_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []catalog.Category
	for rows.Next() {
		var c catalog.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// CreateCategory inserta una nueva categoria.
func (r *CatalogRepository) CreateCategory(ctx context.Context, cat catalog.Category) (catalog.Category, error) {
	if r.pool == nil {
		return catalog.Category{}, catalog.ErrRepositoryNotConfigured
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO categories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`, cat.Name, cat.Description)
	var out catalog.Category
	if err := row.Scan(&out.ID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return catalog.Category{}, err
	}
	return out, nil
}

// UpdateCategory actualiza nombre/descripcion.
func (r *CatalogRepository) UpdateCategory(ctx context.Context, cat catalog.Category) (catalog.Category, error) {
	if r.pool == nil {
		return catalog.Category{}, catalog.ErrRepositoryNotConfigured
	}
	row := r.pool.QueryRow(ctx, `
		UPDATE categories
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, description, created_at, updated_at
	`, cat.Name, cat.Description, cat.ID)
	var out catalog.Category
	if err := row.Scan(&out.ID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return catalog.Category{}, err
	}
	return out, nil
}

// DeleteCategory elimina una categoria por ID.
func (r *CatalogRepository) DeleteCategory(ctx context.Context, id string) error {
	if r.pool == nil {
		return catalog.ErrRepositoryNotConfigured
	}
	_, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	return err
}

// SearchCategories ejecuta una busqueda de texto simple con paginacion.
func (r *CatalogRepository) SearchCategories(ctx context.Context, filter catalog.SearchFilter) ([]catalog.Category, int64, error) {
	if r.pool == nil {
		return nil, 0, catalog.ErrRepositoryNotConfigured
	}
	query := strings.TrimSpace(filter.Query)
	args := []any{}
	where := "1=1"
	if query != "" {
		where = "(name ILIKE $1 OR description ILIKE $1)"
		args = append(args, "%"+query+"%")
	}
	order := "ORDER BY name"
	limit := "LIMIT $%d"
	offset := "OFFSET $%d"
	args = append(args, filter.Limit, filter.Offset)
	limit = fmt.Sprintf(limit, len(args)-1)
	offset = fmt.Sprintf(offset, len(args))

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`SELECT id, name, description, created_at, updated_at FROM categories WHERE %s %s %s`, where, order, limit+" "+offset), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []catalog.Category
	for rows.Next() {
		var c catalog.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64
	err = r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM categories WHERE %s`, where), args[:len(args)-2]...).Scan(&total)
	return items, total, err
}

// ListProducts obtiene productos con query de texto opcional y ordenamiento.
func (r *CatalogRepository) ListProducts(ctx context.Context, filter catalog.ProductFilter) ([]catalog.Product, error) {
	if r.pool == nil {
		return nil, catalog.ErrRepositoryNotConfigured
	}
	where := "1=1"
	args := []any{}
	if strings.TrimSpace(filter.Query) != "" {
		where = "(name ILIKE $1 OR description ILIKE $1)"
		args = append(args, "%"+strings.TrimSpace(filter.Query)+"%")
	}
	order := buildProductOrderClause(filter.SortBy, filter.SortDir)
	args = append(args, filter.Limit, filter.Offset)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, name, description, price::bigint, stock, created_at, updated_at
		FROM products
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, where, order, len(args)-1, len(args)), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []catalog.Product
	for rows.Next() {
		var p catalog.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

// CountProducts devuelve el total de productos que cumplen el filtro.
func (r *CatalogRepository) CountProducts(ctx context.Context, filter catalog.ProductFilter) (int64, error) {
	if r.pool == nil {
		return 0, catalog.ErrRepositoryNotConfigured
	}
	where := "1=1"
	args := []any{}
	if strings.TrimSpace(filter.Query) != "" {
		where = "(name ILIKE $1 OR description ILIKE $1)"
		args = append(args, "%"+strings.TrimSpace(filter.Query)+"%")
	}
	var total int64
	err := r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM products WHERE %s`, where), args...).Scan(&total)
	return total, err
}

// GetProduct obtiene un producto por ID.
func (r *CatalogRepository) GetProduct(ctx context.Context, id string) (catalog.Product, error) {
	if r.pool == nil {
		return catalog.Product{}, catalog.ErrRepositoryNotConfigured
	}
	var p catalog.Product
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, price::bigint, stock, created_at, updated_at
		FROM products
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

// CreateProduct inserta un nuevo producto.
func (r *CatalogRepository) CreateProduct(ctx context.Context, p catalog.Product) (catalog.Product, error) {
	if r.pool == nil {
		return catalog.Product{}, catalog.ErrRepositoryNotConfigured
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO products (name, description, price, stock)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, price::bigint, stock, created_at, updated_at
	`, p.Name, p.Description, p.Price, p.Stock)
	var out catalog.Product
	if err := row.Scan(&out.ID, &out.Name, &out.Description, &out.Price, &out.Stock, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return catalog.Product{}, err
	}
	return out, nil
}

// UpdateProduct actualiza campos de un producto.
func (r *CatalogRepository) UpdateProduct(ctx context.Context, p catalog.Product) (catalog.Product, error) {
	if r.pool == nil {
		return catalog.Product{}, catalog.ErrRepositoryNotConfigured
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return catalog.Product{}, err
	}
	defer tx.Rollback(ctx)

	var original struct {
		Price int64
		Stock int64
	}
	if err := tx.QueryRow(ctx, `SELECT price::bigint, stock FROM products WHERE id = $1 FOR UPDATE`, p.ID).Scan(&original.Price, &original.Stock); err != nil {
		return catalog.Product{}, err
	}
	row := tx.QueryRow(ctx, `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING id, name, description, price::bigint, stock, created_at, updated_at
	`, p.Name, p.Description, p.Price, p.Stock, p.ID)
	var out catalog.Product
	if err := row.Scan(&out.ID, &out.Name, &out.Description, &out.Price, &out.Stock, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return catalog.Product{}, err
	}
	// Guarda historial solo cuando cambia precio o stock.
	if original.Price != out.Price || original.Stock != out.Stock {
		if _, err := tx.Exec(ctx, `
			INSERT INTO product_history (product_id, price, stock)
			VALUES ($1, $2, $3)
		`, out.ID, out.Price, out.Stock); err != nil {
			return catalog.Product{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return catalog.Product{}, err
	}
	return out, nil
}

// DeleteProduct elimina un producto por ID.
func (r *CatalogRepository) DeleteProduct(ctx context.Context, id string) error {
	if r.pool == nil {
		return catalog.ErrRepositoryNotConfigured
	}
	_, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}

// ListProductHistory devuelve historial de precio/stock de un producto.
func (r *CatalogRepository) ListProductHistory(ctx context.Context, id string, filter catalog.ProductHistoryFilter) ([]catalog.ProductHistory, error) {
	if r.pool == nil {
		return nil, catalog.ErrRepositoryNotConfigured
	}
	clauses := []string{"product_id = $1"}
	args := []any{id}
	idx := 2
	if !filter.Start.IsZero() {
		clauses = append(clauses, fmt.Sprintf("changed_at >= $%d", idx))
		args = append(args, filter.Start)
		idx++
	}
	if !filter.End.IsZero() {
		clauses = append(clauses, fmt.Sprintf("changed_at <= $%d", idx))
		args = append(args, filter.End)
		idx++
	}
	where := strings.Join(clauses, " AND ")
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, product_id, price::bigint, stock, changed_at
		FROM product_history
		WHERE %s
		ORDER BY changed_at DESC
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []catalog.ProductHistory
	for rows.Next() {
		var h catalog.ProductHistory
		if err := rows.Scan(&h.ID, &h.ProductID, &h.Price, &h.Stock, &h.ChangedAt); err != nil {
			return nil, err
		}
		items = append(items, h)
	}
	return items, rows.Err()
}

// AssignProductCategory relaciona un producto con una categoria (muchos a muchos).
func (r *CatalogRepository) AssignProductCategory(ctx context.Context, productID, categoryID string) error {
	if r.pool == nil {
		return catalog.ErrRepositoryNotConfigured
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO product_category (product_id, category_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, productID, categoryID)
	return err
}

func buildProductOrderClause(sortBy, sortDir string) string {
	field := "created_at"
	switch sortBy {
	case "name":
		field = "name"
	case "price":
		field = "price"
	case "stock":
		field = "stock"
	case "created_at":
		field = "created_at"
	}
	dir := strings.ToUpper(sortDir)
	if dir != "ASC" {
		dir = "DESC"
	}
	return fmt.Sprintf("ORDER BY %s %s", field, dir)
}
