package mysql

import (
	"context"
	"sync"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// CampaignUserRewardLedgerRepository persists user reward ledger rows.
// Write paths will be added when user reward flow is reimplemented.
type CampaignUserRewardLedgerRepository interface {
	Create(ctx context.Context, row *model.CampaignUserRewardLedger) error
	ListByCampaignAndUser(campaignID, userID int64) ([]model.CampaignUserRewardLedger, error)
}

type campaignUserRewardLedgerRepository struct{}

var (
	campaignUserRewardLedgerRepositoryOnce     sync.Once
	campaignUserRewardLedgerRepositoryInstance CampaignUserRewardLedgerRepository
)

// GetCampaignUserRewardLedgerRepository returns the singleton ledger repository.
func GetCampaignUserRewardLedgerRepository() CampaignUserRewardLedgerRepository {
	campaignUserRewardLedgerRepositoryOnce.Do(func() {
		campaignUserRewardLedgerRepositoryInstance = &campaignUserRewardLedgerRepository{}
	})
	return campaignUserRewardLedgerRepositoryInstance
}

func (r *campaignUserRewardLedgerRepository) db() (*gorm.DB, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	return DB, nil
}

func (r *campaignUserRewardLedgerRepository) Create(ctx context.Context, row *model.CampaignUserRewardLedger) error {
	db, err := r.db()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Create(row).Error
}

func (r *campaignUserRewardLedgerRepository) ListByCampaignAndUser(campaignID, userID int64) ([]model.CampaignUserRewardLedger, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var items []model.CampaignUserRewardLedger
	if err := db.Where("campaign_id = ? AND user_id = ?", campaignID, userID).
		Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
