package mysql

import (
	"sync"

	"github.com/lianjin/campaign-center-api/server/log"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// LandingPageTranslationRepository persists per-language landing copy.
type LandingPageTranslationRepository interface {
	GetByLandingPageAndLang(landingPageID int64, lang string) (*model.CampaignLandingPageTranslation, error)
	ListLangsByLandingPageID(landingPageID int64) ([]string, error)
	Upsert(t *model.CampaignLandingPageTranslation) error
}

type landingPageTranslationRepository struct{}

var (
	landingPageTranslationRepoOnce sync.Once
	landingPageTranslationRepoInst LandingPageTranslationRepository
)

// GetLandingPageTranslationRepository returns the singleton.
func GetLandingPageTranslationRepository() LandingPageTranslationRepository {
	landingPageTranslationRepoOnce.Do(func() {
		landingPageTranslationRepoInst = &landingPageTranslationRepository{}
	})
	return landingPageTranslationRepoInst
}

func (r *landingPageTranslationRepository) db() (*gorm.DB, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	return DB, nil
}

func (r *landingPageTranslationRepository) GetByLandingPageAndLang(
	landingPageID int64, lang string,
) (*model.CampaignLandingPageTranslation, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var row model.CampaignLandingPageTranslation
	q := db.Where("landing_page_id = ? AND lang = ?", landingPageID, lang).First(&row)
	if q.Error != nil {
		if IsNotFound(q.Error) {
			return nil, nil
		}
		return nil, q.Error
	}
	return &row, nil
}

func (r *landingPageTranslationRepository) ListLangsByLandingPageID(landingPageID int64) ([]string, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var langs []string
	err = db.Model(&model.CampaignLandingPageTranslation{}).
		Where("landing_page_id = ?", landingPageID).
		Distinct("lang").
		Order("lang ASC").
		Pluck("lang", &langs).Error
	if err != nil {
		return nil, err
	}
	if langs == nil {
		return []string{}, nil
	}
	return langs, nil
}

func (r *landingPageTranslationRepository) Upsert(t *model.CampaignLandingPageTranslation) error {
	db, err := r.db()
	if err != nil {
		return err
	}
	var cur model.CampaignLandingPageTranslation
	q := db.Where("landing_page_id = ? AND lang = ?", t.LandingPageID, t.Lang).First(&cur)
	if q.Error != nil {
		if !IsNotFound(q.Error) {
			return q.Error
		}
		log.Logger.Infow("landing_page_translation_create",
			"landing_page_id", t.LandingPageID, "lang", t.Lang)
		return db.Create(t).Error
	}
	log.Logger.Infow("landing_page_translation_update",
		"landing_page_id", t.LandingPageID, "lang", t.Lang)
	return db.Model(&cur).Updates(map[string]interface{}{
		"title":       t.Title,
		"description": t.Description,
		"terms":       t.Terms,
		"updated_at":  t.UpdatedAt,
		"updated_by":  t.UpdatedBy,
	}).Error
}
