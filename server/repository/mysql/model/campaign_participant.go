package model

import "time"

// CampaignParticipant maps to table campaign_participants.
type CampaignParticipant struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement"`
	CampaignID   int64      `gorm:"column:campaign_id;index"`
	UserID       int64      `gorm:"column:user_id;index"`
	JoinStatus   string     `gorm:"column:join_status;size:32"`
	TaskStatus   string     `gorm:"column:task_status;size:32"`
	TopupAmount  float64    `gorm:"column:topup_amount;type:decimal(18,2)"`
	RiskStatus   string     `gorm:"column:risk_status;size:32"`
	RewardStatus string     `gorm:"column:reward_status;size:32"`
	RewardAmount float64    `gorm:"column:reward_amount;type:decimal(18,2)"`
	JoinedAt     time.Time  `gorm:"column:joined_at"`
	CompletedAt  *time.Time `gorm:"column:completed_at"`
	RewardedAt   *time.Time `gorm:"column:rewarded_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (CampaignParticipant) TableName() string {
	return "campaign_participants"
}
