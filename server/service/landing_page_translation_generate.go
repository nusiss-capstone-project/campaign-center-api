package service

import (
	"context"
	"strings"

	"github.com/lianjin/campaign-center-api/server/log"
	"github.com/lianjin/campaign-center-api/server/proxy"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

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
		return nil, errTranslationSourceEmpty
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
