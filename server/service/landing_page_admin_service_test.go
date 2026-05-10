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

func TestLandingPageAdminService_CreateLandingPage(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	m.On("Create", mock.MatchedBy(func(p *model.CampaignLandingPage) bool {
		return p.Language == "en-US"
	})).Run(func(args mock.Arguments) {
		p := args.Get(0).(*model.CampaignLandingPage)
		p.ID = 7
	}).Return(nil)

	svc := service.NewLandingPageAdminService(m)
	id, status, err := svc.CreateLandingPage(service.CreateLandingPageParams{
		Language: "en-US", BannerImageURL: "u", Title: "t", Description: "d", Terms: "x",
	})
	require.NoError(t, err)
	require.Equal(t, int64(7), id)
	require.Equal(t, model.LandingPageStatusDraft, status)
}

func TestLandingPageAdminService_UpdateDraft_notDraft(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	m.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{
		ID: 1, Status: model.LandingPageStatusPublished,
	}, nil)

	svc := service.NewLandingPageAdminService(m)
	err := svc.UpdateDraftLandingPage(1, service.CreateLandingPageParams{
		Language: "en", BannerImageURL: "u", Title: "t", Description: "d", Terms: "x",
	})
	require.Error(t, err)
	require.True(t, service.IsLandingPageNotDraft(err))
}

func TestLandingPageAdminService_UpdateDraft_success(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	row := &model.CampaignLandingPage{ID: 2, Status: model.LandingPageStatusDraft}
	m.On("GetByID", int64(2)).Return(row, nil)
	m.On("Update", mock.MatchedBy(func(p *model.CampaignLandingPage) bool {
		return p.Title == "new"
	})).Return(nil)

	svc := service.NewLandingPageAdminService(m)
	err := svc.UpdateDraftLandingPage(2, service.CreateLandingPageParams{
		Language: "en", BannerImageURL: "u", Title: "new", Description: "d", Terms: "x",
	})
	require.NoError(t, err)
}

func TestLandingPageAdminService_ListGetPublish(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	f := mysql.LandingPageListFilter{Page: 1, PageSize: 5}
	m.On("List", f).Return([]model.CampaignLandingPage{{ID: 1}}, int64(1), nil)
	m.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{ID: 1, Title: "x"}, nil)
	m.On("Publish", int64(1), "op").Return(&model.CampaignLandingPage{ID: 1, Status: model.LandingPageStatusPublished}, nil)

	svc := service.NewLandingPageAdminService(m)
	items, total, err := svc.ListLandingPages(f)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), total)

	p, err := svc.GetLandingPage(1)
	require.NoError(t, err)
	require.Equal(t, "x", p.Title)

	pub, err := svc.PublishLandingPage(1, "op")
	require.NoError(t, err)
	require.Equal(t, model.LandingPageStatusPublished, pub.Status)
}

func TestLandingPageAdminService_Get_notFound(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	m.On("GetByID", int64(9)).Return(nil, gorm.ErrRecordNotFound)

	svc := service.NewLandingPageAdminService(m)
	_, err := svc.GetLandingPage(9)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestLandingPageAdminService_Create_error(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	m.On("Create", mock.Anything).Return(errors.New("fail"))

	svc := service.NewLandingPageAdminService(m)
	_, _, err := svc.CreateLandingPage(service.CreateLandingPageParams{
		Language: "en", BannerImageURL: "u", Title: "t", Description: "d", Terms: "x",
	})
	require.Error(t, err)
}
