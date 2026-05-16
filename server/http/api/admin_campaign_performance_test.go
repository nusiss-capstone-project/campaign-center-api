package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestParseDateRange_rejectsInvertedRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(
		http.MethodGet,
		"/?startDate=2026-05-20&endDate=2026-05-10",
		nil,
	)
	c.Request = req

	_, _, err := parseDateRange(c)

	require.EqualError(t, err, "startDate must be before or equal to endDate")
}
