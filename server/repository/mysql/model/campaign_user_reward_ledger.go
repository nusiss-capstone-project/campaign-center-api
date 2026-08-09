package model

import "time"

// CampaignUserRewardLedger maps to table campaign_user_rewards_ledger.
type CampaignUserRewardLedger struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID       int64     `gorm:"column:user_id;not null;uniqueIndex:uk_ledger_user_campaign_rule"`
	CampaignID   int64     `gorm:"column:campaign_id;not null;uniqueIndex:uk_ledger_user_campaign_rule"`
	RuleID       int64     `gorm:"column:rule_id;not null;default:0;uniqueIndex:uk_ledger_user_campaign_rule"`
	RewardStatus string    `gorm:"column:reward_status;size:64;not null;default:pending_distribution"`
	VoucherID    string    `gorm:"column:voucher_id;size:128;not null;default:''"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (CampaignUserRewardLedger) TableName() string {
	return "campaign_user_rewards_ledger"
}
