package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func BalanceCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		mongoStore := store.GetMongoStore()
		if mongoStore == nil || !mongoStore.IsEnabled() {
			c.Next()
			return
		}

		userIDStr, exists := c.Get("userId")
		if !exists {
			c.Next()
			return
		}

		userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
		if err != nil {
			c.Next()
			return
		}

		user, err := mongoStore.GetUserByID(c.Request.Context(), userID)
		if err != nil || user == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Your balance is insufficient."})
			return
		}

		if user.Balance <= 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Your balance is insufficient."})
			return
		}

		c.Next()
	}
}
