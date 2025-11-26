package http

import (
	"net/http"
	"strings"
	"sync"
	"time"

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
	limit           rate.Limit
	burst           int
	mu              sync.Mutex
	clients         map[string]*clientLimiter
	ttl             time.Duration
	cleanupInterval time.Duration
	nextCleanup     time.Time
}

func NewIPRateLimiter(limit rate.Limit, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		limit:           limit,
		burst:           burst,
		clients:         make(map[string]*clientLimiter),
		ttl:             15 * time.Minute,
		cleanupInterval: 5 * time.Minute,
		nextCleanup:     time.Now().Add(5 * time.Minute),
	}
}

type clientLimiter struct {
	limiter *rate.Limiter
	lastUse time.Time
}

func (l *IPRateLimiter) getLimiter(now time.Time, key string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if cl, ok := l.clients[key]; ok {
		cl.lastUse = now
		return cl.limiter
	}
	limiter := rate.NewLimiter(l.limit, l.burst)
	l.clients[key] = &clientLimiter{limiter: limiter, lastUse: now}
	return limiter
}

func (l *IPRateLimiter) Allow(key string) bool {
	now := time.Now()
	l.maybeCleanup(now)
	return l.getLimiter(now, key).Allow()
}

func (l *IPRateLimiter) maybeCleanup(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if now.Before(l.nextCleanup) {
		return
	}
	cutoff := now.Add(-l.ttl)
	for key, cl := range l.clients {
		if cl.lastUse.Before(cutoff) {
			delete(l.clients, key)
		}
	}
	l.nextCleanup = now.Add(l.cleanupInterval)
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
