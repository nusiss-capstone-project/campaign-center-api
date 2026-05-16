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
	ListAvailableCampaigns(userID int64) (*HTTPReply, error)
	GetLandingPageUI(campaignID, userID int64, lang string) (*HTTPReply, error)
	JoinCampaign(campaignID, userID int64) (*HTTPReply, error)
	SimulateTopUp(campaignID, userID int64, amount float64) (*HTTPReply, error)
}

type userCampaignService struct {
	campaigns           mysql.CampaignRepository
	landingPages        mysql.LandingPageRepository
	landingTranslations LandingPageTranslationService
	participants        mysql.ParticipantRepository
	users               mysql.UserRepository
	accounts            AccountService
	rewards             CampaignRewardNotifier
}

var (
	userCampaignServiceOnce sync.Once
	userCampaignServiceInst UserCampaignService
)

// NewUserCampaignService builds a user campaign service with explicit repositories (for tests).
func NewUserCampaignService(
	campaigns mysql.CampaignRepository,
	landingPages mysql.LandingPageRepository,
	landingTranslations LandingPageTranslationService,
	participants mysql.ParticipantRepository,
	users mysql.UserRepository,
	accounts AccountService,
	rewards CampaignRewardNotifier,
) UserCampaignService {
	return &userCampaignService{
		campaigns:           campaigns,
		landingPages:        landingPages,
		landingTranslations: landingTranslations,
		participants:        participants,
		users:               users,
		accounts:            accounts,
		rewards:             rewards,
	}
}

// GetUserCampaignService returns the singleton user campaign service.
func GetUserCampaignService() UserCampaignService {
	userCampaignServiceOnce.Do(func() {
		userCampaignServiceInst = NewUserCampaignService(
			mysql.GetCampaignRepository(),
			mysql.GetLandingPageRepository(),
			GetLandingPageTranslationService(),
			mysql.GetParticipantRepository(),
			mysql.GetUserRepository(),
			GetAccountService(),
			GetCampaignRewardNotifier(),
		)
	})
	return userCampaignServiceInst
}

func (s *userCampaignService) ListAvailableCampaigns(userID int64) (*HTTPReply, error) {
	now := time.Now()
	campaigns, err := s.campaigns.ListPublishedActiveOrUpcoming(now)
	if err != nil {
		return nil, err
	}
	user, err := s.users.GetByID(userID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return campaignListReply([]map[string]any{}, []map[string]any{}), nil
		}
		return nil, err
	}

	visibleCampaigns := make([]model.Campaign, 0, len(campaigns))
	campaignIDs := make([]int64, 0, len(campaigns))
	for _, campaign := range campaigns {
		if campaignEligibilityRejectReason(user, campaign) != "" {
			continue
		}
		visibleCampaigns = append(visibleCampaigns, campaign)
		campaignIDs = append(campaignIDs, campaign.ID)
	}
	participants, err := s.participants.ListByUserAndCampaignIDs(userID, campaignIDs)
	if err != nil {
		return nil, err
	}
	joinedByCampaign := make(map[int64]bool, len(participants))
	for _, participant := range participants {
		if participant.JoinStatus == model.JoinStatusJoined {
			joinedByCampaign[participant.CampaignID] = true
		}
	}

	ongoing := make([]map[string]any, 0)
	upcoming := make([]map[string]any, 0)
	for _, campaign := range visibleCampaigns {
		item := campaignListItem(campaign)
		if !now.Before(campaign.CampaignStartTime) && !now.After(campaign.CampaignEndTime) {
			item["joined"] = joinedByCampaign[campaign.ID]
			ongoing = append(ongoing, item)
			continue
		}
		if now.Before(campaign.CampaignStartTime) {
			upcoming = append(upcoming, item)
		}
	}

	return campaignListReply(ongoing, upcoming), nil
}

func campaignListReply(ongoing, upcoming []map[string]any) *HTTPReply {
	return &HTTPReply{
		HTTPStatus: http.StatusOK,
		Code:       data.CodeSuccess,
		Message:    "success",
		Data: map[string]any{
			"ongoing":  ongoing,
			"upcoming": upcoming,
		},
	}
}

