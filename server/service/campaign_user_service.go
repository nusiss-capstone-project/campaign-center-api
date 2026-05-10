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
	GetLandingPageUI(campaignID, userID int64, language string) (*HTTPReply, error)
	JoinCampaign(campaignID, userID int64) (*HTTPReply, error)
	SimulateTopUp(campaignID, userID int64, amount float64) (*HTTPReply, error)
}

type userCampaignService struct {
	campaigns    mysql.CampaignRepository
	landingPages mysql.LandingPageRepository
	participants mysql.ParticipantRepository
	users        mysql.UserRepository
	rewardTx     mysql.RewardTransactionRepository
}

var (
	userCampaignServiceOnce sync.Once
	userCampaignServiceInst UserCampaignService
)

// NewUserCampaignService builds a user campaign service with explicit repositories (for tests).
func NewUserCampaignService(
	campaigns mysql.CampaignRepository,
	landingPages mysql.LandingPageRepository,
	participants mysql.ParticipantRepository,
	users mysql.UserRepository,
	rewardTx mysql.RewardTransactionRepository,
) UserCampaignService {
	return &userCampaignService{
		campaigns:    campaigns,
		landingPages: landingPages,
		participants: participants,
		users:        users,
		rewardTx:     rewardTx,
	}
}

// GetUserCampaignService returns the singleton user campaign service.
func GetUserCampaignService() UserCampaignService {
	userCampaignServiceOnce.Do(func() {
		userCampaignServiceInst = NewUserCampaignService(
			mysql.GetCampaignRepository(),
			mysql.GetLandingPageRepository(),
			mysql.GetParticipantRepository(),
			mysql.GetUserRepository(),
			mysql.GetRewardTransactionRepository(),
		)
	})
	return userCampaignServiceInst
}

func (s *userCampaignService) GetLandingPageUI(campaignID, userID int64, language string) (*HTTPReply, error) {
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
	if language != "" && lp.Language != language {
		return &HTTPReply{HTTPStatus: http.StatusNotFound, Code: -1, Message: "landing page language mismatch"}, nil
	}
	rules, err := model.ParseRewardRulesJSON(campaign.RewardRules)
	if err != nil {
		return &HTTPReply{HTTPStatus: http.StatusInternalServerError, Code: -1, Message: err.Error()}, nil
	}
	title := replaceLandingTemplates(lp.Title, rules.TopupThreshold, rules.RewardAmount)

	var joined bool
	taskStatus := model.TaskStatusNotStarted
	rewardStatus := model.RewardStatusNotGranted
	if userID > 0 {
		if p, err := s.participants.GetByCampaignAndUser(campaignID, userID); err == nil {
			joined = true
			taskStatus = p.TaskStatus
			rewardStatus = p.RewardStatus
		} else if !mysql.IsNotFound(err) {
			return nil, err
		}
	}

	payload := map[string]any{
		"campaignId":            campaignID,
		"campaignName":          campaign.Name,
		"campaignType":          campaign.Type,
		"status":                campaign.Status,
		"registrationStartTime": campaign.RegistrationStartTime.Format(time.RFC3339),
		"registrationEndTime":   campaign.RegistrationEndTime.Format(time.RFC3339),
		"campaignStartTime":     campaign.CampaignStartTime.Format(time.RFC3339),
		"campaignEndTime":       campaign.CampaignEndTime.Format(time.RFC3339),
		"landingPage": map[string]any{
			"language":       lp.Language,
			"bannerImageUrl": lp.BannerImageURL,
			"title":          title,
			"description":    lp.Description,
			"terms":          lp.Terms,
		},
		"rewardRule": map[string]any{
			"topupThreshold": rules.TopupThreshold,
			"rewardAmount":   rules.RewardAmount,
			"rewardType":     rules.RewardType,
		},
		"userStatus": map[string]any{
			"joined":       joined,
			"taskStatus":   taskStatus,
			"rewardStatus": rewardStatus,
		},
	}
	return &HTTPReply{HTTPStatus: http.StatusOK, Code: data.CodeSuccess, Message: "success", Data: payload}, nil
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

	if user.RiskLevel == model.RiskLevelHigh {
		return s.simulateTopUpManualReview(participant, campaignID, userID, amount)
	}
	return s.simulateTopUpGrantApproved(participant, rules, campaignID, userID, amount, now)
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

func (s *userCampaignService) simulateTopUpManualReview(
	participant *model.CampaignParticipant,
	campaignID, userID int64,
	amount float64,
) (*HTTPReply, error) {
	participant.RiskStatus = model.RiskStatusManualReview
	participant.RewardStatus = model.RewardStatusPendingReview
	participant.RewardAmount = 0
	if err := s.participants.Save(participant); err != nil {
		return nil, err
	}
	return &HTTPReply{
		HTTPStatus: http.StatusOK,
		Code:       data.CodeSuccess,
		Message:    "manual review required",
		Data: map[string]any{
			"campaignId":   campaignID,
			"userId":       userID,
			"topupAmount":  amount,
			"taskStatus":   model.TaskStatusCompleted,
			"riskStatus":   model.RiskStatusManualReview,
			"rewardStatus": model.RewardStatusPendingReview,
			"rewardAmount": 0,
		},
	}, nil
}

func (s *userCampaignService) simulateTopUpGrantApproved(
	participant *model.CampaignParticipant,
	rules model.RewardRulesPayload,
	campaignID, userID int64,
	amount float64,
	now time.Time,
) (*HTTPReply, error) {
	participant.RiskStatus = model.RiskStatusApproved
	participant.RewardStatus = model.RewardStatusGranted
	participant.RewardAmount = rules.RewardAmount
	rewardedAt := now
	participant.RewardedAt = &rewardedAt

	rewardRow := model.RewardTransaction{
		CampaignID:    campaignID,
		UserID:        userID,
		ParticipantID: participant.ID,
		RewardType:    rules.RewardType,
		RewardAmount:  rules.RewardAmount,
		Status:        model.RewardTxnStatusCompleted,
		CreatedAt:     now,
	}
	if err := s.rewardTx.CommitGrantWithParticipant(participant, &rewardRow); err != nil {
		return nil, err
	}

	return &HTTPReply{
		HTTPStatus: http.StatusOK,
		Code:       data.CodeSuccess,
		Message:    "success",
		Data: map[string]any{
			"campaignId":          campaignID,
			"userId":              userID,
			"topupAmount":         amount,
			"taskStatus":          model.TaskStatusCompleted,
			"riskStatus":          model.RiskStatusApproved,
			"rewardStatus":        model.RewardStatusGranted,
			"rewardAmount":        rules.RewardAmount,
			"rewardTransactionId": rewardRow.ID,
		},
	}, nil
}

func replaceLandingTemplates(title string, threshold, reward float64) string {
	s := strings.ReplaceAll(title, "{{threshold}}", formatMoney(threshold))
	return strings.ReplaceAll(s, "{{reward}}", formatMoney(reward))
}

func formatMoney(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
