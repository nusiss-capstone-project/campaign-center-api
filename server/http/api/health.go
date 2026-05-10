package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/http/data"
)

type helloURI struct {
	Client string `uri:"client" binding:"required"`
}

// Ping liveness probe.
// @Summary Ping
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "example: {\"message\":\"pong\"}"
// @Router /ping [get]
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// SayHello sample authenticated-by-client route.
// @Summary Hello (demo)
// @Tags health
// @Produce json
// @Param client path string true "Client type" Enums(merchant, customer)
// @Param name query string false "Name to greet"
// @Success 200 {object} map[string]string "JSON fields: client, message, service"
// @Failure 400 {object} map[string]string "error"
// @Router /{client}/hello [get]
func SayHello(c *gin.Context) {
	var uri helloURI
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		name = "world"
	}
	c.JSON(http.StatusOK, gin.H{
		"client":  uri.Client,
		"message": fmt.Sprintf("Hello %s", name),
		"service": data.ServiceName,
	})
}
