package service

import (
	"net/http"
	"time"

	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

func (s *userCampaignService) buildLandingPageUIPayload(
	campaign *model.Campaign,
	campaignID, userID int64,
	lp *model.CampaignLandingPage,
	resolvedLang, title, descBase, termsBase string,
	rules model.RewardRulesPayload,
) (map[string]any, error) {
	var joined bool
	taskStatus := model.TaskStatusNotStarted
	rewardStatus := model.RewardStatusNotGranted
	if userID > 0 {
		if p, err := s.participants.GetByCampaignAndUser(campaignID, userID); err == nil {
			joined = true
			taskStatus = p.TaskStatus
			rewardStatus = p.RewardStatus
		} else if !mysql.IsNotFound(err) {
			return nil, err
		}
	}
	return map[string]any{
		"campaignId":            campaignID,
		"campaignName":          campaign.Name,
		"campaignType":          campaign.Type,
		"status":                campaign.Status,
		"registrationStartTime": campaign.RegistrationStartTime.Format(time.RFC3339),
		"registrationEndTime":   campaign.RegistrationEndTime.Format(time.RFC3339),
		"campaignStartTime":     campaign.CampaignStartTime.Format(time.RFC3339),
		"campaignEndTime":       campaign.CampaignEndTime.Format(time.RFC3339),
		"landingPage": map[string]any{
			"lang": resolvedLang, "defaultLang": lp.DefaultLang,
			"bannerImageUrl": lp.BannerImageURL,
			"title": title, "description": descBase, "terms": termsBase,
		},
		"rewardRule": map[string]any{
			"topupThreshold": rules.TopupThreshold,
			"rewardAmount":   rules.RewardAmount,
			"rewardType":     rules.RewardType,
		},
		"userStatus": map[string]any{
			"joined": joined, "taskStatus": taskStatus, "rewardStatus": rewardStatus,
		},
	}, nil
}

func landingPageUIReply(payload map[string]any) *HTTPReply {
	return &HTTPReply{HTTPStatus: http.StatusOK, Code: data.CodeSuccess, Message: "success", Data: payload}
}
