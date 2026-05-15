package service

import (
	"context"
	"strings"
	"time"

	"github.com/lianjin/campaign-center-api/server/log"
	"github.com/lianjin/campaign-center-api/server/repository/mysql/model"
)

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
