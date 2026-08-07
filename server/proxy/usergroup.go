package proxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/nusiss-capstone-project/campaign-center-api/server/config"
	ugclient "github.com/nusiss-capstone-project/usergroup-mservice/client"
	"github.com/nusiss-capstone-project/usergroup-mservice/common/usergrouppb"
)

// UsergroupClient wraps MatchUserGroup.
type UsergroupClient interface {
	MatchUserGroup(ctx context.Context, userID, userGroupID int64) (bool, error)
}

type usergroupClient struct {
	raw usergrouppb.UsergroupServiceClient
}

var (
	usergroupOnce sync.Once
	usergroupInst UsergroupClient
)

// GetUsergroupClient returns the singleton usergroup gRPC client.
func GetUsergroupClient() UsergroupClient {
	usergroupOnce.Do(func() {
		cfg := config.Config.UsergroupGrpc
		if cfg == nil {
			cfg = &config.GrpcClientConfig{Host: "127.0.0.1", Port: 50051}
		}
		raw, err := ugclient.GetUsergroupServiceClient(&ugclient.GRpcClientConfig{
			Host: cfg.Host,
			Port: cfg.Port,
		})
		if err != nil {
			panic(err)
		}
		usergroupInst = &usergroupClient{raw: raw}
	})
	return usergroupInst
}

// NewUsergroupClientForTest wires a stub/real client for tests.
func NewUsergroupClientForTest(raw usergrouppb.UsergroupServiceClient) UsergroupClient {
	return &usergroupClient{raw: raw}
}

func (c *usergroupClient) MatchUserGroup(ctx context.Context, userID, userGroupID int64) (bool, error) {
	if userGroupID == 0 {
		return true, nil
	}
	var matched bool
	err := withGRPCCall(ctx, "usergroup.MatchUserGroup", []any{
		"user_id", userID, "user_group_id", userGroupID,
	}, func(callCtx context.Context) error {
		resp, err := c.raw.MatchUserGroup(callCtx, &usergrouppb.MatchUserGroupRequest{
			UserId:      userID,
			UserGroupId: userGroupID,
		})
		if err != nil {
			return err
		}
		if info := resp.GetBaseResponseInfo(); info != nil && info.GetCode() != usergrouppb.ErrorCode_OK {
			return fmt.Errorf("usergroup MatchUserGroup failed: code=%v message=%s", info.GetCode(), info.GetMessage())
		}
		matched = resp.GetMatched()
		return nil
	})
	return matched, err
}
