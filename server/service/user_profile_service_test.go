package service_test

import (
	"testing"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"github.com/lianjin/campaign-center-api/server/service"
	servicemock "github.com/lianjin/campaign-center-api/server/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUserProfileService_GetProfile_masksEmailAndMapsKYC(t *testing.T) {
	createdAt := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	users := servicemock.NewMockUserRepository(t)
	users.On("GetByID", int64(100)).Return(&model.User{
		ID: 100, Name: "alice", KYCStatus: model.KYCStatusPassed, CreatedAt: createdAt,
	}, nil)
	svc := service.NewUserProfileService(users)

	profile, err := svc.GetProfile(100, "alice@example.com")

	require.NoError(t, err)
	require.Equal(t, "alice", profile.Username)
	require.Equal(t, "a***e@example.com", profile.Email)
	require.True(t, profile.KYCChecked)
	require.Equal(t, createdAt.Format(time.RFC3339), profile.RegisteredAt)
}

func TestUserProfileService_GetProfile_notFound(t *testing.T) {
	users := servicemock.NewMockUserRepository(t)
	users.On("GetByID", int64(1)).Return(nil, gorm.ErrRecordNotFound)
	svc := service.NewUserProfileService(users)

	_, err := svc.GetProfile(1, "a@b.com")

	require.Error(t, err)
}

func TestUserProfileService_GetProfile_masksEmailVariants(t *testing.T) {
	createdAt := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name     string
		emailIn  string
		emailOut string
	}{
		{name: "standard", emailIn: "alice@example.com", emailOut: "a***e@example.com"},
		{name: "short local", emailIn: "ab@c.com", emailOut: "a*@c.com"},
		{name: "single char local", emailIn: "a@x.com", emailOut: "*@x.com"},
		{name: "no at sign", emailIn: "secret", emailOut: "s***t"},
		{name: "empty", emailIn: "  ", emailOut: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			users := servicemock.NewMockUserRepository(t)
			users.On("GetByID", int64(1)).Return(&model.User{
				ID: 1, Name: "u", KYCStatus: model.KYCStatusPending, CreatedAt: createdAt,
			}, nil)
			svc := service.NewUserProfileService(users)

			profile, err := svc.GetProfile(1, tc.emailIn)

			require.NoError(t, err)
			require.Equal(t, tc.emailOut, profile.Email)
		})
	}
}
