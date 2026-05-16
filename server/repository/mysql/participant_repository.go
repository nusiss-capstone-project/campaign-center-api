package mysql

import (
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

// ParticipationListFilter filters admin participation list.
type ParticipationListFilter struct {
	CampaignID   int64
	UserID       *int64
	RewardStatus string
	Page         int
	PageSize     int
}

// ParticipantRepository persists campaign participants.
type ParticipantRepository interface {
	GetByCampaignAndUser(campaignID, userID int64) (*model.CampaignParticipant, error)
	Create(p *model.CampaignParticipant) error
	Save(p *model.CampaignParticipant) error
	ListByCampaign(filter ParticipationListFilter) ([]model.CampaignParticipant, int64, error)
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

func (r *participantRepository) ListByCampaign(filter ParticipationListFilter) ([]model.CampaignParticipant, int64, error) {
	db, err := r.db()
	if err != nil {
		return nil, 0, err
	}
	q := db.Model(&model.CampaignParticipant{}).Where("campaign_id = ?", filter.CampaignID)
	if filter.UserID != nil {
		q = q.Where("user_id = ?", *filter.UserID)
	}
	if filter.RewardStatus != "" {
		q = q.Where("reward_status = ?", filter.RewardStatus)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize := filter.Page, filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	var rows []model.CampaignParticipant
	if err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
