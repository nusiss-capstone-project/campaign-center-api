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
	servicemock "github.com/lianjin/campaign-center-api/server/mock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type noopRewardNotifier struct{}

func (noopRewardNotifier) NotifyTopUpReward(service.TopUpRewardEvent) {}

type staticLandingPageTranslationRepo struct {
	row *model.CampaignLandingPageTranslation
}

func (r staticLandingPageTranslationRepo) GetByLandingPageAndLang(
	landingPageID int64, lang string,
) (*model.CampaignLandingPageTranslation, error) {
	if r.row == nil || r.row.LandingPageID != landingPageID || r.row.Lang != lang {
		return nil, nil
	}
	return r.row, nil
}

func (staticLandingPageTranslationRepo) ListLangsByLandingPageID(int64) ([]string, error) {
	return []string{}, nil
}

func (staticLandingPageTranslationRepo) Upsert(*model.CampaignLandingPageTranslation) error {
	return nil
}

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
	translationSvc := service.NewLandingPageTranslationService(lm, trans, nil)
	if am == nil {
		am = &servicemock.MockAccountService{}
	}
	if rn == nil {
		rn = noopRewardNotifier{}
	}
	return service.NewUserCampaignService(cm, lm, translationSvc, pm, um, am, rn)
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

func TestUserCampaignService_GetLandingPageUI_usesTranslationService(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{
		ID: 1, LandingPageID: 10, RewardRules: rewardRulesJSON(t),
	}, nil)
	lm := servicemock.NewMockLandingPageRepository(t)
	lm.On("GetByID", int64(10)).Return(&model.CampaignLandingPage{
		ID: 10, DefaultLang: "en", Title: "Default {{threshold}}", Description: "d", Terms: "t",
	}, nil)
	trans := staticLandingPageTranslationRepo{row: &model.CampaignLandingPageTranslation{
		LandingPageID: 10,
		Lang:          "ja",
		Title:         "JA {{threshold}} {{reward}}",
		Description:   "説明",
		Terms:         "条件",
	}}
	svc := newTestUserCampaignService(t, cm, lm, trans,
		servicemock.NewMockParticipantRepository(t),
		servicemock.NewMockUserRepository(t), nil, nil,
	)

	reply, err := svc.GetLandingPageUI(1, 0, "ja")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, reply.HTTPStatus)
	dataMap := reply.Data.(map[string]any)
	lp := dataMap["landingPage"].(map[string]any)
	require.Equal(t, "ja", lp["lang"])
	require.Equal(t, "JA 100 10", lp["title"])
	require.Equal(t, "説明", lp["description"])
	require.Equal(t, "条件", lp["terms"])
}

func TestUserCampaignService_GetLandingPageUI_replacesLandingTemplateVariables(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{
		ID: 1, LandingPageID: 10,
		CampaignStartTime: time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC),
		CampaignEndTime:   time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		RewardRules: `{"topupThreshold":100,"rewardType":"TOKEN_BONUS","rewardAmount":12.5,` +
			`"rewardCurrency":"USDT","rewardMode":"PERCENTAGE","rewardPercentage":8.5,` +
			`"maxRewardAmount":50,"maxClaimPerUser":2,"minObtainDays":7}`,
	}, nil)
	lm := servicemock.NewMockLandingPageRepository(t)
	lm.On("GetByID", int64(10)).Return(&model.CampaignLandingPage{
		ID:          10,
		DefaultLang: "en-US",
		Title:       "Deposit {{topupThreshold}} {{rewardCurrency}}",
		Description: "Get {{rewardPercentage}}% after {{minObtainDays}} days",
		Terms: "{{campaignStartDate}}-{{campaignEndDate}} {{rewardType}} " +
			"{{rewardMode}} {{rewardAmount}} {{maxRewardAmount}} {{maxClaimPerUser}}",
	}, nil)
	svc := newTestUserCampaignService(t, cm, lm, nil,
		servicemock.NewMockParticipantRepository(t),
		servicemock.NewMockUserRepository(t), nil, nil,
	)

	reply, err := svc.GetLandingPageUI(1, 0, "en-US")
	require.NoError(t, err)
	lp := reply.Data.(map[string]any)["landingPage"].(map[string]any)
	require.Equal(t, "Deposit 100 USDT", lp["title"])
	require.Equal(t, "Get 8.5% after 7 days", lp["description"])
	require.Equal(t, "2026-05-20-2026-06-01 TOKEN_BONUS PERCENTAGE 12.5 50 2", lp["terms"])
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
	um := servicemock.NewMockUserRepository(t)
	um.On("GetByID", int64(5)).Return(&model.User{
		ID: 5, KYCStatus: model.KYCStatusPassed, Segment: model.UserSegmentNewUser, Market: "US",
	}, nil)

	svc := newTestUserCampaignService(t, cm, lm, nil, pm,
		um, nil, nil,
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
	pm.On("Save", mock.MatchedBy(func(p *model.CampaignParticipant) bool {
		return p.RiskStatus == model.RiskStatusApproved &&
			p.RewardStatus == model.RewardStatusPending &&
			p.RewardAmount == 10
	})).Return(nil)
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
	require.Equal(t, model.RewardStatusPending, d["rewardStatus"])
	require.Equal(t, "reward processing", reply.Message)
	rn.AssertExpectations(t)
}

