package server

import (
	"os"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() func(*gin.Context) {
	authToken := os.Getenv("AUTH_TOKEN")
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") != authToken {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		reqToken := c.GetHeader("Authorization")
		if reqToken != authToken {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
		}
		c.Next()
	}
}
