package service

import (
	"strings"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
)

// UserProfile is the user-facing profile payload.
type UserProfile struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	KYCChecked   bool   `json:"kycChecked"`
	RegisteredAt string `json:"registeredAt"`
}

// UserProfileService reads user profile data for web clients.
type UserProfileService interface {
	GetProfile(userID int64, email string) (*UserProfile, error)
}

type userProfileService struct {
	users mysql.UserRepository
}

var (
	userProfileServiceOnce sync.Once
	userProfileServiceInst UserProfileService
)

// NewUserProfileService builds a user profile service with explicit repositories.
func NewUserProfileService(users mysql.UserRepository) UserProfileService {
	return &userProfileService{users: users}
}

// GetUserProfileService returns the singleton user profile service.
func GetUserProfileService() UserProfileService {
	userProfileServiceOnce.Do(func() {
		userProfileServiceInst = NewUserProfileService(mysql.GetUserRepository())
	})
	return userProfileServiceInst
}

func (s *userProfileService) GetProfile(userID int64, email string) (*UserProfile, error) {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return nil, err
	}
	return &UserProfile{
		Username:     user.Name,
		Email:        maskEmail(email),
		KYCChecked:   user.KYCStatus == model.KYCStatusPassed,
		RegisteredAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func maskEmail(email string) string {
	email = strings.TrimSpace(email)
	local, domain, ok := strings.Cut(email, "@")
	if !ok || local == "" || domain == "" {
		return maskString(email)
	}
	return maskString(local) + "@" + domain
}

func maskString(value string) string {
	switch len(value) {
	case 0:
		return ""
	case 1:
		return "*"
	case 2:
		return value[:1] + "*"
	default:
		return value[:1] + "***" + value[len(value)-1:]
	}
}
