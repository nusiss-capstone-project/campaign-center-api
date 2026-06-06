package service

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/event"
	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
)

// UserCampaignService user-facing campaign flows.
type UserCampaignService interface {
	ListAvailableCampaigns(userID int64) (*data.HTTPReply, error)
	GetLandingPageUI(campaignID, userID int64, lang string) (*data.HTTPReply, error)
	JoinCampaign(campaignID, userID int64) (*data.HTTPReply, error)
	SimulateTopUp(campaignID, userID int64, amount float64) (*data.HTTPReply, error)
}

type userCampaignService struct {
	campaigns           mysql.CampaignRepository
	landingPages        mysql.LandingPageRepository
	landingTranslations LandingPageTranslationService
	participants        mysql.ParticipantRepository
	users               mysql.UserRepository
	accounts            AccountService
	rewards             event.CampaignRewardNotifier
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
	rewards event.CampaignRewardNotifier,
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

func (s *userCampaignService) ListAvailableCampaigns(userID int64) (*data.HTTPReply, error) {
	now := time.Now()
	campaigns, err := s.campaigns.ListPublishedActiveOrUpcoming(now)
	if err != nil {
		return nil, err
	}
	user, early, err := s.userForCampaignList(userID)
	if early != nil || err != nil {
		return early, err
	}
	visible, ids := filterEligibleCampaigns(campaigns, user)
	participants, err := s.participants.ListByUserAndCampaignIDs(userID, ids)
	if err != nil {
		return nil, err
	}
	ongoing, upcoming := partitionCampaignListItems(visible, now, joinedByCampaignID(participants))
	return campaignListReply(ongoing, upcoming), nil
}

func (s *userCampaignService) userForCampaignList(userID int64) (*model.User, *data.HTTPReply, error) {
	user, err := s.users.GetByID(userID)
	if err == nil {
		return user, nil, nil
	}
	if mysql.IsNotFound(err) {
		return nil, campaignListReply([]map[string]any{}, []map[string]any{}), nil
	}
	return nil, nil, err
}

func filterEligibleCampaigns(campaigns []model.Campaign, user *model.User) ([]model.Campaign, []int64) {
	visible := make([]model.Campaign, 0, len(campaigns))
	ids := make([]int64, 0, len(campaigns))
	for _, campaign := range campaigns {
		if campaignEligibilityRejectReason(user, campaign) != "" {
			continue
		}
		visible = append(visible, campaign)
		ids = append(ids, campaign.ID)
	}
	return visible, ids
}

func joinedByCampaignID(participants []model.CampaignParticipant) map[int64]bool {
	joined := make(map[int64]bool, len(participants))
	for _, p := range participants {
		if p.JoinStatus == model.JoinStatusJoined {
			joined[p.CampaignID] = true
		}
	}
	return joined
}

func partitionCampaignListItems(
	campaigns []model.Campaign, now time.Time, joined map[int64]bool,
) ([]map[string]any, []map[string]any) {
	ongoing := make([]map[string]any, 0)
	upcoming := make([]map[string]any, 0)
	for _, campaign := range campaigns {
		item := campaignListItem(campaign)
		if campaignIsOngoing(campaign, now) {
			item["joined"] = joined[campaign.ID]
			ongoing = append(ongoing, item)
			continue
		}
		if now.Before(campaign.CampaignStartTime) {
			upcoming = append(upcoming, item)
		}
	}
	return ongoing, upcoming
}

func campaignIsOngoing(campaign model.Campaign, now time.Time) bool {
	return !now.Before(campaign.CampaignStartTime) && !now.After(campaign.CampaignEndTime)
}

func campaignListReply(ongoing, upcoming []map[string]any) *data.HTTPReply {
	return &data.HTTPReply{
		HTTPStatus: http.StatusOK,
		Code:       data.CodeSuccess,
		Message:    MsgSuccess,
		Data: map[string]any{
			"ongoing":  ongoing,
			"upcoming": upcoming,
		},
	}
}

func (s *userCampaignService) GetLandingPageUI(campaignID, userID int64, lang string) (*data.HTTPReply, error) {
	campaign, early, err := s.getCampaignForLanding(campaignID)
	if early != nil || err != nil {
		return early, err
	}
	if early, err = s.assertUserCanViewLanding(campaign, userID); early != nil || err != nil {
		return early, err
	}
	lp, early, err := s.getLandingPageRecord(campaign)
	if early != nil || err != nil {
		return early, err
	}
	rules, err := model.ParseRewardRulesJSON(campaign.RewardRules)
	if err != nil {
		return &data.HTTPReply{HTTPStatus: http.StatusInternalServerError, Code: -1, Message: err.Error()}, nil
	}
	return s.buildLandingPageUIReply(campaign, campaignID, userID, lp, lang, rules)
}

func campaignNotFoundReply() *data.HTTPReply {
	return &data.HTTPReply{HTTPStatus: http.StatusNotFound, Code: -1, Message: MsgCampaignNotFound}
}

func landingNotFoundReply(message string) *data.HTTPReply {
	return &data.HTTPReply{HTTPStatus: http.StatusNotFound, Code: -1, Message: message}
}

func (s *userCampaignService) getCampaignForLanding(campaignID int64) (*model.Campaign, *data.HTTPReply, error) {
	campaign, err := s.campaigns.GetByID(campaignID)
	if err == nil {
		return campaign, nil, nil
	}
	if mysql.IsNotFound(err) {
		return nil, campaignNotFoundReply(), nil
	}
	return nil, nil, err
}

func (s *userCampaignService) assertUserCanViewLanding(campaign *model.Campaign, userID int64) (*data.HTTPReply, error) {
	if userID <= 0 {
		return nil, nil
	}
	user, err := s.users.GetByID(userID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return campaignNotFoundReply(), nil
		}
		return nil, err
	}
	if campaignEligibilityRejectReason(user, *campaign) != "" {
		return campaignNotFoundReply(), nil
	}
	return nil, nil
}

