package model

import "time"

// RewardTransaction maps to table reward_transactions.
type RewardTransaction struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	CampaignID    int64     `gorm:"column:campaign_id;index"`
	UserID        int64     `gorm:"column:user_id;index"`
	ParticipantID int64     `gorm:"column:participant_id;index"`
	RewardType    string    `gorm:"column:reward_type;size:64"`
	RewardAmount  float64   `gorm:"column:reward_amount;type:decimal(18,2)"`
	Status        string    `gorm:"column:status;size:32"`
	Reason        string    `gorm:"column:reason;size:255"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (RewardTransaction) TableName() string {
	return "reward_transactions"
}
