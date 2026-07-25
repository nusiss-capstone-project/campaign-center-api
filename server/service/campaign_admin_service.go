package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
)

// CampaignAdminService admin campaign operations (v2 drafts/versions).
type CampaignAdminService interface {
	CreateCampaign(ctx context.Context, name string) (campaignID int64, status int16, err error)
	CreateVersion(ctx context.Context, campaignID int64) (version int, err error)
	EditVersion(ctx context.Context, campaignID int64, version int, campaign data.CampaignVO) error
	PublishCampaign(ctx context.Context, campaignID int64, operator string) (*data.CampaignVO, error)
	ListCampaigns(req data.CampaignListReq) ([]data.CampaignListVO, int64, error)
	GetCampaign(campaignID int64) (*data.CampaignVO, error)
}

type campaignAdminService struct {
	campaigns mysql.CampaignRepository
	drafts    mysql.CampaignDraftRepository
	rules     mysql.CampaignRewardRuleRepository
}

var (
	campaignAdminServiceOnce sync.Once
	campaignAdminServiceInst CampaignAdminService
)

// NewCampaignAdminService builds a campaign admin service with explicit repositories (for tests).
func NewCampaignAdminService(
	campaigns mysql.CampaignRepository,
	drafts mysql.CampaignDraftRepository,
	rules mysql.CampaignRewardRuleRepository,
) CampaignAdminService {
	return &campaignAdminService{campaigns: campaigns, drafts: drafts, rules: rules}
}

// GetCampaignAdminService returns the singleton campaign admin service.
func GetCampaignAdminService() CampaignAdminService {
	campaignAdminServiceOnce.Do(func() {
		campaignAdminServiceInst = NewCampaignAdminService(
			mysql.GetCampaignRepository(),
			mysql.GetCampaignDraftRepository(),
			mysql.GetCampaignRewardRuleRepository(),
		)
	})
	return campaignAdminServiceInst
}

