package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/lianjin/campaign-center-api/server/http/data"
	"github.com/lianjin/campaign-center-api/server/log"
	"github.com/lianjin/campaign-center-api/server/proxy"
	"github.com/lianjin/campaign-center-api/server/repository/mysql"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

// LandingPageTranslationService LLM generate + persist translations.
type LandingPageTranslationService interface {
	GenerateTranslation(ctx context.Context, p GenerateTranslationParams) (*GeneratedTranslationDTO, error)
	SaveTranslation(ctx context.Context, p SaveTranslationParams) error
	ListTranslatedLangs(ctx context.Context, landingPageID int64) ([]string, error)
	ResolveLandingPageTexts(page *model.CampaignLandingPage, lang string) (*ResolvedLandingPageTexts, error)
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
	Lang        string
	Title       string
	Description string
	Terms       string
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

// ResolvedLandingPageTexts contains landing page copy for the requested language.
type ResolvedLandingPageTexts struct {
	Lang        string
	Title       string
	Description string
	Terms       string
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

func (s *landingPageTranslationService) GenerateTranslation(
	ctx context.Context, p GenerateTranslationParams,
) (*GeneratedTranslationDTO, error) {
	log.Logger.Infow("generate_translation",
		"landing_page_id", p.LandingPageID, "target_lang", p.TargetLang)
	page, err := s.pages.GetByID(p.LandingPageID)
	if err != nil {
		return nil, err
	}
	title, desc, terms := mergedSourceTexts(page, p)
	if strings.TrimSpace(title+desc+terms) == "" {
		return nil, data.ErrTranslationSourceEmpty
	}
	out, err := s.tr.Translate(ctx, proxy.LandingPageTranslateInput{
		SourceLang: p.SourceLang, TargetLang: p.TargetLang,
		Title: title, Description: desc, Terms: terms,
	})
	if err != nil {
		return nil, err
	}
	return &GeneratedTranslationDTO{
		Lang: p.TargetLang, Title: out.Title,
		Description: out.Description, Terms: out.Terms,
	}, nil
}

func (s *landingPageTranslationService) SaveTranslation(ctx context.Context, p SaveTranslationParams) error {
	_ = ctx
	log.Logger.Infow("save_translation", "landing_page_id", p.LandingPageID, "lang", p.Lang)
	if _, err := s.pages.GetByID(p.LandingPageID); err != nil {
		return err
	}
	op := strings.TrimSpace(p.Operator)
	if op == "" {
		op = "system"
	}
	now := time.Now()
	existing, err := s.translations.GetByLandingPageAndLang(p.LandingPageID, p.Lang)
	if err != nil {
		return err
	}
	row := buildTranslationRow(p, op, now, existing)
	return s.translations.Upsert(row)
}

func (s *landingPageTranslationService) ListTranslatedLangs(
	ctx context.Context, landingPageID int64,
) ([]string, error) {
	_ = ctx
	if _, err := s.pages.GetByID(landingPageID); err != nil {
		return nil, err
	}
	return s.translations.ListLangsByLandingPageID(landingPageID)
}

func (s *landingPageTranslationService) ResolveLandingPageTexts(
	page *model.CampaignLandingPage, lang string,
) (*ResolvedLandingPageTexts, error) {
	if lang == "" || lang == page.DefaultLang {
		return defaultLandingPageTexts(page), nil
	}
	tr, err := s.translations.GetByLandingPageAndLang(page.ID, lang)
	if err != nil {
		return nil, err
	}
	if tr == nil {
		return defaultLandingPageTexts(page), nil
	}
	return &ResolvedLandingPageTexts{
		Lang:        lang,
		Title:       tr.Title,
		Description: tr.Description,
		Terms:       tr.Terms,
	}, nil
}

func defaultLandingPageTexts(page *model.CampaignLandingPage) *ResolvedLandingPageTexts {
	return &ResolvedLandingPageTexts{
		Lang:        page.DefaultLang,
		Title:       page.Title,
		Description: page.Description,
		Terms:       page.Terms,
	}
}

func mergedSourceTexts(page *model.CampaignLandingPage, p GenerateTranslationParams) (string, string, string) {
	return coalesceText(p.Title, page.Title),
		coalesceText(p.Description, page.Description),
		coalesceText(p.Terms, page.Terms)
}

func coalesceText(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

func buildTranslationRow(
	p SaveTranslationParams, op string, now time.Time, existing *model.CampaignLandingPageTranslation,
) *model.CampaignLandingPageTranslation {
	row := &model.CampaignLandingPageTranslation{
		LandingPageID: p.LandingPageID,
		Lang:          p.Lang,
		Title:         p.Title,
		Description:   p.Description,
		Terms:         p.Terms,
		UpdatedAt:     now,
		UpdatedBy:     op,
	}
	if existing != nil {
		row.ID = existing.ID
		row.CreatedAt = existing.CreatedAt
		row.CreatedBy = existing.CreatedBy
		return row
	}
	row.CreatedAt = now
	row.CreatedBy = op
	return row
}