func (s *userCampaignService) getLandingPageRecord(campaign *model.Campaign) (*model.CampaignLandingPage, *data.HTTPReply, error) {
	if campaign.LandingPageID == 0 {
		return nil, landingNotFoundReply(MsgLandingPageNotConfigured), nil
	}
	lp, err := s.landingPages.GetByID(campaign.LandingPageID)
	if err == nil {
		return lp, nil, nil
	}
	if mysql.IsNotFound(err) {
		return nil, landingNotFoundReply(MsgLandingPageNotFound), nil
	}
	return nil, nil, err
}

func (s *userCampaignService) buildLandingPageUIReply(
	campaign *model.Campaign,
	campaignID, userID int64,
	lp *model.CampaignLandingPage,
	lang string,
	rules model.RewardRulesPayload,
) (*data.HTTPReply, error) {
	texts, err := s.landingTranslations.ResolveLandingPageTexts(lp, lang)
	if err != nil {
		return nil, err
	}
	templateVars := buildLandingTemplateVars(campaign, rules)
	title := replaceLandingTemplates(texts.Title, templateVars)
	description := replaceLandingTemplates(texts.Description, templateVars)
	terms := replaceLandingTemplates(texts.Terms, templateVars)
	payload, err := s.buildLandingPageUIPayload(landingPageUIInput{
		campaign: campaign, campaignID: campaignID, userID: userID, lp: lp,
		lang: texts.Lang, title: title, description: description, terms: terms, rules: rules,
	})
	if err != nil {
		return nil, err
	}
	return landingPageUIReply(payload), nil
}

func (s *userCampaignService) JoinCampaign(campaignID, userID int64) (*data.HTTPReply, error) {
	campaign, err := s.campaigns.GetByID(campaignID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return &data.HTTPReply{HTTPStatus: http.StatusNotFound, Code: -1, Message: MsgCampaignNotFound}, nil
		}
		return nil, err
	}
	if campaign.Status != model.CampaignStatusPublished {
		return &data.HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeNotEligible,
			Message:    MsgCampaignNotAvailable,
			Data:       map[string]any{"reason": "CAMPAIGN_NOT_PUBLISHED"},
		}, nil
	}
	now := time.Now()
	if now.Before(campaign.RegistrationStartTime) || now.After(campaign.RegistrationEndTime) {
		return &data.HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeNotEligible,
			Message:    MsgUserNotEligible,
			Data:       map[string]any{"reason": "OUTSIDE_REGISTRATION"},
		}, nil
	}

	user, err := s.users.GetByID(userID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return &data.HTTPReply{
				HTTPStatus: http.StatusOK,
				Code:       data.CodeNotEligible,
				Message:    MsgUserNotEligible,
				Data:       map[string]any{"reason": "USER_NOT_FOUND"},
			}, nil
		}
		return nil, err
	}
	if reason := campaignEligibilityRejectReason(user, *campaign); reason != "" {
		return &data.HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeNotEligible,
			Message:    MsgUserNotEligible,
			Data:       map[string]any{"reason": reason},
		}, nil
	}

	if existing, err := s.participants.GetByCampaignAndUser(campaignID, userID); err == nil {
		return &data.HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeSuccess,
			Message:    MsgSuccess,
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
	return &data.HTTPReply{
		HTTPStatus: http.StatusOK,
		Code:       data.CodeSuccess,
		Message:    MsgSuccess,
		Data: map[string]any{
			"campaignId":   campaignID,
			"userId":       userID,
			"joinStatus":   model.JoinStatusJoined,
			"taskStatus":   model.TaskStatusNotStarted,
			"rewardStatus": model.RewardStatusNotGranted,
		},
	}, nil
}

