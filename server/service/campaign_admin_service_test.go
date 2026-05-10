package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"github.com/lianjin/campaign-center-api/server/service"
	servicemock "github.com/lianjin/campaign-center-api/server/service/mock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCampaignAdminService_CreateCampaign(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	m.On("Create", mock.MatchedBy(func(c *model.Campaign) bool {
		return c.Name == "c1" && c.Status == model.CampaignStatusDraft
	})).Run(func(args mock.Arguments) {
		c := args.Get(0).(*model.Campaign)
		c.ID = 99
	}).Return(nil)

	svc := service.NewCampaignAdminService(m)
	id, status, err := svc.CreateCampaign(service.CreateCampaignParams{
		Name:                  "c1",
		Type:                  model.CampaignTypeTopupReward,
		TargetMarket:          "US",
		RegistrationStartTime: time.Now(),
		RegistrationEndTime:   time.Now().Add(time.Hour),
		CampaignStartTime:     time.Now(),
		CampaignEndTime:       time.Now().Add(24 * time.Hour),
		TargetUserSegment:     model.UserSegmentNewUser,
		RewardRules: model.RewardRulesPayload{
			TopupThreshold: 100, RewardAmount: 10, RewardType: model.RewardTypeBonusCredit, MaxClaimPerUser: 1,
		},
	})
	require.NoError(t, err)
	require.Equal(t, int64(99), id)
	require.Equal(t, model.CampaignStatusDraft, status)
}

func TestCampaignAdminService_UpdateDraftCampaign_notDraft(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	m.On("GetByID", int64(1)).Return(&model.Campaign{
		ID:     1,
		Status: model.CampaignStatusPublished,
	}, nil)

	svc := service.NewCampaignAdminService(m)
	err := svc.UpdateDraftCampaign(1, service.UpdateCampaignParams{
		Name: "x", TargetMarket: "US",
		RegistrationStartTime: time.Now(), RegistrationEndTime: time.Now().Add(time.Hour),
		CampaignStartTime: time.Now(), CampaignEndTime: time.Now().Add(24 * time.Hour),
		TargetUserSegment: model.UserSegmentNewUser,
		RewardRules: model.RewardRulesPayload{
			TopupThreshold: 1, RewardAmount: 1, RewardType: "X", MaxClaimPerUser: 1,
		},
	})
	require.Error(t, err)
	require.True(t, service.IsCampaignNotDraft(err))
}

func TestCampaignAdminService_UpdateDraftCampaign_success(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	existing := &model.Campaign{
		ID: 1, Status: model.CampaignStatusDraft, Name: "old",
	}
	m.On("GetByID", int64(1)).Return(existing, nil)
	m.On("Update", mock.MatchedBy(func(c *model.Campaign) bool {
		return c.Name == "new"
	})).Return(nil)

	svc := service.NewCampaignAdminService(m)
	err := svc.UpdateDraftCampaign(1, service.UpdateCampaignParams{
		Name: "new", TargetMarket: "US",
		RegistrationStartTime: time.Now(), RegistrationEndTime: time.Now().Add(time.Hour),
		CampaignStartTime: time.Now(), CampaignEndTime: time.Now().Add(24 * time.Hour),
		TargetUserSegment: model.UserSegmentNewUser,
		RewardRules: model.RewardRulesPayload{
			TopupThreshold: 1, RewardAmount: 1, RewardType: "X", MaxClaimPerUser: 1,
		},
	})
	require.NoError(t, err)
}

func TestCampaignAdminService_ListAndGet(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	filter := mysql.CampaignListFilter{Page: 1, PageSize: 10}
	m.On("List", filter).Return([]model.Campaign{{ID: 1}}, int64(1), nil)
	m.On("GetByID", int64(1)).Return(&model.Campaign{ID: 1, Name: "n"}, nil)

	svc := service.NewCampaignAdminService(m)
	items, total, err := svc.ListCampaigns(filter)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), total)

	c, err := svc.GetCampaign(1)
	require.NoError(t, err)
	require.Equal(t, "n", c.Name)
}

func TestCampaignAdminService_PublishCampaign(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	m.On("Publish", int64(1), "op").Return(&model.Campaign{ID: 1, Status: model.CampaignStatusPublished}, nil)

	svc := service.NewCampaignAdminService(m)
	c, err := svc.PublishCampaign(1, "op")
	require.NoError(t, err)
	require.Equal(t, model.CampaignStatusPublished, c.Status)
}

func TestCampaignAdminService_GetByID_notFound(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	m.On("GetByID", int64(404)).Return(nil, gorm.ErrRecordNotFound)

	svc := service.NewCampaignAdminService(m)
	_, err := svc.GetCampaign(404)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCampaignAdminService_CreateCampaign_repoError(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	m.On("Create", mock.Anything).Return(errors.New("db down"))

	svc := service.NewCampaignAdminService(m)
	_, _, err := svc.CreateCampaign(service.CreateCampaignParams{
		Name: "c", Type: "T", TargetMarket: "US",
		RegistrationStartTime: time.Now(), RegistrationEndTime: time.Now().Add(time.Hour),
		CampaignStartTime: time.Now(), CampaignEndTime: time.Now().Add(24 * time.Hour),
		TargetUserSegment: "S",
		RewardRules:       model.RewardRulesPayload{TopupThreshold: 1, RewardAmount: 1, RewardType: "R", MaxClaimPerUser: 1},
	})
	require.Error(t, err)
}
