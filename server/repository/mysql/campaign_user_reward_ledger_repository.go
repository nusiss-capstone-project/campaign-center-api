package mysql

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// CampaignUserRewardLedgerRepository persists user reward ledger rows.
type CampaignUserRewardLedgerRepository interface {
	Create(ctx context.Context, row *model.CampaignUserRewardLedger) error
	GetByID(id int64) (*model.CampaignUserRewardLedger, error)
	GetByUserCampaignRule(userID, campaignID, ruleID int64) (*model.CampaignUserRewardLedger, error)
	UpdateStatusAndVoucher(ctx context.Context, id int64, status, voucherID string) error
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
	db, err := session(ctx)
	if err != nil {
		return err
	}
	return db.Create(row).Error
}

func (r *campaignUserRewardLedgerRepository) GetByID(id int64) (*model.CampaignUserRewardLedger, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var row model.CampaignUserRewardLedger
	if err := db.Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *campaignUserRewardLedgerRepository) GetByUserCampaignRule(userID, campaignID, ruleID int64) (*model.CampaignUserRewardLedger, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var row model.CampaignUserRewardLedger
	err = db.Where("user_id = ? AND campaign_id = ? AND rule_id = ?", userID, campaignID, ruleID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *campaignUserRewardLedgerRepository) UpdateStatusAndVoucher(ctx context.Context, id int64, status, voucherID string) error {
	db, err := session(ctx)
	if err != nil {
		return err
	}
	updates := map[string]any{
		"reward_status": status,
		"updated_at":    time.Now().UTC(),
	}
	if voucherID != "" {
		updates["voucher_id"] = voucherID
	}
	return db.Model(&model.CampaignUserRewardLedger{}).Where("id = ?", id).Updates(updates).Error
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
