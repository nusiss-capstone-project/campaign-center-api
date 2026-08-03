package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/http/data"
	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
	"github.com/nusiss-capstone-project/campaign-center-api/server/proxy"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql"
	"github.com/nusiss-capstone-project/campaign-center-api/server/repository/mysql/model"
)

// LandingPageTranslationService LLM generate + persist translations.
type LandingPageTranslationService interface {
	GenerateTranslation(ctx context.Context, p GenerateTranslationParams) (*data.GenerateLandingTranslationData, error)
	SaveTranslation(ctx context.Context, p SaveTranslationParams) (*data.PutLandingTranslationData, error)
	ListTranslatedLangs(ctx context.Context, landingPageID int64) (*data.LandingPageTranslatedLangsData, error)
}

// GenerateTranslationParams is the service input for POST .../translations/generate.
type GenerateTranslationParams struct {
	LandingPageID int64
	SourceLang    string
	TargetLang    string
	Title         string
	Description   string
	Terms         string
	Steps         []data.LandingPageRepeatableItemVO
	Faq           []data.LandingPageRepeatableItemVO
}

// SaveTranslationParams is the service input for PUT .../translations/{lang}.
type SaveTranslationParams struct {
	LandingPageID int64
	Lang          string
	Title         string
	Description   string
	Terms         string
	Steps         []data.LandingPageRepeatableItemVO
	Faq           []data.LandingPageRepeatableItemVO
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

func (s *landingPageTranslationService) GenerateTranslation(
	ctx context.Context, p GenerateTranslationParams,
) (*data.GenerateLandingTranslationData, error) {
	log.Logger.Infow("generate_translation",
		"landing_page_id", p.LandingPageID, "target_lang", p.TargetLang)
	page, err := s.pages.GetByID(p.LandingPageID)
	if err != nil {
		return nil, err
	}
	title, desc, terms, steps, faq := mergedSourceContent(page, p)
	if strings.TrimSpace(title+desc+terms) == "" && len(steps) == 0 && len(faq) == 0 {
		return nil, data.ErrTranslationSourceEmpty
	}
	out, err := s.tr.Translate(ctx, proxy.LandingPageTranslateInput{
		SourceLang: p.SourceLang, TargetLang: p.TargetLang,
		Title: title, Description: desc, Terms: terms,
		Steps: steps, Faq: faq,
	})
	if err != nil {
		return nil, err
	}
	return &data.GenerateLandingTranslationData{
		Lang: p.TargetLang, Title: out.Title,
		Description: out.Description, Terms: out.Terms,
		Steps: normalizeRepeatableVO(out.Steps), Faq: normalizeRepeatableVO(out.Faq),
	}, nil
}

func (s *landingPageTranslationService) SaveTranslation(
	ctx context.Context, p SaveTranslationParams,
) (*data.PutLandingTranslationData, error) {
	_ = ctx
	if err := validateLandingPageContent(p.Steps, p.Faq); err != nil {
		return nil, err
	}
	log.Logger.Infow("save_translation", "landing_page_id", p.LandingPageID, "lang", p.Lang)
	if _, err := s.pages.GetByID(p.LandingPageID); err != nil {
		return nil, err
	}
	op := strings.TrimSpace(p.Operator)
	if op == "" {
		op = "system"
	}
	now := time.Now()
	existing, err := s.translations.GetByLandingPageAndLang(p.LandingPageID, p.Lang)
	if err != nil {
		return nil, err
	}
	row := buildTranslationRow(p, op, now, existing)
	if err := s.translations.Upsert(row); err != nil {
		return nil, err
	}
	return &data.PutLandingTranslationData{LandingPageID: p.LandingPageID, Lang: p.Lang}, nil
}

func (s *landingPageTranslationService) ListTranslatedLangs(
	ctx context.Context, landingPageID int64,
) (*data.LandingPageTranslatedLangsData, error) {
	_ = ctx
	if _, err := s.pages.GetByID(landingPageID); err != nil {
		return nil, err
	}
	langs, err := s.translations.ListLangsByLandingPageID(landingPageID)
	if err != nil {
		return nil, err
	}
	if langs == nil {
		langs = []string{}
	}
	return &data.LandingPageTranslatedLangsData{Langs: langs}, nil
}

func mergedSourceContent(
	page *model.CampaignLandingPage, p GenerateTranslationParams,
) (string, string, string, []data.LandingPageRepeatableItemVO, []data.LandingPageRepeatableItemVO) {
	steps := p.Steps
	if len(steps) == 0 {
		steps = toDataRepeatableItems(page.Steps)
	}
	faq := p.Faq
	if len(faq) == 0 {
		faq = toDataRepeatableItems(page.Faq)
	}
	return coalesceText(p.Title, page.Title),
		coalesceText(p.Description, page.Description),
		coalesceText(p.Terms, page.Terms),
		steps,
		faq
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
		Steps:         toModelRepeatableItems(p.Steps),
		Faq:           toModelRepeatableItems(p.Faq),
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
