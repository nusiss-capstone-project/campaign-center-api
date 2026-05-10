package grpc

import (
	"context"
	"testing"

	"github.com/lianjin/campaign-center-api/common/userpb"
)

func TestCampaignCenterServiceSayHello(t *testing.T) {
	svc := &CampaignCenterService{}
	response, err := svc.SayHello(context.Background(), &userpb.HelloRequest{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.GetMessage() != "Hello world" {
		t.Fatalf("expected default greeting, got %q", response.GetMessage())
	}
}
