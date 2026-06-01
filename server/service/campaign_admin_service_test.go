package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
	"github.com/lianjin/campaign-center-api/server/service"
	servicemock "github.com/lianjin/campaign-center-api/server/mock"
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
	require.True(t, data.IsCampaignNotDraft(err))
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

func TestCampaignAdminService_ArchiveCampaign_draft(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	m.On("GetByID", int64(1)).Return(&model.Campaign{ID: 1, Status: model.CampaignStatusDraft}, nil)
	m.On("Archive", int64(1), "op").Return(&model.Campaign{ID: 1, Status: model.CampaignStatusArchive}, nil)

	svc := service.NewCampaignAdminService(m)
	c, err := svc.ArchiveCampaign(1, "op")
	require.NoError(t, err)
	require.Equal(t, model.CampaignStatusArchive, c.Status)
}

func TestCampaignAdminService_ArchiveCampaign_publishedAfterEnd(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	pastStart := time.Now().Add(-48 * time.Hour)
	pastEnd := time.Now().Add(-time.Hour)
	m.On("GetByID", int64(2)).Return(&model.Campaign{
		ID: 2, Status: model.CampaignStatusPublished,
		CampaignStartTime: pastStart, CampaignEndTime: pastEnd,
	}, nil)
	m.On("Archive", int64(2), "op").Return(&model.Campaign{ID: 2, Status: model.CampaignStatusArchive}, nil)

	svc := service.NewCampaignAdminService(m)
	c, err := svc.ArchiveCampaign(2, "op")
	require.NoError(t, err)
	require.Equal(t, model.CampaignStatusArchive, c.Status)
}

func TestCampaignAdminService_ArchiveCampaign_publishedNotYetStarted(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	futureStart := time.Now().Add(time.Hour)
	futureEnd := time.Now().Add(24 * time.Hour)
	m.On("GetByID", int64(5)).Return(&model.Campaign{
		ID: 5, Status: model.CampaignStatusPublished,
		CampaignStartTime: futureStart, CampaignEndTime: futureEnd,
	}, nil)
	m.On("Archive", int64(5), "op").Return(&model.Campaign{ID: 5, Status: model.CampaignStatusArchive}, nil)

	svc := service.NewCampaignAdminService(m)
	c, err := svc.ArchiveCampaign(5, "op")
	require.NoError(t, err)
	require.Equal(t, model.CampaignStatusArchive, c.Status)
}

func TestCampaignAdminService_ArchiveCampaign_publishedDuringActivity(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)
	m.On("GetByID", int64(3)).Return(&model.Campaign{
		ID: 3, Status: model.CampaignStatusPublished,
		CampaignStartTime: past, CampaignEndTime: future,
	}, nil)

	svc := service.NewCampaignAdminService(m)
	_, err := svc.ArchiveCampaign(3, "op")
	require.Error(t, err)
	require.True(t, data.IsCampaignNotArchivable(err))
}

func TestCampaignAdminService_ArchiveCampaign_alreadyArchived(t *testing.T) {
	m := servicemock.NewMockCampaignRepository(t)
	m.On("GetByID", int64(4)).Return(&model.Campaign{ID: 4, Status: model.CampaignStatusArchive}, nil)

	svc := service.NewCampaignAdminService(m)
	_, err := svc.ArchiveCampaign(4, "op")
	require.Error(t, err)
	require.True(t, data.IsCampaignAlreadyArchived(err))
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
