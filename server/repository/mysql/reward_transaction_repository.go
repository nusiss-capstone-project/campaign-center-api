package mysql

import (
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// RewardTransactionRepository persists reward transactions.
type RewardTransactionRepository interface {
	Create(t *model.RewardTransaction) error
	// CommitGrantWithParticipant creates the reward row and updates the participant in a single DB transaction.
	CommitGrantWithParticipant(participant *model.CampaignParticipant, reward *model.RewardTransaction) error
}

type rewardTransactionRepository struct{}

var (
	rewardTransactionRepositoryOnce     sync.Once
	rewardTransactionRepositoryInstance RewardTransactionRepository
)

// GetRewardTransactionRepository returns the singleton reward transaction repository.
func GetRewardTransactionRepository() RewardTransactionRepository {
	rewardTransactionRepositoryOnce.Do(func() {
		rewardTransactionRepositoryInstance = &rewardTransactionRepository{}
	})
	return rewardTransactionRepositoryInstance
}

func (r *rewardTransactionRepository) Create(t *model.RewardTransaction) error {
	if DB == nil {
		return ErrDatabaseDisabled
	}
	return DB.Create(t).Error
}

func (r *rewardTransactionRepository) CommitGrantWithParticipant(participant *model.CampaignParticipant, reward *model.RewardTransaction) error {
	if DB == nil {
		return ErrDatabaseDisabled
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(reward).Error; err != nil {
			return err
		}
		participant.UpdatedAt = time.Now()
		if err := tx.Save(participant).Error; err != nil {
			return err
		}
		return nil
	})
}
