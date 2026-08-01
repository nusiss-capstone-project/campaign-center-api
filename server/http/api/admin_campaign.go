package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
)

// AdminCreateCampaign creates a campaign shell (name only).
// @Summary Create campaign (admin)
// @Tags admin-campaign
// @Accept json
// @Produce json
// @Param body body data.CreateCampaignReq true "Campaign name"
// @Success 200 {object} data.StandardResponse{data=object{campaignId=int}} "success"
// @Failure 400 {object} data.StandardResponse "validation error"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns [post]
func AdminCreateCampaign(c *gin.Context) {
	var req data.CreateCampaignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithContext(c.Request.Context()).Warnw("admin_create_campaign_bad_request", "error", err.Error())
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	id, err := service.GetCampaignAdminService().CreateCampaign(c.Request.Context(), req.Name)
	if err != nil {
		handleCampaignAdminErr(c, err)
		return
	}
	log.WithContext(c.Request.Context()).Infow("admin_create_campaign", "campaign_id", id)
	data.OK(c, gin.H{"campaignId": id})
}

// AdminCreateCampaignVersion creates a new draft version for a campaign.
// If the latest version is still draft, returns that version without creating a new one.
// @Summary Create campaign version (admin)
// @Tags admin-campaign
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Success 200 {object} data.StandardResponse{data=object{campaignId=int,version=int}} "success; returns existing draft version when latest is still draft"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 404 {object} data.StandardResponse "not found"
// @Router /admin/campaigns/{campaignId}/versions [post]
func AdminCreateCampaignVersion(c *gin.Context) {
	campaignID, ok := parseCampaignIDParam(c)
	if !ok {
		return
	}
	version, err := service.GetCampaignAdminService().CreateVersion(c.Request.Context(), campaignID)
	if err != nil {
		handleCampaignAdminErr(c, err)
		return
	}
	log.WithContext(c.Request.Context()).Infow("admin_create_campaign_version", "campaign_id", campaignID, "version", version)
	data.OK(c, gin.H{"campaignId": campaignID, "version": version})
}

// AdminEditCampaignVersion saves draft content for a campaign version.
// @Summary Edit campaign draft version (admin)
// @Tags admin-campaign
// @Accept json
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param version path int true "Version"
// @Param body body data.CampaignVO true "Draft content; read-only fields are ignored"
// @Success 200 {object} data.StandardResponse{data=data.CampaignVO} "updated campaign draft"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 409 {object} data.StandardResponse "not editable"
// @Router /admin/campaigns/{campaignId}/versions/{version} [put]
func AdminEditCampaignVersion(c *gin.Context) {
	campaignID, ok := parseCampaignIDParam(c)
	if !ok {
		return
	}
	version, err := strconv.Atoi(c.Param("version"))
	if err != nil || version < 1 {
		log.WithContext(c.Request.Context()).Warnw("admin_edit_campaign_version_bad_request", "error", service.MsgInvalidVersion)
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidVersion, nil)
		return
	}
	var req data.CampaignVO
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithContext(c.Request.Context()).Warnw("admin_edit_campaign_version_bad_request", "error", err.Error())
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	detail, err := service.GetCampaignAdminService().EditVersion(c.Request.Context(), campaignID, version, req)
	if err != nil {
		handleCampaignAdminErr(c, err)
		return
	}
	log.WithContext(c.Request.Context()).Infow("admin_edit_campaign_version", "campaign_id", campaignID, "version", version)
	data.OK(c, detail)
}

// AdminListCampaigns lists campaigns with optional filters.
// @Summary List campaigns (admin)
// @Tags admin-campaign
// @Produce json
// @Param page query int false "Page (default 1)"
// @Param pageSize query int false "Page size (default 10)"
// @Param status query int false "Campaign status filter"
// @Param campaignId query int false "Campaign ID filter"
// @Success 200 {object} data.StandardResponse{data=object{total=int,items=[]data.CampaignListVO}} "success"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns [get]
func AdminListCampaigns(c *gin.Context) {
	var req data.CampaignListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		log.WithContext(c.Request.Context()).Warnw("admin_list_campaigns_bad_request", "error", err.Error())
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	items, total, err := service.GetCampaignAdminService().ListCampaigns(req)
	if err != nil {
		handleCampaignAdminErr(c, err)
		return
	}
	log.WithContext(c.Request.Context()).Infow("admin_list_campaigns", "total", total, "count", len(items), "page", req.Page, "page_size", req.PageSize)
	data.OK(c, gin.H{"total": total, "items": items})
}

