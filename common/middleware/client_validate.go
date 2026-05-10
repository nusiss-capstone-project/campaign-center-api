package middleware

import "github.com/gin-gonic/gin"

func ValidateClient() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := c.Param("client")
		if client != "merchant" && client != "customer" {
			c.JSON(400, gin.H{"error": "Invalid client type"})
			c.Abort()
			return
		}
		c.Next()
	}
}
