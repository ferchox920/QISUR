package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"catalog-api/internal/catalog"

	goredis "github.com/redis/go-redis/v9"
)

// CachedRepository decora los repositorios de Postgres agregando caching.
// Implementa catalog.ProductRepository y catalog.CategoryRepository.
type CachedRepository struct {
	productRepo  catalog.ProductRepository
	categoryRepo catalog.CategoryRepository
	redis        *goredis.Client
	ttl          time.Duration
}

// NewCachedRepository crea una instancia del decorador.
// Recibe los repositorios originales (Postgres) para usarlos como fallback.
func NewCachedRepository(
	pRepo catalog.ProductRepository,
	cRepo catalog.CategoryRepository,
	rdb *goredis.Client,
	ttl time.Duration,
) *CachedRepository {
	return &CachedRepository{
		productRepo:  pRepo,
		categoryRepo: cRepo,
		redis:        rdb,
		ttl:          ttl,
	}
}

// --- IMPLEMENTACIÓN DE PRODUCTOS ---

// GetProduct busca en caché primero, luego en BD.
func (r *CachedRepository) GetProduct(ctx context.Context, id string) (catalog.Product, error) {
	key := fmt.Sprintf("product:%s", id)

	// 1. Intentar leer de Redis
	val, err := r.redis.Get(ctx, key).Result()
	if err == nil {
		var p catalog.Product
		if jsonErr := json.Unmarshal([]byte(val), &p); jsonErr == nil {
			return p, nil // Hit de caché
		}
	} else if err != goredis.Nil {
		// Loguear error de Redis si es necesario, pero no fallar la request
		// fmt.Printf("Redis error: %v\n", err)
	}

	// 2. Fallback a Postgres
	p, err := r.productRepo.GetProduct(ctx, id)
	if err != nil {
		return catalog.Product{}, err
	}

	// 3. Guardar en Redis (Background para no bloquear)
	go func() {
		data, _ := json.Marshal(p)
		r.redis.Set(context.Background(), key, data, r.ttl)
	}()

	return p, nil
}

// UpdateProduct actualiza en BD y luego invalida (borra) el caché.
func (r *CachedRepository) UpdateProduct(ctx context.Context, p catalog.Product) (catalog.Product, error) {
	updated, err := r.productRepo.UpdateProduct(ctx, p)
	if err != nil {
		return catalog.Product{}, err
	}

	// Invalidation: Borramos la entrada antigua para forzar una recarga en la próxima lectura
	key := fmt.Sprintf("product:%s", p.ID)
	r.redis.Del(ctx, key)

	// Opcional: También podrías actualizar el caché directamente con 'updated'
	return updated, nil
}

// DeleteProduct elimina de BD y del caché.
func (r *CachedRepository) DeleteProduct(ctx context.Context, id string) error {
	err := r.productRepo.DeleteProduct(ctx, id)
	if err != nil {
		return err
	}
	r.redis.Del(ctx, fmt.Sprintf("product:%s", id))
	return nil
}

// CreateProduct pasa directo a BD (no hay caché que leer aun).
func (r *CachedRepository) CreateProduct(ctx context.Context, p catalog.Product) (catalog.Product, error) {
	return r.productRepo.CreateProduct(ctx, p)
}

// ListProducts: El caché de listas es complejo por los filtros.
// Estrategia simple: Generar key basada en filtros con TTL corto.
func (r *CachedRepository) ListProducts(ctx context.Context, filter catalog.ProductFilter) ([]catalog.Product, error) {
	// Generar key única para esta combinación de filtros
	key := fmt.Sprintf("products:list:q=%s:l=%d:o=%d:s=%s:%s",
		filter.Query, filter.Limit, filter.Offset, filter.SortBy, filter.SortDir)

	val, err := r.redis.Get(ctx, key).Result()
	if err == nil {
		var products []catalog.Product
		if jsonErr := json.Unmarshal([]byte(val), &products); jsonErr == nil {
			return products, nil
		}
	}

	products, err := r.productRepo.ListProducts(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Guardar lista con un TTL, ej: 1 minuto (puede ser diferente al TTL de items individuales)
	go func() {
		data, _ := json.Marshal(products)
		r.redis.Set(context.Background(), key, data, r.ttl)
	}()

	return products, nil
}

func (r *CachedRepository) CountProducts(ctx context.Context, filter catalog.ProductFilter) (int64, error) {
	// Se puede cachear similar a ListProducts o pasar directo
	return r.productRepo.CountProducts(ctx, filter)
}

func (r *CachedRepository) ListProductHistory(ctx context.Context, id string, filter catalog.ProductHistoryFilter) ([]catalog.ProductHistory, error) {
	// Historial no suele requerir caché de alta velocidad, pasamos directo
	return r.productRepo.ListProductHistory(ctx, id, filter)
}

func (r *CachedRepository) AssignProductCategory(ctx context.Context, productID, categoryID string) error {
	err := r.productRepo.AssignProductCategory(ctx, productID, categoryID)
	if err == nil {
		// Invalidar caché del producto ya que sus relaciones cambiaron (si se guardan en el struct)
		r.redis.Del(ctx, fmt.Sprintf("product:%s", productID))
	}
	return err
}

// --- IMPLEMENTACIÓN DE CATEGORÍAS ---

// ListCategories es ideal para cachear porque cambia poco.
func (r *CachedRepository) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	key := "categories:all"

	val, err := r.redis.Get(ctx, key).Result()
	if err == nil {
		var cats []catalog.Category
		if jsonErr := json.Unmarshal([]byte(val), &cats); jsonErr == nil {
			return cats, nil
		}
	}

	cats, err := r.categoryRepo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	go func() {
		data, _ := json.Marshal(cats)
		r.redis.Set(context.Background(), key, data, r.ttl)
	}()

	return cats, nil
}

func (r *CachedRepository) CreateCategory(ctx context.Context, cat catalog.Category) (catalog.Category, error) {
	newCat, err := r.categoryRepo.CreateCategory(ctx, cat)
	if err == nil {
		r.redis.Del(ctx, "categories:all") // Invalidar lista completa
	}
	return newCat, err
}

func (r *CachedRepository) UpdateCategory(ctx context.Context, cat catalog.Category) (catalog.Category, error) {
	updated, err := r.categoryRepo.UpdateCategory(ctx, cat)
	if err == nil {
		r.redis.Del(ctx, "categories:all")
		// Si hubiera GetCategory(id), borrar su key aquí
	}
	return updated, err
}

func (r *CachedRepository) DeleteCategory(ctx context.Context, id string) error {
	err := r.categoryRepo.DeleteCategory(ctx, id)
	if err == nil {
		r.redis.Del(ctx, "categories:all")
	}
	return err
}

func (r *CachedRepository) SearchCategories(ctx context.Context, filter catalog.SearchFilter) ([]catalog.Category, int64, error) {
	// Búsquedas complejas suelen ir directo a DB o ElasticSearch
	return r.categoryRepo.SearchCategories(ctx, filter)
}
