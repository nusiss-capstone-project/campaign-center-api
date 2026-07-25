package model

import "time"

// CampaignRewardRule maps to table campaign_reward_rules (flattened bindings).
type CampaignRewardRule struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement"`
	CampaignID       int64     `gorm:"column:campaign_id;not null"`
	RefClient        string    `gorm:"column:ref_client;size:32;not null"`
	RefID            int64     `gorm:"column:ref_id;not null"`
	RewardTemplateID int64     `gorm:"column:reward_template_id;not null;default:0"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (CampaignRewardRule) TableName() string {
	return "campaign_reward_rules"
}
