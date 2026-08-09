package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
)

// AdminListCampaignUsers lists participants for a campaign.
// @Summary List campaign participants (admin)
// @Tags admin-participant
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Success 200 {object} data.StandardResponse{data=data.AdminParticipantListData} "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 404 {object} data.StandardResponse "campaign not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns/{campaignId}/users [get]
func AdminListCampaignUsers(c *gin.Context) {
	campaignID, ok := parseCampaignIDParam(c)
	if !ok {
		return
	}
	out, err := service.GetAdminParticipantService().ListParticipants(campaignID)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "campaign not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, out)
}

// AdminGetCampaignUser returns one participant detail.
// @Summary Get campaign participant detail (admin)
// @Tags admin-participant
// @Produce json
// @Param campaignId path int true "Campaign ID"
// @Param userId path int true "User ID"
// @Success 200 {object} data.StandardResponse{data=data.AdminParticipantVO} "success"
// @Failure 400 {object} data.StandardResponse "bad request"
// @Failure 404 {object} data.StandardResponse "not found"
// @Failure 503 {object} data.StandardResponse "database unavailable"
// @Router /admin/campaigns/{campaignId}/users/{userId} [get]
func AdminGetCampaignUser(c *gin.Context) {
	campaignID, ok := parseCampaignIDParam(c)
	if !ok {
		return
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil || userID < 1 {
		data.JSON(c, http.StatusBadRequest, -1, "invalid userId", nil)
		return
	}
	out, err := service.GetAdminParticipantService().GetParticipant(campaignID, userID)
	if err != nil {
		if mysql.IsNotFound(err) {
			data.JSON(c, http.StatusNotFound, -1, "participant not found", nil)
			return
		}
		handleRepoErr(c, err)
		return
	}
	data.OK(c, out)
}
