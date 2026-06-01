package service_test

import (
	"testing"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"github.com/lianjin/campaign-center-api/server/service"
	servicemock "github.com/lianjin/campaign-center-api/server/mock"
	"github.com/stretchr/testify/require"
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
