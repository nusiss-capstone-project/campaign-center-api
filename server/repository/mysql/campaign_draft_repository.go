package mysql

import (
	"context"
	"sync"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// CampaignDraftSummary is the latest draft version metadata for list views.
type CampaignDraftSummary struct {
	Version int
	Status  string
}

// CampaignDraftRepository persists campaign draft versions.
type CampaignDraftRepository interface {
	Create(ctx context.Context, d *model.CampaignDraft) error
	Update(ctx context.Context, d *model.CampaignDraft) error
	GetByActivityAndVersion(activityID int64, version int) (*model.CampaignDraft, error)
	GetLatestByActivityID(activityID int64) (*model.CampaignDraft, error)
	MaxVersion(activityID int64) (int, error)
	LatestSummariesByActivityIDs(activityIDs []int64) (map[int64]CampaignDraftSummary, error)
}

type campaignDraftRepository struct{}

var (
	campaignDraftRepositoryOnce     sync.Once
	campaignDraftRepositoryInstance CampaignDraftRepository
)

// GetCampaignDraftRepository returns the singleton draft repository.
func GetCampaignDraftRepository() CampaignDraftRepository {
	campaignDraftRepositoryOnce.Do(func() {
		campaignDraftRepositoryInstance = &campaignDraftRepository{}
	})
	return campaignDraftRepositoryInstance
}

func (r *campaignDraftRepository) db() (*gorm.DB, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	return DB, nil
}

func (r *campaignDraftRepository) Create(ctx context.Context, d *model.CampaignDraft) error {
	db, err := session(ctx)
	if err != nil {
		return err
	}
	return db.Create(d).Error
}

func (r *campaignDraftRepository) Update(ctx context.Context, d *model.CampaignDraft) error {
	db, err := session(ctx)
	if err != nil {
		return err
	}
	return db.Model(&model.CampaignDraft{}).Where("id = ?", d.ID).Updates(map[string]interface{}{
		"content":    d.Content,
		"status":     d.Status,
		"updated_at": d.UpdatedAt,
	}).Error
}

func (r *campaignDraftRepository) GetByActivityAndVersion(activityID int64, version int) (*model.CampaignDraft, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var d model.CampaignDraft
	if err := db.Where("activity_id = ? AND version = ?", activityID, version).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *campaignDraftRepository) GetLatestByActivityID(activityID int64) (*model.CampaignDraft, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var d model.CampaignDraft
	if err := db.Where("activity_id = ?", activityID).Order("version DESC").First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *campaignDraftRepository) MaxVersion(activityID int64) (int, error) {
	db, err := r.db()
	if err != nil {
		return 0, err
	}
	var max *int
	if err := db.Model(&model.CampaignDraft{}).
		Where("activity_id = ?", activityID).
		Select("MAX(version)").
		Scan(&max).Error; err != nil {
		return 0, err
	}
	if max == nil {
		return 0, nil
	}
	return *max, nil
}

func (r *campaignDraftRepository) LatestSummariesByActivityIDs(activityIDs []int64) (map[int64]CampaignDraftSummary, error) {
	out := make(map[int64]CampaignDraftSummary, len(activityIDs))
	if len(activityIDs) == 0 {
		return out, nil
	}
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	type row struct {
		ActivityID int64  `gorm:"column:activity_id"`
		Version    int    `gorm:"column:version"`
		Status     string `gorm:"column:status"`
	}
	var rows []row
	// Pick the row that owns MAX(version) per activity_id.
	sub := db.Model(&model.CampaignDraft{}).
		Select("activity_id, MAX(version) AS version").
		Where("activity_id IN ?", activityIDs).
		Group("activity_id")
	if err := db.Table("(?) AS latest", sub).
		Select("d.activity_id, d.version, d.status").
		Joins("JOIN campaign_drafts AS d ON d.activity_id = latest.activity_id AND d.version = latest.version").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, item := range rows {
		out[item.ActivityID] = CampaignDraftSummary{Version: item.Version, Status: item.Status}
	}
	return out, nil
}
