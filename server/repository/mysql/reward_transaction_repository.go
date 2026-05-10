package mysql

import (
	"sync"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// RewardTransactionRepository persists reward transactions.
type RewardTransactionRepository interface {
	Create(t *model.RewardTransaction) error
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
