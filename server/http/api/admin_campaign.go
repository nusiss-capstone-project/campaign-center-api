package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
)

// RewardRulesReq reward rule payload (JSON body fragment).
type RewardRulesReq struct {
	TopupThreshold   float64 `json:"topupThreshold" binding:"required"`
	RewardType       string  `json:"rewardType" binding:"required"`
	RewardAmount     float64 `json:"rewardAmount"`
	RewardCurrency   string  `json:"rewardCurrency"`
	RewardMode       string  `json:"rewardMode"`
	RewardPercentage float64 `json:"rewardPercentage"`
	MaxRewardAmount  float64 `json:"maxRewardAmount"`
	MaxClaimPerUser  int     `json:"maxClaimPerUser" binding:"required"`
	MinObtainDays    int     `json:"minObtainDays"`
}

// CreateCampaignReq POST /admin/campaigns body.
type CreateCampaignReq struct {
	Name                  string         `json:"name" binding:"required"`
	Type                  string         `json:"type" binding:"required"`
	TargetMarket          string         `json:"targetMarket" binding:"required"`
	RegistrationStartTime string         `json:"registrationStartTime" binding:"required"`
	RegistrationEndTime   string         `json:"registrationEndTime" binding:"required"`
	CampaignStartTime     string         `json:"campaignStartTime" binding:"required"`
	CampaignEndTime       string         `json:"campaignEndTime" binding:"required"`
	TargetUserSegment     string         `json:"targetUserSegment" binding:"required"`
	RewardRules           RewardRulesReq `json:"rewardRules" binding:"required"`
	LandingPageID         int64          `json:"landingPageId"`
}

// UpdateCampaignReq PUT /admin/campaigns/:campaignId body.
type UpdateCampaignReq struct {
	Name                  string         `json:"name" binding:"required"`
	TargetMarket          string         `json:"targetMarket" binding:"required"`
	RegistrationStartTime string         `json:"registrationStartTime" binding:"required"`
	RegistrationEndTime   string         `json:"registrationEndTime" binding:"required"`
	CampaignStartTime     string         `json:"campaignStartTime" binding:"required"`
	CampaignEndTime       string         `json:"campaignEndTime" binding:"required"`
	TargetUserSegment     string         `json:"targetUserSegment" binding:"required"`
	RewardRules           RewardRulesReq `json:"rewardRules" binding:"required"`
	LandingPageID         int64          `json:"landingPageId"`
}

// PublishOperatorReq publish action body.
type PublishOperatorReq struct {
	Operator string `json:"operator" binding:"required"`
}

func parseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func rewardRulesPayload(req RewardRulesReq) model.RewardRulesPayload {
	return model.RewardRulesPayload{
		TopupThreshold:   req.TopupThreshold,
		RewardType:       req.RewardType,
		RewardAmount:     req.RewardAmount,
		RewardCurrency:   req.RewardCurrency,
		RewardMode:       req.RewardMode,
		RewardPercentage: req.RewardPercentage,
		MaxRewardAmount:  req.MaxRewardAmount,
		MaxClaimPerUser:  req.MaxClaimPerUser,
		MinObtainDays:    req.MinObtainDays,
	}
}

