package auth

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const ctxUserID = "auth_user_id"
const ctxIsAdmin = "auth_is_admin"

func UserIDFromContext(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

func IsAdminFromContext(c *gin.Context) bool {
	v, ok := c.Get(ctxIsAdmin)
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

func MiddlewareJWT(j *JWTService, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		h := c.GetHeader("Authorization")
		if h == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		const p = "Bearer "
		if !strings.HasPrefix(h, p) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			return
		}

		raw := strings.TrimSpace(strings.TrimPrefix(h, p))
		claims, err := j.Parse(raw)
		if err != nil {
			log.Debug("jwt parse", "err", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxIsAdmin, claims.IsAdmin)
		c.Next()
	}
}

// MiddlewareJWTOptional parses Bearer JWT when present; does not fail if header is missing.
func MiddlewareJWTOptional(j *JWTService, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			c.Next()
			return
		}
		const p = "Bearer "
		if !strings.HasPrefix(h, p) {
			c.Next()
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(h, p))
		claims, err := j.Parse(raw)
		if err != nil {
			log.Debug("optional jwt parse", "err", err)
			c.Next()
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxIsAdmin, claims.IsAdmin)
		c.Next()
	}
}

func MiddlewareAdmin(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdminFromContext(c) {
			log.Warn("admin required", "path", c.FullPath())
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}
		c.Next()
	}
}
