package model

import "time"

// AccountTransaction maps to table account_transaction.
type AccountTransaction struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TransactionNo string    `gorm:"column:transaction_no;size:64;uniqueIndex"`
	UserID        int64     `gorm:"column:user_id;index:idx_user_type_created_at"`
	Currency      string    `gorm:"column:currency;size:16"`
	Amount        float64   `gorm:"column:amount;type:decimal(18,2)"`
	Type          string    `gorm:"column:type;size:32;index:idx_user_type_created_at"`
	Status        string    `gorm:"column:status;size:32"`
	RelatedType   string    `gorm:"column:related_type;size:32;index:idx_related"`
	RelatedID     int64     `gorm:"column:related_id;index:idx_related"`
	BalanceAfter  float64   `gorm:"column:balance_after;type:decimal(18,2)"`
	Remark        string    `gorm:"column:remark;size:255"`
	CreatedAt     time.Time `gorm:"column:created_at;index:idx_user_type_created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (AccountTransaction) TableName() string {
	return "account_transaction"
}
