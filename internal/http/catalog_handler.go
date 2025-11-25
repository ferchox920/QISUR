package http

import (
	"net/http"
	"strconv"

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

func (h *CatalogHandler) ListCategories(c *gin.Context) {
	cats, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toCategoryResponses(cats))
}

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

func (h *CatalogHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	product, err := h.svc.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toProductResponse(product))
}

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

func (h *CatalogHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteProduct(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
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

func parseQueryInt(c *gin.Context, key string, fallback int) int {
	if v, ok := c.GetQuery(key); ok {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}
