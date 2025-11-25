package http

import (
	"net/http"
	"strconv"
	"time"

	"catalog-api/internal/catalog"

	"github.com/gin-gonic/gin"
)

// CatalogHandler orchestrates catalog endpoints.
type CatalogHandler struct {
	svc catalog.Service
}

func NewCatalogHandler(svc catalog.Service) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

// ListCategories godoc
// @Summary List categories
// @Tags Catalog
// @Produce json
// @Success 200 {array} CategoryResponse
// @Router /categories [get]
func (h *CatalogHandler) ListCategories(c *gin.Context) {
	cats, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toCategoryResponses(cats))
}

// CreateCategory godoc
// @Summary Create category
// @Tags Catalog
// @Accept json
// @Produce json
// @Param body body CreateCategoryRequest true "Category payload"
// @Success 201 {object} CategoryResponse
// @Security BearerAuth
// @Router /categories [post]
func (h *CatalogHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat, err := h.svc.CreateCategory(c.Request.Context(), catalog.CreateCategoryInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toCategoryResponse(cat))
}

// UpdateCategory godoc
// @Summary Update category
// @Tags Catalog
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param body body UpdateCategoryRequest true "Category update payload"
// @Success 200 {object} CategoryResponse
// @Security BearerAuth
// @Router /categories/{id} [put]
func (h *CatalogHandler) UpdateCategory(c *gin.Context) {
	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := c.Param("id")
	cat, err := h.svc.UpdateCategory(c.Request.Context(), catalog.UpdateCategoryInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toCategoryResponse(cat))
}

// DeleteCategory godoc
// @Summary Delete category
// @Tags Catalog
// @Param id path string true "Category ID"
// @Success 204
// @Security BearerAuth
// @Router /categories/{id} [delete]
func (h *CatalogHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func toCategoryResponses(cats []catalog.Category) []CategoryResponse {
	out := make([]CategoryResponse, 0, len(cats))
	for _, c := range cats {
		out = append(out, toCategoryResponse(c))
	}
	return out
}

func toCategoryResponse(c catalog.Category) CategoryResponse {
	return CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
	}
}

// ListProducts godoc
// @Summary List products
// @Tags Products
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /products [get]
func (h *CatalogHandler) ListProducts(c *gin.Context) {
	limit := parseQueryInt(c, "limit", 20)
	offset := parseQueryInt(c, "offset", 0)

	products, total, err := h.svc.ListProducts(c.Request.Context(), catalog.ProductFilter{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"products": toProductResponses(products),
	})
}

// GetProduct godoc
// @Summary Get product detail
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} ProductResponse
// @Router /products/{id} [get]
func (h *CatalogHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	product, err := h.svc.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toProductResponse(product))
}

// CreateProduct godoc
// @Summary Create product
// @Tags Products
// @Accept json
// @Produce json
// @Param body body CreateProductRequest true "Product payload"
// @Success 201 {object} ProductResponse
// @Security BearerAuth
// @Router /products [post]
func (h *CatalogHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	product, err := h.svc.CreateProduct(c.Request.Context(), catalog.CreateProductInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toProductResponse(product))
}

// UpdateProduct godoc
// @Summary Update product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param body body UpdateProductRequest true "Product update payload"
// @Success 200 {object} ProductResponse
// @Security BearerAuth
// @Router /products/{id} [put]
func (h *CatalogHandler) UpdateProduct(c *gin.Context) {
	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := c.Param("id")
	product, err := h.svc.UpdateProduct(c.Request.Context(), catalog.UpdateProductInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toProductResponse(product))
}

// DeleteProduct godoc
// @Summary Delete product
// @Tags Products
// @Param id path string true "Product ID"
// @Success 204
// @Security BearerAuth
// @Router /products/{id} [delete]
func (h *CatalogHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteProduct(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Search allows querying products or categories with pagination and sorting.
func (h *CatalogHandler) Search(c *gin.Context) {
	kind := c.Query("type")
	query := c.Query("q")
	limit := parseQueryInt(c, "limit", 20)
	offset := parseQueryInt(c, "offset", 0)
	sortBy := c.Query("sort")
	sortDir := c.Query("order")

	result, err := h.svc.Search(c.Request.Context(), catalog.SearchFilter{
		Kind:    kind,
		Query:   query,
		Limit:   limit,
		Offset:  offset,
		SortBy:  sortBy,
		SortDir: sortDir,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if kind == "category" {
		c.JSON(http.StatusOK, gin.H{
			"total":      result.Total,
			"categories": toCategoryResponses(result.Categories),
		})
		return
	}
	// default to products
	c.JSON(http.StatusOK, gin.H{
		"total":    result.Total,
		"products": toProductResponses(result.Products),
	})
}

// GetProductHistory godoc
// @Summary Product history
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Param start query string false "Start date RFC3339"
// @Param end query string false "End date RFC3339"
// @Success 200 {array} ProductHistoryResponse
// @Router /products/{id}/history [get]
func (h *CatalogHandler) GetProductHistory(c *gin.Context) {
	id := c.Param("id")
	startStr := c.Query("start")
	endStr := c.Query("end")

	var start, end time.Time
	var err error
	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date"})
			return
		}
	}
	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date"})
			return
		}
	}

	items, err := h.svc.GetProductHistory(c.Request.Context(), id, catalog.ProductHistoryFilter{
		Start: start,
		End:   end,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toProductHistoryResponses(items))
}

func toProductResponses(products []catalog.Product) []ProductResponse {
	out := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		out = append(out, toProductResponse(p))
	}
	return out
}

func toProductResponse(p catalog.Product) ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
	}
}

func toProductHistoryResponses(items []catalog.ProductHistory) []ProductHistoryResponse {
	out := make([]ProductHistoryResponse, 0, len(items))
	for _, h := range items {
		out = append(out, ProductHistoryResponse{
			ID:        h.ID,
			ProductID: h.ProductID,
			Price:     h.Price,
			Stock:     h.Stock,
			ChangedAt: h.ChangedAt.Format(time.RFC3339),
		})
	}
	return out
}

func parseQueryInt(c *gin.Context, key string, fallback int) int {
	if v, ok := c.GetQuery(key); ok {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}
