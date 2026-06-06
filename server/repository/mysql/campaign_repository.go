package mysql

import (
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// CampaignListFilter filters admin campaign list.
type CampaignListFilter struct {
	Page     int
	PageSize int
	Status   *int16
	Type     string
}

// CampaignRepository persists campaigns.
type CampaignRepository interface {
	Create(c *model.Campaign) error
	Update(c *model.Campaign) error
	GetByID(id int64) (*model.Campaign, error)
	List(f CampaignListFilter) ([]model.Campaign, int64, error)
	ListPublishedActiveOrUpcoming(now time.Time) ([]model.Campaign, error)
	Publish(id int64, operator string) (*model.Campaign, error)
	Archive(id int64, operator string) (*model.Campaign, error)
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
		"target_market":           c.TargetMarket,
		"registration_start_time": c.RegistrationStartTime,
		"registration_end_time":   c.RegistrationEndTime,
		"campaign_start_time":     c.CampaignStartTime,
		"campaign_end_time":       c.CampaignEndTime,
		"target_user_segment":     c.TargetUserSegment,
		"reward_rules":            c.RewardRules,
		"landing_page_id":         c.LandingPageID,
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
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
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

func (r *campaignRepository) ListPublishedActiveOrUpcoming(now time.Time) ([]model.Campaign, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var items []model.Campaign
	if err := db.Model(&model.Campaign{}).
		Where("status = ? AND campaign_end_time >= ?", model.CampaignStatusPublished, now).
		Order("campaign_start_time ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *campaignRepository) Publish(id int64, operator string) (*model.Campaign, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var updated model.Campaign
	now := time.Now()
	err = db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Campaign{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":     model.CampaignStatusPublished,
			"updated_by": operator,
			"updated_at": now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}
		log := model.AuditLog{
			EntityType:   "campaign",
			EntityID:     id,
			Action:       "PUBLISH",
			OperatorName: operator,
			DetailJSON:   `{"action":"publish"}`,
			CreatedAt:    now,
		}
		return tx.Create(&log).Error
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *campaignRepository) Archive(id int64, operator string) (*model.Campaign, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var updated model.Campaign
	now := time.Now()
	err = db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Campaign{}).Where("id = ? and status <> ?", id, model.CampaignStatusArchive).Updates(map[string]interface{}{
			"status":     model.CampaignStatusArchive,
			"updated_by": operator,
			"updated_at": now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}
		log := model.AuditLog{
			EntityType:   "campaign",
			EntityID:     id,
			Action:       "ARCHIVE",
			OperatorName: operator,
			DetailJSON:   `{"action":"archive"}`,
			CreatedAt:    now,
		}
		return tx.Create(&log).Error
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}
