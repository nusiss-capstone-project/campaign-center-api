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

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

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
