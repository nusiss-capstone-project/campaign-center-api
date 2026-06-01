package service_test

import (
	"testing"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"github.com/lianjin/campaign-center-api/server/service"
	servicemock "github.com/lianjin/campaign-center-api/server/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCampaignPerformanceAdmin_GetSummary_notFound(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(99)).Return(nil, gorm.ErrRecordNotFound)
	svc := service.NewCampaignPerformanceAdminService(cm,
		&stubPerformanceRepo{}, servicemock.NewMockParticipantRepository(t),
	)
	_, err := svc.GetPerformanceSummary(99)
	require.Error(t, err)
}

type stubPerformanceRepo struct {
	summary *mysql.CampaignPerformanceSummary
	daily   []model.CampaignPerformanceDaily
}

func (r stubPerformanceRepo) GetSummary(int64) (*mysql.CampaignPerformanceSummary, error) {
	if r.summary != nil {
		return r.summary, nil
	}
	return &mysql.CampaignPerformanceSummary{}, nil
}

func (r stubPerformanceRepo) ListDaily(int64, time.Time, time.Time) ([]model.CampaignPerformanceDaily, error) {
	return r.daily, nil
}
func (stubPerformanceRepo) IncrementRewardIssued(int64, time.Time, float64, string) error {
	return nil
}

func TestCampaignPerformanceAdmin_GetSummary_success(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{ID: 1}, nil)
	perf := stubPerformanceRepo{}
	perf.summary = &mysql.CampaignPerformanceSummary{
		ParticipantCount: 10, ParticipationCount: 20,
		RewardIssuedCount: 3, RewardIssuedAmount: 99.5,
	}
	svc := service.NewCampaignPerformanceAdminService(cm, &perf,
		servicemock.NewMockParticipantRepository(t))

	out, err := svc.GetPerformanceSummary(1)

	require.NoError(t, err)
	require.Equal(t, int64(1), out["campaignId"])
	require.Equal(t, int64(10), out["participantCount"])
	require.Equal(t, int64(20), out["participationCount"])
	require.Equal(t, int64(3), out["rewardIssuedCount"])
	require.Equal(t, 99.5, out["rewardIssuedAmount"])
	require.Equal(t, model.DefaultCurrency, out["currency"])
}

func TestCampaignPerformanceAdmin_ListDailyPerformance_success(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{ID: 1}, nil)
	statDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	perf := stubPerformanceRepo{}
	perf.daily = []model.CampaignPerformanceDaily{{
		StatDate: statDate, ParticipantCount: 1, ParticipationCount: 2,
		RewardIssuedCount: 1, RewardIssuedAmount: 5, Currency: "USD",
	}}
	svc := service.NewCampaignPerformanceAdminService(cm, &perf,
		servicemock.NewMockParticipantRepository(t))
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC)

	items, err := svc.ListDailyPerformance(1, start, end)

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "2026-05-01", items[0]["date"])
	require.Equal(t, int64(1), items[0]["participantCount"])
	require.Equal(t, "USD", items[0]["currency"])
}

func TestCampaignPerformanceAdmin_ListParticipations_success(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(1)).Return(&model.Campaign{ID: 1}, nil)
	joinedAt := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	pm := servicemock.NewMockParticipantRepository(t)
	filter := mysql.ParticipationListFilter{CampaignID: 1, Page: 1, PageSize: 10}
	pm.On("ListByCampaign", filter).Return([]model.CampaignParticipant{{
		ID: 7, CampaignID: 1, UserID: 100, JoinedAt: joinedAt,
		RewardAmount: 10, RewardStatus: model.RewardStatusGranted,
	}}, int64(1), nil)
	svc := service.NewCampaignPerformanceAdminService(cm, &stubPerformanceRepo{}, pm)

	items, total, err := svc.ListParticipations(filter)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, int64(7), items[0]["participationId"])
	require.Equal(t, int64(100), items[0]["userId"])
	require.Equal(t, joinedAt.Format(time.RFC3339), items[0]["joinAt"])
	require.Equal(t, float64(10), items[0]["rewardAmount"])
}

func TestCampaignPerformanceAdmin_ListDailyPerformance_campaignNotFound(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(99)).Return(nil, gorm.ErrRecordNotFound)
	svc := service.NewCampaignPerformanceAdminService(cm, &stubPerformanceRepo{},
		servicemock.NewMockParticipantRepository(t))

	_, err := svc.ListDailyPerformance(99, time.Now(), time.Now())

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCampaignPerformanceAdmin_ListParticipations_campaignNotFound(t *testing.T) {
	cm := servicemock.NewMockCampaignRepository(t)
	cm.On("GetByID", int64(99)).Return(nil, gorm.ErrRecordNotFound)
	svc := service.NewCampaignPerformanceAdminService(cm, &stubPerformanceRepo{},
		servicemock.NewMockParticipantRepository(t))

	_, _, err := svc.ListParticipations(mysql.ParticipationListFilter{CampaignID: 99})

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
