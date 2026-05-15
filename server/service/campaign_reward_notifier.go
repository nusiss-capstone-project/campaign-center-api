package service

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
	n.ch <- event
}
