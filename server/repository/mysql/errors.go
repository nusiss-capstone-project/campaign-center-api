package mysql

import (
	"errors"

	"gorm.io/gorm"
)

var ErrDatabaseDisabled = errors.New("mysql repository is disabled or not initialized")

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
