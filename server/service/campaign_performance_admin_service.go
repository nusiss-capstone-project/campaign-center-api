package service

import (
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// CampaignPerformanceAdminService admin campaign performance APIs.
type CampaignPerformanceAdminService interface {
	GetPerformanceSummary(campaignID int64) (map[string]any, error)
	ListDailyPerformance(campaignID int64, start, end time.Time) ([]map[string]any, error)
	ListParticipations(filter mysql.ParticipationListFilter) ([]map[string]any, int64, error)
}

type campaignPerformanceAdminService struct {
	campaigns    mysql.CampaignRepository
	performance  mysql.CampaignPerformanceRepository
	participants mysql.ParticipantRepository
}

var (
	campaignPerfAdminOnce sync.Once
	campaignPerfAdminInst CampaignPerformanceAdminService
)

// NewCampaignPerformanceAdminService builds the admin performance service (for tests).
func NewCampaignPerformanceAdminService(
	campaigns mysql.CampaignRepository,
	performance mysql.CampaignPerformanceRepository,
	participants mysql.ParticipantRepository,
) CampaignPerformanceAdminService {
	return &campaignPerformanceAdminService{
		campaigns: campaigns, performance: performance, participants: participants,
	}
}

// GetCampaignPerformanceAdminService returns the singleton admin performance service.
func GetCampaignPerformanceAdminService() CampaignPerformanceAdminService {
	campaignPerfAdminOnce.Do(func() {
		campaignPerfAdminInst = NewCampaignPerformanceAdminService(
			mysql.GetCampaignRepository(),
			mysql.GetCampaignPerformanceRepository(),
			mysql.GetParticipantRepository(),
		)
	})
	return campaignPerfAdminInst
}

func (s *campaignPerformanceAdminService) GetPerformanceSummary(campaignID int64) (map[string]any, error) {
	if _, err := s.campaigns.GetByID(campaignID); err != nil {
		return nil, err
	}
	sum, err := s.performance.GetSummary(campaignID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"campaignId":         campaignID,
		"participantCount":   sum.ParticipantCount,
		"participationCount": sum.ParticipationCount,
		"rewardIssuedCount":  sum.RewardIssuedCount,
		"rewardIssuedAmount": sum.RewardIssuedAmount,
		"currency":           model.DefaultCurrency,
	}, nil
}

func (s *campaignPerformanceAdminService) ListDailyPerformance(
	campaignID int64, start, end time.Time,
) ([]map[string]any, error) {
	if _, err := s.campaigns.GetByID(campaignID); err != nil {
		return nil, err
	}
	rows, err := s.performance.ListDaily(campaignID, start, end)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, map[string]any{
			"date":               row.StatDate.Format("2006-01-02"),
			"participantCount":   row.ParticipantCount,
			"participationCount": row.ParticipationCount,
			"rewardIssuedCount":  row.RewardIssuedCount,
			"rewardIssuedAmount": row.RewardIssuedAmount,
			"currency":           row.Currency,
		})
	}
	return items, nil
}

func (s *campaignPerformanceAdminService) ListParticipations(
	filter mysql.ParticipationListFilter,
) ([]map[string]any, int64, error) {
	if _, err := s.campaigns.GetByID(filter.CampaignID); err != nil {
		return nil, 0, err
	}
	rows, total, err := s.participants.ListByCampaign(filter)
	if err != nil {
		return nil, 0, err
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, participationItem(row))
	}
	return items, total, nil
}

func participationItem(row model.CampaignParticipant) map[string]any {
	item := map[string]any{
		"participationId": row.ID,
		"campaignId":      row.CampaignID,
		"userId":          row.UserID,
		"joinAt":          row.JoinedAt.Format(time.RFC3339),
		"rewardAmount":    row.RewardAmount,
		"rewardStatus":    row.RewardStatus,
		"failureReason":   nil,
	}
	return item
}