// AdminCreateCampaign creates a draft campaign.
// @Summary Create campaign (admin)
// @Tags admin-campaign
// @Accept json
// @Produce json
// @Param body body CreateCampaignReq true "Campaign payload"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "validation error"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns [post]
func AdminCreateCampaign(c *gin.Context) {
	var req CreateCampaignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	regStart, err := parseRFC3339(req.RegistrationStartTime)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidRegistrationStartTime, nil)
		return
	}
	regEnd, err := parseRFC3339(req.RegistrationEndTime)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidRegistrationEndTime, nil)
		return
	}
	cs, err := parseRFC3339(req.CampaignStartTime)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidCampaignStartTime, nil)
		return
	}
	ce, err := parseRFC3339(req.CampaignEndTime)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidCampaignEndTime, nil)
		return
	}
	svc := service.GetCampaignAdminService()
	id, status, err := svc.CreateCampaign(service.CreateCampaignParams{
		Name:                  req.Name,
		Type:                  req.Type,
		TargetMarket:          req.TargetMarket,
		RegistrationStartTime: regStart,
		RegistrationEndTime:   regEnd,
		CampaignStartTime:     cs,
		CampaignEndTime:       ce,
		TargetUserSegment:     req.TargetUserSegment,
		RewardRules:           rewardRulesPayload(req.RewardRules),
		LandingPageID:         req.LandingPageID,
	})
	if err != nil {
		if err == mysql.ErrDatabaseDisabled {
			data.JSON(c, http.StatusServiceUnavailable, -1, err.Error(), nil)
			return
		}
		data.JSON(c, http.StatusInternalServerError, -1, err.Error(), nil)
		return
	}
	data.OK(c, gin.H{"campaignId": id, "status": status})
}

// AdminUpdateCampaign updates a draft campaign.
// @Summary Update campaign (admin)
// @Tags admin-campaign
// @Accept json
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param body body UpdateCampaignReq true "Campaign payload"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 409 {object} data.StandardResponse "not draft"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns/{campaignId} [put]
func AdminUpdateCampaign(c *gin.Context) {
	idStr := c.Param("campaignId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidCampaignID, nil)
		return
	}
	var req UpdateCampaignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	regStart, err := parseRFC3339(req.RegistrationStartTime)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidRegistrationStartTime, nil)
		return
	}
	regEnd, err := parseRFC3339(req.RegistrationEndTime)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidRegistrationEndTime, nil)
		return
	}
	cs, err := parseRFC3339(req.CampaignStartTime)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidCampaignStartTime, nil)
		return
	}
	ce, err := parseRFC3339(req.CampaignEndTime)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidCampaignEndTime, nil)
		return
	}
	svc := service.GetCampaignAdminService()
	err = svc.UpdateDraftCampaign(id, service.UpdateCampaignParams{
		Name:                  req.Name,
		TargetMarket:          req.TargetMarket,
		RegistrationStartTime: regStart,
		RegistrationEndTime:   regEnd,
		CampaignStartTime:     cs,
		CampaignEndTime:       ce,
		TargetUserSegment:     req.TargetUserSegment,
		RewardRules:           rewardRulesPayload(req.RewardRules),
		LandingPageID:         req.LandingPageID,
	})
	if err != nil {
		if data.IsCampaignNotDraft(err) {
			data.JSON(c, http.StatusConflict, -1, err.Error(), nil)
			return
		}
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, service.MsgCampaignNotFound, nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, gin.H{"campaignId": id})
}

// AdminListCampaigns lists campaigns with optional filters.
// @Summary List campaigns (admin)
// @Tags admin-campaign
// @Produce json
// @Param page query int false "Page (default 1)"
// @Param pageSize query int false "Page size (default 10)"
// @Param status query int false "Campaign status filter"
// @Param type query string false "Campaign type e.g. TOPUP_REWARD"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns [get]
func AdminListCampaigns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	var statusPtr *int16
	if s := c.Query("status"); s != "" {
		v, err := strconv.ParseInt(s, 10, 16)
		if err == nil {
			x := int16(v)
			statusPtr = &x
		}
	}
	campaignType := c.Query("type")
	svc := service.GetCampaignAdminService()
	items, total, err := svc.ListCampaigns(mysql.CampaignListFilter{
		Page: page, PageSize: pageSize, Status: statusPtr, Type: campaignType,
	})
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		out = append(out, gin.H{
			"id":                it.ID,
			"name":              it.Name,
			"type":              it.Type,
			"targetMarket":      it.TargetMarket,
			"targetUserSegment": it.TargetUserSegment,
			"status":            it.Status,
			"campaignStartTime": it.CampaignStartTime.Format(time.RFC3339),
			"campaignEndTime":   it.CampaignEndTime.Format(time.RFC3339),
			"landingPageId":     it.LandingPageID,
		})
	}
	data.OK(c, gin.H{"total": total, "items": out})
}

