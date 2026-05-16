package mysql

import (
	"sync"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// UserAuthMappingRepository reads Clerk-to-internal-user mappings.
type UserAuthMappingRepository interface {
	GetByClerkUserID(clerkUserID string) (*model.UserAuthMapping, error)
}

type userAuthMappingRepository struct{}

var (
	userAuthMappingRepoOnce sync.Once
	userAuthMappingRepoInst UserAuthMappingRepository
)

// GetUserAuthMappingRepository returns the singleton mapping repository.
func GetUserAuthMappingRepository() UserAuthMappingRepository {
	userAuthMappingRepoOnce.Do(func() {
		userAuthMappingRepoInst = &userAuthMappingRepository{}
	})
	return userAuthMappingRepoInst
}

func (r *userAuthMappingRepository) GetByClerkUserID(clerkUserID string) (*model.UserAuthMapping, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	var row model.UserAuthMapping
	err := DB.Where("clerk_user_id = ?", clerkUserID).First(&row).Error
	if err != nil {
		if IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}
