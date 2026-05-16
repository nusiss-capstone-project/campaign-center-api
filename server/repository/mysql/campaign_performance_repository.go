package mysql

import (
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CampaignPerformanceSummary aggregates campaign metrics.
type CampaignPerformanceSummary struct {
	ParticipantCount   int64
	ParticipationCount int64
	RewardIssuedCount  int64
	RewardIssuedAmount float64
}

// CampaignPerformanceRepository reads and updates campaign performance stats.
type CampaignPerformanceRepository interface {
	GetSummary(campaignID int64) (*CampaignPerformanceSummary, error)
	ListDaily(campaignID int64, start, end time.Time) ([]model.CampaignPerformanceDaily, error)
	IncrementRewardIssued(campaignID int64, statDate time.Time, amount float64, currency string) error
}

type campaignPerformanceRepository struct{}

var (
	campaignPerformanceRepositoryOnce     sync.Once
	campaignPerformanceRepositoryInstance CampaignPerformanceRepository
)

// GetCampaignPerformanceRepository returns the singleton campaign performance repository.
func GetCampaignPerformanceRepository() CampaignPerformanceRepository {
	campaignPerformanceRepositoryOnce.Do(func() {
		campaignPerformanceRepositoryInstance = &campaignPerformanceRepository{}
	})
	return campaignPerformanceRepositoryInstance
}

func (r *campaignPerformanceRepository) db() (*gorm.DB, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	return DB, nil
}

func (r *campaignPerformanceRepository) GetSummary(campaignID int64) (*CampaignPerformanceSummary, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var out CampaignPerformanceSummary
	if err := db.Model(&model.CampaignParticipant{}).Where("campaign_id = ?", campaignID).
		Select("COUNT(DISTINCT user_id) AS participant_count, COUNT(*) AS participation_count").
		Scan(&out).Error; err != nil {
		return nil, err
	}
	var reward CampaignPerformanceSummary
	if err := db.Model(&model.CampaignParticipant{}).Where("campaign_id = ? AND reward_status = ?",
		campaignID, model.RewardStatusGranted).
		Select("COUNT(*) AS reward_issued_count, COALESCE(SUM(reward_amount), 0) AS reward_issued_amount").
		Scan(&reward).Error; err != nil {
		return nil, err
	}
	out.RewardIssuedCount = reward.RewardIssuedCount
	out.RewardIssuedAmount = reward.RewardIssuedAmount
	return &out, nil
}

func (r *campaignPerformanceRepository) ListDaily(campaignID int64, start, end time.Time) ([]model.CampaignPerformanceDaily, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var rows []model.CampaignPerformanceDaily
	err = db.Where("campaign_id = ? AND stat_date >= ? AND stat_date <= ?",
		campaignID, start, end).Order("stat_date ASC").Find(&rows).Error
	return rows, err
}

func (r *campaignPerformanceRepository) IncrementRewardIssued(
	campaignID int64, statDate time.Time, amount float64, currency string,
) error {
	db, err := r.db()
	if err != nil {
		return err
	}
	day := time.Date(statDate.Year(), statDate.Month(), statDate.Day(), 0, 0, 0, 0, statDate.Location())
	now := time.Now()
	row := model.CampaignPerformanceDaily{
		CampaignID: campaignID, StatDate: day, Currency: currency,
		RewardIssuedCount: 1, RewardIssuedAmount: amount,
		CreatedAt: now, UpdatedAt: now,
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "campaign_id"}, {Name: "stat_date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"reward_issued_count":  gorm.Expr("reward_issued_count + 1"),
			"reward_issued_amount": gorm.Expr("reward_issued_amount + ?", amount),
			"updated_at":           now,
		}),
	}).Create(&row).Error
}
