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

type staticLandingPageTranslationRepo struct {
	row *model.CampaignLandingPageTranslation
}

func (r staticLandingPageTranslationRepo) GetByLandingPageAndLang(landingPageID int64, lang string) (*model.CampaignLandingPageTranslation, error) {
	if r.row == nil || r.row.LandingPageID != landingPageID || r.row.Lang != lang {
		return nil, nil
	}
	return r.row, nil
}

func (r staticLandingPageTranslationRepo) ListLangsByLandingPageID(landingPageID int64) ([]string, error) {
	if r.row == nil || r.row.LandingPageID != landingPageID {
		return nil, nil
	}
	return []string{r.row.Lang}, nil
}

func (r staticLandingPageTranslationRepo) Upsert(*model.CampaignLandingPageTranslation) error {
	return nil
}

func sampleLandingBody() data.LandingPageBody {
	return data.LandingPageBody{
		DefaultLang: "en-US", BannerImageURL: "u", Title: "t", Description: "d", Terms: "x",
		Steps: []data.LandingPageRepeatableItemVO{{Title: "s1", Description: "sd1"}},
		Faq:   []data.LandingPageRepeatableItemVO{{Title: "q1", Description: "a1"}},
	}
}

func TestLandingPageAdminService_CreateLandingPage(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	m.On("Create", mock.MatchedBy(func(p *model.CampaignLandingPage) bool {
		return p.DefaultLang == "en-US" && len(p.Steps) == 1 && p.Steps[0].Title == "s1"
	})).Run(func(args mock.Arguments) {
		p := args.Get(0).(*model.CampaignLandingPage)
		p.ID = 7
	}).Return(nil)

	svc := service.NewLandingPageAdminService(m, noopTrans())
	resp, err := svc.CreateLandingPage(sampleLandingBody())
	require.NoError(t, err)
	require.Equal(t, int64(7), resp.LandingPageID)
	require.Equal(t, model.LandingPageStatusDraft, resp.Status)
	require.Equal(t, "s1", resp.Steps[0].Title)
}

func TestLandingPageAdminService_UpdateDraft_notDraft(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	m.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{
		ID: 1, Status: model.LandingPageStatusPublished,
	}, nil)

	svc := service.NewLandingPageAdminService(m, noopTrans())
	_, err := svc.UpdateDraftLandingPage(1, data.LandingPageBody{
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
		return p.Title == "new" && len(p.Faq) == 1
	})).Return(nil)

	svc := service.NewLandingPageAdminService(m, noopTrans())
	body := data.LandingPageBody{
		DefaultLang: "en", BannerImageURL: "u", Title: "new", Description: "d", Terms: "x",
		Faq: []data.LandingPageRepeatableItemVO{{Title: "q", Description: "a"}},
	}
	resp, err := svc.UpdateDraftLandingPage(2, body)
	require.NoError(t, err)
	require.Equal(t, int64(2), resp.LandingPageID)
	require.Equal(t, "new", resp.Title)
}

func TestLandingPageAdminService_ListGetPublish(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	f := mysql.LandingPageListFilter{Page: 1, PageSize: 5}
	m.On("List", f).Return([]model.CampaignLandingPage{{ID: 1, Title: "x", Status: 1}}, int64(1), nil)
	m.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{
		ID: 1, Title: "x", DefaultLang: "en",
	}, nil)
	m.On("Publish", int64(1), "op").Return(&model.CampaignLandingPage{ID: 1, Status: model.LandingPageStatusPublished}, nil)

	svc := service.NewLandingPageAdminService(m, noopTrans())
	list, err := svc.ListLandingPages(f)
	require.NoError(t, err)
	require.Len(t, list.Items, 1)
	require.Equal(t, int64(1), list.Total)
	require.Equal(t, "x", list.Items[0].Title)

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
		Steps: model.LandingPageRepeatableItems{{Title: "en-step", Description: "en-desc"}},
	}, nil)
	trans := staticLandingPageTranslationRepo{row: &model.CampaignLandingPageTranslation{
		LandingPageID: 1, Lang: "zh-CN",
		Title: "中文标题", Description: "中文描述", Terms: "中文条款",
		Steps: model.LandingPageRepeatableItems{{Title: "中文步骤", Description: "中文说明"}},
	}}
	svc := service.NewLandingPageAdminService(m, trans)

	view, err := svc.GetLandingPage(1, "zh-CN")

	require.NoError(t, err)
	require.Equal(t, "zh-CN", view.Lang)
	require.Equal(t, "中文标题", view.Title)
	require.Equal(t, "中文描述", view.Description)
	require.Equal(t, "中文条款", view.Terms)
	require.Equal(t, "中文步骤", view.Steps[0].Title)
}

func TestLandingPageAdminService_Create_error(t *testing.T) {
	m := servicemock.NewMockLandingPageRepository(t)
	m.On("Create", mock.Anything).Return(errors.New("fail"))

	svc := service.NewLandingPageAdminService(m, noopTrans())
	_, err := svc.CreateLandingPage(data.LandingPageBody{
		DefaultLang: "en", BannerImageURL: "u", Title: "t", Description: "d", Terms: "x",
	})
	require.Error(t, err)
}

func TestLandingPageAdminService_Create_rejectsInvalidSteps(t *testing.T) {
	svc := service.NewLandingPageAdminService(servicemock.NewMockLandingPageRepository(t), noopTrans())
	_, err := svc.CreateLandingPage(data.LandingPageBody{
		DefaultLang: "en", BannerImageURL: "u", Title: "t", Description: "d", Terms: "x",
		Steps: []data.LandingPageRepeatableItemVO{{Title: "", Description: "x"}},
	})
	require.Error(t, err)
}
