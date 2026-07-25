package data

// CampaignRewardRuleVO is the admin reward-rules payload.
type CampaignRewardRuleVO struct {
	ID              int64              `json:"id,omitempty"`
	CampaignID      int64              `json:"campaignId,omitempty"`
	TaskGroupID     int64              `json:"taskGroupId"`
	TaskRewardItems []TaskRewardItemVO `json:"taskRewardItems"`
	TaskGroupReward int64              `json:"taskGroupReward"`
}

// TaskRewardItemVO binds one task to a reward template.
type TaskRewardItemVO struct {
	TaskID             int64  `json:"taskId"`
	TaskName           string `json:"taskName"`
	RewardTemplateID   int64  `json:"rewardTemplateId"`
	RewardTemplateName string `json:"rewardTemplateName"`
}

// TargetUserGroupVO stores an external user-group ID and name.
type TargetUserGroupVO struct {
	ID        int64  `json:"id"`
	GroupName string `json:"groupName"`
}

// BudgetVO stores an external budget project ID and name.
type BudgetVO struct {
	ProjectID   int64  `json:"projectId"`
	ProjectName string `json:"projectName"`
}

// CampaignVO is the shared admin campaign request and response shape.
// Read-only fields in edit requests are ignored by the service.
type CampaignVO struct {
	ID                    int64                `json:"id"`
	Version               int64                `json:"version"`
	Status                int16                `json:"status"`
	Name                  string               `json:"name"`
	Market                string               `json:"market"`
	RegistrationStartTime int64                `json:"registrationStartTime"`
	RegistrationEndTime   int64                `json:"registrationEndTime"`
	CampaignStartTime     int64                `json:"campaignStartTime"`
	CampaignEndTime       int64                `json:"campaignEndTime"`
	TimeZone              string               `json:"timeZone"`
	TargetUserGroups      TargetUserGroupVO    `json:"targetUserGroups"`
	Budget                BudgetVO             `json:"budgets"`
	RewardRules           CampaignRewardRuleVO `json:"rewardRules"`
	LandingPageID         int64                `json:"landingPageId"`
	UpdatedAt             int64                `json:"updatedAt"`
	CreatedAt             int64                `json:"createdAt"`
}

// CampaignListVO is the compact campaign list item.
type CampaignListVO struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	Market            string `json:"market"`
	Status            int16  `json:"status"`
	CampaignStartTime int64  `json:"campaignStartTime"`
	CampaignEndTime   int64  `json:"campaignEndTime"`
	LandingPageID     int64  `json:"landingPageId"`
	UpdatedAt         int64  `json:"updatedAt"`
	CreatedAt         int64  `json:"createdAt"`
}

// CampaignListReq is the admin campaign list query.
type CampaignListReq struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
	Status     *int16 `form:"status"`
	CampaignID *int64 `form:"campaignId"`
}

// CreateCampaignReq is the POST /admin/campaigns body.
type CreateCampaignReq struct {
	Name string `json:"name" binding:"required"`
}

// PublishOperatorReq is the campaign publish body.
type PublishOperatorReq struct {
	Operator string `json:"operator" binding:"required"`
}
