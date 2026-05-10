package mysql

import (
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// ParticipantRepository persists campaign participants.
type ParticipantRepository interface {
	GetByCampaignAndUser(campaignID, userID int64) (*model.CampaignParticipant, error)
	Create(p *model.CampaignParticipant) error
	Save(p *model.CampaignParticipant) error
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
	var p model.CampaignParticipant
	if err := db.Where("campaign_id = ? AND user_id = ?", campaignID, userID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *participantRepository) Create(p *model.CampaignParticipant) error {
	db, err := r.db()
	if err != nil {
		return err
	}
	return db.Create(p).Error
}

func (r *participantRepository) Save(p *model.CampaignParticipant) error {
	db, err := r.db()
	if err != nil {
		return err
	}
	p.UpdatedAt = time.Now()
	return db.Save(p).Error
}
