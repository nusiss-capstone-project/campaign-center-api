package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidateClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/:client", ValidateClient(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	validRecorder := httptest.NewRecorder()
	validRequest := httptest.NewRequest(http.MethodGet, "/web", nil)
	router.ServeHTTP(validRecorder, validRequest)
	if validRecorder.Code != http.StatusOK {
		t.Fatalf("expected valid client to pass, got status %d", validRecorder.Code)
	}

	invalidRecorder := httptest.NewRecorder()
	invalidRequest := httptest.NewRequest(http.MethodGet, "/guest", nil)
	router.ServeHTTP(invalidRecorder, invalidRequest)
	if invalidRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid client to fail, got status %d", invalidRecorder.Code)
	}
}
