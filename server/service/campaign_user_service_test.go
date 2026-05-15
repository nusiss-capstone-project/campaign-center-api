package service_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"github.com/lianjin/campaign-center-api/server/service"
	servicemock "github.com/lianjin/campaign-center-api/server/service/mock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type noopRewardNotifier struct{}

func (noopRewardNotifier) NotifyTopUpReward(service.TopUpRewardEvent) {}

func newTestUserCampaignService(
	t *testing.T,
	cm *servicemock.MockCampaignRepository,
	lm *servicemock.MockLandingPageRepository,
	trans mysql.LandingPageTranslationRepository,
	pm *servicemock.MockParticipantRepository,
	um *servicemock.MockUserRepository,
	am service.AccountService,
	rn service.CampaignRewardNotifier,
) service.UserCampaignService {
	t.Helper()
	if trans == nil {
		trans = mysql.NewNoopLandingPageTranslationRepository()
	}
	if am == nil {
		am = &servicemock.MockAccountService{}
	}
	if rn == nil {
		rn = noopRewardNotifier{}
	}
	return service.NewUserCampaignService(cm, lm, trans, pm, um, am, rn)
}

func defaultRechargeMock(t *testing.T) *servicemock.MockAccountService {
	t.Helper()
	am := &servicemock.MockAccountService{}
	am.On("Recharge", mock.Anything, mock.Anything, mock.Anything).
		Return(&service.RechargeResult{TransactionNo: "TXN_TEST", BalanceAfter: 220}, nil)
	return am
}

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
	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil,
		servicemock.NewMockParticipantRepository(t),
		servicemock.NewMockUserRepository(t), nil, nil,
	)
	reply, err := svc.GetLandingPageUI(1, 0, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, reply.HTTPStatus)
	require.Equal(t, "campaign not found", reply.Message)
}

func TestUserCampaignService_GetLandingPageUI_landingNotConfigured(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{ID: 1, LandingPageID: 0}, nil)
	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil,
		servicemock.NewMockParticipantRepository(t),
		servicemock.NewMockUserRepository(t), nil, nil,
	)
	reply, err := svc.GetLandingPageUI(1, 0, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, reply.HTTPStatus)
	require.Contains(t, reply.Message, "not configured")
}

func TestUserCampaignService_GetLandingPageUI_fallbackWhenMissingTranslation(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{ID: 1, LandingPageID: 10, RewardRules: rewardRulesJSON(t)}, nil)
	lm := servicemock.NewMockLandingPageRepository(t)
	lm.On("GetByID", int64(10)).Return(&model.CampaignLandingPage{
		ID: 10, DefaultLang: "zh-CN", Title: "ZH", Description: "d", Terms: "t",
	}, nil)
	svc := newTestUserCampaignService(t, cm, lm, nil,
		servicemock.NewMockParticipantRepository(t),
		servicemock.NewMockUserRepository(t), nil, nil,
	)
	reply, err := svc.GetLandingPageUI(1, 0, "en-US")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, reply.HTTPStatus)
	dataMap := reply.Data.(map[string]any)
	lp := dataMap["landingPage"].(map[string]any)
	require.Equal(t, "zh-CN", lp["lang"])
	require.Equal(t, "ZH", lp["title"])
}

func TestUserCampaignService_GetLandingPageUI_success(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	now := time.Now()
	camp := publishedCampaign(now.Add(-time.Hour), now.Add(time.Hour))
	cm.On("GetByID", int64(1)).Return(camp, nil)
	lm := servicemock.NewMockLandingPageRepository(t)
	lm.On("GetByID", int64(10)).Return(&model.CampaignLandingPage{
		ID: 10, DefaultLang: "en-US", Title: "Hi {{threshold}} {{reward}}", BannerImageURL: "x",
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(5)).Return(nil, gorm.ErrRecordNotFound)

	svc := newTestUserCampaignService(t, cm, lm, nil, pm,
		servicemock.NewMockUserRepository(t), nil, nil,
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
	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil,
		servicemock.NewMockParticipantRepository(t),
		servicemock.NewMockUserRepository(t), nil, nil,
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

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil, pm, um, nil, nil,
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

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil, pm,
		servicemock.NewMockUserRepository(t), nil, nil,
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

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil, pm,
		servicemock.NewMockUserRepository(t), nil, nil,
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
	am := defaultRechargeMock(t)

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil, pm, um, am, nil,
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
	pm.On("Save", mock.Anything).Return(nil)
	am := defaultRechargeMock(t)
	rn := &servicemock.MockCampaignRewardNotifier{}
	rn.On("NotifyTopUpReward", mock.MatchedBy(func(e service.TopUpRewardEvent) bool {
		return e.ParticipantID == 55 && e.RewardAmount == 10
	}))

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil, pm, um, am, rn,
	)
	reply, err := svc.SimulateTopUp(1, 100, 120)
	require.NoError(t, err)
	require.Equal(t, data.CodeSuccess, reply.Code)
	d := reply.Data.(map[string]any)
	require.Equal(t, "TXN_TEST", d["rechargeTransactionNo"])
	rn.AssertExpectations(t)
}
