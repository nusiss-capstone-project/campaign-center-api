package model

import "time"

// CampaignDraft maps to table campaign_drafts.
type CampaignDraft struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ActivityID int64     `gorm:"column:activity_id;not null"`
	Content    string    `gorm:"column:content;type:text"`
	Version    int       `gorm:"column:version;not null;default:1"`
	Status     string    `gorm:"column:status;size:32;not null;default:draft"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (CampaignDraft) TableName() string {
	return "campaign_drafts"
}
