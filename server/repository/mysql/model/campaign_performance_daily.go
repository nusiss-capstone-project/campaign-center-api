package model

import "time"

// CampaignPerformanceDaily maps to table campaign_performance_daily.
type CampaignPerformanceDaily struct {
	ID                 int64     `gorm:"column:id;primaryKey;autoIncrement"`
	CampaignID         int64     `gorm:"column:campaign_id;uniqueIndex:uk_campaign_date"`
	StatDate           time.Time `gorm:"column:stat_date;type:date;uniqueIndex:uk_campaign_date"`
	ParticipantCount   int64     `gorm:"column:participant_count"`
	ParticipationCount int64     `gorm:"column:participation_count"`
	RewardIssuedCount  int64     `gorm:"column:reward_issued_count"`
	RewardIssuedAmount float64   `gorm:"column:reward_issued_amount;type:decimal(18,2)"`
	RewardFailedCount  int64     `gorm:"column:reward_failed_count"`
	Currency           string    `gorm:"column:currency;size:16"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`
}

func (CampaignPerformanceDaily) TableName() string {
	return "campaign_performance_daily"
}
