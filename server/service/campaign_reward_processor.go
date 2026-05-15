package service

import (
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// CampaignRewardProcessor grants campaign rewards asynchronously.
type CampaignRewardProcessor struct {
	participants mysql.ParticipantRepository
	rewardTx     mysql.RewardTransactionRepository
	accounts     AccountService
	performance  mysql.CampaignPerformanceRepository
}

var (
	rewardProcessorOnce sync.Once
	rewardProcessorInst *CampaignRewardProcessor
	rewardNotifierInst  CampaignRewardNotifier
)

// GetCampaignRewardNotifier returns the singleton notifier (starts worker on first call).
func GetCampaignRewardNotifier() CampaignRewardNotifier {
	rewardProcessorOnce.Do(func() {
		rewardProcessorInst = &CampaignRewardProcessor{
			participants: mysql.GetParticipantRepository(),
			rewardTx:     mysql.GetRewardTransactionRepository(),
			accounts:     GetAccountService(),
			performance:  mysql.GetCampaignPerformanceRepository(),
		}
		ch := make(chan TopUpRewardEvent, 256)
		rewardNotifierInst = &channelCampaignRewardNotifier{ch: ch}
		go rewardProcessorInst.runWorker(ch)
	})
	return rewardNotifierInst
}

// NewCampaignRewardNotifierForTest runs reward handling synchronously.
func NewCampaignRewardNotifierForTest(p *CampaignRewardProcessor) CampaignRewardNotifier {
	return &syncCampaignRewardNotifier{handler: p.HandleTopUpReward}
}

func (p *CampaignRewardProcessor) runWorker(ch <-chan TopUpRewardEvent) {
	for event := range ch {
		_ = p.HandleTopUpReward(event)
	}
}

// HandleTopUpReward processes one top-up reward event.
func (p *CampaignRewardProcessor) HandleTopUpReward(event TopUpRewardEvent) error {
	if event.ManualReview {
		return nil
	}
	participant, err := p.participants.GetByCampaignAndUser(event.CampaignID, event.UserID)
	if err != nil {
		return err
	}
	if participant.RewardStatus == model.RewardStatusGranted {
		return nil
	}
	now := time.Now()
	participant.RiskStatus = model.RiskStatusApproved
	participant.RewardStatus = model.RewardStatusGranted
	participant.RewardAmount = event.RewardAmount
	participant.RewardedAt = &now
	participant.UpdatedAt = now

	rewardRow := model.RewardTransaction{
		CampaignID: event.CampaignID, UserID: event.UserID,
		ParticipantID: participant.ID, RewardType: event.RewardType,
		RewardAmount: event.RewardAmount, Status: model.RewardTxnStatusCompleted,
		CreatedAt: now,
	}
	if err := p.rewardTx.CommitGrantWithParticipant(participant, &rewardRow); err != nil {
		return err
	}
	if _, err := p.accounts.CreditCampaignReward(
		event.UserID, event.CampaignID, event.RewardAmount, model.DefaultCurrency,
	); err != nil {
		return err
	}
	return p.performance.IncrementRewardIssued(
		event.CampaignID, now, event.RewardAmount, model.DefaultCurrency,
	)
}