func TestUserCampaignService_SimulateTopUp_percentageRewardCapped(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{
		ID: 1,
		RewardRules: `{"topupThreshold":100,"rewardType":"TOKEN_BONUS",` +
			`"rewardMode":"PERCENTAGE","rewardPercentage":10,"maxRewardAmount":15,"maxClaimPerUser":1}`,
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(100)).Return(&model.CampaignParticipant{
		ID: 55, RewardStatus: model.RewardStatusNotGranted,
	}, nil)
	um := servicemock.NewMockUserRepository(t)
	um.On("GetByID", int64(100)).Return(&model.User{RiskLevel: "LOW"}, nil)
	pm.On("Save", mock.MatchedBy(func(p *model.CampaignParticipant) bool {
		return p.RewardStatus == model.RewardStatusPending && p.RewardAmount == 15
	})).Return(nil)
	am := defaultRechargeMock(t)
	rn := &servicemock.MockCampaignRewardNotifier{}
	rn.On("NotifyTopUpReward", mock.MatchedBy(func(e service.TopUpRewardEvent) bool {
		return e.ParticipantID == 55 && e.RewardAmount == 15
	}))

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil, pm, um, am, rn,
	)
	reply, err := svc.SimulateTopUp(1, 100, 200)
	require.NoError(t, err)

	d := reply.Data.(map[string]any)
	require.Equal(t, float64(15), d["rewardAmount"])
	require.Equal(t, model.RewardStatusPending, d["rewardStatus"])
	rn.AssertExpectations(t)
}

func TestUserCampaignService_SimulateTopUp_fixedRewardCapped(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{
		ID: 1,
		RewardRules: `{"topupThreshold":100,"rewardType":"TOKEN_BONUS",` +
			`"rewardMode":"FIXED_AMOUNT","rewardAmount":30,"maxRewardAmount":20,"maxClaimPerUser":1}`,
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(100)).Return(&model.CampaignParticipant{
		ID: 55, RewardStatus: model.RewardStatusNotGranted,
	}, nil)
	um := servicemock.NewMockUserRepository(t)
	um.On("GetByID", int64(100)).Return(&model.User{RiskLevel: "LOW"}, nil)
	pm.On("Save", mock.MatchedBy(func(p *model.CampaignParticipant) bool {
		return p.RewardStatus == model.RewardStatusPending && p.RewardAmount == 20
	})).Return(nil)
	am := defaultRechargeMock(t)
	rn := &servicemock.MockCampaignRewardNotifier{}
	rn.On("NotifyTopUpReward", mock.MatchedBy(func(e service.TopUpRewardEvent) bool {
		return e.ParticipantID == 55 && e.RewardAmount == 20
	}))

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil, pm, um, am, rn,
	)
	reply, err := svc.SimulateTopUp(1, 100, 200)
	require.NoError(t, err)

	d := reply.Data.(map[string]any)
	require.Equal(t, float64(20), d["rewardAmount"])
	require.Equal(t, model.RewardStatusPending, d["rewardStatus"])
	rn.AssertExpectations(t)
}

func TestUserCampaignService_SimulateTopUp_pendingRewardNotReenqueued(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{
		ID: 1, RewardRules: `{"topupThreshold":100,"rewardAmount":10,"rewardType":"BONUS_CREDIT","maxClaimPerUser":1}`,
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(100)).Return(&model.CampaignParticipant{
		ID: 55, RewardStatus: model.RewardStatusPending,
	}, nil)

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil, pm,
		servicemock.NewMockUserRepository(t), nil, nil,
	)
	reply, err := svc.SimulateTopUp(1, 100, 120)
	require.NoError(t, err)
	require.Equal(t, data.CodeDuplicateReward, reply.Code)
	require.Equal(t, "Reward already processing", reply.Message)
}

func TestUserCampaignService_SimulateTopUp_invalidRewardModeBeforeRecharge(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{
		ID: 1,
		RewardRules: `{"topupThreshold":100,"rewardType":"TOKEN_BONUS",` +
			`"rewardMode":"UNKNOWN","rewardAmount":30,"maxClaimPerUser":1}`,
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(100)).Return(&model.CampaignParticipant{
		ID: 55, RewardStatus: model.RewardStatusNotGranted,
	}, nil)
	um := servicemock.NewMockUserRepository(t)
	um.On("GetByID", int64(100)).Return(&model.User{RiskLevel: "LOW"}, nil)
	am := &servicemock.MockAccountService{}

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil, pm, um, am, nil,
	)
	reply, err := svc.SimulateTopUp(1, 100, 120)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, reply.HTTPStatus)
	require.Equal(t, "invalid rewardMode: UNKNOWN", reply.Message)
}

