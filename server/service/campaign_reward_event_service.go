package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/kafka"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	"github.com/nusiss-capstone-project/campaign-center-api/server/proxy"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
)

const (
	rewardResultStatusDistributed = "DISTRIBUTED"
	rewardResultStatusFailed      = "FAILED"
	taskCompletionStatusCompleted = "completed"
)

// TaskCompletedEvent matches task-mservice kafka payload.
type TaskCompletedEvent struct {
	TaskID             int    `json:"task_id"`
	UserID             int    `json:"user_id"`
	Status             string `json:"status"`
	GroupID            int    `json:"group_id"`
	CompletedTaskCount int    `json:"completed_task_count"`
	TotalTaskCount     int    `json:"total_task_count"`
}

// RewardDistributionResultEvent matches reward-mservice kafka payload.
type RewardDistributionResultEvent struct {
	ClientRefID       string `json:"client_ref_id"`
	VoucherID         string `json:"voucher_id"`
	Status            string `json:"status"`
	DistributedAmount string `json:"distributed_amount"`
	FailedReason      string `json:"failed_reason"`
}

// CampaignRewardEventService handles task/reward kafka side effects.
type CampaignRewardEventService interface {
	HandleTaskCompleted(ctx context.Context, event TaskCompletedEvent) error
	HandleRewardResult(ctx context.Context, event RewardDistributionResultEvent) error
}

type campaignRewardEventService struct {
	campaigns    mysql.CampaignRepository
	rules        mysql.CampaignRewardRuleRepository
	participants mysql.ParticipantRepository
	ledgers      mysql.CampaignUserRewardLedgerRepository
	reward       proxy.RewardClient
}

var (
	campaignRewardEventOnce sync.Once
	campaignRewardEventInst CampaignRewardEventService
)

// NewCampaignRewardEventService wires dependencies (for tests).
func NewCampaignRewardEventService(
	campaigns mysql.CampaignRepository,
	rules mysql.CampaignRewardRuleRepository,
	participants mysql.ParticipantRepository,
	ledgers mysql.CampaignUserRewardLedgerRepository,
	reward proxy.RewardClient,
) CampaignRewardEventService {
	return &campaignRewardEventService{
		campaigns:    campaigns,
		rules:        rules,
		participants: participants,
		ledgers:      ledgers,
		reward:       reward,
	}
}

// GetCampaignRewardEventService returns the singleton.
func GetCampaignRewardEventService() CampaignRewardEventService {
	campaignRewardEventOnce.Do(func() {
		campaignRewardEventInst = NewCampaignRewardEventService(
			mysql.GetCampaignRepository(),
			mysql.GetCampaignRewardRuleRepository(),
			mysql.GetParticipantRepository(),
			mysql.GetCampaignUserRewardLedgerRepository(),
			proxy.GetRewardClient(),
		)
	})
	return campaignRewardEventInst
}

func (s *campaignRewardEventService) HandleTaskCompleted(ctx context.Context, event TaskCompletedEvent) error {
	if !strings.EqualFold(event.Status, taskCompletionStatusCompleted) {
		log.WithContext(ctx).Infow("task_completed_ignored_status", "status", event.Status)
		return nil
	}
	userID := int64(event.UserID)
	taskID := int64(event.TaskID)
	groupID := int64(event.GroupID)

	rules := make([]model.CampaignRewardRule, 0)
	if taskID > 0 {
		taskRules, err := s.rules.ListByRef(model.RewardRefClientTask, taskID)
		if err != nil {
			return err
		}
		rules = append(rules, taskRules...)
	}
	if groupID > 0 && event.TotalTaskCount > 0 && event.CompletedTaskCount >= event.TotalTaskCount {
		groupRules, err := s.rules.ListByRef(model.RewardRefClientTaskGroup, groupID)
		if err != nil {
			return err
		}
		rules = append(rules, groupRules...)
	}
	if len(rules) == 0 {
		log.WithContext(ctx).Infow("task_completed_no_rules", "task_id", taskID, "group_id", groupID)
		return nil
	}

	for _, rule := range rules {
		if err := s.maybeDistributeForRule(ctx, userID, rule); err != nil {
			return err
		}
	}
	return nil
}

