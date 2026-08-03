package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
)

// AdminGetCampaignPerformanceSummary returns mock aggregated campaign metrics.
// @Summary Get campaign performance summary (admin, mock)
// @Tags admin-campaign-performance
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Router /admin/campaigns/{campaignId}/performance/summary [get]
func AdminGetCampaignPerformanceSummary(c *gin.Context) {
	campaignID, err := parseCampaignID(c)
	if err != nil {
		return
	}
	log.WithContext(c.Request.Context()).Infow("admin_performance_summary_mock", "campaign_id", campaignID)
	data.OK(c, gin.H{
		"participantCount":   0,
		"participationCount": 0,
		"rewardIssuedCount":  0,
		"rewardIssuedAmount": 0,
	})
}

// AdminListCampaignDailyPerformance returns mock daily performance rows.
// @Summary List campaign daily performance (admin, mock)
// @Tags admin-campaign-performance
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param startDate query string true "Start date YYYY-MM-DD"
// @Param endDate query string true "End date YYYY-MM-DD"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Router /admin/campaigns/{campaignId}/performance/daily [get]
func AdminListCampaignDailyPerformance(c *gin.Context) {
	campaignID, err := parseCampaignID(c)
	if err != nil {
		return
	}
	if _, _, err := parseDateRange(c); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	log.WithContext(c.Request.Context()).Infow("admin_performance_daily_mock", "campaign_id", campaignID)
	data.OK(c, gin.H{"items": []any{}})
}

// AdminListCampaignParticipations returns mock participation records.
// @Summary List campaign participations (admin, mock)
// @Tags admin-campaign-performance
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param page query int false "Page (default 1)"
// @Param pageSize query int false "Page size (default 20)"
// @Param userId query int false "Filter by user ID"
// @Param status query string false "Filter by reward status e.g. GRANTED"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Router /admin/campaigns/{campaignId}/participations [get]
func AdminListCampaignParticipations(c *gin.Context) {
	campaignID, err := parseCampaignID(c)
	if err != nil {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	log.WithContext(c.Request.Context()).Infow("admin_participations_mock", "campaign_id", campaignID)
	data.OK(c, gin.H{"items": []any{}, "page": page, "pageSize": pageSize, "total": 0})
}

func parseCampaignID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("campaignId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid campaignId", nil)
		return 0, err
	}
	return id, nil
}

func parseDateRange(c *gin.Context) (time.Time, time.Time, error) {
	startStr, endStr := c.Query("startDate"), c.Query("endDate")
	if startStr == "" || endStr == "" {
		return time.Time{}, time.Time{}, errDateRangeRequired()
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, errInvalidDate("startDate")
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, errInvalidDate("endDate")
	}
	if start.After(end) {
		return time.Time{}, time.Time{}, errInvalidDateRange()
	}
	return start, end, nil
}

type dateRangeError string

func (e dateRangeError) Error() string { return string(e) }

func errDateRangeRequired() error { return dateRangeError("startDate and endDate are required") }

func errInvalidDate(field string) error { return dateRangeError("invalid " + field) }

func errInvalidDateRange() error {
	return dateRangeError("startDate must be before or equal to endDate")
}