func TestUserCampaignService_ListAvailableCampaigns_groupsOngoingAndUpcoming(t *testing.T) {
	now := time.Now()
	ongoing := model.Campaign{
		ID: 1, Name: "Ongoing",
		CampaignStartTime: now.Add(-time.Hour), CampaignEndTime: now.Add(time.Hour),
		TargetMarket: model.MarketUS, TargetUserSegment: model.UserSegmentNewUser,
	}
	upcoming := model.Campaign{
		ID: 2, Name: "Upcoming",
		CampaignStartTime: now.Add(time.Hour), CampaignEndTime: now.Add(2 * time.Hour),
		TargetMarket: model.MarketUS, TargetUserSegment: model.UserSegmentNewUser,
	}
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("ListPublishedActiveOrUpcoming", mock.AnythingOfType("time.Time")).
		Return([]model.Campaign{ongoing, upcoming}, nil)
	um := servicemock.NewMockUserRepository(t)
	um.On("GetByID", int64(100)).Return(&model.User{
		ID: 100, KYCStatus: model.KYCStatusPassed, Segment: model.UserSegmentNewUser, Market: model.MarketUS,
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("ListByUserAndCampaignIDs", int64(100), []int64{1, 2}).
		Return([]model.CampaignParticipant{{CampaignID: 1, JoinStatus: model.JoinStatusJoined}}, nil)

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil, pm,
		um, nil, nil,
	)
	reply, err := svc.ListAvailableCampaigns(100)
	require.NoError(t, err)
	require.Equal(t, data.CodeSuccess, reply.Code)

	payload := reply.Data.(map[string]any)
	ongoingItems := payload["ongoing"].([]map[string]any)
	upcomingItems := payload["upcoming"].([]map[string]any)
	require.Len(t, ongoingItems, 1)
	require.Len(t, upcomingItems, 1)
	require.Equal(t, int64(1), ongoingItems[0]["id"])
	require.Equal(t, "Ongoing", ongoingItems[0]["name"])
	require.Equal(t, true, ongoingItems[0]["joined"])
	require.Equal(t, int64(2), upcomingItems[0]["id"])
	require.NotContains(t, upcomingItems[0], "joined")
}

func TestUserCampaignService_ListAvailableCampaigns_filtersIneligibleCampaigns(t *testing.T) {
	now := time.Now()
	eligible := model.Campaign{
		ID: 1, Name: "Eligible",
		CampaignStartTime: now.Add(-time.Hour), CampaignEndTime: now.Add(time.Hour),
		TargetMarket: model.MarketUS, TargetUserSegment: model.UserSegmentNewUser,
	}
	segmentMismatch := model.Campaign{
		ID: 2, Name: "VIP",
		CampaignStartTime: now.Add(-time.Hour), CampaignEndTime: now.Add(time.Hour),
		TargetMarket: model.MarketUS, TargetUserSegment: model.UserSegmentVIPUser,
	}
	public := model.Campaign{
		ID: 3, Name: "Public",
		CampaignStartTime: now.Add(time.Hour), CampaignEndTime: now.Add(2 * time.Hour),
		TargetMarket: model.MarketGlobal, TargetUserSegment: model.UserSegmentAllUsers,
	}
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("ListPublishedActiveOrUpcoming", mock.AnythingOfType("time.Time")).
		Return([]model.Campaign{eligible, segmentMismatch, public}, nil)
	um := servicemock.NewMockUserRepository(t)
	um.On("GetByID", int64(100)).Return(&model.User{
		ID: 100, KYCStatus: model.KYCStatusPassed, Segment: model.UserSegmentNewUser, Market: model.MarketUS,
	}, nil)
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("ListByUserAndCampaignIDs", int64(100), []int64{1, 3}).
		Return([]model.CampaignParticipant{}, nil)

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil, pm, um, nil, nil,
	)
	reply, err := svc.ListAvailableCampaigns(100)
	require.NoError(t, err)

	payload := reply.Data.(map[string]any)
	ongoingItems := payload["ongoing"].([]map[string]any)
	upcomingItems := payload["upcoming"].([]map[string]any)
	require.Len(t, ongoingItems, 1)
	require.Len(t, upcomingItems, 1)
	require.Equal(t, int64(1), ongoingItems[0]["id"])
	require.Equal(t, int64(3), upcomingItems[0]["id"])
}

func TestUserCampaignService_GetLandingPageUI_hidesIneligibleCampaign(t *testing.T) {
	now := time.Now()
	camp := publishedCampaign(now.Add(-time.Hour), now.Add(time.Hour))
	camp.TargetUserSegment = model.UserSegmentVIPUser
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(camp, nil)
	um := servicemock.NewMockUserRepository(t)
	um.On("GetByID", int64(100)).Return(&model.User{
		ID: 100, KYCStatus: model.KYCStatusPassed, Segment: model.UserSegmentNewUser, Market: model.MarketUS,
	}, nil)

	svc := newTestUserCampaignService(t, cm,
		servicemock.NewMockLandingPageRepository(t), nil,
		servicemock.NewMockParticipantRepository(t), um, nil, nil,
	)
	reply, err := svc.GetLandingPageUI(1, 100, "en-US")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, reply.HTTPStatus)
	require.Equal(t, "campaign not found", reply.Message)
}
