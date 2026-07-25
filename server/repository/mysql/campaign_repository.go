package mysql

import (
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// CampaignListFilter filters admin campaign list.
type CampaignListFilter struct {
	Page       int
	PageSize   int
	Status     *int16
	CampaignID *int64
}

// CampaignRepository persists campaigns.
type CampaignRepository interface {
	Create(c *model.Campaign) error
	Update(c *model.Campaign) error
	GetByID(id int64) (*model.Campaign, error)
	List(f CampaignListFilter) ([]model.Campaign, int64, error)
	UpdateStatus(id int64, status int16, operator string) (*model.Campaign, error)
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

func (r *campaignRepository) Create(c *model.Campaign) error {
	db, err := r.db()
	if err != nil {
		return err
	}
	return db.Create(c).Error
}

func (r *campaignRepository) Update(c *model.Campaign) error {
	db, err := r.db()
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
		"target_user_group_name":  c.TargetUserGroupName,
		"budget_project_id":       c.BudgetProjectID,
		"budget_project_name":     c.BudgetProjectName,
		"landing_page_id":         c.LandingPageID,
		"time_zone":               c.TimeZone,
		"status":                  c.Status,
		"updated_at":              c.UpdatedAt,
		"updated_by":              c.UpdatedBy,
	}).Error
}

func (r *campaignRepository) GetByID(id int64) (*model.Campaign, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var c model.Campaign
	if err := db.Where("id = ?", id).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *campaignRepository) List(f CampaignListFilter) ([]model.Campaign, int64, error) {
	db, err := r.db()
	if err != nil {
		return nil, 0, err
	}
	q := db.Model(&model.Campaign{})
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.CampaignID != nil {
		q = q.Where("id = ?", *f.CampaignID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	ps := f.PageSize
	if ps < 1 {
		ps = 10
	}
	offset := (page - 1) * ps
	var items []model.Campaign
	if err := q.Order("id DESC").Offset(offset).Limit(ps).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *campaignRepository) UpdateStatus(id int64, status int16, operator string) (*model.Campaign, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var updated model.Campaign
	err = db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Campaign{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":     status,
			"updated_by": operator,
			"updated_at": now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("id = ?", id).First(&updated).Error
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}
