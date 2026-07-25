package model

import "time"

// Campaign maps to table campaigns (admin v2).
type Campaign struct {
	ID                    int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name                  string    `gorm:"column:name;size:255"`
	Market                string    `gorm:"column:market;size:64"`
	RegistrationStartTime time.Time `gorm:"column:registration_start_time"`
	RegistrationEndTime   time.Time `gorm:"column:registration_end_time"`
	CampaignStartTime     time.Time `gorm:"column:campaign_start_time"`
	CampaignEndTime       time.Time `gorm:"column:campaign_end_time"`
	TargetUserGroupID     int64     `gorm:"column:target_user_group_id;not null;default:0"`
	BudgetProjectID       int64     `gorm:"column:budget_project_id;not null;default:0"`
	Status                int16     `gorm:"column:status"`
	CreatedAt             time.Time `gorm:"column:created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at"`
	CreatedBy             string    `gorm:"column:created_by;size:255;not null;default:''"`
	UpdatedBy             string    `gorm:"column:updated_by;size:255;not null;default:''"`
	LandingPageID         int64     `gorm:"column:landing_page_id;not null;default:0"`
	TimeZone              string    `gorm:"column:time_zone;size:64;not null;default:''"`
}

func (Campaign) TableName() string {
	return "campaigns"
}
