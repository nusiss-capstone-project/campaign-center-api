package mysql

import (
	"context"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// CampaignRewardRuleRepository persists flattened campaign reward rules.
type CampaignRewardRuleRepository interface {
	ListByCampaignID(campaignID int64) ([]model.CampaignRewardRule, error)
	ListByRef(refClient string, refID int64) ([]model.CampaignRewardRule, error)
	ReplaceByCampaignID(ctx context.Context, campaignID int64, rules []model.CampaignRewardRule) error
	DeleteByCampaignID(ctx context.Context, campaignID int64) error
}

type campaignRewardRuleRepository struct{}

var (
	campaignRewardRuleRepositoryOnce     sync.Once
	campaignRewardRuleRepositoryInstance CampaignRewardRuleRepository
)

// GetCampaignRewardRuleRepository returns the singleton reward rule repository.
func GetCampaignRewardRuleRepository() CampaignRewardRuleRepository {
	campaignRewardRuleRepositoryOnce.Do(func() {
		campaignRewardRuleRepositoryInstance = &campaignRewardRuleRepository{}
	})
	return campaignRewardRuleRepositoryInstance
}

func (r *campaignRewardRuleRepository) db() (*gorm.DB, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	return DB, nil
}

func (r *campaignRewardRuleRepository) ListByCampaignID(campaignID int64) ([]model.CampaignRewardRule, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var items []model.CampaignRewardRule
	if err := db.Where("campaign_id = ?", campaignID).Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *campaignRewardRuleRepository) ListByRef(refClient string, refID int64) ([]model.CampaignRewardRule, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var items []model.CampaignRewardRule
	if err := db.Where("ref_client = ? AND ref_id = ?", refClient, refID).Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *campaignRewardRuleRepository) DeleteByCampaignID(ctx context.Context, campaignID int64) error {
	db, err := session(ctx)
	if err != nil {
		return err
	}
	return db.Where("campaign_id = ?", campaignID).Delete(&model.CampaignRewardRule{}).Error
}

func (r *campaignRewardRuleRepository) ReplaceByCampaignID(ctx context.Context, campaignID int64, rules []model.CampaignRewardRule) error {
	db, err := session(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("campaign_id = ?", campaignID).Delete(&model.CampaignRewardRule{}).Error; err != nil {
			return err
		}
		if len(rules) == 0 {
			return nil
		}
		for i := range rules {
			rules[i].ID = 0
			rules[i].CampaignID = campaignID
			rules[i].CreatedAt = now
			rules[i].UpdatedAt = now
		}
		return tx.Create(&rules).Error
	})
}
