package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
	commonauth "github.com/nusiss-capstone-project/identity-mservice/common/auth"
)

// UserListCampaigns lists published campaigns split into ongoing / upcoming.
// @Summary List available campaigns (user)
// @Tags user-campaign
// @Produce json
// @Param lang query string false "Preferred language for landing-page title; default en"
// @Success 200 {object} data.StandardResponse{data=data.WebCampaignListData} "success"
// @Failure 401 {object} data.StandardResponse "unauthorized"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /web/campaigns [get]
func UserListCampaigns(c *gin.Context) {
	userID, ok := commonauth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}
	lang := c.Query("lang")
	if lang == "" {
		lang = c.DefaultQuery("language", "en")
	}
	payload, err := service.GetWebCampaignService().ListCampaigns(c.Request.Context(), userID, lang)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	log.WithContext(c.Request.Context()).Infow("web_list_campaigns", "user_id", userID, "lang", lang,
		"ongoing", len(payload.Ongoing), "upcoming", len(payload.Upcoming))
	data.OK(c, payload)
}

// UserGetCampaignLanding returns campaign detail with localized landing-page content.
// @Summary Get campaign landing page (user)
// @Tags user-campaign
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param lang query string false "Preferred language; default en"
// @Success 200 {object} data.StandardResponse{data=data.WebCampaignLandingPageData} "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 401 {object} data.StandardResponse "unauthorized"
// @Failure 404 {object} data.StandardResponse "campaign not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /web/campaigns/{campaignId}/landing-page [get]
func UserGetCampaignLanding(c *gin.Context) {
	campaignID, ok := parseWebCampaignID(c)
	if !ok {
		return
	}
	userID, ok := commonauth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}
	lang := c.Query("lang")
	if lang == "" {
		lang = c.DefaultQuery("language", "en")
	}
	payload, err := service.GetWebCampaignService().GetCampaignLanding(c.Request.Context(), campaignID, userID, lang)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "campaign not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	log.WithContext(c.Request.Context()).Infow("web_get_campaign_landing",
		"campaign_id", campaignID, "user_id", userID, "lang", lang)
	data.OK(c, payload)
}

// UserGetCampaignRules returns campaign id, task group id and budget project id.
// @Summary Get campaign rules summary (user)
// @Tags user-campaign
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Success 200 {object} data.StandardResponse{data=data.WebCampaignRulesData} "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 401 {object} data.StandardResponse "unauthorized"
// @Failure 404 {object} data.StandardResponse "campaign not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /web/campaigns/{campaignId}/rules [get]
func UserGetCampaignRules(c *gin.Context) {
	campaignID, ok := parseWebCampaignID(c)
	if !ok {
		return
	}
	if _, ok := commonauth.GetUserID(c.Request.Context()); !ok {
		authError(c)
		return
	}
	payload, err := service.GetWebCampaignService().GetCampaignRules(c.Request.Context(), campaignID)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "campaign not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	log.WithContext(c.Request.Context()).Infow("web_get_campaign_rules", "campaign_id", campaignID)
	data.OK(c, payload)
}

// UserJoinCampaign joins the current user into a published campaign.
// @Summary Join campaign (user)
// @Tags user-campaign
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Success 200 {object} data.StandardResponse{data=data.WebJoinCampaignData} "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 401 {object} data.StandardResponse "unauthorized"
// @Failure 404 {object} data.StandardResponse "campaign not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /web/campaigns/{campaignId}/join [post]
func UserJoinCampaign(c *gin.Context) {
	campaignID, ok := parseWebCampaignID(c)
	if !ok {
		return
	}
	userID, ok := commonauth.GetUserID(c.Request.Context())
	if !ok {
		authError(c)
		return
	}
	payload, err := service.GetWebCampaignService().JoinCampaign(c.Request.Context(), campaignID, userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotEligible) {
			data.BizErr(c, data.CodeNotEligible, service.MsgUserNotEligible, nil)
			return
		}
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "campaign not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	log.WithContext(c.Request.Context()).Infow("web_join_campaign",
		"campaign_id", campaignID, "user_id", userID)
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
