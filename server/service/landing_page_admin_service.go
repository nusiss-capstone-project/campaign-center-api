package service

import (
	"context"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
)

// LandingPageAdminService admin landing page operations.
type LandingPageAdminService interface {
	CreateLandingPage(ctx context.Context, body data.LandingPageBody) (*data.LandingPageCreateResp, error)
	UpdateDraftLandingPage(ctx context.Context, id int64, body data.LandingPageBody) (*data.LandingPageUpdateResp, error)
	ListLandingPages(filter mysql.LandingPageListFilter) (*data.LandingPageListData, error)
	GetLandingPage(id int64, lang string) (*data.LandingPageDetailVO, error)
	PublishLandingPage(ctx context.Context, id int64, operator string) (*data.LandingPagePublishResp, error)
}

// LandingPageDetailView is resolved landing page text for admin display.
type LandingPageDetailView struct {
	ID             int64
	Lang           string
	DefaultLang    string
	BannerImageURL string
	Title          string
	Description    string
	Terms          string
	Steps          []data.LandingPageRepeatableItemVO
	Faq            []data.LandingPageRepeatableItemVO
	Status         int16
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type landingPageAdminService struct {
	pages        mysql.LandingPageRepository
	translations mysql.LandingPageTranslationRepository
}

var (
	landingPageAdminServiceOnce sync.Once
	landingPageAdminServiceInst LandingPageAdminService
)

// NewLandingPageAdminService builds admin service with repositories (for tests).
func NewLandingPageAdminService(
	pages mysql.LandingPageRepository,
	translations mysql.LandingPageTranslationRepository,
) LandingPageAdminService {
	return &landingPageAdminService{pages: pages, translations: translations}
}

// GetLandingPageAdminService returns the singleton landing page admin service.
func GetLandingPageAdminService() LandingPageAdminService {
	landingPageAdminServiceOnce.Do(func() {
		landingPageAdminServiceInst = NewLandingPageAdminService(
			mysql.GetLandingPageRepository(),
			mysql.GetLandingPageTranslationRepository(),
		)
	})
	return landingPageAdminServiceInst
}

func (s *landingPageAdminService) CreateLandingPage(ctx context.Context, body data.LandingPageBody) (*data.LandingPageCreateResp, error) {
	if err := validateLandingPageContent(body.Steps, body.Faq); err != nil {
		return nil, err
	}
	now := time.Now()
	row := model.CampaignLandingPage{
		DefaultLang:    body.DefaultLang,
		BannerImageURL: body.BannerImageURL,
		Title:          body.Title,
		Description:    body.Description,
		Terms:          body.Terms,
		Steps:          toModelRepeatableItems(body.Steps),
		Faq:            toModelRepeatableItems(body.Faq),
		Status:         model.LandingPageStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.pages.Create(ctx, &row); err != nil {
		return nil, err
	}
	log.WithContext(ctx).Infow("landing_page_created", "id", row.ID)
	return landingPageCreateResp(row.ID, row.Status, body), nil
}

func (s *landingPageAdminService) UpdateDraftLandingPage(ctx context.Context, id int64, body data.LandingPageBody) (*data.LandingPageUpdateResp, error) {
	if err := validateLandingPageContent(body.Steps, body.Faq); err != nil {
		return nil, err
	}
	existing, err := s.pages.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing.Status != model.LandingPageStatusDraft {
		return nil, data.ErrLandingPageNotDraft
	}
	now := time.Now()
	existing.DefaultLang = body.DefaultLang
	existing.BannerImageURL = body.BannerImageURL
	existing.Title = body.Title
	existing.Description = body.Description
	existing.Terms = body.Terms
	existing.Steps = toModelRepeatableItems(body.Steps)
	existing.Faq = toModelRepeatableItems(body.Faq)
	existing.UpdatedAt = now
	log.WithContext(ctx).Infow("landing_page_draft_updated", "id", id)
	if err := s.pages.Update(ctx, existing); err != nil {
		return nil, err
	}
	return landingPageUpdateResp(id, body), nil
}

func (s *landingPageAdminService) ListLandingPages(filter mysql.LandingPageListFilter) (*data.LandingPageListData, error) {
	items, total, err := s.pages.List(filter)
	if err != nil {
		return nil, err
	}
	out := make([]data.LandingPageListItemVO, 0, len(items))
	for _, item := range items {
		out = append(out, data.LandingPageListItemVO{
			ID:          item.ID,
			Title:       item.Title,
			Status:      item.Status,
			DefaultLang: item.DefaultLang,
			CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		})
	}
	return &data.LandingPageListData{Total: total, Items: out}, nil
}

func (s *landingPageAdminService) GetLandingPage(id int64, lang string) (*data.LandingPageDetailVO, error) {
	page, err := s.pages.GetByID(id)
	if err != nil {
		return nil, err
	}
	view := landingPageViewFromRow(page)
	if lang == "" || lang == page.DefaultLang {
		view.Lang = page.DefaultLang
		return landingPageDetailVO(view), nil
	}
	tr, err := s.translations.GetByLandingPageAndLang(id, lang)
	if err != nil {
		return nil, err
	}
	if tr != nil {
		applyTranslationToView(view, lang, tr)
		return landingPageDetailVO(view), nil
	}
	view.Lang = page.DefaultLang
	return landingPageDetailVO(view), nil
}

func (s *landingPageAdminService) PublishLandingPage(ctx context.Context, id int64, operator string) (*data.LandingPagePublishResp, error) {
	log.WithContext(ctx).Infow("landing_page_publish", "id", id, "operator", operator)
	updated, err := s.pages.Publish(ctx, id, operator)
	if err != nil {
		return nil, err
	}
	return &data.LandingPagePublishResp{LandingPageID: updated.ID, Status: updated.Status}, nil
}

func landingPageViewFromRow(p *model.CampaignLandingPage) *LandingPageDetailView {
	return &LandingPageDetailView{
		ID:             p.ID,
		DefaultLang:    p.DefaultLang,
		BannerImageURL: p.BannerImageURL,
		Title:          p.Title,
		Description:    p.Description,
		Terms:          p.Terms,
		Steps:          toDataRepeatableItems(p.Steps),
		Faq:            toDataRepeatableItems(p.Faq),
		Status:         p.Status,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func applyTranslationToView(v *LandingPageDetailView, lang string, tr *model.CampaignLandingPageTranslation) {
	v.Lang = lang
	v.Title = tr.Title
	v.Description = tr.Description
	v.Terms = tr.Terms
	v.Steps = toDataRepeatableItems(tr.Steps)
	v.Faq = toDataRepeatableItems(tr.Faq)
}
