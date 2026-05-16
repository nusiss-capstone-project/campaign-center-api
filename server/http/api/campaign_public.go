package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/auth"
	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/service"
)

// JoinCampaignReq POST join body.
type JoinCampaignReq struct{}

// SimulateTopUpReq POST top-up body.
type SimulateTopUpReq struct {
	Amount float64 `json:"amount" binding:"required"`
}

// UserGetCampaignLanding returns landing UI payload for a campaign (template variables resolved).
// @Summary Get campaign landing page (user)
// @Tags user-campaign
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param lang query string false "Preferred language; falls back to default"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /web/campaigns/{campaignId}/landing-page [get]
func UserGetCampaignLanding(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("campaignId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid campaignId", nil)
		return
	}
	userID, ok := auth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}
	lang := c.Query("lang")
	if lang == "" {
		lang = c.Query("language")
	}

	reply, err := service.GetUserCampaignService().GetLandingPageUI(campaignID, userID, lang)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	data.JSON(c, reply.HTTPStatus, reply.Code, reply.Message, reply.Data)
}

// UserJoinCampaign enrolls a user when eligibility checks pass.
// @Summary Join campaign (user)
// @Tags user-campaign
// @Accept json
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Success 200 {object} data.StandardResponse "success or business error code in body"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /web/campaigns/{campaignId}/join [post]
func UserJoinCampaign(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("campaignId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid campaignId", nil)
		return
	}
	userID, ok := auth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}

	reply, err := service.GetUserCampaignService().JoinCampaign(campaignID, userID)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	data.JSON(c, reply.HTTPStatus, reply.Code, reply.Message, reply.Data)
}

// UserSimulateTopUp simulates a top-up: credits account (RECHARGE) then asynchronously processes campaign reward.
// @Summary Simulate top-up with account recharge (user)
// @Tags user-campaign
// @Accept json
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param body body SimulateTopUpReq true "Top-up amount"
// @Success 200 {object} data.StandardResponse "success, manual review, or business error code"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /web/campaigns/{campaignId}/top-up [post]
func UserSimulateTopUp(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("campaignId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid campaignId", nil)
		return
	}
	userID, ok := auth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}
	var req SimulateTopUpReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}

	reply, err := service.GetUserCampaignService().SimulateTopUp(campaignID, userID, req.Amount)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	data.JSON(c, reply.HTTPStatus, reply.Code, reply.Message, reply.Data)
}
