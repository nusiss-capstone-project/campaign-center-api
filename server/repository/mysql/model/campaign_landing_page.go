package model

import "time"

// CampaignLandingPage maps to table campaign_landing_pages.
type CampaignLandingPage struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Language       string    `gorm:"column:language;size:16"`
	BannerImageURL string    `gorm:"column:banner_image_url;size:512"`
	Title          string    `gorm:"column:title;size:255"`
	Description    string    `gorm:"column:description;type:text"`
	Terms          string    `gorm:"column:terms;type:text"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
	Status         int16     `gorm:"column:status"`
	CreatedBy      string    `gorm:"column:created_by;size:255;not null;default:''"`
	UpdatedBy      string    `gorm:"column:updated_by;size:255;not null;default:''"`
}

func (CampaignLandingPage) TableName() string {
	return "campaign_landing_pages"
}
