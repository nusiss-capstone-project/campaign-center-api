package service

import (
	"net/http"

	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

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
		"rewardStatus": model.RewardStatusGranted, "rewardAmount": rules.RewardAmount,
	}, "success")
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
