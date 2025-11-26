package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"catalog-api/internal/catalog"
	"catalog-api/internal/identity"
	"catalog-api/internal/ws"

	"github.com/gin-gonic/gin"
)

type stubTokenValidator struct {
	ctx    AuthContext
	err    error
	tokens []string
}

func (s *stubTokenValidator) Validate(token string) (AuthContext, error) {
	s.tokens = append(s.tokens, token)
	return s.ctx, s.err
}

func TestRouter_Healthz(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := (&RouterFactory{}).Build()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("X-Content-Type-Options"); ct != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options header to be set, got %q", ct)
	}
}

func TestRouter_WebsocketRouteExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := (&RouterFactory{
		WSHub:          ws.NewHub(nil, nil),
		TokenValidator: &stubTokenValidator{},
	}).Build()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", w.Code)
	}
}

func TestRouter_WebsocketRouteValidatesTokenFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator := &stubTokenValidator{ctx: AuthContext{UserID: "u1"}}
	router := (&RouterFactory{
		WSHub:          ws.NewHub(nil, nil),
		TokenValidator: validator,
	}).Build()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws?token=abc123", nil)
	router.ServeHTTP(w, req)

	// No upgrade -> 400, pero el token debe validarse antes.
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-upgraded request, got %d", w.Code)
	}
	if len(validator.tokens) != 1 || validator.tokens[0] != "abc123" {
		t.Fatalf("expected token to be validated from query param, got %+v", validator.tokens)
	}
}

func TestRouter_AdminCategory_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	catSvc := &stubCatalogService{}
	router := (&RouterFactory{
		CatalogHandler: NewCatalogHandler(catSvc, nil),
		TokenValidator: &stubTokenValidator{},
	}).Build()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(`{"name":"Books"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", w.Code)
	}
	if catSvc.createCategoryInput.Name != "" {
		t.Fatalf("handler should not be called when unauthorized")
	}
}

func TestRouter_AdminCategory_ForbiddenForNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	catSvc := &stubCatalogService{}
	validator := &stubTokenValidator{ctx: AuthContext{UserID: "u1", Role: "user"}}
	router := (&RouterFactory{
		CatalogHandler: NewCatalogHandler(catSvc, nil),
		TokenValidator: validator,
	}).Build()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(`{"name":"Books"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin role, got %d", w.Code)
	}
	if len(validator.tokens) != 1 || validator.tokens[0] != "token123" {
		t.Fatalf("expected token to be validated, got %+v", validator.tokens)
	}
}

func TestRouter_AdminCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	catSvc := &stubCatalogService{
		createCategoryResp: catalog.Category{ID: "c1", Name: "Books"},
	}
	validator := &stubTokenValidator{ctx: AuthContext{UserID: "admin", Role: "admin"}}
	router := (&RouterFactory{
		CatalogHandler: NewCatalogHandler(catSvc, nil),
		TokenValidator: validator,
	}).Build()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(`{"name":"Books"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer goodtoken")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if catSvc.createCategoryInput.Name != "Books" {
		t.Fatalf("service should receive payload, got %+v", catSvc.createCategoryInput)
	}
}

func TestRouter_UpdateUser_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	idSvc := &stubIdentityService{
		updateUserResp: sampleUser("u1", "user@example.com"),
	}
	router := (&RouterFactory{
		IdentityHandler: NewIdentityHandler(idSvc),
		TokenValidator:  &stubTokenValidator{},
	}).Build()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/identity/users/me", strings.NewReader(`{"full_name":"New Name"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", w.Code)
	}
	if idSvc.updateUserInput.UserID != "" {
		t.Fatalf("service should not be called when unauthorized")
	}
}

func TestRouter_AdminIdentityRoutesUseRoleMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	idSvc := &stubIdentityService{
		updateRoleResp: sampleUser("u2", "user@example.com"),
	}
	validator := &stubTokenValidator{ctx: AuthContext{UserID: "u1", Role: "user"}}
	router := (&RouterFactory{
		IdentityHandler: NewIdentityHandler(idSvc),
		TokenValidator:  validator,
	}).Build()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/identity/users/u2/role", strings.NewReader(`{"role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer tok")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin role, got %d", w.Code)
	}
	if idSvc.updateRoleInput.Role != "" {
		t.Fatalf("service should not be invoked on forbidden request")
	}
}

func TestRouter_LoginRateLimitedPerIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	idSvc := &stubIdentityService{
		loginResp: identity.AuthToken{Token: "tok"},
	}
	router := (&RouterFactory{
		IdentityHandler: NewIdentityHandler(idSvc),
	}).Build()

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/identity/login", strings.NewReader(`{"email":"a@b.c","password":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.0.2.1:1234"
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on attempt %d, got %d", i+1, w.Code)
		}
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/identity/login", strings.NewReader(`{"email":"a@b.c","password":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.0.2.1:1234"
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after exceeding rate limit, got %d", w.Code)
	}
}
