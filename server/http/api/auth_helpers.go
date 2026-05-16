package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/http/data"
)

func authError(c *gin.Context) {
	data.JSON(c, http.StatusUnauthorized, -1, "Authentication required", nil)
}
