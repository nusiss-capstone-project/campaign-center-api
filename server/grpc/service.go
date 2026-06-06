package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/nusiss-capstone-project/campaign-center-api/common/campaignpb"
)

type CampaignCenterService struct {
	campaignpb.UnimplementedCampaignCenterServiceServer
}

func (s *CampaignCenterService) SayHello(_ context.Context, req *campaignpb.HelloRequest) (*campaignpb.HelloResponse, error) {
	name := strings.TrimSpace(req.GetName())
	if name == "" {
		name = "world"
	}
	return &campaignpb.HelloResponse{Message: fmt.Sprintf("Hello %s", name)}, nil
}