func (s *userCampaignService) GetLandingPageUI(campaignID, userID int64, lang string) (*HTTPReply, error) {
	campaign, err := s.campaigns.GetByID(campaignID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return &HTTPReply{HTTPStatus: http.StatusNotFound, Code: -1, Message: "campaign not found"}, nil
		}
		return nil, err
	}
	if userID > 0 {
		user, err := s.users.GetByID(userID)
		if err != nil {
			if mysql.IsNotFound(err) {
				return &HTTPReply{HTTPStatus: http.StatusNotFound, Code: -1, Message: "campaign not found"}, nil
			}
			return nil, err
		}
		if campaignEligibilityRejectReason(user, *campaign) != "" {
			return &HTTPReply{HTTPStatus: http.StatusNotFound, Code: -1, Message: "campaign not found"}, nil
		}
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
	texts, err := s.landingTranslations.ResolveLandingPageTexts(lp, lang)
	if err != nil {
		return nil, err
	}
	templateVars := buildLandingTemplateVars(campaign, rules)
	title := replaceLandingTemplates(texts.Title, templateVars)
	description := replaceLandingTemplates(texts.Description, templateVars)
	terms := replaceLandingTemplates(texts.Terms, templateVars)
	payload, err := s.buildLandingPageUIPayload(
		campaign, campaignID, userID, lp, texts.Lang, title, description, terms, rules,
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
	if reason := campaignEligibilityRejectReason(user, *campaign); reason != "" {
		return &HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeNotEligible,
			Message:    "User is not eligible for this campaign",
			Data:       map[string]any{"reason": reason},
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
		if IsInvalidAccountInput(err) {
			return &HTTPReply{HTTPStatus: http.StatusBadRequest, Code: -1, Message: err.Error()}, nil
		}
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

func (s *userCampaignService) buildLandingPageUIPayload(
	campaign *model.Campaign,
	campaignID, userID int64,
	lp *model.CampaignLandingPage,
	resolvedLang, title, descBase, termsBase string,
	rules model.RewardRulesPayload,
) (map[string]any, error) {
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
	return map[string]any{
		"campaignId":            campaignID,
		"campaignName":          campaign.Name,
		"campaignType":          campaign.Type,
		"status":                campaign.Status,
		"registrationStartTime": campaign.RegistrationStartTime.Format(time.RFC3339),
		"registrationEndTime":   campaign.RegistrationEndTime.Format(time.RFC3339),
		"campaignStartTime":     campaign.CampaignStartTime.Format(time.RFC3339),
		"campaignEndTime":       campaign.CampaignEndTime.Format(time.RFC3339),
		"landingPage": map[string]any{
			"lang": resolvedLang, "defaultLang": lp.DefaultLang,
			"bannerImageUrl": lp.BannerImageURL,
			"title":          title, "description": descBase, "terms": termsBase,
		},
		"rewardRule": rules,
		"userStatus": map[string]any{
			"joined": joined, "taskStatus": taskStatus, "rewardStatus": rewardStatus,
		},
	}, nil
}

func landingPageUIReply(payload map[string]any) *HTTPReply {
	return &HTTPReply{HTTPStatus: http.StatusOK, Code: data.CodeSuccess, Message: "success", Data: payload}
}

func campaignListItem(campaign model.Campaign) map[string]any {
	return map[string]any{
		"id":        campaign.ID,
		"name":      campaign.Name,
		"startTime": campaign.CampaignStartTime.Format(time.RFC3339),
		"endTime":   campaign.CampaignEndTime.Format(time.RFC3339),
	}
}

func campaignEligibilityRejectReason(user *model.User, campaign model.Campaign) string {
	if user == nil || user.KYCStatus != model.KYCStatusPassed {
		return model.RejectReasonKYCNotPassed
	}
	if campaign.TargetUserSegment != "" &&
		campaign.TargetUserSegment != model.UserSegmentAllUsers &&
		user.Segment != campaign.TargetUserSegment {
		return model.RejectReasonSegment
	}
	if campaign.TargetMarket != "" &&
		campaign.TargetMarket != model.MarketGlobal &&
		user.Market != campaign.TargetMarket {
		return "MARKET_MISMATCH"
	}
	return ""
}

func (s *userCampaignService) simulateTopUpAfterRecharge(
	participant *model.CampaignParticipant,
	rules model.RewardRulesPayload,
	campaignID, userID int64,
	amount float64,
	user *model.User,
	rechargeTxnNo string,
	balanceAfter float64,
) (*HTTPReply, error) {
	if user.RiskLevel == model.RiskLevelHigh {
		return s.simulateTopUpManualReviewWithAccount(
			participant, campaignID, userID, amount, rechargeTxnNo, balanceAfter,
		)
	}
	return s.simulateTopUpEnqueueReward(
		participant, rules, campaignID, userID, amount, rechargeTxnNo, balanceAfter,
	)
}

func (s *userCampaignService) simulateTopUpManualReviewWithAccount(
	participant *model.CampaignParticipant,
	campaignID, userID int64,
	amount float64,
	rechargeTxnNo string,
	balanceAfter float64,
) (*HTTPReply, error) {
	participant.RiskStatus = model.RiskStatusManualReview
	participant.RewardStatus = model.RewardStatusPendingReview
	participant.RewardAmount = 0
	if err := s.participants.Save(participant); err != nil {
		return nil, err
	}
	return topUpReply(campaignID, userID, amount, rechargeTxnNo, balanceAfter, map[string]any{
		"taskStatus": model.TaskStatusCompleted, "riskStatus": model.RiskStatusManualReview,
		"rewardStatus": model.RewardStatusPendingReview, "rewardAmount": 0,
	}, "manual review required")
}

func (s *userCampaignService) simulateTopUpEnqueueReward(
	participant *model.CampaignParticipant,
	rules model.RewardRulesPayload,
	campaignID, userID int64,
	amount float64,
	rechargeTxnNo string,
	balanceAfter float64,
) (*HTTPReply, error) {
	participant.RiskStatus = model.RiskStatusApproved
	participant.RewardStatus = model.RewardStatusPending
	participant.RewardAmount = rules.RewardAmount
	if err := s.participants.Save(participant); err != nil {
		return nil, err
	}
	s.rewards.NotifyTopUpReward(TopUpRewardEvent{
		CampaignID: campaignID, UserID: userID, TopupAmount: amount,
		ParticipantID: participant.ID, ManualReview: false,
		RewardAmount: rules.RewardAmount, RewardType: rules.RewardType,
	})
	return topUpReply(campaignID, userID, amount, rechargeTxnNo, balanceAfter, map[string]any{
		"taskStatus": model.TaskStatusCompleted, "riskStatus": model.RiskStatusApproved,
		"rewardStatus": model.RewardStatusPending, "rewardAmount": rules.RewardAmount,
	}, "reward processing")
}

func topUpReply(
	campaignID, userID int64, amount float64,
	rechargeTxnNo string, balanceAfter float64,
	extra map[string]any,
	message string,
) (*HTTPReply, error) {
	payload := map[string]any{
		"campaignId": campaignID, "userId": userID, "topupAmount": amount,
		"rechargeTransactionNo": rechargeTxnNo, "balanceAfter": balanceAfter,
	}
	for k, v := range extra {
		payload[k] = v
	}
	return &HTTPReply{
		HTTPStatus: http.StatusOK, Code: data.CodeSuccess, Message: message, Data: payload,
	}, nil
}

func buildLandingTemplateVars(campaign *model.Campaign, rules model.RewardRulesPayload) map[string]string {
	vars := map[string]string{
		"topupThreshold":    formatMoney(rules.TopupThreshold),
		"rewardType":        rules.RewardType,
		"rewardAmount":      formatMoney(rules.RewardAmount),
		"rewardCurrency":    rules.RewardCurrency,
		"rewardMode":        rules.RewardMode,
		"rewardPercentage":  formatMoney(rules.RewardPercentage),
		"maxRewardAmount":   formatMoney(rules.MaxRewardAmount),
		"maxClaimPerUser":   strconv.Itoa(rules.MaxClaimPerUser),
		"minObtainDays":     strconv.Itoa(rules.MinObtainDays),
		"campaignStartDate": campaign.CampaignStartTime.Format("2006-01-02"),
		"campaignEndDate":   campaign.CampaignEndTime.Format("2006-01-02"),
	}

	// Keep existing landing page templates working while newer content migrates
	// to the explicit reward rule variable names above.
	vars["threshold"] = vars["topupThreshold"]
	vars["reward"] = vars["rewardAmount"]
	return vars
}

func replaceLandingTemplates(text string, vars map[string]string) string {
	for key, value := range vars {
		text = strings.ReplaceAll(text, "{{"+key+"}}", value)
	}
	return text
}

func formatMoney(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
