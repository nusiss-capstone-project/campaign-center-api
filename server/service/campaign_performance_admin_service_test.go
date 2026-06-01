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

type stubPerformanceRepo struct{}

func (stubPerformanceRepo) GetSummary(int64) (*mysql.CampaignPerformanceSummary, error) {
	return &mysql.CampaignPerformanceSummary{}, nil
}
func (stubPerformanceRepo) ListDaily(int64, time.Time, time.Time) ([]model.CampaignPerformanceDaily, error) {
	return nil, nil
}
func (stubPerformanceRepo) IncrementRewardIssued(int64, time.Time, float64, string) error {
	return nil
}
