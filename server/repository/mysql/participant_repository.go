package mysql

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ParticipantRepository persists campaign join records.
type ParticipantRepository interface {
	GetByCampaignAndUser(campaignID, userID int64) (*model.CampaignParticipant, error)
	ListJoinedCampaignIDs(userID int64, campaignIDs []int64) (map[int64]struct{}, error)
	Join(ctx context.Context, campaignID, userID int64) (*model.CampaignParticipant, error)
}

type participantRepository struct{}

var (
	participantRepositoryOnce     sync.Once
	participantRepositoryInstance ParticipantRepository
)

// GetParticipantRepository returns the singleton participant repository.
func GetParticipantRepository() ParticipantRepository {
	participantRepositoryOnce.Do(func() {
		participantRepositoryInstance = &participantRepository{}
	})
	return participantRepositoryInstance
}

func (r *participantRepository) db() (*gorm.DB, error) {
	if DB == nil {
		return nil, ErrDatabaseDisabled
	}
	return DB, nil
}

func (r *participantRepository) GetByCampaignAndUser(campaignID, userID int64) (*model.CampaignParticipant, error) {
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var row model.CampaignParticipant
	err = db.Where("campaign_id = ? AND user_id = ?", campaignID, userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *participantRepository) ListJoinedCampaignIDs(userID int64, campaignIDs []int64) (map[int64]struct{}, error) {
	out := make(map[int64]struct{})
	if len(campaignIDs) == 0 {
		return out, nil
	}
	db, err := r.db()
	if err != nil {
		return nil, err
	}
	var ids []int64
	if err := db.Model(&model.CampaignParticipant{}).
		Where("user_id = ? AND campaign_id IN ?", userID, campaignIDs).
		Pluck("campaign_id", &ids).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		out[id] = struct{}{}
	}
	return out, nil
}

func (r *participantRepository) Join(ctx context.Context, campaignID, userID int64) (*model.CampaignParticipant, error) {
	db, err := session(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	row := &model.CampaignParticipant{
		CampaignID: campaignID,
		UserID:     userID,
		JoinedAt:   now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "campaign_id"}, {Name: "user_id"}},
		DoNothing: true,
	}).Create(row).Error
	if err != nil {
		return nil, err
	}
	if row.ID != 0 {
		return row, nil
	}
	var existing model.CampaignParticipant
	if err := db.Where("campaign_id = ? AND user_id = ?", campaignID, userID).First(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}
