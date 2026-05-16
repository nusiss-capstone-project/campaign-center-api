package service

import (
	"time"

	"github.com/lianjin/campaign-center-api/server/log"
)

const rewardNotifyTimeout = 100 * time.Millisecond

// CampaignRewardNotifier enqueues campaign reward processing (async in production).
type CampaignRewardNotifier interface {
	NotifyTopUpReward(event TopUpRewardEvent)
}

type syncCampaignRewardNotifier struct {
	handler func(TopUpRewardEvent) error
}

func (n *syncCampaignRewardNotifier) NotifyTopUpReward(event TopUpRewardEvent) {
	_ = n.handler(event)
}

type channelCampaignRewardNotifier struct {
	ch chan TopUpRewardEvent
}

func (n *channelCampaignRewardNotifier) NotifyTopUpReward(event TopUpRewardEvent) {
	select {
	case n.ch <- event:
	case <-time.After(rewardNotifyTimeout):
		log.Logger.Errorw("campaign_reward_event_enqueue_timeout",
			"campaign_id", event.CampaignID,
			"user_id", event.UserID,
			"participant_id", event.ParticipantID,
			"reward_amount", event.RewardAmount,
			"reward_type", event.RewardType,
		)
	}
}
