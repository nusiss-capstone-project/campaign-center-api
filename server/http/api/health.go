package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping liveness probe.
// @Summary Ping
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "example: {\"message\":\"pong\"}"
// @Router /ping [get]
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
