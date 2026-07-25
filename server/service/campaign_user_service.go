package service

import (
	"net/http"
	"sync"

	"github.com/lianjin/campaign-center-api/server/http/data"
)

// UserCampaignService user-facing campaign flows.
// Legacy top-up/join implementation was removed with admin v2 schema;
// Phase 3 replaces HTTP with mocks and deletes remaining coupling.
type UserCampaignService interface {
	ListAvailableCampaigns(userID int64) (*data.HTTPReply, error)
	GetLandingPageUI(campaignID, userID int64, lang string) (*data.HTTPReply, error)
	JoinCampaign(campaignID, userID int64) (*data.HTTPReply, error)
	SimulateTopUp(campaignID, userID int64, amount float64) (*data.HTTPReply, error)
}

type userCampaignServiceStub struct{}

var (
	userCampaignServiceOnce sync.Once
	userCampaignServiceInst UserCampaignService
)

// NewUserCampaignService returns a temporary stub (legacy DAO paths removed).
func NewUserCampaignService(_ ...any) UserCampaignService {
	return &userCampaignServiceStub{}
}

// GetUserCampaignService returns the singleton user campaign service stub.
func GetUserCampaignService() UserCampaignService {
	userCampaignServiceOnce.Do(func() {
		userCampaignServiceInst = &userCampaignServiceStub{}
	})
	return userCampaignServiceInst
}

func userCampaignUnavailable() *data.HTTPReply {
	return &data.HTTPReply{
		HTTPStatus: http.StatusServiceUnavailable,
		Code:       -1,
		Message:    "user campaign APIs are temporarily unavailable pending rewrite",
	}
}

func (s *userCampaignServiceStub) ListAvailableCampaigns(userID int64) (*data.HTTPReply, error) {
	return userCampaignUnavailable(), nil
}

func (s *userCampaignServiceStub) GetLandingPageUI(campaignID, userID int64, lang string) (*data.HTTPReply, error) {
	return userCampaignUnavailable(), nil
}

func (s *userCampaignServiceStub) JoinCampaign(campaignID, userID int64) (*data.HTTPReply, error) {
	return userCampaignUnavailable(), nil
}

func (s *userCampaignServiceStub) SimulateTopUp(campaignID, userID int64, amount float64) (*data.HTTPReply, error) {
	return userCampaignUnavailable(), nil
}
