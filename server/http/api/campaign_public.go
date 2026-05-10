package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/service"
)

// JoinCampaignReq POST join body.
type JoinCampaignReq struct {
	UserID int64 `json:"userId" binding:"required"`
}

// SimulateTopUpReq POST top-up body.
type SimulateTopUpReq struct {
	UserID int64   `json:"userId" binding:"required"`
	Amount float64 `json:"amount" binding:"required"`
}

// UserGetCampaignLanding returns landing UI payload for a campaign (template variables resolved).
// @Summary Get campaign landing page (user)
// @Tags user-campaign
// @Produce json
// @Param client path string true "Client type" Enums(merchant, customer)
// @Param campaignId path int true "Campaign ID"
// @Param userId query int false "User ID for participation status"
// @Param language query string false "Must match landing page language when set"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /{client}/campaigns/{campaignId}/landing-page [get]
func UserGetCampaignLanding(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("campaignId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid campaignId", nil)
		return
	}
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
	language := c.Query("language")

	reply, err := service.GetUserCampaignService().GetLandingPageUI(campaignID, userID, language)
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
// @Param client path string true "Client type" Enums(merchant, customer)
// @Param campaignId path int true "Campaign ID"
// @Param body body JoinCampaignReq true "User id"
// @Success 200 {object} data.StandardResponse "success or business error code in body"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /{client}/campaigns/{campaignId}/join [post]
func UserJoinCampaign(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("campaignId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid campaignId", nil)
		return
	}
	var req JoinCampaignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}

	reply, err := service.GetUserCampaignService().JoinCampaign(campaignID, req.UserID)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	data.JSON(c, reply.HTTPStatus, reply.Code, reply.Message, reply.Data)
}

// UserSimulateTopUp simulates a top-up and may grant rewards or route to manual review.
// @Summary Simulate top-up (user)
// @Tags user-campaign
// @Accept json
// @Produce json
// @Param client path string true "Client type" Enums(merchant, customer)
// @Param campaignId path int true "Campaign ID"
// @Param body body SimulateTopUpReq true "User and amount"
// @Success 200 {object} data.StandardResponse "success, manual review, or business error code"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /{client}/campaigns/{campaignId}/top-up [post]
func UserSimulateTopUp(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("campaignId"), 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, "invalid campaignId", nil)
		return
	}
	var req SimulateTopUpReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}

	reply, err := service.GetUserCampaignService().SimulateTopUp(campaignID, req.UserID, req.Amount)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	data.JSON(c, reply.HTTPStatus, reply.Code, reply.Message, reply.Data)
}
