package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/proxy"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"gorm.io/gorm"
)

const webCampaignListLimit = 1000

// WebCampaignService user-facing campaign list / detail / join.
type WebCampaignService interface {
	ListCampaigns(ctx context.Context, userID int64, lang string) (*data.WebCampaignListData, error)
	GetCampaignLanding(ctx context.Context, campaignID, userID int64, lang string) (*data.WebCampaignLandingPageData, error)
	JoinCampaign(ctx context.Context, campaignID, userID int64) (*data.WebJoinCampaignData, error)
}

type webCampaignService struct {
	campaigns    mysql.CampaignRepository
	pages        mysql.LandingPageRepository
	translations mysql.LandingPageTranslationRepository
	participants mysql.ParticipantRepository
	rules        mysql.CampaignRewardRuleRepository
	usergroup    proxy.UsergroupClient
	task         proxy.TaskClient
}

var (
	webCampaignSvcOnce sync.Once
	webCampaignSvcInst WebCampaignService
)

// NewWebCampaignService wires repositories (for tests).
func NewWebCampaignService(
	campaigns mysql.CampaignRepository,
	pages mysql.LandingPageRepository,
	translations mysql.LandingPageTranslationRepository,
	participants mysql.ParticipantRepository,
	rules mysql.CampaignRewardRuleRepository,
	usergroup proxy.UsergroupClient,
	task proxy.TaskClient,
) WebCampaignService {
	return &webCampaignService{
		campaigns:    campaigns,
		pages:        pages,
		translations: translations,
		participants: participants,
		rules:        rules,
		usergroup:    usergroup,
		task:         task,
	}
}

// GetWebCampaignService returns the singleton.
func GetWebCampaignService() WebCampaignService {
	webCampaignSvcOnce.Do(func() {
		webCampaignSvcInst = NewWebCampaignService(
			mysql.GetCampaignRepository(),
			mysql.GetLandingPageRepository(),
			mysql.GetLandingPageTranslationRepository(),
			mysql.GetParticipantRepository(),
			mysql.GetCampaignRewardRuleRepository(),
			proxy.GetUsergroupClient(),
			proxy.GetTaskClient(),
		)
	})
	return webCampaignSvcInst
}

func (s *webCampaignService) ListCampaigns(ctx context.Context, userID int64, lang string) (*data.WebCampaignListData, error) {
	status := model.CampaignStatusPublished
	campaigns, err := s.campaigns.Find(mysql.CampaignQuery{Status: &status}, 0, webCampaignListLimit)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(campaigns))
	for _, c := range campaigns {
		ids = append(ids, c.ID)
	}
	joinedSet, err := s.participants.ListJoinedCampaignIDs(userID, ids)
	if err != nil {
		return nil, err
	}

	titleByLP, err := s.resolveTitles(campaigns, lang)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	out := &data.WebCampaignListData{
		Ongoing:  make([]data.WebCampaignListItem, 0),
		Upcoming: make([]data.WebCampaignListItem, 0),
	}
	for _, c := range campaigns {
		item := data.WebCampaignListItem{
			ID:                c.ID,
			Title:             titleByLP[c.LandingPageID],
			Market:            c.Market,
			Status:            c.Status,
			CampaignStartTime: timePtrToUnix(c.CampaignStartTime),
			CampaignEndTime:   timePtrToUnix(c.CampaignEndTime),
			LandingPageID:     c.LandingPageID,
			Joined:            false,
		}
		if _, ok := joinedSet[c.ID]; ok {
			item.Joined = true
		}
		switch classifyCampaignWindow(now, c.CampaignStartTime, c.CampaignEndTime) {
		case "ongoing":
			out.Ongoing = append(out.Ongoing, item)
		case "upcoming":
			out.Upcoming = append(out.Upcoming, item)
		}
	}
	return out, nil
}

func (s *webCampaignService) GetCampaignLanding(ctx context.Context, campaignID, userID int64, lang string) (*data.WebCampaignLandingPageData, error) {
	campaign, err := s.requirePublishedCampaign(campaignID)
	if err != nil {
		return nil, err
	}
	lpContent, err := s.loadLandingContent(campaign.LandingPageID, lang)
	if err != nil {
		return nil, err
	}
	part, err := s.participants.GetByCampaignAndUser(campaignID, userID)
	if err != nil {
		return nil, err
	}
	out := &data.WebCampaignLandingPageData{
		CampaignID:  campaign.ID,
		UserID:      userID,
		Name:        campaign.Name,
		Market:      campaign.Market,
		TimeZone:    campaign.TimeZone,
		Joined:      part != nil,
		LandingPage: *lpContent,
	}
	if part != nil {
		out.JoinedAt = timeToUnix(part.JoinedAt)
	}
	return out, nil
}

