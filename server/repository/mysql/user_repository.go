package mysql

import (
	"sync"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// UserRepository reads mock users for eligibility checks.
type UserRepository interface {
	GetByID(id int64) (*model.User, error)
}

type userRepository struct{}

var (
	userRepositoryOnce     sync.Once
	userRepositoryInstance UserRepository
)

// GetUserRepository returns the singleton user repository.
func GetUserRepository() UserRepository {
	userRepositoryOnce.Do(func() {
		userRepositoryInstance = &userRepository{}
	})
	return userRepositoryInstance
}

func (r *userRepository) GetByID(id int64) (*model.User, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	var u model.User
	if err := DB.Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
