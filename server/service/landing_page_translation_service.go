package service

import (
	"context"
	"sync"

	"github.com/lianjin/campaign-center-api/server/proxy"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
)

// LandingPageTranslationService LLM generate + persist translations.
type LandingPageTranslationService interface {
	GenerateTranslation(ctx context.Context, p GenerateTranslationParams) (*GeneratedTranslationDTO, error)
	SaveTranslation(ctx context.Context, p SaveTranslationParams) error
	ListTranslatedLangs(ctx context.Context, landingPageID int64) ([]string, error)
}

// GenerateTranslationParams is the service input for POST .../landing-pages/{id}/translations/generate.
type GenerateTranslationParams struct {
	LandingPageID int64
	SourceLang    string
	TargetLang    string
	Title         string
	Description   string
	Terms         string
}

// GeneratedTranslationDTO is the generate API payload.
type GeneratedTranslationDTO struct {
	Lang          string
	Title         string
	Description   string
	Terms         string
}

// SaveTranslationParams is the service input for PUT .../landing-pages/{id}/translations/{lang}.
type SaveTranslationParams struct {
	LandingPageID int64
	Lang          string
	Title         string
	Description   string
	Terms         string
	Operator      string
}

type landingPageTranslationService struct {
	pages        mysql.LandingPageRepository
	translations mysql.LandingPageTranslationRepository
	tr           proxy.LandingPageTranslator
}

var (
	landingPageTranslationSvcOnce sync.Once
	landingPageTranslationSvcInst LandingPageTranslationService
)

// NewLandingPageTranslationService wires repos and translator (for tests).
func NewLandingPageTranslationService(
	pages mysql.LandingPageRepository,
	translations mysql.LandingPageTranslationRepository,
	tr proxy.LandingPageTranslator,
) LandingPageTranslationService {
	return &landingPageTranslationService{pages: pages, translations: translations, tr: tr}
}

// GetLandingPageTranslationService returns the singleton.
func GetLandingPageTranslationService() LandingPageTranslationService {
	landingPageTranslationSvcOnce.Do(func() {
		landingPageTranslationSvcInst = NewLandingPageTranslationService(
			mysql.GetLandingPageRepository(),
			mysql.GetLandingPageTranslationRepository(),
			proxy.GetLandingPageTranslator(),
		)
	})
	return landingPageTranslationSvcInst
}
