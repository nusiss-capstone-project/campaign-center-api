package service

import (
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/log"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// LandingPageAdminService admin landing page operations.
type LandingPageAdminService interface {
	CreateLandingPage(p CreateLandingPageParams) (id int64, status int16, err error)
	UpdateDraftLandingPage(id int64, p CreateLandingPageParams) error
	ListLandingPages(filter mysql.LandingPageListFilter) ([]model.CampaignLandingPage, int64, error)
	GetLandingPage(id int64, lang string) (*LandingPageDetailView, error)
	PublishLandingPage(id int64, operator string) (*model.CampaignLandingPage, error)
}

// CreateLandingPageParams body fields for create/update landing page.
type CreateLandingPageParams struct {
	DefaultLang    string
	BannerImageURL string
	Title          string
	Description    string
	Terms          string
}

// LandingPageDetailView is resolved landing page text for admin or user display.
type LandingPageDetailView struct {
	ID             int64
	Lang           string
	DefaultLang    string
	BannerImageURL string
	Title          string
	Description    string
	Terms          string
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

func (s *landingPageAdminService) CreateLandingPage(p CreateLandingPageParams) (int64, int16, error) {
	now := time.Now()
	row := model.CampaignLandingPage{
		DefaultLang:    p.DefaultLang,
		BannerImageURL: p.BannerImageURL,
		Title:          p.Title,
		Description:    p.Description,
		Terms:          p.Terms,
		Status:         model.LandingPageStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.pages.Create(&row); err != nil {
		return 0, 0, err
	}
	log.Logger.Infow("landing_page_created", "id", row.ID)
	return row.ID, row.Status, nil
}

func (s *landingPageAdminService) UpdateDraftLandingPage(id int64, p CreateLandingPageParams) error {
	existing, err := s.pages.GetByID(id)
	if err != nil {
		return err
	}
	if existing.Status != model.LandingPageStatusDraft {
		return errLandingPageNotDraft
	}
	now := time.Now()
	existing.DefaultLang = p.DefaultLang
	existing.BannerImageURL = p.BannerImageURL
	existing.Title = p.Title
	existing.Description = p.Description
	existing.Terms = p.Terms
	existing.UpdatedAt = now
	log.Logger.Infow("landing_page_draft_updated", "id", id)
	return s.pages.Update(existing)
}

func (s *landingPageAdminService) ListLandingPages(filter mysql.LandingPageListFilter) ([]model.CampaignLandingPage, int64, error) {
	return s.pages.List(filter)
}

func (s *landingPageAdminService) GetLandingPage(id int64, lang string) (*LandingPageDetailView, error) {
	page, err := s.pages.GetByID(id)
	if err != nil {
		return nil, err
	}
	view := landingPageViewFromRow(page)
	if lang == "" || lang == page.DefaultLang {
		view.Lang = page.DefaultLang
		return view, nil
	}
	tr, err := s.translations.GetByLandingPageAndLang(id, lang)
	if err != nil {
		return nil, err
	}
	if tr != nil {
		applyTranslationToView(view, lang, tr.Title, tr.Description, tr.Terms)
		return view, nil
	}
	view.Lang = page.DefaultLang
	return view, nil
}

func (s *landingPageAdminService) PublishLandingPage(id int64, operator string) (*model.CampaignLandingPage, error) {
	log.Logger.Infow("landing_page_publish", "id", id, "operator", operator)
	return s.pages.Publish(id, operator)
}

func landingPageViewFromRow(p *model.CampaignLandingPage) *LandingPageDetailView {
	return &LandingPageDetailView{
		ID:             p.ID,
		DefaultLang:    p.DefaultLang,
		BannerImageURL: p.BannerImageURL,
		Title:          p.Title,
		Description:    p.Description,
		Terms:          p.Terms,
		Status:         p.Status,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func applyTranslationToView(v *LandingPageDetailView, lang, title, description, terms string) {
	v.Lang = lang
	v.Title = title
	v.Description = description
	v.Terms = terms
}