func (s *campaignRewardEventService) maybeDistributeForRule(ctx context.Context, userID int64, rule model.CampaignRewardRule) error {
	if rule.RewardTemplateID <= 0 {
		return nil
	}
	part, err := s.participants.GetByCampaignAndUser(rule.CampaignID, userID)
	if err != nil {
		return err
	}
	if part == nil {
		log.WithContext(ctx).Infow("task_completed_user_not_joined",
			"campaign_id", rule.CampaignID, "user_id", userID, "rule_id", rule.ID)
		return nil
	}

	existing, err := s.ledgers.GetByUserCampaignRule(userID, rule.CampaignID, rule.ID)
	if err != nil {
		return err
	}
	if existing != nil {
		if strings.TrimSpace(existing.VoucherID) != "" {
			log.WithContext(ctx).Infow("task_completed_duplicate_consumption",
				"ledger_id", existing.ID, "status", existing.RewardStatus, "voucher_id", existing.VoucherID)
			return nil
		}
		// Ledger exists but reward was never accepted (no voucher_id) — retry Distribute.
		log.WithContext(ctx).Infow("task_completed_retry_reward",
			"ledger_id", existing.ID, "status", existing.RewardStatus)
		return s.distributeReward(ctx, userID, rule, existing)
	}

	now := time.Now().UTC()
	ledger := &model.CampaignUserRewardLedger{
		UserID:       userID,
		CampaignID:   rule.CampaignID,
		RuleID:       rule.ID,
		RewardStatus: model.LedgerRewardStatusPendingDistribution,
		VoucherID:    "",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.ledgers.Create(ctx, ledger); err != nil {
		again, getErr := s.ledgers.GetByUserCampaignRule(userID, rule.CampaignID, rule.ID)
		if getErr != nil {
			return err
		}
		if again == nil {
			return err
		}
		if strings.TrimSpace(again.VoucherID) != "" {
			return nil
		}
		return s.distributeReward(ctx, userID, rule, again)
	}

	return s.distributeReward(ctx, userID, rule, ledger)
}

func (s *campaignRewardEventService) distributeReward(
	ctx context.Context, userID int64, rule model.CampaignRewardRule, ledger *model.CampaignUserRewardLedger,
) error {
	campaign, err := s.campaigns.GetByID(rule.CampaignID)
	if err != nil {
		return err
	}
	clientRefID := strconv.FormatInt(ledger.ID, 10)
	result, err := s.reward.Distribute(ctx, proxy.RewardDistributeRequest{
		ClientRefID: clientRefID,
		UserID:      userID,
		ProjectID:   campaign.BudgetProjectID,
		TemplateID:  rule.RewardTemplateID,
	})
	if err != nil {
		_ = s.ledgers.UpdateStatusAndVoucher(ctx, ledger.ID, model.LedgerRewardStatusDistributeFail, "")
		return err
	}
	voucherID := ""
	if result != nil {
		voucherID = result.VoucherID
	}
	return s.ledgers.UpdateStatusAndVoucher(ctx, ledger.ID, model.LedgerRewardStatusDistributing, voucherID)
}

func (s *campaignRewardEventService) HandleRewardResult(ctx context.Context, event RewardDistributionResultEvent) error {
	if strings.TrimSpace(event.ClientRefID) == "" {
		return errors.New("client_ref_id is required")
	}
	ledgerID, err := strconv.ParseInt(event.ClientRefID, 10, 64)
	if err != nil || ledgerID <= 0 {
		return fmt.Errorf("invalid client_ref_id: %s", event.ClientRefID)
	}
	ledger, err := s.ledgers.GetByID(ledgerID)
	if err != nil {
		return err
	}
	status := model.LedgerRewardStatusDistributeFail
	switch strings.ToUpper(event.Status) {
	case rewardResultStatusDistributed:
		status = model.LedgerRewardStatusDistributeSuccess
	case rewardResultStatusFailed:
		status = model.LedgerRewardStatusDistributeFail
	default:
		log.WithContext(ctx).Warnw("reward_result_unknown_status", "status", event.Status, "ledger_id", ledgerID)
		status = model.LedgerRewardStatusDistributeFail
	}
	voucherID := event.VoucherID
	if voucherID == "" {
		voucherID = ledger.VoucherID
	}
	return s.ledgers.UpdateStatusAndVoucher(ctx, ledgerID, status, voucherID)
}

// Parse helpers used by kafka listeners.
func ParseTaskCompletedEvent(msg *kafka.Message) (TaskCompletedEvent, error) {
	var event TaskCompletedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return event, err
	}
	return event, nil
}

func ParseRewardDistributionResultEvent(msg *kafka.Message) (RewardDistributionResultEvent, error) {
	var event RewardDistributionResultEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return event, err
	}
	return event, nil
}
