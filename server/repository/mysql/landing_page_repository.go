package mysql

import (
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// LandingPageListFilter filters admin landing page list.
type LandingPageListFilter struct {
	Page     int
	PageSize int
	Status   *int16
	Language string
}

// LandingPageRepository persists campaign landing pages.
type LandingPageRepository interface {
	Create(p *model.CampaignLandingPage) error
	Update(p *model.CampaignLandingPage) error
	GetByID(id int64) (*model.CampaignLandingPage, error)
	List(f LandingPageListFilter) ([]model.CampaignLandingPage, int64, error)
	Publish(id int64, operator string) (*model.CampaignLandingPage, error)
}

type landingPageRepository struct{}

var (
	landingPageRepositoryOnce     sync.Once
	landingPageRepositoryInstance LandingPageRepository
)

// GetLandingPageRepository returns the singleton landing page repository.
func GetLandingPageRepository() LandingPageRepository {
	landingPageRepositoryOnce.Do(func() {
		landingPageRepositoryInstance = &landingPageRepository{}
	})
	return landingPageRepositoryInstance
}

func (r *landingPageRepository) db() (*gorm.DB, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	return DB, nil
}

func (r *landingPageRepository) Create(p *model.CampaignLandingPage) error {
	db, err := r.db()
	if err != nil {
		return err
	}
	return db.Create(p).Error
}

func (r *landingPageRepository) Update(p *model.CampaignLandingPage) error {
	db, err := r.db()
	if err != nil {
		return err
	}
	return db.Model(&model.CampaignLandingPage{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"language":         p.Language,
		"banner_image_url": p.BannerImageURL,
		"title":            p.Title,
		"description":      p.Description,
		"terms":            p.Terms,
		"updated_at":       p.UpdatedAt,
		"updated_by":       p.UpdatedBy,
	}).Error
}

func (r *landingPageRepository) GetByID(id int64) (*model.CampaignLandingPage, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var p model.CampaignLandingPage
	if err := db.Where("id = ?", id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *landingPageRepository) List(f LandingPageListFilter) ([]model.CampaignLandingPage, int64, error) {
	db, err := r.db()
	if err != nil {
		return nil, 0, err
	}
	q := db.Model(&model.CampaignLandingPage{})
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.Language != "" {
		q = q.Where("language = ?", f.Language)
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
	var items []model.CampaignLandingPage
	if err := q.Order("id DESC").Offset(offset).Limit(ps).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *landingPageRepository) Publish(id int64, operator string) (*model.CampaignLandingPage, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var updated model.CampaignLandingPage
	now := time.Now()
	err = db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.CampaignLandingPage{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":     model.LandingPageStatusPublished,
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
			EntityType:   "landing_page",
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
