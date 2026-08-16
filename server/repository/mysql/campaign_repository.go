package mysql

import (
	"context"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// CampaignQuery narrows campaign queries; nil fields are not constrained.
type CampaignQuery struct {
	Status     *int16
	CampaignID *int64
}

// CampaignRepository persists campaigns.
type CampaignRepository interface {
	Create(ctx context.Context, c *model.Campaign) error
	Update(ctx context.Context, c *model.Campaign) error
	GetByID(id int64) (*model.Campaign, error)
	GetByIDContext(ctx context.Context, id int64) (*model.Campaign, error)
	Count(q CampaignQuery) (int64, error)
	Find(q CampaignQuery, offset, limit int) ([]model.Campaign, error)
	UpdateStatus(ctx context.Context, id int64, status int16, operator string) (*model.Campaign, error)
}

type campaignRepository struct{}

var (
	campaignRepositoryOnce     sync.Once
	campaignRepositoryInstance CampaignRepository
)

// GetCampaignRepository returns the singleton campaign repository.
func GetCampaignRepository() CampaignRepository {
	campaignRepositoryOnce.Do(func() {
		campaignRepositoryInstance = &campaignRepository{}
	})
	return campaignRepositoryInstance
}

func (r *campaignRepository) db() (*gorm.DB, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	return DB, nil
}

func (r *campaignRepository) Create(ctx context.Context, c *model.Campaign) error {
	db, err := session(ctx)
	if err != nil {
		return err
	}
	return db.Create(c).Error
}

func (r *campaignRepository) Update(ctx context.Context, c *model.Campaign) error {
	db, err := session(ctx)
	if err != nil {
		return err
	}
	return db.Model(&model.Campaign{}).Where("id = ?", c.ID).Updates(map[string]interface{}{
		"name":                    c.Name,
		"market":                  c.Market,
		"registration_start_time": c.RegistrationStartTime,
		"registration_end_time":   c.RegistrationEndTime,
		"campaign_start_time":     c.CampaignStartTime,
		"campaign_end_time":       c.CampaignEndTime,
		"target_user_group_id":    c.TargetUserGroupID,
		"budget_project_id":       c.BudgetProjectID,
		"landing_page_id":         c.LandingPageID,
		"time_zone":               c.TimeZone,
		"status":                  c.Status,
		"updated_at":              c.UpdatedAt,
		"updated_by":              c.UpdatedBy,
	}).Error
}

func (r *campaignRepository) GetByID(id int64) (*model.Campaign, error) {
	return r.GetByIDContext(context.Background(), id)
}

func (r *campaignRepository) GetByIDContext(ctx context.Context, id int64) (*model.Campaign, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var c model.Campaign
	if err := db.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *campaignRepository) Count(q CampaignQuery) (int64, error) {
	db, err := r.db()
	if err != nil {
		return 0, err
	}
	var total int64
	if err := campaignQueryScope(db, q).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *campaignRepository) Find(q CampaignQuery, offset, limit int) ([]model.Campaign, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var items []model.Campaign
	if err := campaignQueryScope(db, q).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func campaignQueryScope(db *gorm.DB, q CampaignQuery) *gorm.DB {
	tx := db.Model(&model.Campaign{})
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.CampaignID != nil {
		tx = tx.Where("id = ?", *q.CampaignID)
	}
	return tx
}

func (r *campaignRepository) UpdateStatus(ctx context.Context, id int64, status int16, operator string) (*model.Campaign, error) {
	db, err := session(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	res := db.Model(&model.Campaign{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_by": operator,
		"updated_at": now,
	})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var updated model.Campaign
	if err := db.Where("id = ?", id).First(&updated).Error; err != nil {
		return nil, err
	}
	return &updated, nil
}