func (s *userCampaignService) SimulateTopUp(campaignID, userID int64, amount float64) (*data.HTTPReply, error) {
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
			return &data.HTTPReply{HTTPStatus: http.StatusBadRequest, Code: -1, Message: MsgUserNotFound}, nil
		}
		return nil, err
	}
	rewardAmount, err := calculateTopUpRewardAmount(amount, rules)
	if err != nil {
		return &data.HTTPReply{HTTPStatus: http.StatusBadRequest, Code: -1, Message: err.Error()}, nil
	}

	now := time.Now()
	applyTopUpProgressToParticipant(participant, amount, now)

	recharge, err := s.accounts.Recharge(userID, amount, model.DefaultCurrency)
	if err != nil {
		if data.IsInvalidAccountInput(err) {
			return &data.HTTPReply{HTTPStatus: http.StatusBadRequest, Code: -1, Message: err.Error()}, nil
		}
		return nil, err
	}
	return s.simulateTopUpAfterRecharge(topUpCompletionInput{
		participant: participant, rules: rules, campaignID: campaignID, userID: userID,
		amount: amount, rechargeTxnNo: recharge.TransactionNo, balanceAfter: recharge.BalanceAfter,
		rewardAmount: rewardAmount,
	}, user)
}

// simulateTopUpPrecheck loads campaign and participant and returns an HTTPReply for early business exits.
func (s *userCampaignService) simulateTopUpPrecheck(campaignID, userID int64, amount float64) (
	rules model.RewardRulesPayload,
	participant *model.CampaignParticipant,
	early *data.HTTPReply,
	err error,
) {
	campaign, err := s.campaigns.GetByID(campaignID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return model.RewardRulesPayload{}, nil, &data.HTTPReply{
				HTTPStatus: http.StatusNotFound, Code: -1, Message: MsgCampaignNotFound,
			}, nil
		}
		return model.RewardRulesPayload{}, nil, nil, err
	}
	rules, err = model.ParseRewardRulesJSON(campaign.RewardRules)
	if err != nil {
		return model.RewardRulesPayload{}, nil, &data.HTTPReply{
			HTTPStatus: http.StatusInternalServerError, Code: -1, Message: err.Error(),
		}, nil
	}

	participant, err = s.participants.GetByCampaignAndUser(campaignID, userID)
	if err != nil {
		if mysql.IsNotFound(err) {
			return model.RewardRulesPayload{}, nil, &data.HTTPReply{
				HTTPStatus: http.StatusBadRequest, Code: -1, Message: MsgUserNotJoinedCampaign,
			}, nil
		}
		return model.RewardRulesPayload{}, nil, nil, err
	}
	if participant.RewardStatus == model.RewardStatusGranted {
		return model.RewardRulesPayload{}, nil, &data.HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeDuplicateReward,
			Message:    MsgRewardAlreadyGranted,
			Data: map[string]any{
				"campaignId":   campaignID,
				"userId":       userID,
				"rewardStatus": model.RewardStatusGranted,
			},
		}, nil
	}
	if participant.RewardStatus == model.RewardStatusPending ||
		participant.RewardStatus == model.RewardStatusPendingReview {
		return model.RewardRulesPayload{}, nil, &data.HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeDuplicateReward,
			Message:    MsgRewardAlreadyProcessing,
			Data: map[string]any{
				"campaignId":   campaignID,
				"userId":       userID,
				"rewardStatus": participant.RewardStatus,
			},
		}, nil
	}
	if amount < rules.TopupThreshold {
		return model.RewardRulesPayload{}, nil, &data.HTTPReply{
			HTTPStatus: http.StatusOK,
			Code:       data.CodeTopupNotQualified,
			Message:    MsgTopupAmountNotQualified,
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

type landingPageUIInput struct {
	campaign    *model.Campaign
	campaignID  int64
	userID      int64
	lp          *model.CampaignLandingPage
	lang        string
	title       string
	description string
	terms       string
	rules       model.RewardRulesPayload
}

func (s *userCampaignService) buildLandingPageUIPayload(in landingPageUIInput) (map[string]any, error) {
	var joined bool
	taskStatus := model.TaskStatusNotStarted
	rewardStatus := model.RewardStatusNotGranted
	if in.userID > 0 {
		if p, err := s.participants.GetByCampaignAndUser(in.campaignID, in.userID); err == nil {
			joined = true
			taskStatus = p.TaskStatus
			rewardStatus = p.RewardStatus
		} else if !mysql.IsNotFound(err) {
			return nil, err
		}
	}
	return map[string]any{
		"campaignId":            in.campaignID,
		"campaignName":          in.campaign.Name,
		"campaignType":          in.campaign.Type,
		"status":                in.campaign.Status,
		"registrationStartTime": in.campaign.RegistrationStartTime.Format(time.RFC3339),
		"registrationEndTime":   in.campaign.RegistrationEndTime.Format(time.RFC3339),
		"campaignStartTime":     in.campaign.CampaignStartTime.Format(time.RFC3339),
		"campaignEndTime":       in.campaign.CampaignEndTime.Format(time.RFC3339),
		"landingPage": map[string]any{
			"lang": in.lang, "defaultLang": in.lp.DefaultLang,
			"bannerImageUrl": in.lp.BannerImageURL,
			"title":          in.title, "description": in.description, "terms": in.terms,
		},
		"rewardRule": in.rules,
		"userStatus": map[string]any{
			"joined": joined, "taskStatus": taskStatus, "rewardStatus": rewardStatus,
		},
	}, nil
}

func landingPageUIReply(payload map[string]any) *data.HTTPReply {
	return &data.HTTPReply{HTTPStatus: http.StatusOK, Code: data.CodeSuccess, Message: MsgSuccess, Data: payload}
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

type topUpCompletionInput struct {
	participant   *model.CampaignParticipant
	rules         model.RewardRulesPayload
	campaignID    int64
	userID        int64
	amount        float64
	rechargeTxnNo string
	balanceAfter  float64
	rewardAmount  float64
}

func (s *userCampaignService) simulateTopUpAfterRecharge(
	in topUpCompletionInput, user *model.User,
) (*data.HTTPReply, error) {
	if user.RiskLevel == model.RiskLevelHigh {
		return s.simulateTopUpManualReviewWithAccount(in)
	}
	return s.simulateTopUpEnqueueReward(in)
}

func (s *userCampaignService) simulateTopUpManualReviewWithAccount(in topUpCompletionInput) (*data.HTTPReply, error) {
	in.participant.RiskStatus = model.RiskStatusManualReview
	in.participant.RewardStatus = model.RewardStatusPendingReview
	in.participant.RewardAmount = 0
	if err := s.participants.Save(in.participant); err != nil {
		return nil, err
	}
	return topUpReply(in.campaignID, in.userID, in.amount, in.rechargeTxnNo, in.balanceAfter, map[string]any{
		"taskStatus": model.TaskStatusCompleted, "riskStatus": model.RiskStatusManualReview,
		"rewardStatus": model.RewardStatusPendingReview, "rewardAmount": 0,
	}, MsgManualReviewRequired)
}

func (s *userCampaignService) simulateTopUpEnqueueReward(in topUpCompletionInput) (*data.HTTPReply, error) {
	in.participant.RiskStatus = model.RiskStatusApproved
	in.participant.RewardStatus = model.RewardStatusPending
	in.participant.RewardAmount = in.rewardAmount
	if err := s.participants.Save(in.participant); err != nil {
		return nil, err
	}
	s.rewards.NotifyTopUpReward(event.TopUpRewardEvent{
		CampaignID: in.campaignID, UserID: in.userID, TopupAmount: in.amount,
		ParticipantID: in.participant.ID, ManualReview: false,
		RewardAmount: in.rewardAmount, RewardType: in.rules.RewardType,
	})
	return topUpReply(in.campaignID, in.userID, in.amount, in.rechargeTxnNo, in.balanceAfter, map[string]any{
		"taskStatus": model.TaskStatusCompleted, "riskStatus": model.RiskStatusApproved,
		"rewardStatus": model.RewardStatusPending, "rewardAmount": in.rewardAmount,
	}, MsgRewardProcessing)
}

func calculateTopUpRewardAmount(topupAmount float64, rules model.RewardRulesPayload) (float64, error) {
	mode := strings.TrimSpace(rules.RewardMode)
	if mode == "" {
		mode = model.RewardModeFixedAmount
	}

	var rewardAmount float64
	switch mode {
	case model.RewardModeFixedAmount:
		rewardAmount = rules.RewardAmount
	case model.RewardModePercentage:
		rewardAmount = topupAmount * rules.RewardPercentage / 100
	default:
		return 0, fmt.Errorf(MsgInvalidRewardModeFmt, rules.RewardMode)
	}
	if rules.MaxRewardAmount > 0 && rewardAmount > rules.MaxRewardAmount {
		rewardAmount = rules.MaxRewardAmount
	}
	if rewardAmount < 0 {
		return 0, errors.New(MsgRewardAmountNonNegative)
	}
	return rewardAmount, nil
}

func topUpReply(
	campaignID, userID int64, amount float64,
	rechargeTxnNo string, balanceAfter float64,
	extra map[string]any,
	message string,
) (*data.HTTPReply, error) {
	payload := map[string]any{
		"campaignId": campaignID, "userId": userID, "topupAmount": amount,
		"rechargeTransactionNo": rechargeTxnNo, "balanceAfter": balanceAfter,
	}
	for k, v := range extra {
		payload[k] = v
	}
	return &data.HTTPReply{
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
