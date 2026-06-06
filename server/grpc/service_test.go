package grpc

import (
	"context"
	"testing"

	"github.com/nusiss-capstone-project/campaign-center-api/common/campaignpb"
)

func TestCampaignCenterServiceSayHello(t *testing.T) {
	svc := &CampaignCenterService{}
	response, err := svc.SayHello(context.Background(), &campaignpb.HelloRequest{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.GetMessage() != "Hello world" {
		t.Fatalf("expected default greeting, got %q", response.GetMessage())
	}
}
