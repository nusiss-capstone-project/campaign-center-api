package model

import "time"

// CampaignLandingPageTranslation maps to campaign_landing_page_translations.
type CampaignLandingPageTranslation struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	LandingPageID int64     `gorm:"column:landing_page_id;not null"`
	Lang          string    `gorm:"column:lang;size:16;not null"`
	Title         string    `gorm:"column:title;size:255"`
	Description   string    `gorm:"column:description;type:text"`
	Terms         string    `gorm:"column:terms;type:text"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
	CreatedBy     string    `gorm:"column:created_by;size:255;not null;default:''"`
	UpdatedBy     string    `gorm:"column:updated_by;size:255;not null;default:''"`
}

func (CampaignLandingPageTranslation) TableName() string {
	return "campaign_landing_page_translations"
}
