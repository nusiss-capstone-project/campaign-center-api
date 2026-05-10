package service

import (
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// LandingPageAdminService admin landing page operations.
type LandingPageAdminService interface {
	CreateLandingPage(p CreateLandingPageParams) (id int64, status int16, err error)
	UpdateDraftLandingPage(id int64, p CreateLandingPageParams) error
	ListLandingPages(filter mysql.LandingPageListFilter) ([]model.CampaignLandingPage, int64, error)
	GetLandingPage(id int64) (*model.CampaignLandingPage, error)
	PublishLandingPage(id int64, operator string) (*model.CampaignLandingPage, error)
}

// CreateLandingPageParams body fields for create/update landing page.
type CreateLandingPageParams struct {
	Language       string
	BannerImageURL string
	Title          string
	Description    string
	Terms          string
}

type landingPageAdminService struct {
	pages mysql.LandingPageRepository
}

var (
	landingPageAdminServiceOnce   sync.Once
	landingPageAdminServiceInst LandingPageAdminService
)

// NewLandingPageAdminService builds a landing page admin service with explicit repositories (for tests).
func NewLandingPageAdminService(pages mysql.LandingPageRepository) LandingPageAdminService {
	return &landingPageAdminService{pages: pages}
}

// GetLandingPageAdminService returns the singleton landing page admin service.
func GetLandingPageAdminService() LandingPageAdminService {
	landingPageAdminServiceOnce.Do(func() {
		landingPageAdminServiceInst = NewLandingPageAdminService(mysql.GetLandingPageRepository())
	})
	return landingPageAdminServiceInst
}

func (s *landingPageAdminService) CreateLandingPage(p CreateLandingPageParams) (int64, int16, error) {
	now := time.Now()
	row := model.CampaignLandingPage{
		Language:       p.Language,
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
	existing.Language = p.Language
	existing.BannerImageURL = p.BannerImageURL
	existing.Title = p.Title
	existing.Description = p.Description
	existing.Terms = p.Terms
	existing.UpdatedAt = now
	return s.pages.Update(existing)
}

func (s *landingPageAdminService) ListLandingPages(filter mysql.LandingPageListFilter) ([]model.CampaignLandingPage, int64, error) {
	return s.pages.List(filter)
}

func (s *landingPageAdminService) GetLandingPage(id int64) (*model.CampaignLandingPage, error) {
	return s.pages.GetByID(id)
}

func (s *landingPageAdminService) PublishLandingPage(id int64, operator string) (*model.CampaignLandingPage, error) {
	return s.pages.Publish(id, operator)
}
