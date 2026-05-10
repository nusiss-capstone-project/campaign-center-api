package client

import (
	"fmt"
	"sync"

	"github.com/lianjin/campaign-center-api/common/userpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	conn           *grpc.ClientConn
	client         userpb.CampaignCenterServiceClient
	clientErr      error
	clientSyncOnce sync.Once
)

func GetCampaignCenterServiceClient(config *GRPCClientConfig) (userpb.CampaignCenterServiceClient, error) {
	clientSyncOnce.Do(func() {
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(1024 * 1024)),
		}
		conn, clientErr = grpc.NewClient(fmt.Sprintf("%s:%d", config.Host, config.Port), opts...)
		if clientErr != nil {
			return
		}
		client = userpb.NewCampaignCenterServiceClient(conn)
	})
	return client, clientErr
}

func Destroy() error {
	if conn == nil {
		return nil
	}
	return conn.Close()
}