// AdminGetCampaign returns campaign detail for admin.
// @Summary Get campaign detail (admin)
// @Tags admin-campaign
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Success 200 {object} data.StandardResponse{data=data.CampaignVO} "success"
// @Failure 404 {object} data.StandardResponse "not found"
// @Router /admin/campaigns/{campaignId} [get]
func AdminGetCampaign(c *gin.Context) {
	campaignID, ok := parseCampaignIDParam(c)
	if !ok {
		return
	}
	detail, err := service.GetCampaignAdminService().GetCampaign(campaignID)
	if err != nil {
		handleCampaignAdminErr(c, err)
		return
	}
	log.WithContext(c.Request.Context()).Infow("admin_get_campaign", "campaign_id", campaignID, "status", detail.Status, "version", detail.Version)
	data.OK(c, detail)
}

// AdminPublishCampaign publishes the latest draft version onto campaigns.
// @Summary Publish campaign (admin)
// @Tags admin-campaign
// @Accept json
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param body body data.PublishOperatorReq true "Operator"
// @Success 200 {object} data.StandardResponse{data=data.CampaignVO} "success"
// @Failure 400 {object} data.StandardResponse "validation error"
// @Failure 404 {object} data.StandardResponse "not found"
// @Router /admin/campaigns/{campaignId}/publish [post]
func AdminPublishCampaign(c *gin.Context) {
	campaignID, ok := parseCampaignIDParam(c)
	if !ok {
		return
	}
	var req data.PublishOperatorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithContext(c.Request.Context()).Warnw("admin_publish_campaign_bad_request", "error", err.Error())
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
		return
	}
	detail, err := service.GetCampaignAdminService().PublishCampaign(c.Request.Context(), campaignID, req.Operator)
	if err != nil {
		handleCampaignAdminErr(c, err)
		return
	}
	log.WithContext(c.Request.Context()).Infow("admin_publish_campaign", "campaign_id", campaignID, "operator", req.Operator, "status", detail.Status, "version", detail.Version)
	data.OK(c, detail)
}

func parseCampaignIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("campaignId"), 10, 64)
	if err != nil || id < 1 {
		log.WithContext(c.Request.Context()).Warnw("admin_campaign_bad_request", "error", service.MsgInvalidCampaignID)
		data.JSON(c, http.StatusBadRequest, -1, service.MsgInvalidCampaignID, nil)
		return 0, false
	}
	return id, true
}

func handleCampaignAdminErr(c *gin.Context, err error) {
	logger := log.WithContext(c.Request.Context())
	switch {
	case mysql.IsNotFound(err):
		logger.Warnw("admin_campaign_not_found", "error", err.Error())
		data.JSON(c, http.StatusNotFound, -1, service.MsgCampaignNotFound, nil)
	case data.IsCampaignDraftNotEditable(err):
		logger.Warnw("admin_campaign_draft_not_editable", "error", err.Error())
		data.JSON(c, http.StatusConflict, -1, service.MsgCampaignDraftNotEditable, nil)
	case data.IsCampaignNoDraftToPublish(err):
		logger.Warnw("admin_campaign_no_draft_to_publish", "error", err.Error())
		data.JSON(c, http.StatusBadRequest, -1, service.MsgCampaignNoDraftToPublish, nil)
	case data.IsCampaignPublishInvalid(err):
		logger.Warnw("admin_campaign_publish_invalid", "error", err.Error())
		data.JSON(c, http.StatusBadRequest, -1, err.Error(), nil)
	case err == mysql.ErrDatabaseDisabled:
		logger.Errorw("admin_campaign_db_disabled", "error", err.Error())
		data.JSON(c, http.StatusServiceUnavailable, -1, err.Error(), nil)
	default:
		if err.Error() == service.MsgCampaignNameRequired {
			logger.Warnw("admin_campaign_name_required", "error", err.Error())
			data.JSON(c, http.StatusBadRequest, -1, service.MsgCampaignNameRequired, nil)
			return
		}
		logger.Errorw("admin_campaign_failed", "error", err.Error())
		data.JSON(c, http.StatusInternalServerError, -1, err.Error(), nil)
	}
}

func handleRepoErr(c *gin.Context, err error) {
	if err == mysql.ErrDatabaseDisabled {
		data.JSON(c, http.StatusServiceUnavailable, -1, err.Error(), nil)
		return
	}
	data.JSON(c, http.StatusInternalServerError, -1, err.Error(), nil)
}