func (s *webCampaignService) JoinCampaign(ctx context.Context, campaignID, userID int64) (*data.WebJoinCampaignData, error) {
	campaign, err := s.requirePublishedCampaign(campaignID)
	if err != nil {
		return nil, err
	}

	// Idempotent: already joined → return existing row.
	if existing, err := s.participants.GetByCampaignAndUser(campaignID, userID); err != nil {
		return nil, err
	} else if existing != nil {
		return &data.WebJoinCampaignData{
			CampaignID: campaignID,
			UserID:     userID,
			Joined:     true,
			JoinedAt:   timeToUnix(existing.JoinedAt),
			Message:    "joined",
		}, nil
	}

	matched, err := s.usergroup.MatchUserGroup(ctx, userID, campaign.TargetUserGroupID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, ErrUserNotEligible
	}

	taskGroupID, err := s.resolveTaskGroupID(campaignID)
	if err != nil {
		return nil, err
	}
	if taskGroupID > 0 {
		if err := s.task.EnrollTaskGroup(ctx, userID, taskGroupID); err != nil {
			return nil, fmt.Errorf("enroll task group: %w", err)
		}
	}

	row, err := s.participants.Join(ctx, campaignID, userID)
	if err != nil {
		return nil, err
	}
	return &data.WebJoinCampaignData{
		CampaignID: campaignID,
		UserID:     userID,
		Joined:     true,
		JoinedAt:   timeToUnix(row.JoinedAt),
		Message:    "joined",
	}, nil
}

func (s *webCampaignService) resolveTaskGroupID(campaignID int64) (int64, error) {
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

func (s *webCampaignService) requirePublishedCampaign(campaignID int64) (*model.Campaign, error) {
	campaign, err := s.campaigns.GetByID(campaignID)
	if err != nil {
		return nil, err
	}
	if campaign.Status != model.CampaignStatusPublished {
		return nil, gorm.ErrRecordNotFound
	}
	return campaign, nil
}

func (s *webCampaignService) resolveTitles(campaigns []model.Campaign, lang string) (map[int64]string, error) {
	out := make(map[int64]string)
	seen := make(map[int64]struct{})
	for _, c := range campaigns {
		lpID := c.LandingPageID
		if lpID == 0 {
			continue
		}
		if _, ok := seen[lpID]; ok {
			continue
		}
		seen[lpID] = struct{}{}
		content, err := s.loadLandingContent(lpID, lang)
		if err != nil {
			if mysql.IsNotFound(err) {
				out[lpID] = ""
				continue
			}
			return nil, err
		}
		out[lpID] = content.Title
	}
	return out, nil
}

func (s *webCampaignService) loadLandingContent(landingPageID int64, lang string) (*data.WebLandingPageContent, error) {
	if landingPageID == 0 {
		return &data.WebLandingPageContent{
			Lang:  lang,
			Steps: []data.LandingPageRepeatableItemVO{},
			Faq:   []data.LandingPageRepeatableItemVO{},
		}, nil
	}
	page, err := s.pages.GetByID(landingPageID)
	if err != nil {
		return nil, err
	}
	view := landingPageViewFromRow(page)
	resolvedLang := page.DefaultLang
	if lang != "" && lang != page.DefaultLang {
		tr, trErr := s.translations.GetByLandingPageAndLang(landingPageID, lang)
		if trErr != nil {
			return nil, trErr
		}
		if tr != nil {
			applyTranslationToView(view, lang, tr)
			resolvedLang = lang
		}
	}
	if view.Lang == "" {
		view.Lang = resolvedLang
	}
	return &data.WebLandingPageContent{
		Lang:           view.Lang,
		BannerImageURL: view.BannerImageURL,
		Title:          view.Title,
		Description:    view.Description,
		Terms:          view.Terms,
		Steps:          normalizeRepeatableVO(view.Steps),
		Faq:            normalizeRepeatableVO(view.Faq),
	}, nil
}

// classifyCampaignWindow returns ongoing | upcoming | ended (ended is omitted from list).
func classifyCampaignWindow(now time.Time, start, end *time.Time) string {
	if start != nil && now.Before(start.UTC()) {
		return "upcoming"
	}
	if end != nil && now.After(end.UTC()) {
		return "ended"
	}
	if start == nil && end == nil {
		return "ongoing"
	}
	return "ongoing"
}
