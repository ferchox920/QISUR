package http

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
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

// RateLimitMiddleware aplica limitacion por IP usando un IPRateLimiter compartido.
func RateLimitMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}
		if !limiter.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}

// IPRateLimiter gestiona limitadores por IP.
type IPRateLimiter struct {
	limit   rate.Limit
	burst   int
	mu      sync.Mutex
	clients map[string]*rate.Limiter
}

func NewIPRateLimiter(limit rate.Limit, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		limit:   limit,
		burst:   burst,
		clients: make(map[string]*rate.Limiter),
	}
}

func (l *IPRateLimiter) getLimiter(key string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if limiter, ok := l.clients[key]; ok {
		return limiter
	}
	limiter := rate.NewLimiter(l.limit, l.burst)
	l.clients[key] = limiter
	return limiter
}

func (l *IPRateLimiter) Allow(key string) bool {
	return l.getLimiter(key).Allow()
}

// SecurityHeadersMiddleware inyecta cabeceras defensivas basicas.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		c.Next()
	}
}
