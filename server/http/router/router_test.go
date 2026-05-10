package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter()

	pingRecorder := httptest.NewRecorder()
	pingRequest := httptest.NewRequest(http.MethodGet, "/campaign-center-api/v1/ping", nil)
	r.ServeHTTP(pingRecorder, pingRequest)
	if pingRecorder.Code != http.StatusOK {
		t.Fatalf("expected ping status 200, got %d", pingRecorder.Code)
	}

	helloRecorder := httptest.NewRecorder()
	helloRequest := httptest.NewRequest(http.MethodGet, "/campaign-center-api/v1/customer/hello?name=Copilot", nil)
	r.ServeHTTP(helloRecorder, helloRequest)
	if helloRecorder.Code != http.StatusOK {
		t.Fatalf("expected hello status 200, got %d", helloRecorder.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(helloRecorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if body["client"] != "customer" {
		t.Fatalf("expected client in response, got %q", body["client"])
	}

	invalidRecorder := httptest.NewRecorder()
	invalidRequest := httptest.NewRequest(http.MethodGet, "/campaign-center-api/v1/guest/hello", nil)
	r.ServeHTTP(invalidRecorder, invalidRequest)
	if invalidRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid client status 400, got %d", invalidRecorder.Code)
	}
}
