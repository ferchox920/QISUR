package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"catalog-api/internal/catalog"
	"catalog-api/internal/ws"

	"github.com/gin-gonic/gin"
)

type stubCatalogService struct {
	listCategoriesResp []catalog.Category
	listCategoriesErr  error

	createCategoryInput catalog.CreateCategoryInput
	createCategoryResp  catalog.Category
	createCategoryErr   error

	updateCategoryInput catalog.UpdateCategoryInput
	updateCategoryResp  catalog.Category
	updateCategoryErr   error

	deleteCategoryID  string
	deleteCategoryErr error

	listProductsFilter catalog.ProductFilter
	listProductsResp   []catalog.Product
	listProductsTotal  int64
	listProductsErr    error

	getProductID   string
	getProductResp catalog.Product
	getProductErr  error

	createProductInput catalog.CreateProductInput
	createProductResp  catalog.Product
	createProductErr   error

	updateProductInput catalog.UpdateProductInput
	updateProductResp  catalog.Product
	updateProductErr   error

	deleteProductID  string
	deleteProductErr error

	assignProductCategoryProductID  string
	assignProductCategoryCategoryID string
	assignProductCategoryErr        error

	searchFilter catalog.SearchFilter
	searchResp   catalog.SearchResult
	searchErr    error

	historyProductID string
	historyFilter    catalog.ProductHistoryFilter
	historyResp      []catalog.ProductHistory
	historyErr       error
}

func (s *stubCatalogService) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	return s.listCategoriesResp, s.listCategoriesErr
}

func (s *stubCatalogService) CreateCategory(ctx context.Context, input catalog.CreateCategoryInput) (catalog.Category, error) {
	s.createCategoryInput = input
	return s.createCategoryResp, s.createCategoryErr
}

func (s *stubCatalogService) UpdateCategory(ctx context.Context, input catalog.UpdateCategoryInput) (catalog.Category, error) {
	s.updateCategoryInput = input
	return s.updateCategoryResp, s.updateCategoryErr
}

func (s *stubCatalogService) DeleteCategory(ctx context.Context, id string) error {
	s.deleteCategoryID = id
	return s.deleteCategoryErr
}

func (s *stubCatalogService) ListProducts(ctx context.Context, filter catalog.ProductFilter) ([]catalog.Product, int64, error) {
	s.listProductsFilter = filter
	return s.listProductsResp, s.listProductsTotal, s.listProductsErr
}

func (s *stubCatalogService) GetProduct(ctx context.Context, id string) (catalog.Product, error) {
	s.getProductID = id
	return s.getProductResp, s.getProductErr
}

func (s *stubCatalogService) CreateProduct(ctx context.Context, input catalog.CreateProductInput) (catalog.Product, error) {
	s.createProductInput = input
	return s.createProductResp, s.createProductErr
}

func (s *stubCatalogService) UpdateProduct(ctx context.Context, input catalog.UpdateProductInput) (catalog.Product, error) {
	s.updateProductInput = input
	return s.updateProductResp, s.updateProductErr
}

func (s *stubCatalogService) DeleteProduct(ctx context.Context, id string) error {
	s.deleteProductID = id
	return s.deleteProductErr
}

func (s *stubCatalogService) Search(ctx context.Context, filter catalog.SearchFilter) (catalog.SearchResult, error) {
	s.searchFilter = filter
	return s.searchResp, s.searchErr
}

func (s *stubCatalogService) GetProductHistory(ctx context.Context, id string, filter catalog.ProductHistoryFilter) ([]catalog.ProductHistory, error) {
	s.historyProductID = id
	s.historyFilter = filter
	return s.historyResp, s.historyErr
}

func (s *stubCatalogService) AssignProductCategory(ctx context.Context, productID, categoryID string) error {
	s.assignProductCategoryProductID = productID
	s.assignProductCategoryCategoryID = categoryID
	return s.assignProductCategoryErr
}

type testRecordingEmitter struct {
	events []string
	data   []interface{}
}

func (r *testRecordingEmitter) Emit(event string, data interface{}) {
	r.events = append(r.events, event)
	r.data = append(r.data, data)
}

func TestCreateCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{
		createCategoryResp: catalog.Category{ID: "c1", Name: "Books", Description: "All books"},
	}
	em := &testRecordingEmitter{}
	h := NewCatalogHandler(svc, em)

	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"name":"Books","description":"All books"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.CreateCategory(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if svc.createCategoryInput.Name != "Books" {
		t.Fatalf("expected service to receive name Books, got %s", svc.createCategoryInput.Name)
	}
	var resp CategoryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != "c1" || resp.Name != "Books" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if len(em.events) != 1 || em.events[0] != ws.EventCategoryCreated {
		t.Fatalf("expected category created event, got %+v", em.events)
	}
}

func TestCreateCategory_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{}
	em := &testRecordingEmitter{}
	h := NewCatalogHandler(svc, em)

	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"description":"no name"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.CreateCategory(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if svc.createCategoryInput.Name != "" {
		t.Fatalf("expected service not to be called, got %+v", svc.createCategoryInput)
	}
	if len(em.events) != 0 {
		t.Fatalf("expected no events, got %+v", em.events)
	}
}

func TestUpdateCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{
		updateCategoryResp: catalog.Category{ID: "c1", Name: "Updated", Description: "Desc"},
	}
	em := &testRecordingEmitter{}
	h := NewCatalogHandler(svc, em)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "c1"}}
	req := httptest.NewRequest(http.MethodPut, "/categories/c1", strings.NewReader(`{"name":"Updated","description":"Desc"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.UpdateCategory(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if svc.updateCategoryInput.ID != "c1" || svc.updateCategoryInput.Name != "Updated" {
		t.Fatalf("service received wrong input: %+v", svc.updateCategoryInput)
	}
	if len(em.events) != 1 || em.events[0] != ws.EventCategoryUpdated {
		t.Fatalf("expected update event, got %+v", em.events)
	}
}

func TestDeleteCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{}
	em := &testRecordingEmitter{}
	h := NewCatalogHandler(svc, em)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "c9"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/categories/c9", nil)

	h.DeleteCategory(c)

	if status := c.Writer.Status(); status != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", status)
	}
	if svc.deleteCategoryID != "c9" {
		t.Fatalf("expected service delete ID c9, got %s", svc.deleteCategoryID)
	}
	if len(em.events) != 1 || em.events[0] != ws.EventCategoryDeleted {
		t.Fatalf("expected delete event, got %+v", em.events)
	}
}

func TestListCategories_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{listCategoriesErr: errors.New("boom")}
	h := NewCatalogHandler(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/categories", nil)

	h.ListCategories(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestListProducts_UsesQueryDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{
		listProductsResp: []catalog.Product{
			{ID: "p1", Name: "Pen", Description: "Blue", Price: 100, Stock: 5},
		},
		listProductsTotal: 1,
	}
	h := NewCatalogHandler(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/products?limit=abc&offset=3", nil)
	c.Request = req

	h.ListProducts(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if svc.listProductsFilter.Limit != 20 || svc.listProductsFilter.Offset != 3 {
		t.Fatalf("expected fallback limit 20 and offset 3, got %+v", svc.listProductsFilter)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["total"].(float64) != 1 {
		t.Fatalf("expected total 1, got %v", body["total"])
	}
}

func TestGetProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{
		getProductResp: catalog.Product{ID: "p1", Name: "Pen", Description: "Blue", Price: 10, Stock: 2},
	}
	h := NewCatalogHandler(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "p1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/products/p1", nil)

	h.GetProduct(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if svc.getProductID != "p1" {
		t.Fatalf("service called with wrong id %s", svc.getProductID)
	}
	var resp ProductResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != "p1" || resp.Name != "Pen" {
		t.Fatalf("unexpected response %+v", resp)
	}
}

func TestCreateProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{
		createProductResp: catalog.Product{ID: "p1", Name: "Pen", Description: "Blue", Price: 10, Stock: 2},
	}
	em := &testRecordingEmitter{}
	h := NewCatalogHandler(svc, em)

	body := `{"name":"Pen","description":"Blue","price":10,"stock":2}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.CreateProduct(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if svc.createProductInput.Name != "Pen" || svc.createProductInput.Price != 10 {
		t.Fatalf("service received %+v", svc.createProductInput)
	}
	if len(em.events) != 1 || em.events[0] != ws.EventProductCreated {
		t.Fatalf("expected product created event, got %+v", em.events)
	}
}

func TestUpdateProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{
		updateProductResp: catalog.Product{ID: "p1", Name: "Pen", Description: "Red", Price: 12, Stock: 5},
	}
	em := &recordingEmitter{}
	h := NewCatalogHandler(svc, em)

	body := `{"name":"Pen","description":"Red","price":12,"stock":5}`
	req := httptest.NewRequest(http.MethodPut, "/products/p1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "p1"}}
	c.Request = req

	h.UpdateProduct(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if svc.updateProductInput.ID != "p1" || svc.updateProductInput.Price != 12 {
		t.Fatalf("service received %+v", svc.updateProductInput)
	}
	if len(em.events) != 1 || em.events[0] != ws.EventProductUpdated {
		t.Fatalf("expected product updated event, got %+v", em.events)
	}
}

func TestDeleteProduct_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{deleteProductErr: errors.New("fail")}
	em := &recordingEmitter{}
	h := NewCatalogHandler(svc, em)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "p1"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/products/p1", nil)

	h.DeleteProduct(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if len(em.events) != 0 {
		t.Fatalf("expected no event when delete fails, got %+v", em.events)
	}
}

func TestAddProductCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{}
	em := &recordingEmitter{}
	h := NewCatalogHandler(svc, em)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		{Key: "id", Value: "p1"},
		{Key: "categoryId", Value: "c1"},
	}
	c.Request = httptest.NewRequest(http.MethodPost, "/products/p1/categories/c1", nil)

	h.AddProductCategory(c)

	if status := c.Writer.Status(); status != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", status)
	}
	if svc.assignProductCategoryProductID != "p1" || svc.assignProductCategoryCategoryID != "c1" {
		t.Fatalf("service received wrong IDs %+v/%+v", svc.assignProductCategoryProductID, svc.assignProductCategoryCategoryID)
	}
	if len(em.events) != 1 || em.events[0] != ws.EventProductCategoryAssigned {
		t.Fatalf("expected assignment event, got %+v", em.events)
	}
}

