package proxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/nusiss-capstone-project/campaign-center-api/server/config"
	rewardclient "github.com/nusiss-capstone-project/reward-mservice/client"
	"github.com/nusiss-capstone-project/reward-mservice/common/rewardpb"
)

// RewardDistributeRequest is the outbound reward call payload.
type RewardDistributeRequest struct {
	ClientRefID string
	UserID      int64
	ProjectID   int64
	TemplateID  int64
}

// RewardDistributeResult is the sync response from reward-mservice.
type RewardDistributeResult struct {
	ClientRefID string
	VoucherID   string
}

// RewardClient wraps RewardDistribution.
type RewardClient interface {
	Distribute(ctx context.Context, req RewardDistributeRequest) (*RewardDistributeResult, error)
}

type rewardClient struct {
	raw rewardpb.RewardServiceClient
}

var (
	rewardOnce sync.Once
	rewardInst RewardClient
)

// GetRewardClient returns the singleton reward gRPC client.
func GetRewardClient() RewardClient {
	rewardOnce.Do(func() {
		cfg := config.Config.RewardGrpc
		if cfg == nil {
			cfg = &config.GrpcClientConfig{Host: "127.0.0.1", Port: 50053}
		}
		raw, err := rewardclient.GetRewardServiceClient(&rewardclient.GRpcClientConfig{
			Host: cfg.Host,
			Port: cfg.Port,
		})
		if err != nil {
			panic(err)
		}
		rewardInst = &rewardClient{raw: raw}
	})
	return rewardInst
}

// NewRewardClientForTest wires a stub/real client for tests.
func NewRewardClientForTest(raw rewardpb.RewardServiceClient) RewardClient {
	return &rewardClient{raw: raw}
}

func (c *rewardClient) Distribute(ctx context.Context, req RewardDistributeRequest) (*RewardDistributeResult, error) {
	resp, err := c.raw.Reward(ctx, &rewardpb.RewardDistributionRequest{
		ClientRefId: req.ClientRefID,
		UserId:      uint64(req.UserID),
		ProjectId:   uint64(req.ProjectID),
		TemplateId:  uint64(req.TemplateID),
		BaseInfo: &rewardpb.BaseRequestInfo{
			From: "campaign-center-api",
			To:   "reward-mservice",
		},
	})
	if err != nil {
		return nil, err
	}
	if info := resp.GetBaseInfo(); info != nil && info.GetCode() != rewardpb.ErrorCode_OK {
		return nil, fmt.Errorf("reward Distribute failed: code=%v message=%s", info.GetCode(), info.GetMessage())
	}
	return &RewardDistributeResult{
		ClientRefID: resp.GetClientRefId(),
		VoucherID:   resp.GetVoucherId(),
	}, nil
}
