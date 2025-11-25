package http

import (
	"net/http"

	"catalog-api/internal/identity"

	"github.com/gin-gonic/gin"
)

// IdentityHandler orchestrates identity-related HTTP endpoints.
type IdentityHandler struct {
	svc identity.Service
}

func NewIdentityHandler(svc identity.Service) *IdentityHandler {
	return &IdentityHandler{svc: svc}
}

func (h *IdentityHandler) RegisterClient(c *gin.Context) {
	var req RegisterClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.svc.RegisterClient(c.Request.Context(), identity.RegisterUserInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toIdentityResponse(user))
}

func (h *IdentityHandler) RegisterUser(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.svc.RegisterStandardUser(c.Request.Context(), identity.RegisterUserInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toIdentityResponse(user))
}

func (h *IdentityHandler) VerifyUser(c *gin.Context) {
	var req VerifyUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.VerifyUser(c.Request.Context(), identity.VerifyUserInput{
		UserID: req.UserID,
		Code:   req.Code,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *IdentityHandler) BlockUser(c *gin.Context) {
	var req BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.Param("id")
	adminID := c.GetString("admin_id") // TODO: replace with real auth context extraction.

	if err := h.svc.BlockUser(c.Request.Context(), identity.BlockUserInput{
		AdminID: adminID,
		UserID:  userID,
		Reason:  req.Reason,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *IdentityHandler) UpdateUser(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.Param("id")
	updaterID := c.GetString("user_id")

	updated, err := h.svc.UpdateUser(c.Request.Context(), identity.UpdateUserInput{
		UserID:    identity.UserID(userID),
		UpdaterID: updaterID,
		FullName:  req.FullName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toIdentityResponse(updated))
}

func (h *IdentityHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.svc.Login(c.Request.Context(), identity.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token.Token})
}

func (h *IdentityHandler) UpdateUserRole(c *gin.Context) {
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.Param("id")
	adminID := c.GetString("user_id")

	updated, err := h.svc.UpdateUserRole(c.Request.Context(), identity.UpdateUserRoleInput{
		AdminID: adminID,
		UserID:  identity.UserID(userID),
		Role:    identity.RoleName(req.Role),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toIdentityResponse(updated))
}

func toIdentityResponse(u identity.User) IdentityResponse {
	return IdentityResponse{
		ID:         u.ID,
		Email:      u.Email,
		FullName:   u.FullName,
		Role:       string(u.Role),
		Status:     string(u.Status),
		IsVerified: u.IsVerified,
	}
}
