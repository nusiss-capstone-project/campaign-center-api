package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lianjin/campaign-center-api/server/auth"
	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/log"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"github.com/lianjin/campaign-center-api/server/service"
)

// UserListCampaigns returns a mock list of ongoing/upcoming campaigns.
// @Summary List available campaigns (user, mock)
// @Tags user-campaign
// @Produce json
// @Success 200 {object} data.StandardResponse{data=data.WebCampaignListData} "success"
// @Failure 401 {object} data.StandardResponse "unauthorized"
// @Router /web/campaigns [get]
func UserListCampaigns(c *gin.Context) {
	userID, ok := auth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}
	now := time.Now().Unix()
	payload := data.WebCampaignListData{
		Ongoing: []data.WebCampaignListItem{{
			ID: 1001, Name: "Summer Deposit Bonus", Market: model.MarketSG,
			Status: model.CampaignStatusPublished,
			CampaignStartTime: now - 86400, CampaignEndTime: now + 86400*30,
			LandingPageID: 2001, Joined: false,
		}},
		Upcoming: []data.WebCampaignListItem{{
			ID: 1002, Name: "Autumn Welcome", Market: model.MarketSG,
			Status: model.CampaignStatusPublished,
			CampaignStartTime: now + 86400*7, CampaignEndTime: now + 86400*37,
			LandingPageID: 2002, Joined: false,
		}},
	}
	log.WithContext(c.Request.Context()).Infow("web_list_campaigns_mock", "user_id", userID,
		"ongoing", len(payload.Ongoing), "upcoming", len(payload.Upcoming))
	data.OK(c, payload)
}

// UserGetCampaignLanding returns a mock landing-page UI payload.
// @Summary Get campaign landing page (user, mock)
// @Tags user-campaign
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param lang query string false "Preferred language; default en"
// @Success 200 {object} data.StandardResponse{data=data.WebCampaignLandingPageData} "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 401 {object} data.StandardResponse "unauthorized"
// @Router /web/campaigns/{campaignId}/landing-page [get]
func UserGetCampaignLanding(c *gin.Context) {
	campaignID, ok := parseWebCampaignID(c)
	if !ok {
		return
	}
	userID, ok := auth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}
	lang := c.Query("lang")
	if lang == "" {
		lang = c.DefaultQuery("language", "en")
	}
	payload := data.WebCampaignLandingPageData{
		CampaignID: campaignID,
		UserID:     userID,
		Name:       "Summer Deposit Bonus",
		Market:     model.MarketSG,
		TimeZone:   "Asia/Singapore",
		Joined:     false,
		LandingPage: data.WebLandingPageContent{
			Lang:           lang,
			BannerImageURL: "https://cdn.example.com/banner.png",
			Title:          "Deposit and get a bonus",
			Description:    "Join the campaign and complete the deposit task.",
			Terms:          "One reward per user.",
		},
		Participation: data.WebParticipationStatus{
			Joined:       false,
			TaskStatus:   model.TaskStatusNotStarted,
			RewardStatus: model.RewardStatusNotGranted,
		},
	}
	log.WithContext(c.Request.Context()).Infow("web_get_campaign_landing_mock",
		"campaign_id", campaignID, "user_id", userID, "lang", lang)
	data.OK(c, payload)
}

// UserJoinCampaign returns a mock successful join response.
// @Summary Join campaign (user, mock)
// @Tags user-campaign
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Success 200 {object} data.StandardResponse{data=data.WebJoinCampaignData} "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 401 {object} data.StandardResponse "unauthorized"
// @Router /web/campaigns/{campaignId}/join [post]
func UserJoinCampaign(c *gin.Context) {
	campaignID, ok := parseWebCampaignID(c)
	if !ok {
		return
	}
	userID, ok := auth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}
	payload := data.WebJoinCampaignData{
		CampaignID: campaignID,
		UserID:     userID,
		Joined:     true,
		JoinedAt:   time.Now().Unix(),
		Message:    "joined (mock)",
	}
	log.WithContext(c.Request.Context()).Infow("web_join_campaign_mock",
		"campaign_id", campaignID, "user_id", userID)
	data.OK(c, payload)
}

// UserDepositCampaign returns a mock deposit / task-complete response.
// @Summary Deposit for campaign task (user, mock)
// @Tags user-campaign
// @Accept json
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param body body data.WebDepositReq true "Deposit amount"
// @Success 200 {object} data.StandardResponse{data=data.WebDepositData} "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 401 {object} data.StandardResponse "unauthorized"
// @Router /web/campaigns/{campaignId}/deposit [post]
func UserDepositCampaign(c *gin.Context) {
	campaignID, ok := parseWebCampaignID(c)
	if !ok {
		return
	}
	userID, ok := auth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}
	var req data.WebDepositReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithContext(c.Request.Context()).Warnw("web_deposit_campaign_bad_request", "error", err.Error())
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	currency := req.Currency
	if currency == "" {
		currency = model.DefaultCurrency
	}
	payload := data.WebDepositData{
		CampaignID:   campaignID,
		UserID:       userID,
		Amount:       req.Amount,
		Currency:     currency,
		TaskStatus:   model.TaskStatusCompleted,
		RewardStatus: model.LedgerRewardStatusPendingDistribution,
		LedgerID:     9001,
		Message:      "deposit accepted (mock)",
	}
	log.WithContext(c.Request.Context()).Infow("web_deposit_campaign_mock",
		"campaign_id", campaignID, "user_id", userID, "amount", req.Amount, "currency", currency)
	data.OK(c, payload)
}

func parseWebCampaignID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("campaignId"), 10, 64)
	if err != nil || id < 1 {
		log.WithContext(c.Request.Context()).Warnw("web_campaign_bad_request", "error", service.MsgInvalidCampaignID)
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidCampaignID, nil)
		return 0, false
	}
	return id, true
}