// AdminGetCampaign returns campaign detail for admin.
// @Summary Get campaign detail (admin)
// @Tags admin-campaign
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns/{campaignId} [get]
func AdminGetCampaign(c *gin.Context) {
	idStr := c.Param("campaignId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidCampaignID, nil)
		return
	}
	svc := service.GetCampaignAdminService()
	campaign, err := svc.GetCampaign(id)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, service.MsgCampaignNotFound, nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	rules, err := model.ParseRewardRulesJSON(campaign.RewardRules)
	if err != nil {
		data.JSON(c, http.StatusInternalServerError, -1, err.Error(), nil)
		return
	}
	data.OK(c, gin.H{
		"id":                    campaign.ID,
		"name":                  campaign.Name,
		"type":                  campaign.Type,
		"targetMarket":          campaign.TargetMarket,
		"registrationStartTime": campaign.RegistrationStartTime.Format(time.RFC3339),
		"registrationEndTime":   campaign.RegistrationEndTime.Format(time.RFC3339),
		"campaignStartTime":     campaign.CampaignStartTime.Format(time.RFC3339),
		"campaignEndTime":       campaign.CampaignEndTime.Format(time.RFC3339),
		"targetUserSegment":     campaign.TargetUserSegment,
		"rewardRules": gin.H{
			"topupThreshold":   rules.TopupThreshold,
			"rewardType":       rules.RewardType,
			"rewardAmount":     rules.RewardAmount,
			"rewardCurrency":   rules.RewardCurrency,
			"rewardMode":       rules.RewardMode,
			"rewardPercentage": rules.RewardPercentage,
			"maxRewardAmount":  rules.MaxRewardAmount,
			"maxClaimPerUser":  rules.MaxClaimPerUser,
			"minObtainDays":    rules.MinObtainDays,
		},
		"status":        campaign.Status,
		"landingPageId": campaign.LandingPageID,
	})
}

// AdminPublishCampaign publishes a campaign.
// @Summary Publish campaign (admin)
// @Tags admin-campaign
// @Accept json
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param body body PublishOperatorReq true "Operator"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns/{campaignId}/publish [post]
func AdminPublishCampaign(c *gin.Context) {
	idStr := c.Param("campaignId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidCampaignID, nil)
		return
	}
	var req PublishOperatorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	svc := service.GetCampaignAdminService()
	updated, err := svc.PublishCampaign(id, req.Operator)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, service.MsgCampaignNotFound, nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, gin.H{"campaignId": updated.ID, "status": updated.Status})
}

// AdminArchiveCampaign archives a draft campaign, or a published campaign when current time is outside the active window (before start or after end).
// @Summary Archive campaign (admin)
// @Tags admin-campaign
// @Accept json
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param body body PublishOperatorReq true "Operator"
// @Success 200 {object} data.StandardResponse "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 409 {object} data.StandardResponse "not eligible or already archived"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns/{campaignId}/archive [post]
func AdminArchiveCampaign(c *gin.Context) {
	idStr := c.Param("campaignId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidCampaignID, nil)
		return
	}
	var req PublishOperatorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	svc := service.GetCampaignAdminService()
	updated, err := svc.ArchiveCampaign(id, req.Operator)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, service.MsgCampaignNotFound, nil)
			return
		}
		if data.IsCampaignAlreadyArchived(err) || data.IsCampaignNotArchivable(err) {
			data.JSON(c, http.StatusConflict, -1, err.Error(), nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, gin.H{"campaignId": updated.ID, "status": updated.Status})
}

func handleRepoErr(c *gin.Context, err error) {
	if err == mysql.ErrDatabaseDisabled {
		data.JSON(c, http.StatusServiceUnavailable, -1, err.Error(), nil)
		return
	}
	data.JSON(c, http.StatusInternalServerError, -1, err.Error(), nil)
}
