package service_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"github.com/lianjin/campaign-center-api/server/service"
	servicemock "github.com/lianjin/campaign-center-api/server/service/mock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func rewardRulesJSON(t *testing.T) string {
	t.Helper()
	b, err := json.Marshal(model.RewardRulesPayload{
		TopupThreshold: 100, RewardAmount: 10, RewardType: model.RewardTypeBonusCredit, MaxClaimPerUser: 1,
	})
	require.NoError(t, err)
	return string(b)
}

func publishedCampaign(regStart, regEnd time.Time) *model.Campaign {
	return &model.Campaign{
		ID:                    1,
		Status:                model.CampaignStatusPublished,
		TargetMarket:          "US",
		TargetUserSegment:     model.UserSegmentNewUser,
		LandingPageID:         10,
		RegistrationStartTime: regStart,
		RegistrationEndTime:   regEnd,
		CampaignStartTime:     regStart,
		CampaignEndTime:       regEnd.Add(24 * time.Hour),
		RewardRules:           `{"topupThreshold":100,"rewardAmount":10,"rewardType":"BONUS_CREDIT","maxClaimPerUser":1}`,
		Name:                  "C",
		Type:                  model.CampaignTypeTopupReward,
	}
}

func TestUserCampaignService_GetLandingPageUI_campaignNotFound(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(nil, gorm.ErrRecordNotFound)
	svc := service.NewUserCampaignService(cm,
		servicemock.NewMockLandingPageRepository(t),
		servicemock.NewMockParticipantRepository(t),
		servicemock.NewMockUserRepository(t),
		servicemock.NewMockRewardTransactionRepository(t),
	)
	reply, err := svc.GetLandingPageUI(1, 0, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, reply.HTTPStatus)
	require.Equal(t, "campaign not found", reply.Message)
}

func TestUserCampaignService_GetLandingPageUI_landingNotConfigured(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{ID: 1, LandingPageID: 0}, nil)
	svc := service.NewUserCampaignService(cm,
		servicemock.NewMockLandingPageRepository(t),
		servicemock.NewMockParticipantRepository(t),
		servicemock.NewMockUserRepository(t),
		servicemock.NewMockRewardTransactionRepository(t),
	)
	reply, err := svc.GetLandingPageUI(1, 0, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, reply.HTTPStatus)
	require.Contains(t, reply.Message, "not configured")
}

func TestUserCampaignService_GetLandingPageUI_languageMismatch(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{ID: 1, LandingPageID: 10, RewardRules: rewardRulesJSON(t)}, nil)
	lm := servicemock.NewMockLandingPageRepository(t)
	lm.On("GetByID", int64(10)).Return(&model.CampaignLandingPage{ID: 10, Language: "zh-CN"}, nil)
	svc := service.NewUserCampaignService(cm, lm,
		servicemock.NewMockParticipantRepository(t),
		servicemock.NewMockUserRepository(t),
		servicemock.NewMockRewardTransactionRepository(t),
	)
	reply, err := svc.GetLandingPageUI(1, 0, "en-US")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, reply.HTTPStatus)
}

func TestUserCampaignService_GetLandingPageUI_success(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	now := time.Now()
	camp := publishedCampaign(now.Add(-time.Hour), now.Add(time.Hour))
	cm.On("GetByID", int64(1)).Return(camp, nil)
	lm := servicemock.NewMockLandingPageRepository(t)
	lm.On("GetByID", int64(10)).Return(&model.CampaignLandingPage{
		ID: 10, Language: "en-US", Title: "Hi {{threshold}} {{reward}}", BannerImageURL: "x",
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(5)).Return(nil, gorm.ErrRecordNotFound)

	svc := service.NewUserCampaignService(cm, lm, pm,
		servicemock.NewMockUserRepository(t),
		servicemock.NewMockRewardTransactionRepository(t),
	)
	reply, err := svc.GetLandingPageUI(1, 5, "en-US")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, reply.HTTPStatus)
	require.Equal(t, data.CodeSuccess, reply.Code)
	dataMap, ok := reply.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "C", dataMap["campaignName"])
}

func TestUserCampaignService_JoinCampaign_notPublished(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{ID: 1, Status: model.CampaignStatusDraft}, nil)
	svc := service.NewUserCampaignService(cm,
		servicemock.NewMockLandingPageRepository(t),
		servicemock.NewMockParticipantRepository(t),
		servicemock.NewMockUserRepository(t),
		servicemock.NewMockRewardTransactionRepository(t),
	)
	reply, err := svc.JoinCampaign(1, 100)
	require.NoError(t, err)
	require.Equal(t, data.CodeNotEligible, reply.Code)
}

