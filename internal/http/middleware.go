package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TokenValidator valida tokens de auth y devuelve un contexto de auth.
type TokenValidator interface {
	Validate(token string) (AuthContext, error)
}

// AuthContext captura identidad extraida del token.
type AuthContext struct {
	UserID string
	Role   string
}

// AuthMiddleware autentica peticiones usando bearer tokens.
func AuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := bearerTokenFromHeader(c.Request)
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		ctx, err := validator.Validate(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// propagar identidad hacia los handlers.
		c.Set("user_id", ctx.UserID)
		c.Set("role", ctx.Role)
		c.Next()
	}
}

func bearerTokenFromHeader(r *http.Request) string {
	authz := r.Header.Get("Authorization")
	if authz == "" {
		return ""
	}
	lower := strings.ToLower(authz)
	raw := strings.TrimSpace(authz)
	if strings.HasPrefix(lower, "bearer ") {
		raw = strings.TrimSpace(authz[7:])
	}
	return raw
}

// RoleMiddleware autoriza peticiones segun roles permitidos.
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
