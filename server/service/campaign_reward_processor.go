package service

import (
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/event"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
)

// CampaignRewardProcessor grants campaign rewards asynchronously.
type CampaignRewardProcessor struct {
	participants mysql.ParticipantRepository
	rewardTx     mysql.RewardTransactionRepository
	accounts     AccountService
	performance  mysql.CampaignPerformanceRepository
}

// NewCampaignRewardProcessor builds a processor with explicit dependencies (for tests).
func NewCampaignRewardProcessor(
	participants mysql.ParticipantRepository,
	rewardTx mysql.RewardTransactionRepository,
	accounts AccountService,
	performance mysql.CampaignPerformanceRepository,
) *CampaignRewardProcessor {
	return &CampaignRewardProcessor{
		participants: participants,
		rewardTx:     rewardTx,
		accounts:     accounts,
		performance:  performance,
	}
}

var (
	rewardProcessorOnce sync.Once
	rewardProcessorInst *CampaignRewardProcessor
	rewardNotifierInst  event.CampaignRewardNotifier
)

// GetCampaignRewardNotifier returns the singleton notifier (starts worker on first call).
func GetCampaignRewardNotifier() event.CampaignRewardNotifier {
	rewardProcessorOnce.Do(func() {
		rewardProcessorInst = &CampaignRewardProcessor{
			participants: mysql.GetParticipantRepository(),
			rewardTx:     mysql.GetRewardTransactionRepository(),
			accounts:     GetAccountService(),
			performance:  mysql.GetCampaignPerformanceRepository(),
		}
		ch := make(chan event.TopUpRewardEvent, 256)
		rewardNotifierInst = event.NewChannelCampaignRewardNotifier(ch)
		go rewardProcessorInst.runWorker(ch)
	})
	return rewardNotifierInst
}

// NewCampaignRewardNotifierForTest runs reward handling synchronously.
func NewCampaignRewardNotifierForTest(p *CampaignRewardProcessor) event.CampaignRewardNotifier {
	return event.NewSyncCampaignRewardNotifier(p.HandleTopUpReward)
}

func (p *CampaignRewardProcessor) runWorker(ch <-chan event.TopUpRewardEvent) {
	for event := range ch {
		if err := p.HandleTopUpReward(event); err != nil {
			log.Logger.Errorw("campaign_reward_event_failed",
				"error", err,
				"campaign_id", event.CampaignID,
				"user_id", event.UserID,
				"participant_id", event.ParticipantID,
				"topup_amount", event.TopupAmount,
				"reward_amount", event.RewardAmount,
				"reward_type", event.RewardType,
				"manual_review", event.ManualReview,
			)
		}
	}
}

// HandleTopUpReward processes one top-up reward event.
func (p *CampaignRewardProcessor) HandleTopUpReward(ev event.TopUpRewardEvent) error {
	if ev.ManualReview {
		return nil
	}
	participant, err := p.participants.GetByCampaignAndUser(ev.CampaignID, ev.UserID)
	if err != nil {
		return err
	}
	if participant.RewardStatus == model.RewardStatusGranted {
		return nil
	}
	now := time.Now()
	participant.RiskStatus = model.RiskStatusApproved
	participant.RewardStatus = model.RewardStatusGranted
	participant.RewardAmount = ev.RewardAmount
	participant.RewardedAt = &now
	participant.UpdatedAt = now

	rewardRow := model.RewardTransaction{
		CampaignID: ev.CampaignID, UserID: ev.UserID,
		ParticipantID: participant.ID, RewardType: ev.RewardType,
		RewardAmount: ev.RewardAmount, Status: model.RewardTxnStatusCompleted,
		CreatedAt: now,
	}
	if err := p.rewardTx.CommitGrantWithParticipant(participant, &rewardRow); err != nil {
		return err
	}
	if _, err := p.accounts.CreditCampaignReward(
		ev.UserID, ev.CampaignID, ev.RewardAmount, model.DefaultCurrency,
	); err != nil {
		return err
	}
	return p.performance.IncrementRewardIssued(
		ev.CampaignID, now, ev.RewardAmount, model.DefaultCurrency,
	)
}
