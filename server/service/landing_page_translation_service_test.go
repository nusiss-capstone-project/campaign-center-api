package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	servicemock "github.com/nusiss-capstone-project/campaign-center-api/server/mock"
	"github.com/nusiss-capstone-project/campaign-center-api/server/proxy"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
	"github.com/nusiss-capstone-project/campaign-center-api/server/service"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubTranslator struct {
	out *proxy.LandingPageTranslateOutput
	err error
}

func (t stubTranslator) Translate(context.Context, proxy.LandingPageTranslateInput) (*proxy.LandingPageTranslateOutput, error) {
	return t.out, t.err
}

type translationRepoStub struct {
	existing *model.CampaignLandingPageTranslation
	upserted *model.CampaignLandingPageTranslation
	langs    []string
}

func (r *translationRepoStub) GetByLandingPageAndLang(int64, string) (*model.CampaignLandingPageTranslation, error) {
	return r.existing, nil
}

func (r *translationRepoStub) ListLangsByLandingPageID(int64) ([]string, error) {
	return r.langs, nil
}

func (r *translationRepoStub) Upsert(row *model.CampaignLandingPageTranslation) error {
	r.upserted = row
	return nil
}

func TestLandingPageTranslationService_GenerateTranslation_emptySource(t *testing.T) {
	pages := servicemock.NewMockLandingPageRepository(t)
	pages.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{ID: 1}, nil)
	svc := service.NewLandingPageTranslationService(pages, &translationRepoStub{}, stubTranslator{})

	_, err := svc.GenerateTranslation(context.Background(), service.GenerateTranslationParams{
		LandingPageID: 1, SourceLang: "en", TargetLang: "zh-CN",
	})

	require.Error(t, err)
	require.True(t, data.IsTranslationSourceEmpty(err))
}

func TestLandingPageTranslationService_GenerateTranslation_success(t *testing.T) {
	pages := servicemock.NewMockLandingPageRepository(t)
	pages.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{
		ID: 1, Title: "Hello", Description: "Desc", Terms: "Terms",
	}, nil)
	tr := stubTranslator{out: &proxy.LandingPageTranslateOutput{
		Title: "你好", Description: "描述", Terms: "条款",
	}}
	svc := service.NewLandingPageTranslationService(pages, &translationRepoStub{}, tr)

	out, err := svc.GenerateTranslation(context.Background(), service.GenerateTranslationParams{
		LandingPageID: 1, SourceLang: "en", TargetLang: "zh-CN",
	})

	require.NoError(t, err)
	require.Equal(t, "zh-CN", out.Lang)
	require.Equal(t, "你好", out.Title)
}

func TestLandingPageTranslationService_GenerateTranslation_mergesRequestOverrides(t *testing.T) {
	pages := servicemock.NewMockLandingPageRepository(t)
	pages.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{
		ID: 1, Title: "page title", Description: "page desc", Terms: "page terms",
	}, nil)
	captured := &proxy.LandingPageTranslateInput{}
	svc := service.NewLandingPageTranslationService(pages, &translationRepoStub{}, translatorCapture{
		captured: captured,
		inner:    stubTranslator{out: &proxy.LandingPageTranslateOutput{}},
	})

	_, err := svc.GenerateTranslation(context.Background(), service.GenerateTranslationParams{
		LandingPageID: 1, SourceLang: "en", TargetLang: "zh-CN",
		Title: "override title",
	})
	require.NoError(t, err)
	require.Equal(t, "override title", captured.Title)
	require.Equal(t, "page desc", captured.Description)
}

type translatorCapture struct {
	captured *proxy.LandingPageTranslateInput
	inner    proxy.LandingPageTranslator
}

func (t translatorCapture) Translate(ctx context.Context, in proxy.LandingPageTranslateInput) (*proxy.LandingPageTranslateOutput, error) {
	*t.captured = in
	return t.inner.Translate(ctx, in)
}

