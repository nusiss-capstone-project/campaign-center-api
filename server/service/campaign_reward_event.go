package service

// TopUpRewardEvent is dispatched after a successful campaign top-up recharge.
type TopUpRewardEvent struct {
	CampaignID    int64
	UserID        int64
	TopupAmount   float64
	ParticipantID int64
	ManualReview  bool
	RewardAmount  float64
	RewardType    string
}
