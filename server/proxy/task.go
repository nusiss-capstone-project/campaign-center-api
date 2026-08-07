package proxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/nusiss-capstone-project/campaign-center-api/server/config"
	taskclient "github.com/nusiss-capstone-project/task-mservice/client"
	"github.com/nusiss-capstone-project/task-mservice/common/taskpb"
)

// TaskClient wraps EnrollTask.
type TaskClient interface {
	EnrollTaskGroup(ctx context.Context, userID, taskGroupID int64) error
}

type taskClient struct {
	raw taskpb.TaskServiceClient
}

var (
	taskOnce sync.Once
	taskInst TaskClient
)

// GetTaskClient returns the singleton task gRPC client.
func GetTaskClient() TaskClient {
	taskOnce.Do(func() {
		cfg := config.Config.TaskGrpc
		if cfg == nil {
			cfg = &config.GrpcClientConfig{Host: "127.0.0.1", Port: 50052}
		}
		raw, err := taskclient.GetTaskServiceClient(&taskclient.GRpcClientConfig{
			Host: cfg.Host,
			Port: cfg.Port,
		})
		if err != nil {
			panic(err)
		}
		taskInst = &taskClient{raw: raw}
	})
	return taskInst
}

// NewTaskClientForTest wires a stub/real client for tests.
func NewTaskClientForTest(raw taskpb.TaskServiceClient) TaskClient {
	return &taskClient{raw: raw}
}

func (c *taskClient) EnrollTaskGroup(ctx context.Context, userID, taskGroupID int64) error {
	if taskGroupID <= 0 {
		return fmt.Errorf("task_group_id is required")
	}
	return withGRPCCall(ctx, "task.EnrollTask", []any{
		"user_id", userID, "task_group_id", taskGroupID,
	}, func(callCtx context.Context) error {
		resp, err := c.raw.EnrollTask(callCtx, &taskpb.EnrollTaskRequest{
			UserId:      userID,
			TaskGroupId: taskGroupID,
		})
		if err != nil {
			return err
		}
		if base := resp.GetBase(); base != nil && base.GetCode() != taskpb.ErrorCode_OK {
			return fmt.Errorf("task EnrollTask failed: code=%v message=%s", base.GetCode(), base.GetMessage())
		}
		return nil
	})
}
