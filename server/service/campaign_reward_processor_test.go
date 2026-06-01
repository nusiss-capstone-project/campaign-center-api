package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/lianjin/campaign-center-api/server/event"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"github.com/lianjin/campaign-center-api/server/service"
	servicemock "github.com/lianjin/campaign-center-api/server/mock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubRewardTxRepo struct {
	commitErr error
	committed bool
}

func (r *stubRewardTxRepo) Create(*model.RewardTransaction) error { return nil }

func (r *stubRewardTxRepo) CommitGrantWithParticipant(
	*model.CampaignParticipant, *model.RewardTransaction,
) error {
	r.committed = true
	return r.commitErr
}

type stubPerfRepoForReward struct {
	incremented bool
	err         error
}

func (r *stubPerfRepoForReward) GetSummary(int64) (*mysql.CampaignPerformanceSummary, error) {
	return nil, nil
}

func (r *stubPerfRepoForReward) ListDaily(int64, time.Time, time.Time) ([]model.CampaignPerformanceDaily, error) {
	return nil, nil
}

func (r *stubPerfRepoForReward) IncrementRewardIssued(int64, time.Time, float64, string) error {
	r.incremented = true
	return r.err
}

func newRewardProcessor(
	t *testing.T,
	pm *servicemock.MockParticipantRepository,
	rt *stubRewardTxRepo,
	am service.AccountService,
	perf *stubPerfRepoForReward,
) *service.CampaignRewardProcessor {
	t.Helper()
	if rt == nil {
		rt = &stubRewardTxRepo{}
	}
	if perf == nil {
		perf = &stubPerfRepoForReward{}
	}
	return service.NewCampaignRewardProcessor(pm, rt, am, perf)
}

func TestCampaignRewardProcessor_HandleTopUpReward_manualReviewSkips(t *testing.T) {
	pm := servicemock.NewMockParticipantRepository(t)
	p := newRewardProcessor(t, pm, nil, nil, nil)

	err := p.HandleTopUpReward(event.TopUpRewardEvent{ManualReview: true})

	require.NoError(t, err)
	pm.AssertNotCalled(t, "GetByCampaignAndUser", mock.Anything, mock.Anything)
}

func TestCampaignRewardProcessor_HandleTopUpReward_alreadyGranted(t *testing.T) {
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(2)).Return(&model.CampaignParticipant{
		ID: 9, RewardStatus: model.RewardStatusGranted,
	}, nil)
	rt := &stubRewardTxRepo{}
	p := newRewardProcessor(t, pm, rt, nil, nil)

	err := p.HandleTopUpReward(event.TopUpRewardEvent{
		CampaignID: 1, UserID: 2, RewardAmount: 10, RewardType: model.RewardTypeBonusCredit,
	})

	require.NoError(t, err)
	require.False(t, rt.committed)
}

func TestCampaignRewardProcessor_HandleTopUpReward_success(t *testing.T) {
	pm := servicemock.NewMockParticipantRepository(t)
	participant := &model.CampaignParticipant{ID: 9, CampaignID: 1, UserID: 2}
	pm.On("GetByCampaignAndUser", int64(1), int64(2)).Return(participant, nil)

	rt := &stubRewardTxRepo{}
	perf := &stubPerfRepoForReward{}
	am := servicemock.NewMockAccountService(t)
	am.On("CreditCampaignReward", int64(2), int64(1), float64(10), model.DefaultCurrency).
		Return(&service.RechargeResult{TransactionNo: "TXN1", BalanceAfter: 110}, nil)

	p := newRewardProcessor(t, pm, rt, am, perf)
	err := p.HandleTopUpReward(event.TopUpRewardEvent{
		CampaignID: 1, UserID: 2, RewardAmount: 10, RewardType: model.RewardTypeBonusCredit,
	})

	require.NoError(t, err)
	require.True(t, rt.committed)
	require.True(t, perf.incremented)
	require.Equal(t, model.RewardStatusGranted, participant.RewardStatus)
	require.Equal(t, float64(10), participant.RewardAmount)
	am.AssertExpectations(t)
}

func TestCampaignRewardProcessor_HandleTopUpReward_participantLookupFails(t *testing.T) {
	lookupErr := errors.New("db down")
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(2)).Return(nil, lookupErr)
	p := newRewardProcessor(t, pm, nil, nil, nil)

	err := p.HandleTopUpReward(event.TopUpRewardEvent{CampaignID: 1, UserID: 2})

	require.ErrorIs(t, err, lookupErr)
}

func TestCampaignRewardProcessor_HandleTopUpReward_commitFails(t *testing.T) {
	commitErr := errors.New("commit failed")
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(2)).Return(&model.CampaignParticipant{ID: 9}, nil)
	rt := &stubRewardTxRepo{commitErr: commitErr}
	p := newRewardProcessor(t, pm, rt, nil, nil)

	err := p.HandleTopUpReward(event.TopUpRewardEvent{
		CampaignID: 1, UserID: 2, RewardAmount: 10, RewardType: model.RewardTypeBonusCredit,
	})

	require.ErrorIs(t, err, commitErr)
}

func TestCampaignRewardProcessor_HandleTopUpReward_creditFails(t *testing.T) {
	creditErr := errors.New("credit failed")
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(2)).Return(&model.CampaignParticipant{ID: 9}, nil)
	rt := &stubRewardTxRepo{}
	am := servicemock.NewMockAccountService(t)
	am.On("CreditCampaignReward", int64(2), int64(1), float64(10), model.DefaultCurrency).
		Return(nil, creditErr)
	p := newRewardProcessor(t, pm, rt, am, nil)

	err := p.HandleTopUpReward(event.TopUpRewardEvent{
		CampaignID: 1, UserID: 2, RewardAmount: 10, RewardType: model.RewardTypeBonusCredit,
	})

	require.ErrorIs(t, err, creditErr)
}

func TestCampaignRewardProcessor_HandleTopUpReward_incrementPerformanceFails(t *testing.T) {
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(2)).Return(&model.CampaignParticipant{ID: 9}, nil)
	rt := &stubRewardTxRepo{}
	perfErr := errors.New("stats unavailable")
	perf := &stubPerfRepoForReward{err: perfErr}
	am := servicemock.NewMockAccountService(t)
	am.On("CreditCampaignReward", int64(2), int64(1), float64(10), model.DefaultCurrency).
		Return(&service.RechargeResult{}, nil)
	p := newRewardProcessor(t, pm, rt, am, perf)

	err := p.HandleTopUpReward(event.TopUpRewardEvent{
		CampaignID: 1, UserID: 2, RewardAmount: 10, RewardType: model.RewardTypeBonusCredit,
	})

	require.ErrorIs(t, err, perfErr)
	require.True(t, rt.committed)
}

func TestNewCampaignRewardNotifierForTest_runsSync(t *testing.T) {
	pm := servicemock.NewMockParticipantRepository(t)
	pm.On("GetByCampaignAndUser", int64(1), int64(2)).Return(nil, gorm.ErrRecordNotFound)
	p := newRewardProcessor(t, pm, nil, nil, nil)
	n := service.NewCampaignRewardNotifierForTest(p)

	n.NotifyTopUpReward(event.TopUpRewardEvent{CampaignID: 1, UserID: 2})
	pm.AssertExpectations(t)
}
