package mysql

import (
	"sync"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// AccountTransactionListFilter lists account transactions for a user.
type AccountTransactionListFilter struct {
	UserID   int64
	Type     string
	CursorID int64
	Limit    int
}

// AccountTransactionRepository reads account transactions.
type AccountTransactionRepository interface {
	SumAmountByType(userID int64, currency, txnType string) (float64, error)
	List(filter AccountTransactionListFilter) ([]model.AccountTransaction, error)
}

type accountTransactionRepository struct{}

var (
	accountTransactionRepositoryOnce     sync.Once
	accountTransactionRepositoryInstance AccountTransactionRepository
)

// GetAccountTransactionRepository returns the singleton account transaction repository.
func GetAccountTransactionRepository() AccountTransactionRepository {
	accountTransactionRepositoryOnce.Do(func() {
		accountTransactionRepositoryInstance = &accountTransactionRepository{}
	})
	return accountTransactionRepositoryInstance
}

func (r *accountTransactionRepository) db() (*gorm.DB, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	return DB, nil
}

func (r *accountTransactionRepository) SumAmountByType(userID int64, currency, txnType string) (float64, error) {
	db, err := r.db()
	if err != nil {
		return 0, err
	}
	var total float64
	q := db.Model(&model.AccountTransaction{}).
		Where("user_id = ? AND currency = ? AND type = ? AND status = ?",
			userID, currency, txnType, model.AccountTxnStatusSuccess)
	if err := q.Select("COALESCE(SUM(amount), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *accountTransactionRepository) List(filter AccountTransactionListFilter) ([]model.AccountTransaction, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	q := db.Where("user_id = ?", filter.UserID)
	if filter.Type != "" {
		q = q.Where("type = ?", filter.Type)
	}
	if filter.CursorID > 0 {
		q = q.Where("id < ?", filter.CursorID)
	}
	var rows []model.AccountTransaction
	if err := q.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
