package model

import "time"

// UserAccount maps to table user_account.
type UserAccount struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64     `gorm:"column:user_id;uniqueIndex:uk_user_currency"`
	Currency  string    `gorm:"column:currency;size:16;uniqueIndex:uk_user_currency"`
	Balance   float64   `gorm:"column:balance;type:decimal(18,2)"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (UserAccount) TableName() string {
	return "user_account"
}
