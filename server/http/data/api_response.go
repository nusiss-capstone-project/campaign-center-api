package data

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard API envelope per phase1 design.
const (
	CodeSuccess           = 0
	CodeNotEligible       = 4001
	CodeTopupNotQualified = 4002
	CodeDuplicateReward   = 4003
)

// JSON writes the common {code,message,data} response.
func JSON(c *gin.Context, httpStatus int, code int, message string, payload any) {
	c.JSON(httpStatus, gin.H{
		"code":    code,
		"message": message,
		"data":    payload,
	})
}

// OK sends HTTP 200 with code 0.
func OK(c *gin.Context, payload any) {
	JSON(c, http.StatusOK, CodeSuccess, "success", payload)
}

// BizErr sends HTTP 200 with a non-zero business code (matches doc style).
func BizErr(c *gin.Context, code int, message string, payload any) {
	JSON(c, http.StatusOK, code, message, payload)
}
