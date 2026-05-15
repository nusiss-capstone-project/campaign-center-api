package service

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// UserCampaignService user-facing campaign flows.
type UserCampaignService interface {
	GetLandingPageUI(campaignID, userID int64, lang string) (*HTTPReply, error)
	JoinCampaign(campaignID, userID int64) (*HTTPReply, error)
	SimulateTopUp(campaignID, userID int64, amount float64) (*HTTPReply, error)
}

type userCampaignService struct {
	campaigns    mysql.CampaignRepository
	landingPages mysql.LandingPageRepository
	translations mysql.LandingPageTranslationRepository
	participants mysql.ParticipantRepository
	users        mysql.UserRepository
	accounts     AccountService
	rewards      CampaignRewardNotifier
}

var (
	userCampaignServiceOnce sync.Once
	userCampaignServiceInst UserCampaignService
)

// NewUserCampaignService builds a user campaign service with explicit repositories (for tests).
func NewUserCampaignService(
	campaigns mysql.CampaignRepository,
	landingPages mysql.LandingPageRepository,
	translations mysql.LandingPageTranslationRepository,
	participants mysql.ParticipantRepository,
	users mysql.UserRepository,
	accounts AccountService,
	rewards CampaignRewardNotifier,
) UserCampaignService {
	return &userCampaignService{
		campaigns:    campaigns,
		landingPages: landingPages,
		translations: translations,
		participants: participants,
		users:        users,
		accounts:     accounts,
		rewards:      rewards,
	}
}

// GetUserCampaignService returns the singleton user campaign service.
func GetUserCampaignService() UserCampaignService {
	userCampaignServiceOnce.Do(func() {
		userCampaignServiceInst = NewUserCampaignService(
			mysql.GetCampaignRepository(),
			mysql.GetLandingPageRepository(),
			mysql.GetLandingPageTranslationRepository(),
			mysql.GetParticipantRepository(),
			mysql.GetUserRepository(),
			GetAccountService(),
			GetCampaignRewardNotifier(),
		)
	})
	return userCampaignServiceInst
}

func (s *userCampaignService) GetLandingPageUI(campaignID, userID int64, lang string) (*HTTPReply, error) {
	campaign, err := s.campaigns.GetByID(campaignID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return &HTTPReply{HTTPStatus: http.StatusNotFound, Code: -1, Message: "campaign not found"}, nil
		}
		return nil, err
	}
	if campaign.LandingPageID == 0 {
		return &HTTPReply{HTTPStatus: http.StatusNotFound, Code: -1, Message: "landing page not configured"}, nil
	}
	lp, err := s.landingPages.GetByID(campaign.LandingPageID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return &HTTPReply{HTTPStatus: http.StatusNotFound, Code: -1, Message: "landing page not found"}, nil
		}
		return nil, err
	}
	rules, err := model.ParseRewardRulesJSON(campaign.RewardRules)
	if err != nil {
		return &HTTPReply{HTTPStatus: http.StatusInternalServerError, Code: -1, Message: err.Error()}, nil
	}
	titleBase, descBase, termsBase, resolvedLang, err := s.resolveLandingPageTexts(lp, lang)
	if err != nil {
		return nil, err
	}
	title := replaceLandingTemplates(titleBase, rules.TopupThreshold, rules.RewardAmount)
	payload, err := s.buildLandingPageUIPayload(
		campaign, campaignID, userID, lp, resolvedLang, title, descBase, termsBase, rules,
	)
	if err != nil {
		return nil, err
	}
	return landingPageUIReply(payload), nil
}

func (s *userCampaignService) JoinCampaign(campaignID, userID int64) (*HTTPReply, error) {
	campaign, err := s.campaigns.GetByID(campaignID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return &HTTPReply{HTTPStatus: http.StatusNotFound, Code: -1, Message: "campaign not found"}, nil
		}
		return nil, err
	}
	if campaign.Status != model.CampaignStatusPublished {
		return &HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeNotEligible,
			Message:    "campaign not available",
			Data:       map[string]any{"reason": "CAMPAIGN_NOT_PUBLISHED"},
		}, nil
	}
	now := time.Now()
	if now.Before(campaign.RegistrationStartTime) || now.After(campaign.RegistrationEndTime) {
		return &HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeNotEligible,
			Message:    "User is not eligible for this campaign",
			Data:       map[string]any{"reason": "OUTSIDE_REGISTRATION"},
		}, nil
	}

	user, err := s.users.GetByID(userID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return &HTTPReply{
				HTTPStatus: http.StatusOK,
				Code:       data.CodeNotEligible,
				Message:    "User is not eligible for this campaign",
				Data:       map[string]any{"reason": "USER_NOT_FOUND"},
			}, nil
		}
		return nil, err
	}
	if user.KYCStatus != model.KYCStatusPassed {
		return &HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeNotEligible,
			Message:    "User is not eligible for this campaign",
			Data:       map[string]any{"reason": model.RejectReasonKYCNotPassed},
		}, nil
	}
	if user.Segment != campaign.TargetUserSegment {
		return &HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeNotEligible,
			Message:    "User is not eligible for this campaign",
			Data:       map[string]any{"reason": model.RejectReasonSegment},
		}, nil
	}
	if user.Market != campaign.TargetMarket {
		return &HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeNotEligible,
			Message:    "User is not eligible for this campaign",
			Data:       map[string]any{"reason": "MARKET_MISMATCH"},
		}, nil
	}

	if existing, err := s.participants.GetByCampaignAndUser(campaignID, userID); err == nil {
		return &HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeSuccess,
			Message:    "success",
			Data: map[string]any{
				"campaignId":   campaignID,
				"userId":       userID,
				"joinStatus":   existing.JoinStatus,
				"taskStatus":   existing.TaskStatus,
				"rewardStatus": existing.RewardStatus,
			},
		}, nil
	} else if !mysql.IsNotFound(err) {
		return nil, err
	}

	p := model.CampaignParticipant{
		CampaignID:   campaignID,
		UserID:       userID,
		JoinStatus:   model.JoinStatusJoined,
		TaskStatus:   model.TaskStatusNotStarted,
		RewardStatus: model.RewardStatusNotGranted,
		JoinedAt:     now,
		UpdatedAt:    now,
	}
	if err := s.participants.Create(&p); err != nil {
		return nil, err
	}
	return &HTTPReply{
		HTTPStatus: http.StatusOK,
		Code:       data.CodeSuccess,
		Message:    "success",
		Data: map[string]any{
			"campaignId":   campaignID,
			"userId":       userID,
			"joinStatus":   model.JoinStatusJoined,
			"taskStatus":   model.TaskStatusNotStarted,
			"rewardStatus": model.RewardStatusNotGranted,
		},
	}, nil
}