func TestLandingPageTranslationService_SaveTranslation_insert(t *testing.T) {
	pages := servicemock.NewMockLandingPageRepository(t)
	pages.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{ID: 1}, nil)
	repo := &translationRepoStub{}
	svc := service.NewLandingPageTranslationService(pages, repo, stubTranslator{})

	err := svc.SaveTranslation(context.Background(), service.SaveTranslationParams{
		LandingPageID: 1, Lang: "zh-CN",
		Title: "t", Description: "d", Terms: "x", Operator: "admin",
	})

	require.NoError(t, err)
	require.NotNil(t, repo.upserted)
	require.Equal(t, "zh-CN", repo.upserted.Lang)
	require.Equal(t, "admin", repo.upserted.CreatedBy)
}

func TestLandingPageTranslationService_SaveTranslation_updatePreservesCreated(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	pages := servicemock.NewMockLandingPageRepository(t)
	pages.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{ID: 1}, nil)
	repo := &translationRepoStub{existing: &model.CampaignLandingPageTranslation{
		ID: 5, LandingPageID: 1, Lang: "zh-CN",
		CreatedAt: created, CreatedBy: "seed",
	}}
	svc := service.NewLandingPageTranslationService(pages, repo, stubTranslator{})

	err := svc.SaveTranslation(context.Background(), service.SaveTranslationParams{
		LandingPageID: 1, Lang: "zh-CN",
		Title: "new", Description: "d", Terms: "x",
	})

	require.NoError(t, err)
	require.Equal(t, int64(5), repo.upserted.ID)
	require.Equal(t, created, repo.upserted.CreatedAt)
	require.Equal(t, "seed", repo.upserted.CreatedBy)
}

func TestLandingPageTranslationService_ListTranslatedLangs(t *testing.T) {
	pages := servicemock.NewMockLandingPageRepository(t)
	pages.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{ID: 1}, nil)
	repo := &translationRepoStub{langs: []string{"zh-CN", "ja"}}
	svc := service.NewLandingPageTranslationService(pages, repo, stubTranslator{})

	langs, err := svc.ListTranslatedLangs(context.Background(), 1)

	require.NoError(t, err)
	require.Equal(t, []string{"zh-CN", "ja"}, langs)
}

func TestLandingPageTranslationService_ListTranslatedLangs_pageNotFound(t *testing.T) {
	pages := servicemock.NewMockLandingPageRepository(t)
	pages.On("GetByID", int64(9)).Return(nil, gorm.ErrRecordNotFound)
	svc := service.NewLandingPageTranslationService(pages, &translationRepoStub{}, stubTranslator{})

	_, err := svc.ListTranslatedLangs(context.Background(), 9)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestLandingPageTranslationService_SaveTranslation_pageNotFound(t *testing.T) {
	pages := servicemock.NewMockLandingPageRepository(t)
	pages.On("GetByID", int64(9)).Return(nil, gorm.ErrRecordNotFound)
	svc := service.NewLandingPageTranslationService(pages, &translationRepoStub{}, stubTranslator{})

	err := svc.SaveTranslation(context.Background(), service.SaveTranslationParams{
		LandingPageID: 9, Lang: "zh-CN", Title: "t", Description: "d", Terms: "x",
	})

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestLandingPageTranslationService_GenerateTranslation_propagatesTranslatorError(t *testing.T) {
	trErr := errors.New("llm down")
	pages := servicemock.NewMockLandingPageRepository(t)
	pages.On("GetByID", int64(1)).Return(&model.CampaignLandingPage{
		ID: 1, Title: "t", Description: "d", Terms: "x",
	}, nil)
	svc := service.NewLandingPageTranslationService(pages, &translationRepoStub{}, stubTranslator{err: trErr})

	_, err := svc.GenerateTranslation(context.Background(), service.GenerateTranslationParams{
		LandingPageID: 1, SourceLang: "en", TargetLang: "zh-CN",
	})

	require.ErrorIs(t, err, trErr)
}
