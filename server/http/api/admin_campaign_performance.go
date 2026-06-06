package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
)

// AdminGetCampaignPerformanceSummary returns aggregated campaign metrics.
// @Summary Get campaign performance summary (admin)
// @Tags admin-campaign-performance
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 404 {object} data.StandardResponse "campaign not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns/{campaignId}/performance/summary [get]
func AdminGetCampaignPerformanceSummary(c *gin.Context) {
	campaignID, err := parseCampaignID(c)
	if err != nil {
		return
	}
	payload, err := service.GetCampaignPerformanceAdminService().GetPerformanceSummary(campaignID)
	if err != nil {
		handleCampaignPerfErr(c, err)
		return
	}
	data.OK(c, payload)
}

// AdminListCampaignDailyPerformance returns daily performance rows.
// @Summary List campaign daily performance (admin)
// @Tags admin-campaign-performance
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param startDate query string true "Start date YYYY-MM-DD"
// @Param endDate query string true "End date YYYY-MM-DD"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 404 {object} data.StandardResponse "campaign not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns/{campaignId}/performance/daily [get]
func AdminListCampaignDailyPerformance(c *gin.Context) {
	campaignID, err := parseCampaignID(c)
	if err != nil {
		return
	}
	start, end, err := parseDateRange(c)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	items, err := service.GetCampaignPerformanceAdminService().ListDailyPerformance(campaignID, start, end)
	if err != nil {
		handleCampaignPerfErr(c, err)
		return
	}
	data.OK(c, gin.H{"items": items})
}

// AdminListCampaignParticipations lists participation records.
// @Summary List campaign participations (admin)
// @Tags admin-campaign-performance
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param page query int false "Page (default 1)"
// @Param pageSize query int false "Page size (default 20)"
// @Param userId query int false "Filter by user ID"
// @Param status query string false "Filter by reward status e.g. GRANTED"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 404 {object} data.StandardResponse "campaign not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns/{campaignId}/participations [get]
func AdminListCampaignParticipations(c *gin.Context) {
	campaignID, err := parseCampaignID(c)
	if err != nil {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	filter := mysql.ParticipationListFilter{
		CampaignID: campaignID, Page: page, PageSize: pageSize,
		RewardStatus: c.Query("status"),
	}
	if uid := c.Query("userId"); uid != "" {
		v, parseErr := strconv.ParseInt(uid, 10, 64)
		if parseErr != nil {
			data.JSON(c, http.StatusBadRequest, -1, "invalid userId", nil)
			return
		}
		filter.UserID = &v
	}
	items, total, err := service.GetCampaignPerformanceAdminService().ListParticipations(filter)
	if err != nil {
		handleCampaignPerfErr(c, err)
		return
	}
	data.OK(c, gin.H{"items": items, "page": page, "pageSize": pageSize, "total": total})
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

func handleCampaignPerfErr(c *gin.Context, err error) {
	if mysql.IsNotFound(err) {
		data.JSON(c, http.StatusNotFound, -1, "campaign not found", nil)
		return
	}
	handleRepoErr(c, err)
}

type dateRangeError string

func (e dateRangeError) Error() string { return string(e) }

func errDateRangeRequired() error { return dateRangeError("startDate and endDate are required") }

func errInvalidDate(field string) error { return dateRangeError("invalid " + field) }

func errInvalidDateRange() error {
	return dateRangeError("startDate must be before or equal to endDate")
}