func (s *campaignAdminService) CreateCampaign(ctx context.Context, name string) (int64, int16, error) {
	name = trimCampaignName(name)
	if name == "" {
		return 0, 0, fmt.Errorf("%s", MsgCampaignNameRequired)
	}
	now := time.Now()
	campaign := model.Campaign{
		Name:      name,
		Status:    model.CampaignStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.campaigns.Create(ctx, &campaign); err != nil {
		return 0, 0, err
	}
	return campaign.ID, campaign.Status, nil
}

func (s *campaignAdminService) CreateVersion(ctx context.Context, campaignID int64) (int, error) {
	if _, err := s.campaigns.GetByID(campaignID); err != nil {
		return 0, err
	}
	maxVersion, err := s.drafts.MaxVersion(campaignID)
	if err != nil {
		return 0, err
	}
	content, err := s.initialDraftContent(campaignID, maxVersion)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	draft := model.CampaignDraft{
		ActivityID: campaignID,
		Content:    content,
		Version:    maxVersion + 1,
		Status:     model.CampaignDraftStatusDraft,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.drafts.Create(ctx, &draft); err != nil {
		return 0, err
	}
	return draft.Version, nil
}

func (s *campaignAdminService) EditVersion(ctx context.Context, campaignID int64, version int, campaign data.CampaignVO) error {
	draft, err := s.drafts.GetByActivityAndVersion(campaignID, version)
	if err != nil {
		return err
	}
	if draft.Status != model.CampaignDraftStatusDraft {
		return data.ErrCampaignDraftNotEditable
	}
	raw, err := json.Marshal(campaignDraftVO(campaign))
	if err != nil {
		return err
	}
	draft.Content = string(raw)
	draft.UpdatedAt = time.Now()
	return s.drafts.Update(ctx, draft)
}

func (s *campaignAdminService) PublishCampaign(ctx context.Context, campaignID int64, operator string) (*data.CampaignVO, error) {
	draft, err := s.drafts.GetLatestByActivityID(campaignID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return nil, data.ErrCampaignNoDraftToPublish
		}
		return nil, err
	}
	if draft.Status != model.CampaignDraftStatusDraft {
		return nil, data.ErrCampaignNoDraftToPublish
	}
	content, err := parseDraftContent(draft.Content)
	if err != nil {
		return nil, err
	}
	if err := validatePublishContent(content); err != nil {
		return nil, err
	}
	campaign, err := s.campaigns.GetByID(campaignID)
	if err != nil {
		return nil, err
	}
	applyContentToCampaign(campaign, content, operator)
	if err := s.campaigns.Update(ctx, campaign); err != nil {
		return nil, err
	}
	if err := s.rules.ReplaceByCampaignID(ctx, campaignID, flattenRewardRules(campaignID, content.RewardRules)); err != nil {
		return nil, err
	}
	draft.Status = model.CampaignDraftStatusPublished
	draft.UpdatedAt = time.Now()
	if err := s.drafts.Update(ctx, draft); err != nil {
		return nil, err
	}
	return s.GetCampaign(campaignID)
}

func (s *campaignAdminService) ListCampaigns(req data.CampaignListReq) ([]data.CampaignListVO, int64, error) {
	query := mysql.CampaignQuery{Status: req.Status, CampaignID: req.CampaignID}
	total, err := s.campaigns.Count(query)
	if err != nil {
		return nil, 0, err
	}
	offset, limit := pageToOffsetLimit(req.Page, req.PageSize)
	items, err := s.campaigns.Find(query, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	out := make([]data.CampaignListVO, 0, len(items))
	for _, item := range items {
		out = append(out, campaignToListVO(item))
	}
	return out, total, nil
}

func (s *campaignAdminService) GetCampaign(campaignID int64) (*data.CampaignVO, error) {
	campaign, err := s.campaigns.GetByID(campaignID)
	if err != nil {
		return nil, err
	}
	detail := campaignToVO(campaign)
	draft, err := s.drafts.GetLatestByActivityID(campaignID)
	if err == nil {
		detail.Version = int64(draft.Version)
		content, err := parseDraftContent(draft.Content)
		if err != nil {
			return nil, err
		}
		applyDraftVO(detail, content)
	} else if !mysql.IsNotFound(err) {
		return nil, err
	}
	return detail, nil
}

func (s *campaignAdminService) initialDraftContent(campaignID int64, maxVersion int) (string, error) {
	if maxVersion > 0 {
		prev, err := s.drafts.GetByActivityAndVersion(campaignID, maxVersion)
		if err != nil {
			return "", err
		}
		if prev.Content != "" {
			return prev.Content, nil
		}
	}
	campaign, err := s.campaigns.GetByID(campaignID)
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(data.CampaignVO{Name: campaign.Name})
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func trimCampaignName(name string) string {
	return strings.TrimSpace(name)
}

func parseDraftContent(raw string) (data.CampaignVO, error) {
	var content data.CampaignVO
	if raw == "" {
		return content, nil
	}
	if err := json.Unmarshal([]byte(raw), &content); err != nil {
		return content, err
	}
	return content, nil
}

func validatePublishContent(content data.CampaignVO) error {
	if content.Budget.ProjectID <= 0 {
		return fmt.Errorf("%w: budget projectId must be > 0", data.ErrCampaignPublishInvalid)
	}
	if content.RewardRules.TaskGroupID <= 0 || content.RewardRules.TaskGroupReward <= 0 {
		return fmt.Errorf("%w: taskGroupId and taskGroupReward must be > 0", data.ErrCampaignPublishInvalid)
	}
	for _, item := range content.RewardRules.TaskRewardItems {
		if item.TaskID <= 0 || item.RewardTemplateID <= 0 {
			return fmt.Errorf("%w: taskId and rewardTemplateId must be > 0", data.ErrCampaignPublishInvalid)
		}
	}
	return nil
}
