package listener

import (
	"context"

	"github.com/nusiss-capstone-project/campaign-center-api/server/kafka"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
)

func handleTaskCompleted(ctx context.Context, msg *kafka.Message) error {
	event, err := service.ParseTaskCompletedEvent(msg)
	if err != nil {
		log.WithContext(ctx).Errorw("task_completed_unmarshal_failed", "error", err)
		return err
	}
	log.WithContext(ctx).Infow("task_completed_received",
		"task_id", event.TaskID, "user_id", event.UserID, "group_id", event.GroupID, "status", event.Status)
	return service.GetCampaignRewardEventService().HandleTaskCompleted(ctx, event)
}

func handleRewardDistributionResult(ctx context.Context, msg *kafka.Message) error {
	event, err := service.ParseRewardDistributionResultEvent(msg)
	if err != nil {
		log.WithContext(ctx).Errorw("reward_result_unmarshal_failed", "error", err)
		return err
	}
	log.WithContext(ctx).Infow("reward_result_received",
		"client_ref_id", event.ClientRefID, "status", event.Status, "voucher_id", event.VoucherID)
	return service.GetCampaignRewardEventService().HandleRewardResult(ctx, event)
}
