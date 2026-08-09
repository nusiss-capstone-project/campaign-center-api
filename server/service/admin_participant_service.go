package service

import (
	"sync"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

const defaultParticipantRiskLevel = "LOW"

// AdminParticipantService admin participant list / detail.
type AdminParticipantService interface {
	ListParticipants(campaignID int64) (*data.AdminParticipantListData, error)
	GetParticipant(campaignID, userID int64) (*data.AdminParticipantVO, error)
}

type adminParticipantService struct {
	campaigns    mysql.CampaignRepository
	rules        mysql.CampaignRewardRuleRepository
	participants mysql.ParticipantRepository
}

var (
	adminParticipantOnce sync.Once
	adminParticipantInst AdminParticipantService
)

// NewAdminParticipantService wires repositories (for tests).
func NewAdminParticipantService(
	campaigns mysql.CampaignRepository,
	rules mysql.CampaignRewardRuleRepository,
	participants mysql.ParticipantRepository,
) AdminParticipantService {
	return &adminParticipantService{
		campaigns:    campaigns,
		rules:        rules,
		participants: participants,
	}
}

// GetAdminParticipantService returns the singleton.
func GetAdminParticipantService() AdminParticipantService {
	adminParticipantOnce.Do(func() {
		adminParticipantInst = NewAdminParticipantService(
			mysql.GetCampaignRepository(),
			mysql.GetCampaignRewardRuleRepository(),
			mysql.GetParticipantRepository(),
		)
	})
	return adminParticipantInst
}

func (s *adminParticipantService) ListParticipants(campaignID int64) (*data.AdminParticipantListData, error) {
	campaign, err := s.requireCampaign(campaignID)
	if err != nil {
		return nil, err
	}
	if campaign.Status != model.CampaignStatusPublished {
		return &data.AdminParticipantListData{
			Campaign: data.AdminParticipantCampaignVO{
				ID:   campaign.ID,
				Name: campaign.Name,
			},
			Participants: []data.AdminParticipantVO{},
		}, nil
	}
	taskGroupID, err := s.resolveTaskGroupID(campaignID)
	if err != nil {
		return nil, err
	}
	rows, err := s.participants.ListByCampaignID(campaignID)
	if err != nil {
		return nil, err
	}
	items := make([]data.AdminParticipantVO, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAdminParticipantVO(row))
	}
	return &data.AdminParticipantListData{
		Campaign: data.AdminParticipantCampaignVO{
			ID:          campaign.ID,
			Name:        campaign.Name,
			TaskGroupID: taskGroupID,
			ProjectID:   campaign.BudgetProjectID,
		},
		Participants: items,
	}, nil
}

func (s *adminParticipantService) GetParticipant(campaignID, userID int64) (*data.AdminParticipantVO, error) {
	if _, err := s.requireCampaign(campaignID); err != nil {
		return nil, err
	}
	row, err := s.participants.GetByCampaignAndUser(campaignID, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, gorm.ErrRecordNotFound
	}
	vo := toAdminParticipantVO(*row)
	return &vo, nil
}

func (s *adminParticipantService) requireCampaign(campaignID int64) (*model.Campaign, error) {
	campaign, err := s.campaigns.GetByID(campaignID)
	if err != nil {
		return nil, err
	}
	return campaign, nil
}

func (s *adminParticipantService) resolveTaskGroupID(campaignID int64) (int64, error) {
	rules, err := s.rules.ListByCampaignID(campaignID)
	if err != nil {
		return 0, err
	}
	for _, rule := range rules {
		if rule.RefClient == model.RewardRefClientTaskGroup && rule.RefID > 0 {
			return rule.RefID, nil
		}
	}
	return 0, nil
}

func toAdminParticipantVO(row model.CampaignParticipant) data.AdminParticipantVO {
	return data.AdminParticipantVO{
		UserID:    row.UserID,
		JoinedAt:  timeToUnix(row.JoinedAt),
		RiskLevel: defaultParticipantRiskLevel,
	}
}