func (s *userCampaignService) SimulateTopUp(campaignID, userID int64, amount float64) (*HTTPReply, error) {
	rules, participant, early, err := s.simulateTopUpPrecheck(campaignID, userID, amount)
	if err != nil {
		return nil, err
	}
	if early != nil {
		return early, nil
	}

	user, err := s.users.GetByID(userID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return &HTTPReply{HTTPStatus: http.StatusBadRequest, Code: -1, Message: "user not found"}, nil
		}
		return nil, err
	}

	now := time.Now()
	applyTopUpProgressToParticipant(participant, amount, now)

	recharge, err := s.accounts.Recharge(userID, amount, model.DefaultCurrency)
	if err != nil {
		return nil, err
	}
	return s.simulateTopUpAfterRecharge(
		participant, rules, campaignID, userID, amount, user,
		recharge.TransactionNo, recharge.BalanceAfter,
	)
}

// simulateTopUpPrecheck loads campaign and participant and returns an HTTPReply for early business exits.
func (s *userCampaignService) simulateTopUpPrecheck(campaignID, userID int64, amount float64) (
	rules model.RewardRulesPayload,
	participant *model.CampaignParticipant,
	early *HTTPReply,
	err error,
) {
	campaign, err := s.campaigns.GetByID(campaignID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return model.RewardRulesPayload{}, nil, &HTTPReply{
				HTTPStatus: http.StatusNotFound, Code: -1, Message: "campaign not found",
			}, nil
		}
		return model.RewardRulesPayload{}, nil, nil, err
	}
	rules, err = model.ParseRewardRulesJSON(campaign.RewardRules)
	if err != nil {
		return model.RewardRulesPayload{}, nil, &HTTPReply{
			HTTPStatus: http.StatusInternalServerError, Code: -1, Message: err.Error(),
		}, nil
	}

	participant, err = s.participants.GetByCampaignAndUser(campaignID, userID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return model.RewardRulesPayload{}, nil, &HTTPReply{
				HTTPStatus: http.StatusBadRequest, Code: -1, Message: "user has not joined this campaign",
			}, nil
		}
		return model.RewardRulesPayload{}, nil, nil, err
	}
	if participant.RewardStatus == model.RewardStatusGranted {
		return model.RewardRulesPayload{}, nil, &HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeDuplicateReward,
			Message:    "Reward already granted",
			Data: map[string]any{
				"campaignId":   campaignID,
				"userId":       userID,
				"rewardStatus": model.RewardStatusGranted,
			},
		}, nil
	}
	if amount < rules.TopupThreshold {
		return model.RewardRulesPayload{}, nil, &HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeTopupNotQualified,
			Message:    "Top-up amount does not meet campaign requirement",
			Data: map[string]any{
				"requiredAmount": rules.TopupThreshold,
				"actualAmount":   amount,
				"taskStatus":     model.TaskStatusNotQualified,
				"rewardStatus":   model.RewardStatusNotGranted,
			},
		}, nil
	}
	return rules, participant, nil, nil
}

func applyTopUpProgressToParticipant(participant *model.CampaignParticipant, amount float64, now time.Time) {
	completedAt := now
	participant.TopupAmount = amount
	participant.TaskStatus = model.TaskStatusCompleted
	participant.CompletedAt = &completedAt
	participant.UpdatedAt = now
}

func replaceLandingTemplates(title string, threshold, reward float64) string {
	s := strings.ReplaceAll(title, "{{threshold}}", formatMoney(threshold))
	return strings.ReplaceAll(s, "{{reward}}", formatMoney(reward))
}

func formatMoney(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
