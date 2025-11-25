package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TokenValidator validates auth tokens and returns an auth context.
type TokenValidator interface {
	Validate(token string) (AuthContext, error)
}

// AuthContext captures identity extracted from the token.
type AuthContext struct {
	UserID string
	Role   string
}

// AuthMiddleware authenticates requests using bearer tokens.
func AuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authz := c.GetHeader("Authorization")
		if authz == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		lower := strings.ToLower(authz)
		raw := strings.TrimSpace(authz)
		if strings.HasPrefix(lower, "bearer ") {
			raw = strings.TrimSpace(authz[7:])
		}
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		ctx, err := validator.Validate(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// propagate identity for handlers.
		c.Set("user_id", ctx.UserID)
		c.Set("role", ctx.Role)
		c.Next()
	}
}

// RoleMiddleware authorizes requests by allowed roles.
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}
	return func(c *gin.Context) {
		roleVal, ok := c.Get("role")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing role"})
			return
		}
		roleStr, _ := roleVal.(string)
		if _, allowed := roleSet[roleStr]; !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
