package model

import "time"

// CampaignParticipant maps to table campaign_participants (join-only).
type CampaignParticipant struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement"`
	CampaignID int64     `gorm:"column:campaign_id;uniqueIndex:uk_participant_campaign_user;not null"`
	UserID     int64     `gorm:"column:user_id;uniqueIndex:uk_participant_campaign_user;not null"`
	JoinedAt   time.Time `gorm:"column:joined_at;not null;default:CURRENT_TIMESTAMP(3)"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;autoCreateTime;default:CURRENT_TIMESTAMP(3)"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;autoUpdateTime;default:CURRENT_TIMESTAMP(3)"`
}

func (CampaignParticipant) TableName() string {
	return "campaign_participants"
}
