package mock

import (
	"github.com/lianjin/campaign-center-api/server/service"
	mock "github.com/stretchr/testify/mock"
)

// MockCampaignRewardNotifier is a testify mock for CampaignRewardNotifier.
type MockCampaignRewardNotifier struct {
	mock.Mock
}

func (_m *MockCampaignRewardNotifier) NotifyTopUpReward(event service.TopUpRewardEvent) {
	_m.Called(event)
}
