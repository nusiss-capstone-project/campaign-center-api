package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func authError(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    "UNAUTHORIZED",
		"message": "Authentication required",
	})
}
