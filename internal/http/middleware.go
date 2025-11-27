package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

		// Contexto del request para Redis
		if !limiter.Allow(c.Request.Context(), ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}

// IPRateLimiter gestiona limitadores usando Redis.
type IPRateLimiter struct {
	redisClient *redis.Client
	limit       int           // Peticiones permitidas
	window      time.Duration // Ventana de tiempo (ej. 1 minuto)
}

// NewIPRateLimiter crea el limitador respaldado por Redis.
func NewIPRateLimiter(rdb *redis.Client, limit int, window time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		redisClient: rdb,
		limit:       limit,
		window:      window,
	}
}

// Allow verifica si la IP puede realizar la petición.
func (l *IPRateLimiter) Allow(ctx context.Context, ip string) bool {
	// Clave única por IP y ventana de tiempo (simple fixed window)
	// Ejemplo: rate:192.168.1.1 (con expiración)
	key := fmt.Sprintf("rate:%s", ip)

	// Incrementamos el contador
	count, err := l.redisClient.Incr(ctx, key).Result()
	if err != nil {
		// En caso de error de Redis, permitimos el tráfico para no bloquear el servicio (Fail Open)
		return true
	}

	// Si es la primera petición, establecemos el TTL
	if count == 1 {
		l.redisClient.Expire(ctx, key, l.window)
	}

	return count <= int64(l.limit)
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
