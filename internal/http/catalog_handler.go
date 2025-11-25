package http

import (
	"net/http"

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