func TestSearch_Category(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{
		searchResp: catalog.SearchResult{
			Categories: []catalog.Category{{ID: "c1", Name: "Books"}},
			Total:      1,
		},
	}
	h := NewCatalogHandler(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/search?type=category&q=bo&limit=2&offset=1&sort=name&order=asc", nil)
	c.Request = req

	h.Search(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if svc.searchFilter.Kind != "category" || svc.searchFilter.Limit != 2 || svc.searchFilter.Offset != 1 {
		t.Fatalf("service search filter %+v", svc.searchFilter)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := body["categories"]; !ok {
		t.Fatalf("expected categories key in response")
	}
}

func TestGetProductHistory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	svc := &stubCatalogService{
		historyResp: []catalog.ProductHistory{
			{ID: "h1", ProductID: "p1", Price: 10, Stock: 1, ChangedAt: t1},
			{ID: "h2", ProductID: "p1", Price: 12, Stock: 2, ChangedAt: t2},
		},
	}
	h := NewCatalogHandler(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/products/p1/history?start="+t1.Format(time.RFC3339)+"&end="+t2.Format(time.RFC3339), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "p1"}}

	h.GetProductHistory(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if svc.historyProductID != "p1" || svc.historyFilter.Start.IsZero() || svc.historyFilter.End.IsZero() {
		t.Fatalf("service history filter %+v %+v", svc.historyProductID, svc.historyFilter)
	}
	var resp []ProductHistoryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 2 || resp[0].ChangedAt == "" {
		t.Fatalf("unexpected history response %+v", resp)
	}
}

func TestGetProductHistory_InvalidDate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubCatalogService{}
	h := NewCatalogHandler(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/products/p1/history?start=bad-date", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "p1"}}

	h.GetProductHistory(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid date, got %d", w.Code)
	}
	if svc.historyProductID != "" {
		t.Fatalf("service should not be called on invalid date")
	}
}
