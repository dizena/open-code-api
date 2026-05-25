package userapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/auth"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/store"
)

type JWTMiddleware struct {
	secret     string
	mongoStore *store.MongoStore
}

func NewJWTMiddleware(secret string, mongoStore *store.MongoStore) *JWTMiddleware {
	return &JWTMiddleware{secret: secret, mongoStore: mongoStore}
}

func (m *JWTMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(m.secret, tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("account", claims.Account)

		// Query database in real-time for current role and status
		if m.mongoStore != nil {
			user, err := m.mongoStore.GetUserByAccount(c.Request.Context(), claims.Account)
			if err != nil || user == nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				c.Abort()
				return
			}
			if user.Status != store.UserStatusActive {
				c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
				c.Abort()
				return
			}
			c.Set("role", string(user.Role))
		} else {
			c.Set("role", claims.Role)
		}

		c.Next()
	}
}

func (m *JWTMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return c.Query("token")
}
