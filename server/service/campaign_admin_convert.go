package service

import (
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
)

// campaignDraftVO keeps only fields editable through a campaign version.
func campaignDraftVO(campaign data.CampaignVO) data.CampaignVO {
	return data.CampaignVO{
		Name:                  campaign.Name,
		Market:                campaign.Market,
		RegistrationStartTime: campaign.RegistrationStartTime,
		RegistrationEndTime:   campaign.RegistrationEndTime,
		CampaignStartTime:     campaign.CampaignStartTime,
		CampaignEndTime:       campaign.CampaignEndTime,
		TimeZone:              campaign.TimeZone,
		TargetUserGroups:      campaign.TargetUserGroups,
		Budget:                campaign.Budget,
		RewardRules:           campaign.RewardRules,
		LandingPageID:         campaign.LandingPageID,
	}
}

func applyContentToCampaign(campaign *model.Campaign, content data.CampaignVO, operator string) {
	campaign.Name = content.Name
	campaign.Market = content.Market
	campaign.RegistrationStartTime = unixToTime(content.RegistrationStartTime)
	campaign.RegistrationEndTime = unixToTime(content.RegistrationEndTime)
	campaign.CampaignStartTime = unixToTime(content.CampaignStartTime)
	campaign.CampaignEndTime = unixToTime(content.CampaignEndTime)
	campaign.TargetUserGroupID = content.TargetUserGroups.ID
	campaign.BudgetProjectID = content.Budget.ProjectID
	campaign.LandingPageID = content.LandingPageID
	campaign.TimeZone = content.TimeZone
	campaign.Status = model.CampaignStatusPublished
	campaign.UpdatedAt = time.Now()
	campaign.UpdatedBy = operator
}

func flattenRewardRules(campaignID int64, rules data.CampaignRewardRuleVO) []model.CampaignRewardRule {
	out := make([]model.CampaignRewardRule, 0, len(rules.TaskRewardItems)+1)
	out = append(out, model.CampaignRewardRule{
		CampaignID:       campaignID,
		RefClient:        model.RewardRefClientTaskGroup,
		RefID:            rules.TaskGroupID,
		RewardTemplateID: rules.TaskGroupReward,
	})
	for _, item := range rules.TaskRewardItems {
		out = append(out, model.CampaignRewardRule{
			CampaignID:       campaignID,
			RefClient:        model.RewardRefClientTask,
			RefID:            item.TaskID,
			RewardTemplateID: item.RewardTemplateID,
		})
	}
	return out
}

func campaignToListVO(campaign model.Campaign, version int) data.CampaignListVO {
	return data.CampaignListVO{
		ID:        campaign.ID,
		Name:      campaign.Name,
		Status:    campaign.Status,
		Version:   int64(version),
		CreatedAt: timeToUnix(campaign.CreatedAt),
		UpdatedAt: timeToUnix(campaign.UpdatedAt),
	}
}

func campaignToVO(campaign *model.Campaign) *data.CampaignVO {
	return &data.CampaignVO{
		ID:                    campaign.ID,
		Status:                campaign.Status,
		Name:                  campaign.Name,
		Market:                campaign.Market,
		RegistrationStartTime: timePtrToUnix(campaign.RegistrationStartTime),
		RegistrationEndTime:   timePtrToUnix(campaign.RegistrationEndTime),
		CampaignStartTime:     timePtrToUnix(campaign.CampaignStartTime),
		CampaignEndTime:       timePtrToUnix(campaign.CampaignEndTime),
		TimeZone:              campaign.TimeZone,
		TargetUserGroups: data.TargetUserGroupVO{
			ID: campaign.TargetUserGroupID,
		},
		Budget: data.BudgetVO{
			ProjectID: campaign.BudgetProjectID,
		},
		LandingPageID: campaign.LandingPageID,
		UpdatedAt:     timeToUnix(campaign.UpdatedAt),
		CreatedAt:     timeToUnix(campaign.CreatedAt),
	}
}

func applyDraftVO(campaign *data.CampaignVO, draft data.CampaignVO) {
	readOnly := *campaign
	*campaign = campaignDraftVO(draft)
	campaign.ID = readOnly.ID
	campaign.Version = readOnly.Version
	campaign.Status = readOnly.Status
	campaign.CreatedAt = readOnly.CreatedAt
	campaign.UpdatedAt = readOnly.UpdatedAt
}

func unixToTime(seconds int64) *time.Time {
	if seconds <= 0 {
		return nil
	}
	t := time.Unix(seconds, 0).UTC()
	return &t
}

func timeToUnix(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.Unix()
}

func timePtrToUnix(value *time.Time) int64 {
	if value == nil || value.IsZero() {
		return 0
	}
	return value.Unix()
}
