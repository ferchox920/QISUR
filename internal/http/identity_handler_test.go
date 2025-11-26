package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"catalog-api/internal/identity"

	"github.com/gin-gonic/gin"
)

type stubIdentityService struct {
	registerClientInput identity.RegisterUserInput
	registerClientResp  identity.User
	registerClientErr   error

	registerUserInput identity.RegisterUserInput
	registerUserResp  identity.User
	registerUserErr   error

	verifyInput identity.VerifyUserInput
	verifyErr   error

	blockInput identity.BlockUserInput
	blockErr   error

	loginInput identity.LoginInput
	loginResp  identity.AuthToken
	loginErr   error

	updateUserInput identity.UpdateUserInput
	updateUserResp  identity.User
	updateUserErr   error

	updateRoleInput identity.UpdateUserRoleInput
	updateRoleResp  identity.User
	updateRoleErr   error
}

func (s *stubIdentityService) RegisterClient(ctx context.Context, input identity.RegisterUserInput) (identity.User, error) {
	s.registerClientInput = input
	return s.registerClientResp, s.registerClientErr
}

func (s *stubIdentityService) RegisterStandardUser(ctx context.Context, input identity.RegisterUserInput) (identity.User, error) {
	s.registerUserInput = input
	return s.registerUserResp, s.registerUserErr
}

func (s *stubIdentityService) VerifyUser(ctx context.Context, input identity.VerifyUserInput) error {
	s.verifyInput = input
	return s.verifyErr
}

func (s *stubIdentityService) BlockUser(ctx context.Context, input identity.BlockUserInput) error {
	s.blockInput = input
	return s.blockErr
}

func (s *stubIdentityService) Login(ctx context.Context, input identity.LoginInput) (identity.AuthToken, error) {
	s.loginInput = input
	return s.loginResp, s.loginErr
}

func (s *stubIdentityService) SeedAdmin(ctx context.Context, seed identity.AdminSeedInput) error {
	return nil
}

func (s *stubIdentityService) UpdateUser(ctx context.Context, input identity.UpdateUserInput) (identity.User, error) {
	s.updateUserInput = input
	return s.updateUserResp, s.updateUserErr
}

func (s *stubIdentityService) UpdateUserRole(ctx context.Context, input identity.UpdateUserRoleInput) (identity.User, error) {
	s.updateRoleInput = input
	return s.updateRoleResp, s.updateRoleErr
}

func sampleUser(id, email string) identity.User {
	return identity.User{
		ID:         identity.UserID(id),
		Email:      email,
		FullName:   "John Doe",
		Role:       identity.RoleUser,
		Status:     identity.UserStatusPendingVerification,
		IsVerified: false,
	}
}

func TestRegisterClient_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubIdentityService{
		registerClientResp: sampleUser("u1", "client@example.com"),
	}
	h := NewIdentityHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/identity/users/client", strings.NewReader(`{"email":"client@example.com","password":"password123","full_name":"Client"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.RegisterClient(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if svc.registerClientInput.Email != "client@example.com" || svc.registerClientInput.FullName != "Client" {
		t.Fatalf("service received wrong input: %+v", svc.registerClientInput)
	}
	var resp IdentityResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != "u1" || resp.Email != "client@example.com" {
		t.Fatalf("unexpected response %+v", resp)
	}
}

func TestRegisterClient_BadPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubIdentityService{}
	h := NewIdentityHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/identity/users/client", strings.NewReader(`{"email":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.RegisterClient(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if svc.registerClientInput.Email != "" {
		t.Fatalf("service should not be called on bad payload")
	}
}

func TestRegisterUser_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubIdentityService{registerUserErr: errors.New("duplicate")}
	h := NewIdentityHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/identity/users", strings.NewReader(`{"email":"user@example.com","password":"password123","full_name":"User"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.RegisterUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if svc.registerUserInput.Email != "user@example.com" {
		t.Fatalf("service should be called with email, got %+v", svc.registerUserInput)
	}
}

func TestVerifyUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubIdentityService{}
	h := NewIdentityHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/identity/verify", strings.NewReader(`{"user_id":"u1","code":"123456"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.VerifyUser(c)

	if status := c.Writer.Status(); status != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", status)
	}
	if svc.verifyInput.UserID != "u1" || svc.verifyInput.Code != "123456" {
		t.Fatalf("service received wrong verify input: %+v", svc.verifyInput)
	}
}

func TestVerifyUser_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubIdentityService{verifyErr: errors.New("invalid")}
	h := NewIdentityHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/identity/verify", strings.NewReader(`{"user_id":"u1","code":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.VerifyUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlockUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubIdentityService{}
	h := NewIdentityHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "admin-1")
	c.Params = gin.Params{{Key: "id", Value: "u2"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/identity/users/u2/block", strings.NewReader(`{"reason":"abuse"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.BlockUser(c)

	if c.Writer.Status() != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", c.Writer.Status())
	}
	if svc.blockInput.AdminID != "admin-1" || svc.blockInput.UserID != "u2" {
		t.Fatalf("service received wrong block input %+v", svc.blockInput)
	}
}

func TestUpdateUser_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubIdentityService{}
	h := NewIdentityHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/identity/users/me", strings.NewReader(`{"full_name":"New Name"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateUser(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if svc.updateUserInput.UserID != "" {
		t.Fatalf("service should not be called when unauthorized")
	}
}

func TestUpdateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubIdentityService{
		updateUserResp: sampleUser("u1", "user@example.com"),
	}
	h := NewIdentityHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "u1")
	c.Request = httptest.NewRequest(http.MethodPut, "/identity/users/me", strings.NewReader(`{"full_name":"New Name"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if svc.updateUserInput.UserID != "u1" || svc.updateUserInput.UpdaterID != "u1" || svc.updateUserInput.FullName != "New Name" {
		t.Fatalf("service received wrong update input %+v", svc.updateUserInput)
	}
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubIdentityService{
		loginResp: identity.AuthToken{Token: "jwt123"},
	}
	h := NewIdentityHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/identity/login", strings.NewReader(`{"email":"user@example.com","password":"password123"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if svc.loginInput.Email != "user@example.com" {
		t.Fatalf("service login input %+v", svc.loginInput)
	}
	var resp LoginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token != "jwt123" {
		t.Fatalf("unexpected token %s", resp.Token)
	}
}

func TestLogin_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubIdentityService{loginErr: errors.New("invalid")}
	h := NewIdentityHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/identity/login", strings.NewReader(`{"email":"user@example.com","password":"wrong"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestUpdateUserRole_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubIdentityService{
		updateRoleResp: sampleUser("u2", "member@example.com"),
	}
	h := NewIdentityHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "admin-1")
	c.Params = gin.Params{{Key: "id", Value: "u2"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/identity/users/u2/role", strings.NewReader(`{"role":"admin"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateUserRole(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if svc.updateRoleInput.AdminID != "admin-1" || string(svc.updateRoleInput.UserID) != "u2" || svc.updateRoleInput.Role != identity.RoleName("admin") {
		t.Fatalf("service received wrong role input %+v", svc.updateRoleInput)
	}
}
