package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/lianjin/campaign-center-api/common/userpb"
)

type CampaignCenterService struct {
	userpb.UnimplementedCampaignCenterServiceServer
}

func (s *CampaignCenterService) SayHello(_ context.Context, req *userpb.HelloRequest) (*userpb.HelloResponse, error) {
	name := strings.TrimSpace(req.GetName())
	if name == "" {
		name = "world"
	}
	return &userpb.HelloResponse{Message: fmt.Sprintf("Hello %s", name)}, nil
}
