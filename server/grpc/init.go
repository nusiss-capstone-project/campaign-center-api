package grpc

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/lianjin/campaign-center-api/common/userpb"
	"github.com/lianjin/campaign-center-api/server/config"
	"github.com/lianjin/campaign-center-api/server/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	grpcpkg "google.golang.org/grpc"
)

func Init(exitSig chan os.Signal) {
	ipPort := fmt.Sprintf("%s:%d", config.Config.GrpcConfig.Host, config.Config.GrpcConfig.Port)
	listener, err := net.Listen("tcp", ipPort)
	if err != nil {
		log.Logger.Errorf("Failed to listen: %v", err)
		if exitSig != nil {
			exitSig <- os.Interrupt
		}
		return
	}
	opts := []grpcpkg.ServerOption{
		grpcpkg.ConnectionTimeout(time.Duration(config.Config.GrpcConfig.ConnectTimeout) * time.Second),
		grpcpkg.MaxConcurrentStreams(uint32(config.Config.GrpcConfig.MaxPoolSize)),
		grpcpkg.MaxRecvMsgSize(1024 * 1024),
		grpcpkg.MaxSendMsgSize(1024 * 1024),
		grpcpkg.StatsHandler(otelgrpc.NewServerHandler()),
	}
	grpcServer := grpcpkg.NewServer(opts...)
	userpb.RegisterCampaignCenterServiceServer(grpcServer, &CampaignCenterService{})

	log.Logger.Infof("gRPC server is running on %s", ipPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Logger.Errorf("Failed to serve gRPC: %v", err)
		if exitSig != nil {
			exitSig <- os.Interrupt
		}
	}
}
