package http

// ListCategoriesDoc godoc
// @Summary List categories
// @Tags Catalog
// @Produce json
// @Success 200 {array} CategoryResponse
// @Router /categories [get]
func ListCategoriesDoc() {}

// CreateCategoryDoc godoc
// @Summary Create category
// @Tags Catalog
// @Accept json
// @Produce json
// @Param body body CreateCategoryRequest true "Category payload"
// @Success 201 {object} CategoryResponse
// @Security BearerAuth
// @Router /categories [post]
func CreateCategoryDoc() {}

// UpdateCategoryDoc godoc
// @Summary Update category
// @Tags Catalog
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param body body UpdateCategoryRequest true "Category update payload"
// @Success 200 {object} CategoryResponse
// @Security BearerAuth
// @Router /categories/{id} [put]
func UpdateCategoryDoc() {}

// DeleteCategoryDoc godoc
// @Summary Delete category
// @Tags Catalog
// @Param id path string true "Category ID"
// @Success 204
// @Security BearerAuth
// @Router /categories/{id} [delete]
func DeleteCategoryDoc() {}

// ListProductsDoc godoc
// @Summary List products
// @Tags Products
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /products [get]
func ListProductsDoc() {}

// GetProductDoc godoc
// @Summary Get product detail
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} ProductResponse
// @Router /products/{id} [get]
func GetProductDoc() {}

// CreateProductDoc godoc
// @Summary Create product
// @Tags Products
// @Accept json
// @Produce json
// @Param body body CreateProductRequest true "Product payload"
// @Success 201 {object} ProductResponse
// @Security BearerAuth
// @Router /products [post]
func CreateProductDoc() {}

// UpdateProductDoc godoc
// @Summary Update product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param body body UpdateProductRequest true "Product update payload"
// @Success 200 {object} ProductResponse
// @Security BearerAuth
// @Router /products/{id} [put]
func UpdateProductDoc() {}

// DeleteProductDoc godoc
// @Summary Delete product
// @Tags Products
// @Param id path string true "Product ID"
// @Success 204
// @Security BearerAuth
// @Router /products/{id} [delete]
func DeleteProductDoc() {}

// ProductHistoryDoc godoc
// @Summary Product history
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Param start query string false "Start date RFC3339"
// @Param end query string false "End date RFC3339"
// @Success 200 {array} ProductHistoryResponse
// @Router /products/{id}/history [get]
func ProductHistoryDoc() {}

// SearchDoc godoc
// @Summary Search products or categories
// @Tags Search
// @Produce json
// @Param type query string true "product or category"
// @Param q query string false "Search query"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param sort query string false "Sort field"
// @Param order query string false "Sort order asc|desc"
// @Success 200 {object} map[string]interface{}
// @Router /search [get]
func SearchDoc() {}