func TestUserCampaignService_JoinCampaign_success(t *testing.T) {
	now := time.Now()
	camp := publishedCampaign(now.Add(-time.Hour), now.Add(time.Hour))
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(camp, nil)
	um := servicemock.NewMockUserRepository(t)
	um.On("GetByID", int64(100)).Return(&model.User{
		ID: 100, KYCStatus: model.KYCStatusPassed, Segment: model.UserSegmentNewUser, Market: "US",
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(100)).Return(nil, gorm.ErrRecordNotFound)
	pm.On("Create", mock.MatchedBy(func(p *model.CampaignParticipant) bool {
		return p.UserID == 100 && p.JoinStatus == model.JoinStatusJoined
	})).Return(nil)

	svc := service.NewUserCampaignService(cm,
		servicemock.NewMockLandingPageRepository(t), pm, um,
		servicemock.NewMockRewardTransactionRepository(t),
	)
	reply, err := svc.JoinCampaign(1, 100)
	require.NoError(t, err)
	require.Equal(t, data.CodeSuccess, reply.Code)
}

func TestUserCampaignService_SimulateTopUp_notJoined(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{
		ID: 1, RewardRules: `{"topupThreshold":100,"rewardAmount":10,"rewardType":"BONUS_CREDIT","maxClaimPerUser":1}`,
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(100)).Return(nil, gorm.ErrRecordNotFound)

	svc := service.NewUserCampaignService(cm,
		servicemock.NewMockLandingPageRepository(t), pm,
		servicemock.NewMockUserRepository(t),
		servicemock.NewMockRewardTransactionRepository(t),
	)
	reply, err := svc.SimulateTopUp(1, 100, 120)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, reply.HTTPStatus)
}

func TestUserCampaignService_SimulateTopUp_belowThreshold(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{
		ID: 1, RewardRules: `{"topupThreshold":100,"rewardAmount":10,"rewardType":"BONUS_CREDIT","maxClaimPerUser":1}`,
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(100)).Return(&model.CampaignParticipant{
		ID: 1, RewardStatus: model.RewardStatusNotGranted,
	}, nil)

	svc := service.NewUserCampaignService(cm,
		servicemock.NewMockLandingPageRepository(t), pm,
		servicemock.NewMockUserRepository(t),
		servicemock.NewMockRewardTransactionRepository(t),
	)
	reply, err := svc.SimulateTopUp(1, 100, 50)
	require.NoError(t, err)
	require.Equal(t, data.CodeTopupNotQualified, reply.Code)
}

func TestUserCampaignService_SimulateTopUp_manualReview(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{
		ID: 1, RewardRules: `{"topupThreshold":100,"rewardAmount":10,"rewardType":"BONUS_CREDIT","maxClaimPerUser":1}`,
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(100)).Return(&model.CampaignParticipant{
		ID: 55, RewardStatus: model.RewardStatusNotGranted,
	}, nil)
	um := servicemock.NewMockUserRepository(t)
	um.On("GetByID", int64(100)).Return(&model.User{RiskLevel: model.RiskLevelHigh}, nil)
	pm.On("Save", mock.Anything).Return(nil)

	svc := service.NewUserCampaignService(cm,
		servicemock.NewMockLandingPageRepository(t), pm, um,
		servicemock.NewMockRewardTransactionRepository(t),
	)
	reply, err := svc.SimulateTopUp(1, 100, 120)
	require.NoError(t, err)
	require.Equal(t, "manual review required", reply.Message)
}

func TestUserCampaignService_SimulateTopUp_granted(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{
		ID: 1, RewardRules: `{"topupThreshold":100,"rewardAmount":10,"rewardType":"BONUS_CREDIT","maxClaimPerUser":1}`,
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(100)).Return(&model.CampaignParticipant{
		ID: 55, RewardStatus: model.RewardStatusNotGranted,
	}, nil)
	um := servicemock.NewMockUserRepository(t)
	um.On("GetByID", int64(100)).Return(&model.User{RiskLevel: "LOW"}, nil)
	rm := servicemock.NewMockRewardTransactionRepository(t)
	rm.On("Create", mock.MatchedBy(func(tx *model.RewardTransaction) bool {
		return tx.ParticipantID == 55
	})).Run(func(args mock.Arguments) {
		tx := args.Get(0).(*model.RewardTransaction)
		tx.ID = 999
	}).Return(nil)
	pm.On("Save", mock.Anything).Return(nil)

	svc := service.NewUserCampaignService(cm,
		servicemock.NewMockLandingPageRepository(t), pm, um, rm,
	)
	reply, err := svc.SimulateTopUp(1, 100, 120)
	require.NoError(t, err)
	require.Equal(t, data.CodeSuccess, reply.Code)
	d := reply.Data.(map[string]any)
	require.EqualValues(t, 999, d["rewardTransactionId"])
}
