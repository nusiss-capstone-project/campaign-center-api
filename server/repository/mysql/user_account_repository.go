package mysql

import (
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserAccountRepository persists user account balances.
type UserAccountRepository interface {
	GetByUserAndCurrency(userID int64, currency string) (*model.UserAccount, error)
	// CreditWithTransaction atomically increases balance and inserts an account_transaction row.
	CreditWithTransaction(txn *model.AccountTransaction) (balanceAfter float64, err error)
}

type userAccountRepository struct{}

var (
	userAccountRepositoryOnce     sync.Once
	userAccountRepositoryInstance UserAccountRepository
)

// GetUserAccountRepository returns the singleton user account repository.
func GetUserAccountRepository() UserAccountRepository {
	userAccountRepositoryOnce.Do(func() {
		userAccountRepositoryInstance = &userAccountRepository{}
	})
	return userAccountRepositoryInstance
}

func (r *userAccountRepository) db() (*gorm.DB, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	return DB, nil
}

func (r *userAccountRepository) GetByUserAndCurrency(userID int64, currency string) (*model.UserAccount, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var acc model.UserAccount
	err = db.Where("user_id = ? AND currency = ?", userID, currency).First(&acc).Error
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *userAccountRepository) CreditWithTransaction(txn *model.AccountTransaction) (float64, error) {
	db, err := r.db()
	if err != nil {
		return 0, err
	}
	var balanceAfter float64
	err = db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var acc model.UserAccount
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND currency = ?", txn.UserID, txn.Currency).
			First(&acc).Error
		if err == gorm.ErrRecordNotFound {
			acc = model.UserAccount{
				UserID: txn.UserID, Currency: txn.Currency,
				Balance: 0, CreatedAt: now, UpdatedAt: now,
			}
			if err := tx.Create(&acc).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		acc.Balance += txn.Amount
		acc.UpdatedAt = now
		if err := tx.Save(&acc).Error; err != nil {
			return err
		}
		txn.BalanceAfter = acc.Balance
		txn.CreatedAt = now
		txn.UpdatedAt = now
		if err := tx.Create(txn).Error; err != nil {
			return err
		}
		balanceAfter = acc.Balance
		return nil
	})
	return balanceAfter, err
}
