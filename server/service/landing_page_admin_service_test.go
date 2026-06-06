package service_test

import (
	"errors"
	"testing"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	servicemock "github.com/nusiss-capstone-project/campaign-center-api/server/mock"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func noopTrans() mysql.LandingPageTranslationRepository {
	return mysql.NewNoopLandingPageTranslationRepository()
}

func TestLandingPageAdminService_CreateLandingPage(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	m.On("Create", mock.MatchedBy(func(p *model.CampaignLandingPage) bool {
		return p.DefaultLang == "en-US"
	})).Run(func(args mock.Arguments) {
		p := args.Get(0).(*model.CampaignLandingPage)
		p.ID = 7
	}).Return(nil)

	svc := service.NewLandingPageAdminService(m, noopTrans())
	id, status, err := svc.CreateLandingPage(service.CreateLandingPageParams{
		DefaultLang: "en-US", BannerImageURL: "u", Title: "t", Description: "d", Terms: "x",
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

	svc := service.NewLandingPageAdminService(m, noopTrans())
	err := svc.UpdateDraftLandingPage(1, service.CreateLandingPageParams{
		DefaultLang: "en", BannerImageURL: "u", Title: "t", Description: "d", Terms: "x",
	})
	require.Error(t, err)
	require.True(t, data.IsLandingPageNotDraft(err))
}

func TestLandingPageAdminService_UpdateDraft_success(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	row := &model.CampaignLandingPage{ID: 2, Status: model.LandingPageStatusDraft}
	m.On("GetByID", int64(2)).Return(row, nil)
	m.On("Update", mock.MatchedBy(func(p *model.CampaignLandingPage) bool {
		return p.Title == "new"
	})).Return(nil)

	svc := service.NewLandingPageAdminService(m, noopTrans())
	err := svc.UpdateDraftLandingPage(2, service.CreateLandingPageParams{
		DefaultLang: "en", BannerImageURL: "u", Title: "new", Description: "d", Terms: "x",
	})
	require.NoError(t, err)
}

func TestLandingPageAdminService_ListGetPublish(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	f := mysql.LandingPageListFilter{Page: 1, PageSize: 5}
	m.On("List", f).Return([]model.CampaignLandingPage{{ID: 1}}, int64(1), nil)
	m.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{
		ID: 1, Title: "x", DefaultLang: "en",
	}, nil)
	m.On("Publish", int64(1), "op").Return(&model.CampaignLandingPage{ID: 1, Status: model.LandingPageStatusPublished}, nil)

	svc := service.NewLandingPageAdminService(m, noopTrans())
	items, total, err := svc.ListLandingPages(f)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), total)

	p, err := svc.GetLandingPage(1, "")
	require.NoError(t, err)
	require.Equal(t, "x", p.Title)
	require.Equal(t, "en", p.Lang)

	pub, err := svc.PublishLandingPage(1, "op")
	require.NoError(t, err)
	require.Equal(t, model.LandingPageStatusPublished, pub.Status)
}

func TestLandingPageAdminService_Get_notFound(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	m.On("GetByID", int64(9)).Return(nil, gorm.ErrRecordNotFound)

	svc := service.NewLandingPageAdminService(m, noopTrans())
	_, err := svc.GetLandingPage(9, "")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestLandingPageAdminService_GetLandingPage_usesTranslation(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	m.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{
		ID: 1, DefaultLang: "en", Title: "default title", Description: "d", Terms: "t",
	}, nil)
	trans := staticLandingPageTranslationRepo{row: &model.CampaignLandingPageTranslation{
		LandingPageID: 1, Lang: "zh-CN",
		Title: "中文标题", Description: "中文描述", Terms: "中文条款",
	}}
	svc := service.NewLandingPageAdminService(m, trans)

	view, err := svc.GetLandingPage(1, "zh-CN")

	require.NoError(t, err)
	require.Equal(t, "zh-CN", view.Lang)
	require.Equal(t, "中文标题", view.Title)
	require.Equal(t, "中文描述", view.Description)
	require.Equal(t, "中文条款", view.Terms)
}

func TestLandingPageAdminService_Create_error(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	m.On("Create", mock.Anything).Return(errors.New("fail"))

	svc := service.NewLandingPageAdminService(m, noopTrans())
	_, _, err := svc.CreateLandingPage(service.CreateLandingPageParams{
		DefaultLang: "en", BannerImageURL: "u", Title: "t", Description: "d", Terms: "x",
	})
	require.Error(t, err)
}
