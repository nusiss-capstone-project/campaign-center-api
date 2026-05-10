package service

import (
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// CampaignAdminService admin campaign operations.
type CampaignAdminService interface {
	CreateCampaign(p CreateCampaignParams) (campaignID int64, status int16, err error)
	UpdateDraftCampaign(id int64, p UpdateCampaignParams) error
	ListCampaigns(filter mysql.CampaignListFilter) ([]model.Campaign, int64, error)
	GetCampaign(id int64) (*model.Campaign, error)
	PublishCampaign(id int64, operator string) (*model.Campaign, error)
}

// CreateCampaignParams input for creating a campaign.
type CreateCampaignParams struct {
	Name                  string
	Type                  string
	TargetMarket          string
	RegistrationStartTime time.Time
	RegistrationEndTime   time.Time
	CampaignStartTime     time.Time
	CampaignEndTime       time.Time
	TargetUserSegment     string
	RewardRules           model.RewardRulesPayload
	LandingPageID         int64
}

// UpdateCampaignParams input for updating a draft campaign.
type UpdateCampaignParams struct {
	Name                  string
	TargetMarket          string
	RegistrationStartTime time.Time
	RegistrationEndTime   time.Time
	CampaignStartTime     time.Time
	CampaignEndTime       time.Time
	TargetUserSegment     string
	RewardRules           model.RewardRulesPayload
	LandingPageID         int64
}

type campaignAdminService struct {
	campaigns mysql.CampaignRepository
}

var (
	campaignAdminServiceOnce   sync.Once
	campaignAdminServiceInst CampaignAdminService
)

// NewCampaignAdminService builds a campaign admin service with explicit repositories (for tests).
func NewCampaignAdminService(campaigns mysql.CampaignRepository) CampaignAdminService {
	return &campaignAdminService{campaigns: campaigns}
}

// GetCampaignAdminService returns the singleton campaign admin service.
func GetCampaignAdminService() CampaignAdminService {
	campaignAdminServiceOnce.Do(func() {
		campaignAdminServiceInst = NewCampaignAdminService(mysql.GetCampaignRepository())
	})
	return campaignAdminServiceInst
}

func (s *campaignAdminService) CreateCampaign(p CreateCampaignParams) (int64, int16, error) {
	rulesJSON, err := model.MarshalRewardRulesPayload(p.RewardRules)
	if err != nil {
		return 0, 0, err
	}
	now := time.Now()
	campaign := model.Campaign{
		Name:                  p.Name,
		Type:                  p.Type,
		TargetMarket:          p.TargetMarket,
		RegistrationStartTime: p.RegistrationStartTime,
		RegistrationEndTime:   p.RegistrationEndTime,
		CampaignStartTime:     p.CampaignStartTime,
		CampaignEndTime:       p.CampaignEndTime,
		TargetUserSegment:     p.TargetUserSegment,
		RewardRules:           rulesJSON,
		Status:                model.CampaignStatusDraft,
		CreatedAt:             now,
		UpdatedAt:             now,
		LandingPageID:         p.LandingPageID,
	}
	if err := s.campaigns.Create(&campaign); err != nil {
		return 0, 0, err
	}
	return campaign.ID, campaign.Status, nil
}

func (s *campaignAdminService) UpdateDraftCampaign(id int64, p UpdateCampaignParams) error {
	existing, err := s.campaigns.GetByID(id)
	if err != nil {
		return err
	}
	if existing.Status != model.CampaignStatusDraft {
		return errCampaignNotDraft
	}
	rulesJSON, err := model.MarshalRewardRulesPayload(p.RewardRules)
	if err != nil {
		return err
	}
	now := time.Now()
	existing.Name = p.Name
	existing.TargetMarket = p.TargetMarket
	existing.RegistrationStartTime = p.RegistrationStartTime
	existing.RegistrationEndTime = p.RegistrationEndTime
	existing.CampaignStartTime = p.CampaignStartTime
	existing.CampaignEndTime = p.CampaignEndTime
	existing.TargetUserSegment = p.TargetUserSegment
	existing.RewardRules = rulesJSON
	existing.LandingPageID = p.LandingPageID
	existing.UpdatedAt = now
	return s.campaigns.Update(existing)
}

func (s *campaignAdminService) ListCampaigns(filter mysql.CampaignListFilter) ([]model.Campaign, int64, error) {
	return s.campaigns.List(filter)
}

func (s *campaignAdminService) GetCampaign(id int64) (*model.Campaign, error) {
	return s.campaigns.GetByID(id)
}

func (s *campaignAdminService) PublishCampaign(id int64, operator string) (*model.Campaign, error) {
	return s.campaigns.Publish(id, operator)
}
